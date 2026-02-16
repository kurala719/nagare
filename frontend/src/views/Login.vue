<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">{{ $t('auth.signInTitle') }}</h2>
      <el-form :model="form" @submit.prevent>
        <el-form-item :label="$t('auth.username')">
          <el-input v-model="form.username" autocomplete="username" />
        </el-form-item>
        <el-form-item :label="$t('auth.password')">
          <el-input v-model="form.password" type="password" autocomplete="current-password" show-password />
        </el-form-item>
        <el-button type="primary" class="auth-submit" :loading="loading" @click="onLogin">
          {{ $t('auth.login') }}
        </el-button>
        <div class="auth-links">
          <router-link to="/register">{{ $t('auth.createAccount') }}</router-link>
          <router-link to="/reset-password">{{ $t('auth.resetLink') }}</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { loginUser } from '@/api/users'
import { setToken } from '@/utils/auth'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const loading = ref(false)
const form = reactive({
  username: '',
  password: ''
})

const onLogin = async () => {
  if (!form.username || !form.password) {
    ElMessage.warning(t('auth.enterCredentials'))
    return
  }
  loading.value = true
  try {
    const { data } = await loginUser(form)
    const token = data?.data?.token || data?.token
    if (!token) {
      throw new Error(t('auth.missingToken'))
    }
    setToken(token)
    const redirect = route.query.redirect || '/dashboard'
    router.replace(redirect)
  } catch (err) {
    ElMessage.error(err?.response?.data?.error || err.message || t('auth.loginFailed'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f0f5ff, #ffffff);
}

.auth-card {
  width: 380px;
  padding: 18px 24px;
}

.auth-title {
  margin-bottom: 16px;
  text-align: center;
}

.auth-submit {
  width: 100%;
}

.auth-links {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  font-size: 12px;
}
</style>
