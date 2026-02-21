/*引入axios*/

import axios from 'axios'
import router from '../router'
import { ElMessageBox } from 'element-plus'
import { getToken, clearToken } from './auth'
import i18n from '../i18n'

// Determine API base URL based on environment
const getApiBaseURL = () => {
  // In development, use the proxy (empty baseURL)
  // In production or preview, use the full backend URL
  if (import.meta.env.DEV) {
    return ''
  }
  // For preview or production, communicate directly with backend
  return import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
}

const request = axios.create({
    baseURL: getApiBaseURL(),
    timeout: 30000, // 30 seconds timeout
    withCredentials: false, // 表示请求可以携带cookie
    headers: {
        'X-Tunnel-Skip-AntiPhishing-Page': 'true'
    }
})

request.interceptors.request.use((config) => {
    // Ensure URL starts with /api/v1 if it doesn't have it
    if (config.url && !config.url.startsWith('/api/v1')) {
        config.url = `/api/v1${config.url.startsWith('/') ? '' : '/'}${config.url}`
    }
    
    const token = getToken()
    if (token) {
        config.headers = config.headers || {}
        config.headers.Authorization = `Bearer ${token}`
    } else if (config.headers && config.headers.Authorization) {
        delete config.headers.Authorization
    }
    return config
})

let isAuthAlertOpen = false

// Map backend error messages to i18n keys
const mapErrorToI18nKey = (errorMessage) => {
    if (!errorMessage) return null
    
    const errorMap = {
        'invalid email format': 'common.invalidEmail',
        'password must be at least 8 characters and include 3 of: lowercase, uppercase, digits, special characters': 'common.weakPassword',
        'username must be 3-32 characters, alphanumeric with underscores/hyphens only': 'common.invalidUsername',
        'invalid input': 'common.invalidInput',
        'resource not found': 'common.operationFailed',
        'unauthorized': 'common.unauthorizedTitle',
        'forbidden': 'common.accessDeniedTitle',
        'resource already exists': 'common.operationFailed'
    }
    
    const lowerMsg = errorMessage.toLowerCase()
    return errorMap[lowerMsg] || null
}

request.interceptors.response.use(
    (response) => response.data,
    async (error) => {
        const status = error?.response?.status
        if (status === 401 || status === 403) {
            const isForbidden = status === 403
            if (!isForbidden) {
                clearToken()
            }
            const redirect = router.currentRoute.value.fullPath
            const t = i18n?.global?.t || ((key) => key)
            if (!isAuthAlertOpen) {
                isAuthAlertOpen = true
                await ElMessageBox.alert(
                    isForbidden
                        ? t('common.insufficientPrivileges')
                        : t('common.sessionExpired'),
                    isForbidden ? t('common.accessDeniedTitle') : t('common.unauthorizedTitle'),
                    {
                        confirmButtonText: t('common.ok'),
                        type: 'warning',
                    }
                )
                isAuthAlertOpen = false
            }
            if (!isForbidden) {
                router.replace({ path: '/login', query: { redirect } })
            }
        }
        
        // Translate backend error messages to i18n keys
        if (error?.response?.data?.error) {
            const i18nKey = mapErrorToI18nKey(error.response.data.error)
            if (i18nKey) {
                const t = i18n?.global?.t || ((key) => key)
                error.response.data.translatedError = t(i18nKey)
            }
        }
        
        return Promise.reject(error)
    }
)
//前端采用export.default，在写后端代码时用module.export

export { mapErrorToI18nKey }
export default request
