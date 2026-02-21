<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('profile.title') }}</h1>
      <p class="page-subtitle">{{ profile.username }}</p>
    </div>

    <el-card class="profile-card">
      <div class="profile-header">
        <div class="avatar-block">
          <div class="avatar-container">
            <el-avatar :size="100" :src="pendingAvatarUrl || profile.avatar || undefined">
              <span class="avatar-fallback">{{ avatarInitials }}</span>
            </el-avatar>
            <div
              class="avatar-edit-overlay"
              role="button"
              tabindex="0"
              @click="openAvatarPicker"
              @keydown.enter.prevent="openAvatarPicker"
              @keydown.space.prevent="openAvatarPicker"
            >
              <el-icon><Edit /></el-icon>
            </div>
            <input
              ref="avatarInput"
              class="avatar-input"
              type="file"
              accept="image/*"
              @change="handleAvatarFileChange"
            />
          </div>
          <p class="avatar-help">{{ $t('profile.avatarHelp') }}</p>
        </div>
        <div class="profile-meta">
          <h2 class="profile-name">{{ profile.nickname || profile.username || '-' }}</h2>
          <el-tag effect="dark" class="role-tag">{{ profile.role || roleLabel }}</el-tag>
        </div>
      </div>

      <div class="profile-form-section">
        <el-form ref="profileFormRef" :model="form" :rules="rules" label-width="120px" label-position="top">
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
              <el-form-item :label="$t('profile.email')" prop="email">
                <el-input v-model="form.email" :prefix-icon="Message" />
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item :label="$t('profile.phone')">
                <el-input v-model="form.phone" :prefix-icon="Phone" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="24">
            <el-col :md="12">
              <el-form-item :label="$t('profile.address')">
                <el-input v-model="form.address" :prefix-icon="Location" />
              </el-form-item>
            </el-col>
            <el-col :md="12">
              <el-form-item :label="$t('profile.qq') || 'QQ'">
                <el-input v-model="form.qq" />
              </el-form-item>
            </el-col>
          </el-row>

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
import { Edit, Message, Phone, Location } from '@element-plus/icons-vue'
import { getUserProfile, updateUserProfile, uploadAvatar } from '@/api/users'
import { getUserClaims, getUserPrivileges } from '@/utils/auth'

const saving = ref(false)
const uploading = ref(false)
const uploadProgress = ref(0)
const profileFormRef = ref(null)
const avatarInput = ref(null)
const pendingAvatarFile = ref(null)
const pendingAvatarUrl = ref('')
const { t } = useI18n()
const profile = reactive({
  username: '',
  nickname: '',
  email: '',
  phone: '',
  avatar: '',
  address: '',
  introduction: '',
  role: '',
  qq: ''
})

const form = reactive({
  nickname: '',
  email: '',
  phone: '',
  avatar: '',
  address: '',
  introduction: '',
  qq: ''
})

const validateEmail = (rule, value, callback) => {
  if (!value) {
    callback()
    return
  }
  const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailPattern.test(String(value).trim())) {
    callback(new Error(t('profile.emailInvalid')))
    return
  }
  callback()
}

const rules = {
  email: [{ validator: validateEmail, trigger: 'blur' }]
}

const roleLabel = computed(() => {
  const privilege = getUserPrivileges()
  if (privilege >= 3) return t('profile.roles.superadmin')
  if (privilege >= 2) return t('profile.roles.admin')
  if (privilege >= 1) return t('profile.roles.user')
  return t('profile.roles.unauthorized')
})

const avatarInitials = computed(() => {
  const source = (profile.nickname || profile.username || '').trim()
  if (!source) return '?'
  const parts = source.split(/\s+/)
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
})

const resolveBackendOrigin = () => {
  if (!import.meta.env.DEV) return ''
  if (typeof window === 'undefined') return ''
  const { protocol, hostname, port } = window.location
  if (port === '8080') return ''
  return `${protocol}//${hostname}:8080`
}

const normalizeAvatarUrl = (value) => {
  if (!value) return ''
  const trimmed = String(value).trim()
  if (!trimmed) return ''
  if (/^(https?:|data:|blob:)/i.test(trimmed)) return trimmed
  const prefixed = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  const backendOrigin = resolveBackendOrigin()
  if (backendOrigin && prefixed.startsWith('/avatars/')) {
    return `${backendOrigin}${prefixed}`
  }
  return prefixed
}

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
    const normalizedAvatar = normalizeAvatarUrl(payload?.avatar)
    Object.assign(profile, payload)
    profile.avatar = normalizedAvatar
    Object.assign(form, {
      nickname: payload?.nickname || '',
      email: payload?.email || '',
      phone: payload?.phone || '',
      avatar: normalizedAvatar,
      address: payload?.address || '',
      introduction: payload?.introduction || '',
      qq: payload?.qq || payload?.QQ || ''
    })
  } catch (err) {
    if (err?.response?.status !== 404) {
      ElMessage.error(err?.response?.data?.error || err.message || t('profile.loadFailed'))
    }
  } finally {
    setDefaultsFromClaims()
  }
}

const openAvatarPicker = () => {
  if (uploading.value) return
  avatarInput.value?.click()
}

const clearPendingAvatar = () => {
  if (pendingAvatarUrl.value) {
    URL.revokeObjectURL(pendingAvatarUrl.value)
  }
  pendingAvatarUrl.value = ''
  pendingAvatarFile.value = null
  uploadProgress.value = 0
}

const handleAvatarFileChange = (event) => {
  const file = event.target.files?.[0]
  if (!file) return

  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    ElMessage.error(t('profile.avatarUploadFailed') || 'Invalid avatar file type')
    event.target.value = ''
    return
  }

  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error(t('profile.avatarTooLarge') || 'Avatar file is too large (max 5MB)')
    event.target.value = ''
    return
  }

  clearPendingAvatar()
  pendingAvatarFile.value = file
  pendingAvatarUrl.value = URL.createObjectURL(file)
  event.target.value = ''
}

const onSave = async () => {
  saving.value = true
  try {
    if (profileFormRef.value) {
      await profileFormRef.value.validate()
    }
    let avatarURL = form.avatar
    if (pendingAvatarFile.value) {
      uploading.value = true
      const formData = new FormData()
      formData.append('avatar', pendingAvatarFile.value)
      const response = await uploadAvatar(formData, (progressEvent) => {
        if (!progressEvent.total) return
        uploadProgress.value = Math.round((progressEvent.loaded / progressEvent.total) * 100)
      })
      avatarURL = normalizeAvatarUrl(response?.data?.avatar_url || response?.avatar_url)
      if (!avatarURL) {
        throw new Error('Missing avatar URL in response')
      }
    }

    await updateUserProfile({ ...form, avatar: avatarURL })
    ElMessage.success(t('profile.updated'))
    await loadProfile()
    clearPendingAvatar()
  } catch (err) {
    ElMessage.error(err?.response?.data?.error || err.message || t('profile.saveFailed'))
  } finally {
    saving.value = false
    uploading.value = false
  }
}

const onReset = () => {
  Object.assign(form, {
    nickname: profile.nickname || '',
    email: profile.email || '',
    phone: profile.phone || '',
    avatar: profile.avatar || '',
    address: profile.address || '',
    introduction: profile.introduction || '',
    qq: profile.qq || ''
  })
  clearPendingAvatar()
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

.avatar-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.avatar-container {
  position: relative;
  display: inline-block;
}

.avatar-input {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
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

.avatar-fallback {
  font-weight: 700;
  font-size: 18px;
  color: var(--text-muted);
}

.avatar-help {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-tertiary);
}

</style>
