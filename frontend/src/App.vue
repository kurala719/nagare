<template>
  <el-config-provider :locale="elementLocale">
    <div class="app-wrapper">
      <router-view v-if="isStandaloneLayout" />
      <MainLayout v-else />
    </div>
  </el-config-provider>
</template>

<script>
import { defineComponent, computed, ref, onMounted, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElConfigProvider } from 'element-plus'
import en from 'element-plus/dist/locale/en.mjs'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

const MainLayout = defineAsyncComponent(() => import('@/layout/MainLayout.vue'))

export default defineComponent({
  name: 'App',
  components: {
    ElConfigProvider,
    MainLayout
  },
  setup() {
    const route = useRoute()
    const { locale } = useI18n()
    
    const elementLocale = computed(() => {
      try {
        return (locale.value === 'zh-CN' ? zhCn : en) || en;
      } catch (e) {
        return en;
      }
    })

    const isStandaloneLayout = computed(() => {
      // Default to true only for specific routes
      const layout = route.meta?.layout
      return layout === 'auth' || layout === 'status'
    })

    onMounted(() => {
      const storedTheme = localStorage.getItem('nagare_theme')
      const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      const useDark = storedTheme ? storedTheme === 'dark' : prefersDark
      
      const html = document.documentElement
      const body = document.body
      if (useDark) {
        html.classList.add('dark')
        body.classList.add('theme-dark')
      } else {
        html.classList.add('light')
        body.classList.add('theme-light')
      }

      const storedLocale = localStorage.getItem('nagare_locale')
      if (storedLocale) {
        locale.value = storedLocale
      }
    })

    return {
      elementLocale,
      isStandaloneLayout
    }
  }
})
</script>

<style>
/* Global Styles */
#app, .app-wrapper {
  width: 100%;
  height: 100vh;
  margin: 0;
  padding: 0;
  background-color: var(--color-background, #f8fafc);
}

:root {
  --text-strong: #1e293b;
  --border-1: #e2e8f0;
  --surface-3: #f1f5f9;
}

html.dark {
  --text-strong: #f8fafc;
  --border-1: #334155;
  --surface-3: #1e293b;
}

body {
  margin: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Scrollbar Styles */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

.dark ::-webkit-scrollbar-thumb {
  background: #475569;
}

.dark ::-webkit-scrollbar-thumb:hover {
  background: #64748b;
}
</style>
