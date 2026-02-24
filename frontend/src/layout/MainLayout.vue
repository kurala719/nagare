<template>
  <el-container class="layout-container">
    <el-header>
      <div class="header-content">
        <div class="header-left">
          <div class="toolbar-link menu-toggle" @click="toggleSidebar" :title="isSidebarCollapsed ? t('common.expand') : t('common.collapse')">
            <el-icon :size="20">
              <Expand v-if="isSidebarCollapsed" />
              <Fold v-else />
            </el-icon>
          </div>
          <div class="app-title">
            <el-icon :size="24" style="margin-right: 8px"><Monitor /></el-icon>
            <span>Nagare</span>
          </div>
          <div class="global-search-container">
            <GlobalSearch />
          </div>
        </div>
        <div class="toolbar">
          <SiteMessageCenter />
          <el-divider direction="vertical" />
          <el-dropdown @command="setLanguage">
            <span class="toolbar-link">
              <el-icon>
                <Setting />
              </el-icon>
              <span style="margin-left: 8px">{{ t('common.language') }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="en">English</el-dropdown-item>
                <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-divider direction="vertical" />
          <div class="toolbar-link" @click="toggleTheme">
            <el-icon>
              <Moon v-if="isDarkMode" />
              <Sunny v-else />
            </el-icon>
            <span style="margin-left: 8px">{{ isDarkMode ? t('common.night') : t('common.day') }}</span>
          </div>
          <el-divider direction="vertical" />
          <el-dropdown trigger="click" @command="handleUserCommand">
            <span class="toolbar-link">
              <el-avatar :size="24" :src="currentUserAvatar || undefined">
                <span class="toolbar-avatar-fallback">{{ currentUserInitials }}</span>
              </el-avatar>
              <span class="toolbar-username">{{ currentUserLabel }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-if="!isAuthenticated" command="login">{{ t('auth.login') }}</el-dropdown-item>
                <el-dropdown-item v-if="!isAuthenticated" command="register">{{ t('auth.register') }}</el-dropdown-item>
                <el-dropdown-item v-if="isAuthenticated" command="reset">{{ t('auth.reset') }}</el-dropdown-item>
                <el-dropdown-item v-if="isAuthenticated" command="profile">{{ t('menu.profile') }}</el-dropdown-item>
                <el-dropdown-item v-if="isAuthenticated" command="logout">{{ t('auth.logout') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>
    
    <el-container class="main-layout">
      <div
        v-if="isSidebarCollapsed"
        class="sidebar-edge-toggle"
        @click="toggleSidebar"
        :title="t('common.expand')"
      >
        <el-icon :size="18">
          <DArrowRight />
        </el-icon>
      </div>
      <el-aside
        :width="isSidebarCollapsed ? '0px' : '200px'"
        :class="['sidebar-transition', { 'is-collapsed': isSidebarCollapsed }]"
      >
        <el-scrollbar>
          <el-menu :default-active="$route.path" :collapse="isSidebarCollapsed" router>
            <template v-for="item in visibleMenuItems" :key="item.key">
              <el-sub-menu v-if="item.children" :index="item.key">
                <template #title>
                  <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                  <span>{{ t(item.label) }}</span>
                </template>
                <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
                  {{ t(child.label) }}
                </el-menu-item>
              </el-sub-menu>
              <el-menu-item v-else :index="item.path">
                <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
                <template #title>{{ t(item.label) }}</template>
              </el-menu-item>
            </template>
          </el-menu>
        </el-scrollbar>
      </el-aside>

      <el-main>
        <div class="main-container">
          <div class="content-area">
            <router-view v-slot="{ Component }">
              <component :is="Component" />
            </router-view>
          </div>
          <div
            v-if="isChatBarCollapsed"
            class="chat-edge-toggle"
            @click="toggleChatBar"
            :title="t('common.showChat')"
          >
            <el-icon :size="18">
              <DArrowLeft />
            </el-icon>
          </div>
          <div :class="['chat-bar-container', { 'is-collapsed': isChatBarCollapsed }]">
            <div
              v-if="!isChatBarCollapsed"
              class="chat-bar-toggle"
              @click="toggleChatBar"
              :title="t('common.hideChat')"
            >
              <el-icon :size="18">
                <DArrowRight />
              </el-icon>
            </div>
            <div class="chat-bar-content">
              <el-scrollbar>
                <SideBarChat />
              </el-scrollbar>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import { defineComponent, ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Expand, Fold, DArrowLeft, DArrowRight, Monitor, Setting, Moon, Sunny, Tools, PieChart, InfoFilled } from '@element-plus/icons-vue'
import SideBarChat from '@/components/Customed/SideBarChat.vue'
import GlobalSearch from '@/components/GlobalSearch.vue'
import SiteMessageCenter from '@/components/SiteMessageCenter.vue'
import { getUserProfile } from '@/api/users'
import { getUserPrivileges, getToken, clearToken, getUserClaims } from '@/utils/auth'

export default defineComponent({
  name: 'MainLayout',
  components: {
    GlobalSearch,
    SideBarChat,
    SiteMessageCenter,
    Expand,
    Fold,
    DArrowLeft,
    DArrowRight,
    Monitor,
    Setting,
    Moon,
    Sunny,
    Tools,
    PieChart,
    InfoFilled
  },
  setup() {
    const route = useRoute()
    const router = useRouter()
    const { t, locale } = useI18n()
    const userPrivilege = ref(getUserPrivileges())
    const authState = ref(!!getToken())
    const isAuthenticated = computed(() => authState.value)
    const isDarkMode = ref(false)
    const isSidebarCollapsed = ref(false)
    const isChatBarCollapsed = ref(false)
    
    const currentUserName = ref('')
    const currentUserAvatar = ref('')

    const currentUserLabel = computed(() => {
      if (!authState.value) return t('auth.guest')
      if (currentUserName.value) return currentUserName.value
      const claims = getUserClaims()
      return claims?.username || t('auth.guest')
    })

    const currentUserInitials = computed(() => {
      const source = (currentUserName.value || '').trim()
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

    const loadCurrentUserProfile = async () => {
      if (!authState.value) {
        currentUserAvatar.value = ''
        currentUserName.value = ''
        return
      }
      try {
        const { data } = await getUserProfile()
        const payload = data?.data || data
        currentUserAvatar.value = normalizeAvatarUrl(payload?.avatar || '')
        currentUserName.value = payload?.nickname || payload?.username || ''
      } catch {
        currentUserAvatar.value = ''
        currentUserName.value = ''
      }
    }

    const menuItems = computed(() => [
      { key: 'dashboard', path: '/dashboard', label: 'menu.databoard', minPrivilege: 1, icon: 'DataBoard' },
      { key: 'insights_group', label: 'menu.insights', minPrivilege: 1, icon: 'PieChart', children: [
        { key: 'systemStatus', path: '/system-status', label: 'menu.systemStatus', minPrivilege: 1 },
        { key: 'metricRacing', path: '/metric-racing', label: 'menu.metricRacing', minPrivilege: 1 },
        { key: 'analytics', path: '/analytics', label: 'menu.analytics', minPrivilege: 1 },
        { key: 'reports', path: '/reports', label: 'menu.reports', minPrivilege: 2 },
      ]},
      { key: 'inventory_group', label: 'menu.inventory', minPrivilege: 1, icon: 'Box', children: [
        { key: 'group', path: '/group', label: 'menu.group', minPrivilege: 1 },
        { key: 'host', path: '/host', label: 'menu.host', minPrivilege: 1 },
      ]},
      { key: 'observability_group', label: 'menu.observability', minPrivilege: 1, icon: 'Odometer', children: [
        { key: 'metrics', path: '/item', label: 'menu.item', minPrivilege: 1 },
        { key: 'monitor', path: '/monitor', label: 'menu.monitor', minPrivilege: 1 },
        { key: 'alarm', path: '/alarm', label: 'menu.alarm', minPrivilege: 1 },
      ]},
      { key: 'alerting_group', label: 'menu.alerting', minPrivilege: 1, icon: 'Bell', children: [
        { key: 'alert', path: '/alert', label: 'menu.alert', minPrivilege: 1 },
        { key: 'trigger', path: '/trigger', label: 'menu.trigger', minPrivilege: 2 },
        { key: 'action', path: '/action', label: 'menu.action', minPrivilege: 2 },
      ]},
      { key: 'tooling_group', label: 'menu.tooling', minPrivilege: 1, icon: 'Tools', children: [
        { key: 'terminal', path: '/terminal', label: 'menu.terminal', minPrivilege: 1 },
        { key: 'ansible', path: '/ansible/playbooks', label: 'menu.ansible', minPrivilege: 2 },
        { key: 'kb', path: '/knowledge-base', label: 'menu.kb', minPrivilege: 1 },
      ]},
      { key: 'system_group', label: 'menu.system', minPrivilege: 2, icon: 'Setting', children: [
        { key: 'media', path: '/media', label: 'menu.media', minPrivilege: 2 },
        { key: 'provider', path: '/provider', label: 'menu.provider', minPrivilege: 2 },
        { key: 'user', path: '/user', label: 'menu.user', minPrivilege: 2 },
        { key: 'log', path: '/log', label: 'menu.log', minPrivilege: 2 },
        { key: 'auditLog', path: '/audit-log', label: 'menu.auditLog', minPrivilege: 2 },
        { key: 'registerApplication', path: '/register-application', label: 'menu.registerApplication', minPrivilege: 3 },
        { key: 'retention', path: '/retention', label: 'retention.title', minPrivilege: 3 },
        { key: 'configuration', path: '/config-settings', label: 'menu.configuration', minPrivilege: 3 },
      ]},
      { key: 'profile', path: '/profile', label: 'menu.profile', minPrivilege: 1, icon: 'User' },
      { key: 'about', path: '/about', label: 'menu.about', minPrivilege: 1, icon: 'InfoFilled' },
    ])

    const visibleMenuItems = computed(() => {
      const privilege = userPrivilege.value
      return menuItems.value
        .filter((item) => privilege >= item.minPrivilege)
        .map((item) => {
          if (!item.children) return item
          const visibleChildren = item.children.filter((child) => privilege >= child.minPrivilege)
          if (visibleChildren.length === 0) return null
          return { ...item, children: visibleChildren }
        })
        .filter(Boolean)
    })

    const setLanguage = (lang) => {
      locale.value = lang
      localStorage.setItem('nagare_locale', lang)
    }

    const applyTheme = (dark) => {
      isDarkMode.value = dark
      const html = document.documentElement
      const body = document.body
      
      html.classList.remove('dark', 'light')
      body.classList.remove('theme-dark', 'theme-light')
      
      if (dark) {
        html.classList.add('dark')
        body.classList.add('theme-dark')
      } else {
        html.classList.add('light')
        body.classList.add('theme-light')
      }
      
      localStorage.setItem('nagare_theme', dark ? 'dark' : 'light')
    }

    const toggleTheme = () => {
      applyTheme(!isDarkMode.value)
    }

    const toggleSidebar = () => {
      isSidebarCollapsed.value = !isSidebarCollapsed.value
      localStorage.setItem('nagare_sidebar_collapsed', isSidebarCollapsed.value ? 'true' : 'false')
    }

    const toggleChatBar = () => {
      isChatBarCollapsed.value = !isChatBarCollapsed.value
      localStorage.setItem('nagare_chatbar_collapsed', isChatBarCollapsed.value ? 'true' : 'false')
    }

    const handleUserCommand = (command) => {
      switch (command) {
        case 'profile': router.push('/profile'); break;
        case 'login': router.push('/login'); break;
        case 'register': router.push('/register'); break;
        case 'reset': router.push('/reset-password'); break;
        case 'logout':
          clearToken();
          authState.value = false;
          router.replace('/login');
          break;
      }
    }

    watch(
      () => route.fullPath,
      () => {
        authState.value = !!getToken()
        userPrivilege.value = getUserPrivileges()
        loadCurrentUserProfile()
      }
    )

    onMounted(() => {
      const stored = localStorage.getItem('nagare_theme')
      const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      const useDark = stored ? stored === 'dark' : prefersDark
      applyTheme(useDark)

      const sidebarCollapsed = localStorage.getItem('nagare_sidebar_collapsed')
      if (sidebarCollapsed) isSidebarCollapsed.value = sidebarCollapsed === 'true'

      const chatBarCollapsed = localStorage.getItem('nagare_chatbar_collapsed')
      if (chatBarCollapsed) isChatBarCollapsed.value = chatBarCollapsed === 'true'

      const updateAuth = () => {
        authState.value = !!getToken()
        userPrivilege.value = getUserPrivileges()
        loadCurrentUserProfile()
      }
      window.addEventListener('auth-changed', updateAuth)
      onBeforeUnmount(() => {
        window.removeEventListener('auth-changed', updateAuth)
      })

      loadCurrentUserProfile()
    })

    return {
      t,
      visibleMenuItems,
      isAuthenticated,
      isDarkMode,
      isSidebarCollapsed,
      isChatBarCollapsed,
      currentUserLabel,
      currentUserName,
      currentUserAvatar,
      currentUserInitials,
      setLanguage,
      toggleTheme,
      toggleSidebar,
      toggleChatBar,
      handleUserCommand
    }
  }
})
</script>

<style scoped>
.layout-container {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.layout-container > .el-container {
  height: calc(100% - 60px);
  overflow: hidden;
}

.main-layout {
  position: relative;
}

.el-header {
  position: relative;
  background: var(--surface-1);
  color: var(--text-strong);
  height: 64px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border-1);
  box-shadow: var(--shadow-soft);
  backdrop-filter: var(--surface-blur);
  z-index: 100;
  padding: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 24px;
  flex: 1;
}

.global-search-container {
  flex: 1;
  max-width: 460px;
}

.app-title {
  display: flex;
  align-items: center;
  font-size: 22px;
  font-weight: 700;
  color: var(--brand-600);
  letter-spacing: -0.5px;
  font-family: var(--font-display);
}

.menu-toggle {
  padding: 10px;
  border-radius: 12px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid transparent;
  cursor: pointer;
  background: var(--surface-2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.menu-toggle:hover {
  background: var(--surface-3);
  border-color: var(--brand-200);
  transform: scale(1.05);
  color: var(--brand-600);
}

.el-aside {
  background: var(--surface-1);
  backdrop-filter: var(--surface-blur);
  border-right: 1px solid var(--border-1);
  height: 100%;
  flex-shrink: 0;
  min-width: 0;
  overflow: hidden;
  box-shadow: 4px 0 24px rgba(15, 23, 42, 0.04);
}

.el-aside :deep(.el-menu) {
  background: transparent;
  border-right: none;
  padding: 12px;
}

.el-aside :deep(.el-menu-item), .el-aside :deep(.el-sub-menu__title) {
  transition: all 0.3s ease;
  margin: 4px 0;
  border-radius: 12px;
  height: 48px;
  line-height: 48px;
  color: var(--text-muted);
}

.el-aside :deep(.el-menu-item:hover), .el-aside :deep(.el-sub-menu__title:hover) {
  background: var(--brand-50);
  color: var(--brand-600);
}

.el-aside :deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, var(--brand-500) 0%, var(--brand-600) 100%);
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(37, 99, 235, 0.2);
}

.sidebar-transition {
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-edge-toggle {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 28px;
  height: 48px;
  background: var(--brand-600);
  border-radius: 0 8px 8px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 20;
  transition: all 0.3s ease;
  box-shadow: 2px 0 10px rgba(37, 99, 235, 0.3);
}

.content-area {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  background: transparent;
}

.el-main {
  padding: 0 !important;
  overflow: hidden;
  height: 100%;
}

.main-container {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  position: relative;
}

.chat-bar-container {
  position: relative;
  width: 360px;
  height: 100%;
  border-left: 1px solid var(--border-1);
  background: var(--surface-1);
  backdrop-filter: var(--surface-blur);
  transition: width 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: visible;
  box-shadow: -4px 0 24px rgba(15, 23, 42, 0.04);
}

.chat-bar-container.is-collapsed {
  width: 0;
  border-left: none;
  box-shadow: none;
}

.chat-edge-toggle {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 48px;
  background: var(--brand-600);
  border-radius: 8px 0 0 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 20;
  transition: all 0.3s ease;
  box-shadow: -2px 0 10px rgba(37, 99, 235, 0.3);
}

.chat-edge-toggle:hover {
  background: var(--brand-700);
  transform: translateY(-50%) scale(1.05);
}

.chat-bar-toggle {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 32px;
  height: 48px;
  background: var(--brand-600);
  border-radius: 0 8px 8px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 10;
  transition: all 0.3s ease;
  box-shadow: 2px 0 10px rgba(37, 99, 235, 0.3);
}

.chat-bar-content {
  width: 360px;
  height: 100%;
  overflow: hidden;
  opacity: 1;
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
}

.toolbar-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-strong);
  padding: 0 12px;
  border-radius: 10px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  font-size: 14px;
  font-weight: 500;
  height: 40px;
  box-sizing: border-box;
}

.toolbar-link:hover {
  background: var(--surface-2);
  color: var(--brand-600);
}

.toolbar-link .el-icon {
  font-size: 18px;
}

.toolbar-username {
  margin-left: 10px;
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
}

:deep(.el-divider--vertical) {
  border-color: var(--border-1);
  height: 16px;
  margin: 0 8px;
  opacity: 0.6;
}
</style>
