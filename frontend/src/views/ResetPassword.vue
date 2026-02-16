<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">{{ $t('auth.resetTitle') }}</h2>
      <el-form :model="form" @submit.prevent>
        <el-form-item :label="$t('auth.oldPassword')">
          <el-input v-model="form.oldPassword" type="password" autocomplete="current-password" show-password />
        </el-form-item>
        <el-form-item :label="$t('auth.newPassword')">
          <el-input v-model="form.newPassword" type="password" autocomplete="new-password" show-password />
        </el-form-item>
        <el-form-item :label="$t('auth.confirmNewPassword')">
          <el-input v-model="form.confirm" type="password" autocomplete="new-password" show-password />
        </el-form-item>
        <el-button type="primary" class="auth-submit" :loading="loading" @click="onReset">
          {{ $t('auth.reset') }}
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
import { resetPassword } from '@/api/users'

const router = useRouter()
const { t } = useI18n()
const loading = ref(false)
const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirm: ''
})

const onReset = async () => {
  if (!form.oldPassword || !form.newPassword) {
    ElMessage.warning(t('auth.completeAllFields'))
    return
  }
  if (form.newPassword !== form.confirm) {
    ElMessage.warning(t('auth.passwordsMismatch'))
    return
  }
  loading.value = true
  try {
    await resetPassword({
      old_password: form.oldPassword,
      new_password: form.newPassword
    })
    ElMessage.success(t('auth.resetSuccess'))
    router.replace('/login')
  } catch (err) {
    ElMessage.error(err?.response?.data?.error || err.message || t('auth.resetFailed'))
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
  width: 400px;
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
