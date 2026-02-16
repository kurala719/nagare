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
          <el-dropdown @command="setLanguage">
            <span class="toolbar-link">
              <el-icon style="margin-right: 8px; margin-top: 1px">
                <setting />
              </el-icon>
              {{ t('common.language') }}
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
            <el-icon style="margin-right: 6px; margin-top: 1px">
              <moon v-if="isDarkMode" />
              <sunny v-else />
            </el-icon>
            {{ isDarkMode ? t('common.night') : t('common.day') }}
          </div>
          <el-divider direction="vertical" />
          <el-dropdown trigger="click" @command="handleUserCommand">
            <span class="toolbar-link">
              <el-avatar :size="28" />
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
import { Expand, Fold, DArrowLeft, DArrowRight, Monitor, Setting, Moon, Sunny } from '@element-plus/icons-vue'
import SideBarChat from '@/components/Customed/SideBarChat.vue'
import GlobalSearch from '@/components/GlobalSearch.vue'
import { getUserPrivileges, getToken, clearToken, getUserClaims } from '@/utils/auth'

export default defineComponent({
  name: 'MainLayout',
  components: {
    GlobalSearch,
    SideBarChat,
    Expand,
    Fold,
    DArrowLeft,
    DArrowRight,
    Monitor,
    Setting,
    Moon,
    Sunny
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
    
    const currentUserLabel = computed(() => {
      if (!authState.value) return t('auth.guest')
      const claims = getUserClaims()
      return claims?.username || t('auth.guest')
    })

    const menuItems = computed(() => [
      { key: 'dashboard', path: '/dashboard', label: 'menu.databoard', minPrivilege: 1, icon: 'DataBoard' },
      { key: 'monitor_group', label: 'menu.monitor', minPrivilege: 1, icon: 'Monitor', children: [
        { key: 'monitor', path: '/monitor', label: 'menu.monitor', minPrivilege: 1 },
        { key: 'host', path: '/host', label: 'menu.host', minPrivilege: 1 },
        { key: 'site', path: '/site', label: 'menu.site', minPrivilege: 1 },
        { key: 'item', path: '/item', label: 'menu.item', minPrivilege: 1 },
      ]},
      { key: 'alert_group', label: 'menu.alert', minPrivilege: 1, icon: 'Bell', children: [
        { key: 'alert', path: '/alert', label: 'menu.alert', minPrivilege: 1 },
        { key: 'trigger', path: '/trigger', label: 'menu.trigger', minPrivilege: 2 },
        { key: 'action', path: '/action', label: 'menu.action', minPrivilege: 2 },
      ]},
      { key: 'media_group', label: 'menu.media', minPrivilege: 2, icon: 'Message', children: [
        { key: 'provider', path: '/provider', label: 'menu.provider', minPrivilege: 2 },
        { key: 'media', path: '/media', label: 'menu.media', minPrivilege: 2 },
        { key: 'mediaType', path: '/media-type', label: 'menu.mediaType', minPrivilege: 2 },
      ]},
      { key: 'system_group', label: 'menu.system', minPrivilege: 2, icon: 'Setting', children: [
        { key: 'user', path: '/user', label: 'menu.user', minPrivilege: 2 },
        { key: 'log', path: '/log', label: 'menu.log', minPrivilege: 2 },
        { key: 'registerApplication', path: '/register-application', label: 'menu.registerApplication', minPrivilege: 3 },
        { key: 'system', path: '/system', label: 'menu.system', minPrivilege: 3 },
      ]},
      { key: 'profile', path: '/profile', label: 'menu.profile', minPrivilege: 1, icon: 'User' },
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
      }
      window.addEventListener('auth-changed', updateAuth)
      onBeforeUnmount(() => {
        window.removeEventListener('auth-changed', updateAuth)
      })
    })

    return {
      t,
      visibleMenuItems,
      isAuthenticated,
      isDarkMode,
      isSidebarCollapsed,
      isChatBarCollapsed,
      currentUserLabel,
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
  background: rgba(255, 255, 255, 0.9);
  color: var(--text-strong);
  height: 60px;
  flex-shrink: 0;
  border-bottom: 1px solid var(--border-1);
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(12px);
  z-index: 100;
  padding: 0;
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  flex: 1;
}

.global-search-container {
  flex: 1;
  max-width: 400px;
  margin-left: 20px;
}

.app-title {
  display: flex;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-strong);
  letter-spacing: 0.5px;
}

.menu-toggle {
  padding: 8px;
  border-radius: 8px;
  transition: all 0.3s ease;
  border: 1px solid transparent;
  cursor: pointer;
}

.menu-toggle:hover {
  background: var(--surface-3);
  border-color: var(--border-1);
  transform: scale(1.1);
}

.el-aside {
  color: var(--el-text-color-primary);
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%);
  border-right: 1px solid var(--border-1);
  height: 100%;
  flex-shrink: 0;
  min-width: 0;
  overflow: hidden;
  box-shadow: 2px 0 14px rgba(15, 23, 42, 0.06);
}

.el-aside :deep(.el-menu) {
  background: transparent;
  border-right: none;
}

.el-aside :deep(.el-menu-item), .el-aside :deep(.el-sub-menu__title) {
  transition: all 0.3s ease;
  margin: 4px 8px;
  border-radius: 8px;
}

.el-aside :deep(.el-menu-item:hover), .el-aside :deep(.el-sub-menu__title:hover) {
  background: rgba(37, 99, 235, 0.08);
  transform: translateX(4px);
}

.el-aside :deep(.el-menu-item.is-active) {
  background: linear-gradient(180deg, #2563eb 0%, #1d4ed8 100%);
  color: #ffffff;
  font-weight: 600;
}

.sidebar-transition {
  transition: width 0.3s ease;
}

.sidebar-transition.is-collapsed {
  border-right: none;
  box-shadow: none;
}

.sidebar-transition.is-collapsed :deep(.el-scrollbar) {
  opacity: 0;
  pointer-events: none;
}

.sidebar-edge-toggle {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 36px;
  height: 56px;
  background: linear-gradient(180deg, #2563eb 0%, #1d4ed8 100%);
  border-radius: 0 10px 10px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 20;
  transition: all 0.3s ease;
  box-shadow: 2px 6px 16px rgba(37, 99, 235, 0.3);
}

.sidebar-edge-toggle:hover {
  background: linear-gradient(180deg, #1d4ed8 0%, #1e40af 100%);
  transform: translateY(-50%) scale(1.05);
  box-shadow: 2px 6px 18px rgba(30, 64, 175, 0.35);
}

.el-main {
  padding: 0;
  height: 100%;
  overflow: hidden;
}

.main-container {
  display: flex;
  height: 100%;
  width: 100%;
  position: relative;
  overflow: visible;
}

.content-area {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  background: transparent;
  padding: 20px;
}

.chat-bar-container {
  position: relative;
  width: 320px;
  height: 100%;
  border-left: 1px solid var(--border-1);
  background: linear-gradient(180deg, #ffffff 0%, #f1f5f9 100%);
  transition: width 0.3s ease;
  overflow: visible;
  box-shadow: -2px 0 14px rgba(15, 23, 42, 0.06);
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
  width: 48px;
  height: 64px;
  background: linear-gradient(180deg, #2563eb 0%, #1d4ed8 100%);
  border-radius: 12px 0 0 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 20;
  transition: all 0.3s ease;
  box-shadow: -2px 6px 16px rgba(37, 99, 235, 0.3);
}

.chat-edge-toggle:hover {
  background: linear-gradient(180deg, #1d4ed8 0%, #1e40af 100%);
  transform: translateY(-50%) scale(1.05);
  box-shadow: -2px 6px 18px rgba(30, 64, 175, 0.35);
}

.chat-bar-toggle {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 48px;
  height: 64px;
  background: linear-gradient(180deg, #2563eb 0%, #1d4ed8 100%);
  border-radius: 0 12px 12px 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  z-index: 10;
  transition: all 0.3s ease;
  box-shadow: 2px 6px 16px rgba(37, 99, 235, 0.3);
}

.chat-bar-toggle:hover {
  background: linear-gradient(180deg, #1d4ed8 0%, #1e40af 100%);
  transform: translateY(-50%) scale(1.05);
  box-shadow: 2px 6px 18px rgba(30, 64, 175, 0.35);
}

.chat-bar-content {
  width: 320px;
  height: 100%;
  overflow: hidden;
  padding-left: 48px;
  opacity: 1;
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.chat-bar-container.is-collapsed .chat-bar-content {
  opacity: 0;
  pointer-events: none;
  transform: translateX(100%);
}

.toolbar {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  height: 100%;
}

.toolbar-link {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  color: var(--text-strong);
  padding: 8px 12px;
  border-radius: 8px;
  transition: all 0.3s ease;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid transparent;
}

.toolbar-link:hover {
  background: var(--surface-3);
  border-color: var(--border-1);
  transform: translateY(-2px);
}

.toolbar-username {
  margin-left: 8px;
  font-size: 13px;
  font-weight: 500;
}

:deep(.el-divider--vertical) {
  background: rgba(15, 23, 42, 0.12);
  height: 20px;
  margin: 0;
}
</style>