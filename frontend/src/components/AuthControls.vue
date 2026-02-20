<template>
  <div class="auth-controls">
    <el-dropdown @command="setLanguage">
      <span class="control-btn">
        <el-icon><Setting /></el-icon>
        <span>{{ currentLanguageLabel }}</span>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="en">English</el-dropdown-item>
          <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <div class="control-btn theme-toggle" @click="toggleTheme">
      <el-icon>
        <Moon v-if="isDarkMode" />
        <Sunny v-else />
      </el-icon>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Setting, Moon, Sunny } from '@element-plus/icons-vue'

const { locale } = useI18n()
const isDarkMode = ref(false)

const currentLanguageLabel = computed(() => {
  return locale.value === 'zh-CN' ? '中文' : 'English'
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

onMounted(() => {
  const storedTheme = localStorage.getItem('nagare_theme')
  const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(storedTheme ? storedTheme === 'dark' : prefersDark)
})
</script>

<style scoped>
.auth-controls {
  position: absolute;
  top: 24px;
  right: 24px;
  display: flex;
  gap: 12px;
  z-index: 100;
}

.control-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--surface-1);
  border: 1px solid var(--border-1);
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--text-strong);
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: var(--shadow-soft);
}

.control-btn:hover {
  background: var(--surface-2);
  border-color: var(--brand-200);
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.theme-toggle {
  padding: 8px;
  width: 40px;
  justify-content: center;
}
</style>
