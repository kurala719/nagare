/*引入axios*/

import axios from 'axios'
import router from '../router'
import { ElMessageBox } from 'element-plus'
import { getToken, clearToken } from './auth'
import i18n from '../i18n'
const request = axios.create({
    baseURL: '', // Handled by interceptor to avoid double prefixing
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
        return Promise.reject(error)
    }
)
//前端采用export.default，在写后端代码时用module.export

export default request
