<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('reports.title') }}</h1>
      <p class="page-subtitle">{{ $t('reports.subtitle') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-select v-model="filterType" :placeholder="$t('reports.filterType')" clearable style="width: 150px">
          <el-option label="Weekly" value="weekly" />
          <el-option label="Monthly" value="monthly" />
        </el-select>
      </div>
      <div class="action-group">
        <el-dropdown v-if="selectedRows.length > 0" class="batch-actions">
          <el-button type="warning">
            {{ $t('common.selectedCount', { count: selectedRows.length }) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="Delete" @click="handleBulkDelete" style="color: var(--el-color-danger)">
                {{ $t('common.bulkDelete') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <el-button @click="configDialogVisible = true" :icon="Setting">{{ $t('reports.config') }}</el-button>
        <el-dropdown split-button type="primary" @click="generateWeekly" @command="handleGenerateCommand">
          {{ $t('reports.generateWeekly') }}
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="monthly">{{ $t('reports.generateMonthly') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <el-table :data="reports" border style="width: 100%" v-loading="loading" @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" :label="$t('reports.reportTitle')" min-width="200" />
      <el-table-column prop="report_type" :label="$t('reports.type')" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.report_type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="generated_at" :label="$t('reports.generatedAt')" width="180">
        <template #default="{ row }">
          {{ row.generated_at ? new Date(row.generated_at).toLocaleString() : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('reports.status')" width="120">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('common.actions')" width="150" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" type="primary" :icon="Download" @click="download(row)" :disabled="row.status !== 'completed'" />
            <el-button size="small" type="danger" :icon="Delete" @click="remove(row)" />
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- Config Dialog -->
    <el-dialog v-model="configDialogVisible" :title="$t('reports.configTitle')" width="500px">
      <el-form :model="configForm" label-width="160px">
        <el-divider content-position="left">Weekly Report</el-divider>
        <el-form-item :label="$t('reports.autoGenerate')">
          <el-switch v-model="configForm.auto_generate_weekly" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item :label="$t('reports.generateDay')">
          <el-select v-model="configForm.weekly_generate_day">
            <el-option v-for="day in days" :key="day" :label="day" :value="day" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('reports.generateTime')">
          <el-time-select v-model="configForm.weekly_generate_time" start="00:00" step="01:00" end="23:00" />
        </el-form-item>

        <el-divider content-position="left">Monthly Report</el-divider>
        <el-form-item :label="$t('reports.autoGenerate')">
          <el-switch v-model="configForm.auto_generate_monthly" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item :label="$t('reports.generateDate')">
          <el-input-number v-model="configForm.monthly_generate_date" :min="1" :max="28" />
        </el-form-item>
        <el-form-item :label="$t('reports.generateTime')">
          <el-time-select v-model="configForm.monthly_generate_time" start="00:00" step="01:00" end="23:00" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveConfig">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive, watch } from 'vue'
import { Download, Delete, Setting, ArrowDown } from '@element-plus/icons-vue'
import { 
  fetchReports, 
  generateWeeklyReport, 
  generateMonthlyReport, 
  deleteReport, 
  bulkDeleteReports,
  getReportConfig, 
  updateReportConfig 
} from '@/api/reports'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'
import { getToken } from '@/utils/auth'

const { t } = useI18n()
const reports = ref([])
const loading = ref(false)
const filterType = ref('')
const configDialogVisible = ref(false)
const days = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday']
const selectedRows = ref([])

const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

const handleBulkDelete = () => {
  if (selectedRows.value.length === 0) return
  
  ElMessageBox.confirm(
    t('common.bulkDeleteConfirmText', { count: selectedRows.value.length }),
    t('common.bulkDeleteConfirmTitle'),
    { type: 'warning' }
  ).then(async () => {
    try {
      const ids = selectedRows.value.map(row => row.id)
      await bulkDeleteReports(ids)
      ElMessage.success(t('common.bulkDeleteSuccess', { count: selectedRows.value.length }))
      loadReports()
    } catch (err) {
      ElMessage.error(t('common.bulkDeleteFailed'))
    }
  }).catch(() => {})
}

const configForm = reactive({
  auto_generate_weekly: 0,
  weekly_generate_day: 'Monday',
  weekly_generate_time: '09:00',
  auto_generate_monthly: 0,
  monthly_generate_date: 1,
  monthly_generate_time: '09:00',
})

const loadReports = async () => {
  loading.value = true
  try {
    const res = await fetchReports({ type: filterType.value })
    if (res && res.success) {
      reports.value = res.data || []
    }
  } catch (e) {
    ElMessage.error(t('reports.loadFailed') || 'Failed to load reports')
  } finally {
    loading.value = false
  }
}

const loadConfig = async () => {
  try {
    const res = await getReportConfig()
    if (res && res.success) {
      Object.assign(configForm, res.data)
    }
  } catch (e) {
    console.error(e)
  }
}

const saveConfig = async () => {
  try {
    const res = await updateReportConfig(configForm)
    if (res && res.success) {
      ElMessage.success(t('common.saveSuccess') || 'Configuration saved')
      configDialogVisible.value = false
    }
  } catch (e) {
    ElMessage.error(t('common.saveFailed') || 'Failed to save configuration')
  }
}

const generateWeekly = async () => {
  try {
    await generateWeeklyReport()
    ElMessage.success(t('reports.generationStarted') || 'Weekly report generation started')
    loadReports()
  } catch (e) {
    ElMessage.error(t('reports.generationFailed') || 'Failed to start generation')
  }
}

const handleGenerateCommand = async (cmd) => {
  if (cmd === 'monthly') {
    try {
      await generateMonthlyReport()
      ElMessage.success(t('reports.generationStarted') || 'Monthly report generation started')
      loadReports()
    } catch (e) {
      ElMessage.error(t('reports.generationFailed') || 'Failed to start generation')
    }
  }
}

const download = (row) => {
  const token = getToken()
  window.open(`/api/v1/reports/${row.id}/download?token=${token}`, '_blank')
}

const remove = async (row) => {
  ElMessageBox.confirm(t('reports.deleteConfirm') || 'Delete this report?', t('common.warning') || 'Warning', { type: 'warning' })
    .then(async () => {
      await deleteReport(row.id)
      ElMessage.success(t('common.deleteSuccess') || 'Report deleted')
      loadReports()
    })
}

const getStatusType = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'danger'
  return 'info'
}

watch(filterType, () => {
  loadReports()
})

onMounted(() => {
  loadReports()
  loadConfig()
})
</script>
