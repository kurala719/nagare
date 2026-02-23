<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="title-section">
        <h1 class="page-title">{{ $t('retention.title') }}</h1>
        <p class="page-subtitle">{{ $t('retention.subtitle') }}</p>
      </div>
      <div class="action-group">
        <el-button type="danger" plain @click="handleManualCleanup" :loading="cleaning" :icon="Delete">
          {{ $t('retention.manualCleanup') }}
        </el-button>
        <el-button @click="loadPolicies" :icon="Refresh">
          {{ $t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card class="retention-card" v-loading="loading">
      <el-table :data="policies" stripe style="width: 100%">
        <el-table-column prop="data_type" :label="$t('retention.dataType')" width="200">
          <template #default="{ row }">
            <span v-if="row.data_type" class="data-type-label">{{ $t(`retention.types.${row.data_type}`) || row.data_type }}</span>
            <code v-if="row.data_type" class="data-type-code">{{ row.data_type }}</code>
          </template>
        </el-table-column>
        
        <el-table-column prop="retention_days" :label="$t('retention.retentionDays')" width="180" align="center">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-input-number v-model="editForm.retention_days" :min="0" size="small" />
            </template>
            <template v-else>
              <el-tag :type="row.retention_days === 0 ? 'info' : 'primary'">
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
              <el-icon :color="row.enabled === 1 ? '#67C23A' : '#F56C6C'">
                <CircleCheckFilled v-if="row.enabled === 1" />
                <CircleCloseFilled v-else />
              </el-icon>
            </template>
          </template>
        </el-table-column>

        <el-table-column prop="description" :label="$t('retention.description')">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-input v-model="editForm.description" size="small" />
            </template>
            <template v-else>
              <span class="description-text">{{ row.description || '-' }}</span>
            </template>
          </template>
        </el-table-column>

        <el-table-column :label="$t('retention.actions')" width="150" align="right">
          <template #default="{ row }">
            <template v-if="editingId === (row.id || row.data_type)">
              <el-button link type="primary" @click="savePolicy" :loading="saving">
                {{ $t('retention.save') }}
              </el-button>
              <el-button link @click="cancelEdit">
                {{ $t('retention.cancel') }}
              </el-button>
            </template>
            <template v-else>
              <el-button link type="primary" @click="startEdit(row)">
                {{ $t('retention.edit') }}
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <div class="info-section">
      <el-alert
        title="Automated Cleanup"
        type="info"
        description="The system automatically performs data retention cleanup every day at 2:00 AM based on these policies."
        show-icon
        :closable="false"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Delete, Refresh, CircleCheckFilled, CircleCloseFilled 
} from '@element-plus/icons-vue'
import api from '@/api'

const policies = ref([])
const loading = ref(false)
const cleaning = ref(false)
const saving = ref(false)
const editingId = ref(null)

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
    // The interceptor returns response.data (the APIResponse object)
    if (res && res.success) {
      policies.value = res.data || []
      
      // Ensure all supported types are present (even if not in DB yet)
      const supportedTypes = [
        { type: 'logs', desc: 'System and service logs' },
        { type: 'alerts', desc: 'Alert history' },
        { type: 'audit_logs', desc: 'User operational logs' },
        { type: 'item_history', desc: 'Metric history data' },
        { type: 'host_history', desc: 'Host status history' },
        { type: 'network_history', desc: 'Network health score history' },
        { type: 'chat', desc: 'AI chat messages' },
        { type: 'ansible_jobs', desc: 'Ansible execution logs' },
        { type: 'reports', desc: 'Generated PDF reports' },
        { type: 'site_messages', desc: 'User notifications' }
      ]
      
      const existingTypes = policies.value.map(p => p.data_type)
      supportedTypes.forEach(st => {
        if (!existingTypes.includes(st.type)) {
          policies.value.push({
            id: 0, // Mark as new
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
    }
  } catch (err) {
    console.error('Failed to load retention policies:', err)
  } finally {
    loading.value = false
  }
}

const startEdit = (row) => {
  editingId.value = row.id || row.data_type // Use data_type if id is 0
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
      'Are you sure you want to perform manual data cleanup now according to the enabled policies?',
      'Manual Cleanup',
      {
        confirmButtonText: 'Perform Cleanup',
        cancelButtonText: 'Cancel',
        type: 'warning'
      }
    )
    
    cleaning.value = true
    const res = await api.retention.performRetentionCleanup()
    if (res && res.success) {
      const counts = res.data || {}
      let message = 'Cleanup performed successfully. '
      const details = Object.entries(counts).map(([type, count]) => `${type}: ${count} rows`).join(', ')
      if (details) message += 'Records removed: ' + details
      else message += 'No records needed cleaning.'
      
      ElMessage.success(message)
    }
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('Cleanup failed')
    }
  } finally {
    cleaning.value = false
  }
}

onMounted(() => {
  loadPolicies()
})
</script>

<style scoped>
.retention-card {
  margin-bottom: 24px;
  border-radius: var(--radius-lg);
}

.data-type-label {
  display: block;
  font-weight: 600;
  color: var(--text-strong);
}

.data-type-code {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  font-family: monospace;
}

.description-text {
  font-size: 13px;
  color: var(--text-muted);
}

.info-section {
  max-width: 600px;
}
</style>
