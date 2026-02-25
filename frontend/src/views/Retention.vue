<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="header-main">
        <h1 class="page-title">{{ $t('retention.title') }}</h1>
        <div class="header-info">
          <p class="page-subtitle">{{ $t('retention.subtitle') }}</p>
          <div class="refresh-info" v-if="lastUpdated">
            <span class="last-updated">{{ $t('dashboard.summaryLastUpdated') }}: {{ lastUpdated }}</span>
            <el-tag v-if="autoRefreshEnabled" size="small" type="success" effect="plain" class="auto-refresh-tag">
              <el-icon class="is-loading"><Refresh /></el-icon>
              {{ $t('retention.autoRefreshing') }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-tag type="info" effect="plain">
          <el-icon style="vertical-align: middle; margin-right: 4px;"><Timer /></el-icon>
          {{ $t('retention.nextCleanup') }}
        </el-tag>
      </div>
      <div class="action-group">
        <el-switch
          v-model="autoRefreshEnabled"
          style="margin-right: 16px"
          :active-text="$t('common.autoRefresh') || 'Auto-refresh'"
          @change="handleAutoRefreshChange"
        />
        <el-button type="danger" plain @click="handleManualCleanup" :loading="cleaning" :icon="Delete">
          {{ $t('retention.manualCleanup') }}
        </el-button>
        <el-button type="primary" @click="loadPolicies" :icon="Refresh" :loading="loading">
          {{ $t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card class="retention-card animate-slide-up" v-loading="loading" shadow="hover">
      <el-table :data="policies" stripe style="width: 100%" header-cell-class-name="table-header">
        <el-table-column prop="data_type" :label="$t('retention.dataType')" min-width="200">
          <template #default="{ row }">
            <div class="data-type-cell">
              <span v-if="row.data_type" class="data-type-label">{{ $t(`retention.types.${row.data_type}`) || row.data_type }}</span>
              <code v-if="row.data_type" class="data-type-code">{{ row.data_type }}</code>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="retention_days" :label="$t('retention.retentionDays')" width="180" align="center">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-input-number v-model="editForm.retention_days" :min="0" size="small" />
            </template>
            <template v-else>
              <el-tag :type="row.retention_days === 0 ? 'info' : 'primary'" effect="light">
                {{ row.retention_days === 0 ? $t('retention.forever') : `${row.retention_days} ${$t('common.day')}` }}
              </el-tag>
            </template>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" :label="$t('retention.enabled')" width="120" align="center">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-switch v-model="editForm.enabled" :active-value="1" :inactive-value="0" />
            </template>
            <template v-else>
              <el-tag :type="row.enabled === 1 ? 'success' : 'danger'" size="small" effect="dark">
                {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
            </template>
          </template>
        </el-table-column>

        <el-table-column prop="description" :label="$t('retention.description')" min-width="250">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-input v-model="editForm.description" size="small" :placeholder="$t('retention.enterDescription')" />
            </template>
            <template v-else>
              <span class="description-text">
                {{ $t(`retention.typeDescriptions.${row.data_type}`) !== `retention.typeDescriptions.${row.data_type}` ? $t(`retention.typeDescriptions.${row.data_type}`) : (row.description || '-') }}
              </span>
            </template>
          </template>
        </el-table-column>

        <el-table-column :label="$t('retention.actions')" width="180" align="right" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <template v-if="editingId === (row.id || row.data_type)">
                <el-button link type="primary" @click="savePolicy" :loading="saving" :icon="Check">
                  {{ $t('retention.save') }}
                </el-button>
                <el-button link @click="cancelEdit" :icon="Close">
                  {{ $t('retention.cancel') }}
                </el-button>
              </template>
              <template v-else>
                <el-button link type="primary" @click="startEdit(row)" :icon="Edit">
                  {{ $t('retention.edit') }}
                </el-button>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="info-section animate-slide-up" style="animation-delay: 0.2s">
      <el-alert
        :title="$t('retention.automatedCleanupTitle') || 'Automated Cleanup'"
        type="info"
        :description="$t('retention.automatedCleanupDesc') || 'The system automatically performs data retention cleanup every day at 2:00 AM based on these policies.'"
        show-icon
        :closable="false"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Delete, Refresh, Timer, Edit, Check, Close
} from '@element-plus/icons-vue'
import api from '@/api'

const { t } = useI18n()
const policies = ref([])
const loading = ref(false)
const cleaning = ref(false)
const saving = ref(false)
const editingId = ref(null)
const lastUpdated = ref('')
const autoRefreshEnabled = ref(true)
let refreshInterval = null

const editForm = reactive({
  id: null,
  data_type: '',
  retention_days: 30,
  enabled: 1,
  description: ''
})

const loadPolicies = async () => {
  loading.value = true
  try {
    const res = await api.retention.fetchRetentionPolicies()
    if (res && res.success) {
      policies.value = res.data || []
      
      // Ensure all supported types are present (even if not in DB yet)
      const supportedTypes = [
        { type: 'logs', desc: t('retention.typeDescriptions.logs') },
        { type: 'alerts', desc: t('retention.typeDescriptions.alerts') },
        { type: 'audit_logs', desc: t('retention.typeDescriptions.audit_logs') },
        { type: 'item_history', desc: t('retention.typeDescriptions.item_history') },
        { type: 'host_history', desc: t('retention.typeDescriptions.host_history') },
        { type: 'network_history', desc: t('retention.typeDescriptions.network_history') },
        { type: 'chat', desc: t('retention.typeDescriptions.chat') },
        { type: 'ansible_jobs', desc: t('retention.typeDescriptions.ansible_jobs') },
        { type: 'reports', desc: t('retention.typeDescriptions.reports') },
        { type: 'site_messages', desc: t('retention.typeDescriptions.site_messages') }
      ]
      
      const existingTypes = policies.value.map(p => p.data_type)
      supportedTypes.forEach(st => {
        if (!existingTypes.includes(st.type)) {
          policies.value.push({
            id: 0, 
            data_type: st.type,
            retention_days: 30,
            enabled: 0,
            description: st.desc
          })
        }
      })
      
      // Sort: logs, alerts, audit_logs first, then others
      const priority = { 'logs': 1, 'alerts': 2, 'audit_logs': 3 }
      policies.value.sort((a, b) => {
        const pa = priority[a.data_type] || 99
        const pb = priority[b.data_type] || 99
        if (pa !== pb) return pa - pb
        return a.data_type.localeCompare(b.data_type)
      })
      lastUpdated.value = new Date().toLocaleString()
    }
  } catch (err) {
    console.error('Failed to load retention policies:', err)
  } finally {
    loading.value = false
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshInterval = setInterval(() => {
    if (!loading.value && !editingId.value) {
      loadPolicies()
    }
  }, 30000)
}

const stopAutoRefresh = () => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

const handleAutoRefreshChange = (val) => {
  if (val) startAutoRefresh()
  else stopAutoRefresh()
}

const startEdit = (row) => {
  editingId.value = row.id || row.data_type 
  editForm.id = row.id
  editForm.data_type = row.data_type
  editForm.retention_days = row.retention_days
  editForm.enabled = row.enabled
  editForm.description = row.description
}

const cancelEdit = () => {
  editingId.value = null
}

const savePolicy = async () => {
  saving.value = true
  try {
    const res = await api.retention.updateRetentionPolicy(editForm)
    if (res && res.success) {
      ElMessage.success(res.message || 'Policy updated')
      editingId.value = null
      loadPolicies()
    }
  } catch (err) {
    ElMessage.error('Failed to update policy')
  } finally {
    saving.value = false
  }
}

const handleManualCleanup = async () => {
  try {
    await ElMessageBox.confirm(
      t('retention.cleanupConfirmText'),
      t('retention.cleanupConfirm'),
      {
        confirmButtonText: t('retention.performCleanup'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    
    cleaning.value = true
    const res = await api.retention.performRetentionCleanup()
    if (res && res.success) {
      const counts = res.data || {}
      let message = t('retention.cleanupSuccessMsg') + ' '
      const details = Object.entries(counts).map(([type, count]) => `${t('retention.types.' + type) || type}: ${count} rows`).join(', ')
      if (details) message += t('retention.recordsRemoved') + ' ' + details
      else message += t('retention.noRecordsCleaned')
      
      ElMessage.success(message)
    }
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error(t('retention.cleanupFailedMsg'))
    }
  } finally {
    cleaning.value = false
  }
}

onMounted(() => {
  loadPolicies()
  if (autoRefreshEnabled.value) startAutoRefresh()
})

onBeforeUnmount(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-main {
  display: flex;
  flex-direction: column;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 4px;
}

.refresh-info {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.auto-refresh-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.retention-card {
  margin-bottom: 24px;
}

.data-type-cell {
  display: flex;
  flex-direction: column;
}

.data-type-label {
  font-weight: 600;
  color: var(--text-strong);
}

.data-type-code {
  font-size: 11px;
  color: var(--text-muted);
  font-family: var(--font-mono, monospace);
  margin-top: 2px;
}

.description-text {
  font-size: 13px;
  color: var(--text-muted);
}

.info-section {
  max-width: 800px;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
