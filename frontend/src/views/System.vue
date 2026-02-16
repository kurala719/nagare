<template>
  <div class="system-page">
    <div class="page-header">
      <h2>{{ $t('system.title') || 'System' }}</h2>
    </div>

    <el-tabs v-model="activeTabs" class="system-tabs">
      <!-- Configuration Tab -->
      <el-tab-pane :label="$t('system.configuration') || 'Configuration'" name="config">
        <div v-if="configLoading" style="text-align: center; padding: 40px;">
          <el-icon class="is-loading" size="50" color="#409EFF">
            <Loading />
          </el-icon>
          <p style="margin-top: 16px; color: #909399;">{{ $t('system.loading') || 'Loading configuration...' }}</p>
        </div>

        <el-alert 
          v-if="configError && !configLoading" 
          :title="configError" 
          type="error" 
          show-icon
          style="margin: 20px;"
          :closable="false"
        >
          <template #default>
            <el-button size="small" @click="loadConfiguration">{{ $t('system.retry') || 'Retry' }}</el-button>
          </template>
        </el-alert>

        <el-form 
          v-if="!configLoading && !configError && config" 
          :model="config" 
          label-width="200px"
          style="padding: 20px; max-width: 800px;"
        >
          <!-- System Configuration Section -->
          <div style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.systemSettings') || 'System Settings' }}</h4>

            <el-form-item :label="$t('system.systemName') || 'System Name'" v-if="config.system.system_name !== undefined">
              <el-input 
                v-model="config.system.system_name" 
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.ipAddress') || 'IP Address'" v-if="config.system.ip_address !== undefined">
              <el-input 
                v-model="config.system.ip_address" 
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.availability') || 'Availability'" v-if="config.system.availability !== undefined">
              <el-switch 
                v-model="config.system.availability" 
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.port') || 'Port'" v-if="config.system.port !== undefined">
              <el-input 
                v-model="config.system.port" 
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- Database Configuration Section -->
          <div v-if="config.database" style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.databaseSettings') || 'Database Settings' }}</h4>

            <el-form-item :label="$t('system.dbVersion') || 'Version'" v-if="config.database.version !== undefined">
              <el-input 
                v-model="config.database.version"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.dbHost') || 'Host'" v-if="config.database.host !== undefined">
              <el-input 
                v-model="config.database.host"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.dbPort') || 'Port'" v-if="config.database.port !== undefined">
              <el-input 
                v-model="config.database.port"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.dbUser') || 'Username'" v-if="config.database.username !== undefined">
              <el-input 
                v-model="config.database.username"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.dbPassword') || 'Password'" v-if="config.database.password !== undefined">
              <el-input 
                v-model="config.database.password"
                type="password"
                show-password
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.dbName') || 'Database Name'" v-if="config.database.database_name !== undefined">
              <el-input 
                v-model="config.database.database_name"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- Sync Configuration Section -->
          <div v-if="config.sync" style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.syncSettings') || 'Sync Settings' }}</h4>

            <el-form-item :label="$t('system.syncEnabled') || 'Enabled'" v-if="config.sync.enabled !== undefined">
              <el-switch v-model="config.sync.enabled" :disabled="!isEditing" />
            </el-form-item>

            <el-form-item :label="$t('system.syncInterval') || 'Interval (seconds)'" v-if="config.sync.interval_seconds !== undefined">
              <el-input 
                v-model="config.sync.interval_seconds"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.syncConcurrency') || 'Concurrency'" v-if="config.sync.concurrency !== undefined">
              <el-input 
                v-model="config.sync.concurrency"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- Status Check Configuration Section -->
          <div v-if="config.status_check" style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.statusCheckSettings') || 'Status Check Settings' }}</h4>

            <el-form-item :label="$t('system.statusCheckEnabled') || 'Enabled'" v-if="config.status_check.enabled !== undefined">
              <el-switch v-model="config.status_check.enabled" :disabled="!isEditing" />
            </el-form-item>

            <el-form-item :label="$t('system.statusCheckInterval') || 'Interval (seconds)'" v-if="config.status_check.interval_seconds !== undefined">
              <el-input 
                v-model="config.status_check.interval_seconds"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.statusCheckConcurrency') || 'Concurrency'" v-if="config.status_check.concurrency !== undefined">
              <el-input 
                v-model="config.status_check.concurrency"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- MCP Configuration Section -->
          <div v-if="config.mcp" style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.mcpSettings') || 'MCP Settings' }}</h4>

            <el-form-item :label="$t('system.mcpEnabled') || 'Enabled'" v-if="config.mcp.enabled !== undefined">
              <el-switch v-model="config.mcp.enabled" :disabled="!isEditing" />
            </el-form-item>

            <el-form-item :label="$t('system.mcpApiKey') || 'API Key'" v-if="config.mcp.api_key !== undefined">
              <el-input 
                v-model="config.mcp.api_key"
                type="password"
                show-password
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.mcpMaxConcurrency') || 'Max Concurrency'" v-if="config.mcp.max_concurrency !== undefined">
              <el-input 
                v-model="config.mcp.max_concurrency"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- AI Configuration Section -->
          <div v-if="config.ai" style="margin-bottom: 30px;">
            <h4 style="margin-bottom: 16px; color: #303133;">{{ $t('system.aiSettings') || 'AI Settings' }}</h4>

            <el-form-item :label="$t('system.aiEnabled') || 'Analysis Enabled'" v-if="config.ai.analysis_enabled !== undefined">
              <el-switch v-model="config.ai.analysis_enabled" :disabled="!isEditing" />
            </el-form-item>

            <el-form-item :label="$t('system.aiProviderId') || 'Provider ID'" v-if="config.ai.provider_id !== undefined">
              <el-input 
                v-model="config.ai.provider_id"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.aiModel') || 'Model'" v-if="config.ai.model !== undefined">
              <el-input 
                v-model="config.ai.model"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.aiTimeout') || 'Timeout (seconds)'" v-if="config.ai.analysis_timeout_seconds !== undefined">
              <el-input 
                v-model="config.ai.analysis_timeout_seconds"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>

            <el-form-item :label="$t('system.aiMinSeverity') || 'Min Severity'" v-if="config.ai.analysis_min_severity !== undefined">
              <el-input 
                v-model="config.ai.analysis_min_severity"
                type="number"
                :disabled="!isEditing"
              />
            </el-form-item>
          </div>

          <!-- Action Buttons -->
          <div style="margin-top: 30px; display: flex; gap: 12px;">
            <el-button 
              v-if="!isEditing"
              type="primary" 
              @click="isEditing = true"
            >
              {{ $t('system.edit') || 'Edit' }}
            </el-button>
            <template v-else>
              <el-button 
                type="primary" 
                @click="saveConfiguration"
                :loading="configSaving"
              >
                {{ $t('system.save') || 'Save' }}
              </el-button>
              <el-button @click="cancelEdit">{{ $t('system.cancel') || 'Cancel' }}</el-button>
            </template>
            <el-button
              :loading="configPersisting"
              @click="persistConfiguration"
            >
              {{ $t('system.persist') || 'Save to Disk' }}
            </el-button>
            <el-button 
              @click="reloadConfiguration"
              :loading="configReloading"
            >
              {{ $t('system.reload') || 'Reload' }}
            </el-button>
          </div>
        </el-form>
      </el-tab-pane>

      <!-- Display Tab -->
      <el-tab-pane :label="$t('system.display') || 'Display'" name="display">
        <div style="padding: 20px; max-width: 800px;">
          <el-descriptions :column="1" border>
            <el-descriptions-item :label="$t('system.currentTheme') || 'Current Theme'">
              <span>{{ currentTheme === 'dark' ? $t('system.darkMode') || 'Dark' : $t('system.lightMode') || 'Light' }}</span>
              <el-button 
                link 
                size="small" 
                @click="toggleTheme"
                style="margin-left: 12px;"
              >
                {{ $t('system.switch') || 'Switch' }}
              </el-button>
            </el-descriptions-item>

            <el-descriptions-item :label="$t('system.currentLocale') || 'Current Language'">
              <span>{{ currentLocale === 'en' ? 'English' : '中文' }}</span>
              <el-button 
                link 
                size="small" 
                @click="toggleLocale"
                style="margin-left: 12px;"
              >
                {{ $t('system.switch') || 'Switch' }}
              </el-button>
            </el-descriptions-item>

            <el-descriptions-item :label="$t('system.sidebarCollapsed') || 'Sidebar Collapsed'">
              {{ sidebarCollapsed ? $t('common.yes') || 'Yes' : $t('common.no') || 'No' }}
            </el-descriptions-item>

            <el-descriptions-item :label="$t('system.chatbarCollapsed') || 'Chat Bar Collapsed'">
              {{ chatbarCollapsed ? $t('common.yes') || 'Yes' : $t('common.no') || 'No' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { authFetch } from '@/utils/authFetch'

const { t, locale } = useI18n()

// State
const activeTabs = ref('config')
const config = ref(null)
const isEditing = ref(false)
const configLoading = ref(false)
const configError = ref(null)
const configSaving = ref(false)
const configReloading = ref(false)
const configPersisting = ref(false)

// Frontend display settings
const currentTheme = ref('light')
const currentLocale = ref('en')
const sidebarCollapsed = ref(false)
const chatbarCollapsed = ref(false)

// Backup for cancel operation
let configBackup = null

const createEmptyConfig = () => ({
  system: {},
  database: {},
  sync: {},
  status_check: {},
  mcp: {},
  ai: {},
})

const normalizeConfig = (data) => {
  const raw = data?.data || data?.config || data?.settings || data || {}
  const base = createEmptyConfig()
  return {
    ...base,
    ...raw,
    system: { ...base.system, ...(raw.system || {}) },
    database: { ...base.database, ...(raw.database || {}) },
    sync: { ...base.sync, ...(raw.sync || {}) },
    status_check: { ...base.status_check, ...(raw.status_check || {}) },
    mcp: { ...base.mcp, ...(raw.mcp || {}) },
    ai: { ...base.ai, ...(raw.ai || {}) },
  }
}

// Configuration API calls
const loadConfiguration = async () => {
  configLoading.value = true
  configError.value = null
  try {
    const response = await authFetch('/api/v1/config/', {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json().catch(() => ({}))
    if (data?.success === false) {
      throw new Error(data.error || 'Failed to load configuration')
    }
    config.value = normalizeConfig(data)
    configBackup = JSON.parse(JSON.stringify(config.value))
  } catch (err) {
    configError.value = err.message || 'Failed to load configuration'
    console.error('Error loading configuration:', err)
  } finally {
    configLoading.value = false
  }
}

const saveConfiguration = async () => {
  configSaving.value = true
  try {
    const response = await authFetch('/api/v1/config/', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify(config.value),
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json().catch(() => ({}))
    if (data?.success === false) {
      throw new Error(data.error || 'Failed to save configuration')
    }
    ElMessage.success(t('system.saveSuccess') || 'Configuration saved successfully')
    configBackup = JSON.parse(JSON.stringify(config.value))
    isEditing.value = false
  } catch (err) {
    ElMessage.error(err.message || 'Failed to save configuration')
    console.error('Error saving configuration:', err)
  } finally {
    configSaving.value = false
  }
}

const persistConfiguration = async () => {
  configPersisting.value = true
  try {
    const response = await authFetch('/api/v1/config/save', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json().catch(() => ({}))
    if (data?.success === false) {
      throw new Error(data.error || 'Failed to persist configuration')
    }
    ElMessage.success(t('system.persistSuccess') || 'Configuration saved to disk')
  } catch (err) {
    ElMessage.error(err.message || 'Failed to persist configuration')
    console.error('Error persisting configuration:', err)
  } finally {
    configPersisting.value = false
  }
}

const reloadConfiguration = async () => {
  configReloading.value = true
  try {
    const response = await authFetch('/api/v1/config/reload', {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
      },
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const data = await response.json().catch(() => ({}))
    if (data?.success === false) {
      throw new Error(data.error || 'Failed to reload configuration')
    }
    ElMessage.success(t('system.reloadSuccess') || 'Configuration reloaded successfully')
    await loadConfiguration()
  } catch (err) {
    ElMessage.error(err.message || 'Failed to reload configuration')
    console.error('Error reloading configuration:', err)
  } finally {
    configReloading.value = false
  }
}

const cancelEdit = () => {
  if (configBackup) {
    config.value = JSON.parse(JSON.stringify(configBackup))
  }
  isEditing.value = false
}


// Frontend display settings management
const loadDisplaySettings = () => {
  const theme = localStorage.getItem('nagare_theme') || 'light'
  const loc = localStorage.getItem('nagare_locale') || 'en'
  const sidebarState = localStorage.getItem('nagare_sidebar_collapsed') === 'true'
  const chatbarState = localStorage.getItem('nagare_chatbar_collapsed') === 'true'

  currentTheme.value = theme
  currentLocale.value = loc
  sidebarCollapsed.value = sidebarState
  chatbarCollapsed.value = chatbarState
}

const applyTheme = (dark) => {
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
}

const toggleTheme = () => {
  const newTheme = currentTheme.value === 'light' ? 'dark' : 'light'
  currentTheme.value = newTheme
  localStorage.setItem('nagare_theme', newTheme)
  applyTheme(newTheme === 'dark')
  ElMessage.success(`Theme switched to ${newTheme}`)
}

const toggleLocale = () => {
  const newLocale = currentLocale.value === 'en' ? 'zh-CN' : 'en'
  currentLocale.value = newLocale
  localStorage.setItem('nagare_locale', newLocale)
  locale.value = newLocale
  ElMessage.success(`Language switched to ${newLocale === 'en' ? 'English' : '中文'}`)
}

onMounted(() => {
  loadConfiguration()
  loadDisplaySettings()
})
</script>

<style scoped>
.system-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
}

.system-tabs {
  background: white;
  border-radius: 4px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

.system-tabs :deep(.el-tabs__content) {
  padding: 0;
}

h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

</style>