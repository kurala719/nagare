<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('profile.title') }}</h1>
      <p class="page-subtitle">{{ profile.username }}</p>
    </div>

    <el-card class="profile-card">
      <div class="profile-header">
        <div class="avatar-container">
          <el-avatar :size="100" :src="profile.avatar" />
          <div class="avatar-edit-overlay" v-if="profile.avatar">
            <el-icon><Edit /></el-icon>
          </div>
        </div>
        <div class="profile-meta">
          <h2 class="profile-name">{{ profile.nickname || profile.username || '-' }}</h2>
          <el-tag effect="dark" class="role-tag">{{ profile.role || roleLabel }}</el-tag>
        </div>
      </div>

      <div class="profile-form-section">
        <el-form :model="form" label-width="120px" label-position="top">
          <el-row :gutter="24">
            <el-col :md="12">
              <el-form-item :label="$t('profile.username')">
                <el-input v-model="profile.username" disabled />
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item :label="$t('profile.nickname')">
                <el-input v-model="form.nickname" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="24">
            <el-col :md="12">
              <el-form-item :label="$t('profile.email')">
                <el-input v-model="form.email" :prefix-icon="Message" />
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item :label="$t('profile.phone')">
                <el-input v-model="form.phone" :prefix-icon="Phone" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item :label="$t('profile.avatar')">
            <el-input v-model="form.avatar" placeholder="https://example.com/avatar.png" :prefix-icon="Link" />
          </el-form-item>

          <el-form-item :label="$t('profile.address')">
            <el-input v-model="form.address" :prefix-icon="Location" />
          </el-form-item>

          <el-form-item :label="$t('profile.introduction')">
            <el-input v-model="form.introduction" type="textarea" :autosize="{ minRows: 4, maxRows: 8 }" />
          </el-form-item>

          <div class="form-actions">
            <el-button type="primary" :loading="saving" @click="onSave" class="save-btn">
              {{ $t('profile.save') }}
            </el-button>
            <el-button @click="onReset" plain>{{ $t('profile.reset') }}</el-button>
          </div>
        </el-form>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Edit, Message, Phone, Link, Location } from '@element-plus/icons-vue'
import { getUserProfile, updateUserProfile } from '@/api/users'
import { getUserClaims, getUserPrivileges } from '@/utils/auth'

const saving = ref(false)
const { t } = useI18n()
const profile = reactive({
  username: '',
  nickname: '',
  email: '',
  phone: '',
  avatar: '',
  address: '',
  introduction: '',
  role: ''
})

const form = reactive({
  nickname: '',
  email: '',
  phone: '',
  avatar: '',
  address: '',
  introduction: ''
})

const roleLabel = computed(() => {
  const privilege = getUserPrivileges()
  if (privilege >= 3) return t('profile.roles.superadmin')
  if (privilege >= 2) return t('profile.roles.admin')
  if (privilege >= 1) return t('profile.roles.user')
  return t('profile.roles.unauthorized')
})

const setDefaultsFromClaims = () => {
  const claims = getUserClaims()
  if (claims?.username && !profile.username) {
    profile.username = claims.username
  }
}

const loadProfile = async () => {
  try {
    const { data } = await getUserProfile()
    const payload = data?.data || data
    Object.assign(profile, payload)
    Object.assign(form, {
      nickname: payload?.nickname || '',
      email: payload?.email || '',
      phone: payload?.phone || '',
      avatar: payload?.avatar || '',
      address: payload?.address || '',
      introduction: payload?.introduction || ''
    })
  } catch (err) {
    if (err?.response?.status !== 404) {
      ElMessage.error(err?.response?.data?.error || err.message || t('profile.loadFailed'))
    }
  } finally {
    setDefaultsFromClaims()
  }
}

const onSave = async () => {
  saving.value = true
  try {
    await updateUserProfile({ ...form })
    ElMessage.success(t('profile.updated'))
    await loadProfile()
  } catch (err) {
    ElMessage.error(err?.response?.data?.error || err.message || t('profile.saveFailed'))
  } finally {
    saving.value = false
  }
}

const onReset = () => {
  Object.assign(form, {
    nickname: profile.nickname || '',
    email: profile.email || '',
    phone: profile.phone || '',
    avatar: profile.avatar || '',
    address: profile.address || '',
    introduction: profile.introduction || ''
  })
}

onMounted(() => {
  setDefaultsFromClaims()
  loadProfile()
})
</script>

<style scoped>
.profile-card {
  max-width: 800px;
  margin: 0 auto;
  border: 1px solid var(--border-1);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 32px;
  padding: 24px;
  background: var(--surface-2);
  border-bottom: 1px solid var(--border-1);
  margin: -20px -20px 24px -20px;
}

.avatar-container {
  position: relative;
  display: inline-block;
}

.avatar-edit-overlay {
  position: absolute;
  bottom: 0;
  right: 0;
  background: var(--brand-600);
  color: white;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
}

.profile-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.profile-name {
  margin: 0;
  font-size: 24px;
  font-weight: 800;
  color: var(--text-strong);
  font-family: var(--font-display);
}

.role-tag {
  align-self: flex-start;
  text-transform: uppercase;
  font-size: 11px;
  letter-spacing: 1px;
}

.profile-form-section {
  padding: 8px 12px;
}

.form-actions {
  margin-top: 32px;
  display: flex;
  gap: 12px;
  padding-top: 24px;
  border-top: 1px solid var(--border-1);
}

.save-btn {
  padding-left: 40px;
  padding-right: 40px;
}

:deep(.el-form-item__label) {
  font-weight: 700;
  font-size: 13px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
</style>
