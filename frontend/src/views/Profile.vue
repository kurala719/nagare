<template>
  <div class="profile-page">
    <el-card class="profile-card">
      <div class="profile-header">
        <el-avatar :size="80" :src="profile.avatar" />
        <div class="profile-meta">
          <h2>{{ profile.nickname || profile.username || '-' }}</h2>
          <p class="role">{{ profile.role || roleLabel }}</p>
        </div>
      </div>

      <el-divider />

      <el-form :model="form" label-width="120px">
        <el-form-item :label="$t('profile.username')">
          <el-input v-model="profile.username" disabled />
        </el-form-item>
        <el-form-item :label="$t('profile.nickname')">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item :label="$t('profile.email')">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item :label="$t('profile.phone')">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item :label="$t('profile.avatar')">
          <el-input v-model="form.avatar" />
        </el-form-item>
        <el-form-item :label="$t('profile.address')">
          <el-input v-model="form.address" />
        </el-form-item>
        <el-form-item :label="$t('profile.introduction')">
          <el-input v-model="form.introduction" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="onSave">{{ $t('profile.save') }}</el-button>
          <el-button @click="onReset">{{ $t('profile.reset') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getUserInformation, updateUserInformation, createUserInformation } from '@/api/userInformation'
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
    const { data } = await getUserInformation()
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
    // Try to update first
    try {
      await updateUserInformation({ ...form })
      ElMessage.success(t('profile.updated'))
    } catch (err) {
      // If update fails (404), create new profile
      if (err?.response?.status === 404) {
        await createUserInformation({ ...form })
        ElMessage.success(t('profile.created'))
      } else {
        throw err
      }
    }
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
.profile-page {
  padding: 24px;
}

.profile-card {
  max-width: 720px;
  margin: 0 auto;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.profile-meta h2 {
  margin: 0;
}

.role {
  margin: 4px 0 0;
  color: #6b7280;
}
</style>
