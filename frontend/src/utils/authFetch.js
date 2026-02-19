import router from '../router'
import { ElMessageBox } from 'element-plus'
import { getToken, clearToken } from './auth'
import i18n from '../i18n'

let isAuthAlertOpen = false

export async function authFetch(url, options = {}) {
  let finalUrl = url
  if (typeof finalUrl === 'string' && finalUrl.startsWith('/') && !finalUrl.startsWith('/api/')) {
    finalUrl = `/api/v1${finalUrl}`
  } else if (typeof finalUrl === 'string' && finalUrl.startsWith('/api/') && !finalUrl.startsWith('/api/v1/')) {
    finalUrl = `/api/v1${finalUrl.slice('/api'.length)}`
  }
  const headers = new Headers(options.headers || {})
  headers.set('X-Tunnel-Skip-AntiPhishing-Page', 'true')
  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  const response = await fetch(finalUrl, { ...options, headers })
  if (response.status === 401 || response.status === 403) {
    const isForbidden = response.status === 403
    if (!isForbidden) {
      clearToken()
    }
    const redirect = router.currentRoute.value.fullPath
    const t = i18n?.global?.t || ((key) => key)
    if (!isAuthAlertOpen) {
      isAuthAlertOpen = true
      await ElMessageBox.alert(
        isForbidden ? t('common.insufficientPrivileges') : t('common.sessionExpired'),
        isForbidden ? t('common.accessDeniedTitle') : t('common.unauthorizedTitle'),
        {
        confirmButtonText: t('common.ok'),
        type: 'warning',
        }
      )
      isAuthAlertOpen = false
    }
    if (!isForbidden) {
      router.push({ path: '/login', query: { redirect } })
    }
  }
  return response
}
