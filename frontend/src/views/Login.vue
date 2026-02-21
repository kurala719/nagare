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
              <el-icon><Cpu /></el-icon>
              <span>Asset Management</span>
            </div>
            <div class="feature-item animate-slide-up delay-4" style="animation-delay: 0.5s">
              <el-icon><Odometer /></el-icon>
              <span>Deep Observability</span>
            </div>
            <div class="feature-item animate-slide-up delay-4" style="animation-delay: 0.6s">
              <el-icon><MagicStick /></el-icon>
              <span>AI Guard Analysis</span>
            </div>
          </div>
        </div>
        <div class="auth-aside-footer animate-fade-in" style="animation-delay: 0.8s">
          &copy; 2026 Nagare Project. All rights reserved.
        </div>
      </div>

      <!-- Right side: Login Form -->
      <div class="auth-main">
        <div class="login-box animate-slide-up delay-2">
          <div class="login-header">
            <h2 class="login-title">{{ $t('auth.welcomeBack') }}</h2>
            <p class="login-subtitle">{{ $t('auth.signInTitle') }}</p>
          </div>

          <el-form 
            ref="loginFormRef"
            :model="form" 
            :rules="rules"
            label-position="top"
            @keyup.enter="onLogin"
          >
            <el-form-item :label="$t('auth.username')" prop="username">
              <el-input 
                v-model="form.username" 
                placeholder="Enter your username"
                :prefix-icon="User"
                autocomplete="username"
              />
            </el-form-item>
            
            <el-form-item :label="$t('auth.password')" prop="password">
              <el-input 
                v-model="form.password" 
                type="password" 
                placeholder="Enter your password"
                :prefix-icon="Lock"
                autocomplete="current-password" 
                show-password 
              />
            </el-form-item>

            <div class="login-options">
              <el-checkbox v-model="rememberMe">{{ $t('auth.rememberMe') }}</el-checkbox>
              <router-link to="/reset-password" class="forgot-link">{{ $t('auth.resetLink') }}</router-link>
            </div>

            <el-button 
              type="primary" 
              class="submit-btn" 
              :loading="loading" 
              @click="onLogin"
            >
              {{ $t('auth.login') }}
            </el-button>

            <div class="register-hint">
              <span>Don't have an account?</span>
              <router-link to="/register">{{ $t('auth.createAccount') }}</router-link>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { User, Lock, Monitor, Cpu, Odometer, MagicStick } from '@element-plus/icons-vue'
import AuthControls from '@/components/AuthControls.vue'
import AnimatedBackground from '@/components/Customed/AnimatedBackground.vue'
import { loginUser } from '@/api/users'
import { setToken } from '@/utils/auth'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const loading = ref(false)
const loginFormRef = ref(null)
const rememberMe = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: t('auth.username') + ' is required', trigger: 'blur' }],
  password: [{ required: true, message: t('auth.password') + ' is required', trigger: 'blur' }]
}

const onLogin = async () => {
  if (!loginFormRef.value) return
  
  try {
    await loginFormRef.value.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    const response = await loginUser(form)
    const data = response.data || response
    const token = data?.data?.token || data?.token
    
    if (!token) {
      throw new Error(t('auth.missingToken'))
    }
    
    // Store token - this triggers notifyAuthChanged()
    setToken(token)
    
    if (rememberMe.value) {
      localStorage.setItem('nagare_remembered_user', form.username)
    } else {
      localStorage.removeItem('nagare_remembered_user')
    }

    ElMessage.success(t('auth.welcomeBack'))
    
    // Crucial: Wait for state to settle
    await nextTick()
    
    const redirect = route.query.redirect || '/dashboard'
    console.log('Redirecting to:', redirect)
    
    // Attempt router redirect
    try {
      await router.push(redirect)
    } catch (routerError) {
      console.warn('Router push failed, trying window.location:', routerError)
      // Fallback if router fails
      window.location.hash = redirect
    }
    
  } catch (err) {
    console.error('Login error:', err)
    const errorMsg = err?.response?.data?.error || err.message || t('auth.loginFailed')
    ElMessage.error(errorMsg)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const remembered = localStorage.getItem('nagare_remembered_user')
  if (remembered) {
    form.username = remembered
    rememberMe.value = true
  }
})
</script>

<style scoped>
.auth-wrapper {
  min-height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  /* background: transparent !important; Removed to allow base color fallback if needed */
}

.auth-container {
  display: flex;
  width: 1000px;
  height: 600px;
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

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.forgot-link {
  font-size: 14px;
  color: var(--brand-600);
  text-decoration: none;
}

.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 700;
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
