import { createRouter,createWebHashHistory } from "vue-router";
import { ElMessageBox } from 'element-plus'
import i18n from '../i18n'
import { getToken, getUserPrivileges } from '@/utils/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { layout: 'auth' }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/Register.vue'),
    meta: { layout: 'auth' }
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('@/views/ResetPassword.vue'),
    meta: { requiresAuth: false, layout: 'auth' }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/Dashboard.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/alert',
    name: 'Alert',
    component: () => import('@/views/Alert.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/group',
    name: 'Group',
    component: () => import('@/views/Group.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/group/:id/detail',
    name: 'GroupDetail',
    component: () => import('@/views/GroupDetail.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/host',
    name: 'Host',
    component: () => import('@/views/Host.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/host/:id/detail',
    name: 'HostDetail',
    component: () => import('@/views/HostDetail.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/host/:id/terminal',
    name: 'Terminal',
    component: () => import('@/views/Terminal.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/terminal',
    name: 'MaintenanceTerminal',
    component: () => import('@/views/Terminal.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/site-messages',
    name: 'SiteMessages',
    component: () => import('@/views/SiteMessage.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/monitor',
    name: 'Monitor',
    component: () => import('@/views/Monitor.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/knowledge-base',
    name: 'KnowledgeBase',
    component: () => import('@/views/KnowledgeBase.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/alarm',
    name: 'Alarm',
    component: () => import('@/views/Alarm.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/item',
    name: 'Item',
    component: () => import('@/views/Item.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/item/:id/detail',
    name: 'ItemDetail',
    component: () => import('@/views/ItemDetail.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/host/:hostId/items',
    name: 'HostItems',
    redirect: (to) => ({
      path: '/item',
      query: { hostId: to.params.hostId }
    })
  },
  {
    path: '/provider',
    name: 'Provider',
    component: () => import('@/views/Provider.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/media',
    name: 'Media',
    component: () => import('@/views/Media.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/action',
    name: 'Action',
    component: () => import('@/views/Action.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/trigger',
    name: 'Trigger',
    component: () => import('@/views/Trigger.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/log',
    name: 'Log',
    component: () => import('@/views/Log.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/audit-log',
    name: 'AuditLog',
    component: () => import('@/views/AuditLog.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/system-status',
    name: 'SystemStatus',
    component: () => import('@/views/SystemStatus.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/analytics',
    name: 'Analytics',
    component: () => import('@/views/Analytics.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/user',
    name: 'User',
    component: () => import('@/views/User.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/register-application',
    name: 'RegisterApplication',
    component: () => import('@/views/RegisterApplication.vue'),
    meta: { requiresAuth: true, minPrivilege: 3 }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/Profile.vue'),
    meta: { requiresAuth: true, minPrivilege: 1 }
  },
  {
    path: '/config-settings',
    name: 'Configuration',
    component: () => import('@/views/Configuration.vue'),
    meta: { requiresAuth: true, minPrivilege: 3 }
  },
  {
    path: '/retention',
    name: 'Retention',
    component: () => import('@/views/Retention.vue'),
    meta: { requiresAuth: true, minPrivilege: 3 }
  },
  {
    path: '/status',
    name: 'StatusPage',
    component: () => import('@/views/StatusPage.vue'),
    meta: { requiresAuth: false, layout: 'status' }
  },
  {
    path: '/reports',
    name: 'Reports',
    component: () => import('@/views/Reports.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/ansible/playbooks',
    name: 'AnsiblePlaybooks',
    component: () => import('@/views/AnsiblePlaybook.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/ansible/jobs',
    name: 'AnsibleJobs',
    component: () => import('@/views/AnsibleJob.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/ansible/jobs/:id',
    name: 'AnsibleJobDetail',
    component: () => import('@/views/AnsibleJob.vue'),
    meta: { requiresAuth: true, minPrivilege: 2 }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  },
  {
    path: '/',
    redirect:'/dashboard'
  }
]

const router = createRouter({
  history:createWebHashHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const requiresAuth = to.meta?.requiresAuth
  const token = getToken()
  
  console.log(`Routing to: ${to.path}, RequiresAuth: ${requiresAuth}, HasToken: ${!!token}`)

  if (!requiresAuth) {
    return next()
  }
  if (!token) {
    const t = i18n?.global?.t || ((key) => key)
    await ElMessageBox.alert(t('common.loginToContinue'), t('common.unauthorizedTitle'), {
      confirmButtonText: t('common.ok'),
      type: 'warning',
    })
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  const minPrivilege = Number(to.meta?.minPrivilege ?? 0)
  const userPrivilege = getUserPrivileges()
  if (userPrivilege < minPrivilege) {
    const t = i18n?.global?.t || ((key) => key)
    await ElMessageBox.alert(t('common.insufficientPrivileges'), t('common.accessDeniedTitle'), {
      confirmButtonText: t('common.ok'),
      type: 'warning',
    })
    return next(false)
  }
  return next()
})


export default router
