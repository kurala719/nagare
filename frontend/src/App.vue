<template>
  <el-config-provider :locale="elementLocale">
    <router-view v-if="isAuthLayout" />
    <MainLayout v-else />
  </el-config-provider>
</template>

<script>
import { defineComponent, computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElConfigProvider } from 'element-plus'
import en from 'element-plus/dist/locale/en.mjs'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import MainLayout from '@/layout/MainLayout.vue'

export default defineComponent({
  name: 'App',
  components: {
    ElConfigProvider,
    MainLayout
  },
  setup() {
    const route = useRoute()
    const { locale } = useI18n()
    const elementLocale = computed(() => (locale.value === 'zh-CN' ? zhCn : en))
    const isAuthLayout = computed(() => route.meta?.layout === 'auth')

    onMounted(() => {
      // Initialize theme from local storage
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

      // Initialize locale
      const storedLocale = localStorage.getItem('nagare_locale')
      if (storedLocale) {
        locale.value = storedLocale
      }
    })

    return {
      elementLocale,
      isAuthLayout
    }
  }
})
</script>

<style>
/* Global Styles */
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
