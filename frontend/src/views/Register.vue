<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">{{ $t('auth.registerTitle') }}</h2>
      <el-form :model="form" @submit.prevent>
        <el-form-item :label="$t('auth.username')">
          <el-input v-model="form.username" autocomplete="username" />
        </el-form-item>
        <el-form-item :label="$t('auth.password')">
          <el-input v-model="form.password" type="password" autocomplete="new-password" show-password />
        </el-form-item>
        <el-form-item :label="$t('auth.confirmPassword')">
          <el-input v-model="form.confirm" type="password" autocomplete="new-password" show-password />
        </el-form-item>
        <el-button type="primary" class="auth-submit" :loading="loading" @click="onRegister">
          {{ $t('auth.submit') }}
        </el-button>
        <div class="auth-links">
          <router-link to="/login">{{ $t('auth.backToLogin') }}</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { registerUser } from '@/api/users'

const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const form = reactive({
  username: '',
  password: '',
  confirm: ''
})

const onRegister = async () => {
  if (!form.username || !form.password) {
    ElMessage.warning(t('auth.enterCredentials'))
    return
  }
  if (form.password !== form.confirm) {
    ElMessage.warning(t('auth.passwordsMismatch'))
    return
  }
  loading.value = true
  try {
    await registerUser({ username: form.username, password: form.password })
    ElMessage.success(t('auth.applicationSubmitted'))
    router.replace('/login')
  } catch (err) {
    ElMessage.error(err?.response?.data?.error || err.message || t('auth.registrationFailed'))
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
  margin-top: 12px;
  font-size: 12px;
  text-align: center;
}
</style>
