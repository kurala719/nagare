<template>
  <div class="auth-wrapper">
    <AnimatedBackground />
    <AuthControls />

    <div class="auth-container animate-scale-in">
      <!-- Left side: Brand/Illustration -->
      <div class="auth-aside animate-fade-in">
        <div class="brand-content">
          <div class="logo-wrapper animate-slide-up delay-1 animate-float">
            <el-icon :size="48" color="#ffffff"><Monitor /></el-icon>
          </div>
          <h1 class="brand-name animate-slide-up delay-2">Nagare</h1>
          <p class="brand-tagline animate-slide-up delay-3">{{ $t('auth.tagline') }}</p>
          
          <div class="feature-list">
            <div class="feature-item animate-slide-up delay-4">
              <el-icon><Lock /></el-icon>
              <span>Security First</span>
            </div>
            <div class="feature-item animate-slide-up delay-4" style="animation-delay: 0.5s">
              <el-icon><Key /></el-icon>
              <span>Encrypted Credentials</span>
            </div>
            <div class="feature-item animate-slide-up delay-4" style="animation-delay: 0.6s">
              <el-icon><Finished /></el-icon>
              <span>Access Restoration</span>
            </div>
          </div>
        </div>
        <div class="auth-aside-footer animate-fade-in" style="animation-delay: 0.8s">
          &copy; 2026 Nagare Project. All rights reserved.
        </div>
      </div>

      <!-- Right side: Reset Password Form -->
      <div class="auth-main">
        <div class="login-box animate-slide-up delay-2">
          <div class="login-header">
            <h2 class="login-title">Reset Password</h2>
            <p class="login-subtitle">{{ $t('auth.resetTitle') }}</p>
          </div>

          <el-form 
            ref="resetFormRef"
            :model="form" 
            :rules="rules"
            label-position="top"
            @keyup.enter="onReset"
          >
            <el-form-item :label="$t('auth.username')" prop="username">
              <el-input 
                v-model="form.username" 
                placeholder="Enter your username"
                :prefix-icon="User"
                autocomplete="username"
              />
            </el-form-item>
            
            <el-form-item :label="$t('auth.newPassword')" prop="newPassword">
              <el-input 
                v-model="form.newPassword" 
                type="password" 
                placeholder="Enter new password"
                :prefix-icon="Key"
                autocomplete="new-password" 
                show-password 
              />
            </el-form-item>

            <el-form-item :label="$t('auth.confirmNewPassword')" prop="confirm">
              <el-input 
                v-model="form.confirm" 
                type="password" 
                placeholder="Confirm new password"
                :prefix-icon="Check"
                autocomplete="new-password" 
                show-password 
              />
            </el-form-item>

            <el-button 
              type="primary" 
              class="submit-btn" 
              :loading="loading" 
              @click="onReset"
            >
              {{ $t('auth.submitResetRequest') || 'Submit Reset Request' }}
            </el-button>

            <div class="register-hint">
              <span>Remembered your password?</span>
              <router-link to="/login">{{ $t('auth.backToLogin') }}</router-link>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Lock, Key, Check, Monitor, Finished, User } from '@element-plus/icons-vue'
import AuthControls from '@/components/AuthControls.vue'
import AnimatedBackground from '@/components/Customed/AnimatedBackground.vue'
import request from '@/utils/request'

const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const resetFormRef = ref(null)

const form = reactive({
  username: '',
  newPassword: '',
  confirm: ''
})

const validatePass2 = (rule, value, callback) => {
  if (value === '') {
    callback(new Error(t('auth.completeAllFields')))
  } else if (value !== form.newPassword) {
    callback(new Error(t('auth.passwordsMismatch')))
  } else {
    callback()
  }
}

const validatePasswordStrength = (rule, value, callback) => {
  if (!value) {
    callback(new Error(t('auth.newPassword') + ' is required'))
    return
  }

  const trimmed = String(value)
  if (trimmed.length < 8 || /\s/.test(trimmed)) {
    callback(new Error(t('auth.passwordRules')))
    return
  }

  let score = 0
  if (/[a-z]/.test(trimmed)) score += 1
  if (/[A-Z]/.test(trimmed)) score += 1
  if (/[0-9]/.test(trimmed)) score += 1
  if (/[^A-Za-z0-9]/.test(trimmed)) score += 1

  if (score < 3) {
    callback(new Error(t('auth.passwordTooWeak')))
    return
  }

  callback()
}

const rules = {
  username: [{ required: true, message: t('auth.username') + ' is required', trigger: 'blur' }],
  newPassword: [{ validator: validatePasswordStrength, trigger: 'blur' }],
  confirm: [{ validator: validatePass2, trigger: 'blur' }]
}

const onReset = async () => {
  if (!resetFormRef.value) return
  
  try {
    await resetFormRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await request({
      url: '/auth/reset-request',
      method: 'POST',
      data: {
        username: form.username,
        password: form.newPassword
      }
    })
    ElMessage.success(t('auth.resetApplicationSubmitted') || 'Reset request submitted. Please wait for admin approval.')
    router.replace('/login')
  } catch (err) {
    const errorMsg = err?.response?.data?.error || err.message || t('auth.resetFailed')
    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrapper {
  min-height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background: transparent !important;
}

.auth-container {
  display: flex;
  width: 1000px;
  height: 650px;
  background: var(--surface-1);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-lg);
  border: 1px solid var(--border-1);
  position: relative;
  z-index: 10;
}

.auth-aside {
  flex: 1;
  background: linear-gradient(135deg, var(--brand-600) 0%, var(--brand-700) 100%);
  padding: 48px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  color: white;
  position: relative;
  overflow: hidden;
}

.auth-aside::before {
  content: '';
  position: absolute;
  top: -100px;
  right: -100px;
  width: 300px;
  height: 300px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
}

.logo-wrapper {
  width: 80px;
  height: 80px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  backdrop-filter: blur(10px);
}

.brand-name {
  font-size: 42px;
  font-weight: 800;
  margin: 0 0 12px 0;
  letter-spacing: -1px;
  font-family: var(--font-display);
}

.brand-tagline {
  font-size: 18px;
  opacity: 0.9;
  line-height: 1.5;
  max-width: 300px;
}

.feature-list {
  margin-top: 48px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  font-weight: 500;
  font-size: 16px;
}

.feature-item .el-icon {
  background: rgba(255, 255, 255, 0.2);
  padding: 8px;
  border-radius: 10px;
}

.auth-aside-footer {
  font-size: 13px;
  opacity: 0.7;
}

.auth-main {
  flex: 1.2;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 48px;
  background: var(--surface-1);
}

.login-box {
  width: 100%;
  max-width: 360px;
}

.login-header {
  margin-bottom: 32px;
}

.login-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-strong);
  margin: 0 0 8px 0;
}

.login-subtitle {
  color: var(--text-muted);
  margin: 0;
}

.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 700;
  margin-top: 12px;
  margin-bottom: 24px;
}

.register-hint {
  text-align: center;
  font-size: 14px;
  color: var(--text-muted);
}

.register-hint a {
  margin-left: 8px;
  color: var(--brand-600);
  font-weight: 600;
  text-decoration: none;
}

/* Responsive */
@media (max-width: 900px) {
  .auth-aside {
    display: none;
  }
  .auth-container {
    width: 90%;
    max-width: 450px;
    height: auto;
  }
  .auth-main {
    padding: 40px 24px;
  }
}
</style>
