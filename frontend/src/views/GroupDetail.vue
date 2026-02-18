<template>
  <div class="detail-page">
    <div class="detail-header">
      <div>
        <h2>{{ group.name || $t('groups.detailTitle') }}</h2>
        <p class="subtitle">{{ group.description || '-' }}</p>
      </div>
      <el-button @click="$router.back()">{{ $t('common.back') }}</el-button>
    </div>

    <el-row :gutter="16" class="stats-row">
      <el-col :xs="12" :md="6">
        <el-card>
          <div class="stat-label">{{ $t('groups.totalHosts') }}</div>
          <div class="stat-value">{{ summary.total_hosts }}</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card>
          <div class="stat-label">{{ $t('groups.activeHosts') }}</div>
          <div class="stat-value">{{ summary.active_hosts }}</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card>
          <div class="stat-label">{{ $t('groups.errorHosts') }}</div>
          <div class="stat-value">{{ summary.error_hosts }}</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card>
          <div class="stat-label">{{ $t('groups.totalItems') }}</div>
          <div class="stat-value">{{ summary.total_items }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>{{ $t('groups.hostStatusChart') }}</template>
          <div ref="statusChartRef" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="12">
        <el-card>
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
              <span>{{ $t('groups.hosts') }}</span>
              <el-button link size="small" @click="showColumnDialog = true">{{ $t('common.columns') }}</el-button>
            </div>
          </template>
          <el-table :data="hosts" height="320" border>
            <el-table-column v-if="visibleColumns.includes('name')" prop="name" :label="$t('hosts.name')" min-width="160" sortable />
            <el-table-column v-if="visibleColumns.includes('ip_addr')" prop="ip_addr" :label="$t('hosts.ip')" min-width="140" sortable />
            <el-table-column v-if="visibleColumns.includes('status')" prop="status" :label="$t('hosts.status')" min-width="120" sortable>
              <template #default="{ row }">
                <el-tooltip :content="row.status_reason || statusLabel(row.status)" placement="top">
                  <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
                </el-tooltip>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>

  <!-- Columns Dialog -->
  <el-dialog v-model="showColumnDialog" :title="$t('common.columns')" width="400px">
    <el-transfer v-model="visibleColumns" :data="availableColumns" />
    <template #footer>
      <el-button @click="showColumnDialog = false">{{ $t('common.close') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { onMounted, ref, watch, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import { fetchGroupDetail } from '@/api/groups'

const route = useRoute()
const { t } = useI18n()
const group = ref({})
const summary = ref({ total_hosts: 0, active_hosts: 0, error_hosts: 0, syncing_hosts: 0, total_items: 0 })
const hosts = ref([])
const statusChartRef = ref(null)
const showColumnDialog = ref(false)
let statusChart

// Column configuration for hosts table
const availableColumns = [
  { key: 'name', label: t('hosts.name') },
  { key: 'ip_addr', label: t('hosts.ip') },
  { key: 'status', label: t('hosts.status') },
]

const visibleColumns = ref(['name', 'ip_addr', 'status'])

const loadVisibleColumns = () => {
  const saved = localStorage.getItem('groupDetailColumns')
  if (saved) {
    try {
      visibleColumns.value = JSON.parse(saved)
    } catch (e) {
      visibleColumns.value = ['name', 'ip_addr', 'status']
    }
  }
}

const saveVisibleColumns = () => {
  localStorage.setItem('groupDetailColumns', JSON.stringify(visibleColumns.value))
}

watch(visibleColumns, saveVisibleColumns, { deep: true })

const statusLabel = (status) => {
  switch (status) {
    case 1:
      return t('common.statusActive')
    case 2:
      return t('common.statusError')
    case 3:
      return t('common.statusSyncing')
    default:
      return t('common.statusInactive')
  }
}

const statusTag = (status) => {
  switch (status) {
    case 1:
      return 'success'
    case 2:
      return 'danger'
    case 3:
      return 'warning'
    default:
      return 'info'
  }
}

const buildStatusChart = () => {
  if (!statusChartRef.value) return
  if (!statusChart) {
    statusChart = echarts.init(statusChartRef.value)
  }
  const data = [
    { name: t('common.statusActive'), value: summary.value.active_hosts || 0 },
    { name: t('common.statusError'), value: summary.value.error_hosts || 0 },
    { name: t('common.statusSyncing'), value: summary.value.syncing_hosts || 0 },
  ]
  statusChart.setOption({
    tooltip: { trigger: 'item' },
    series: [
      {
        type: 'pie',
        radius: ['35%', '70%'],
        data,
        label: { formatter: '{b}: {c}' }
      }
    ]
  })
}

const loadData = async () => {
  const groupId = Number(route.params.id)
  if (!groupId) return
  const resp = await fetchGroupDetail(groupId)
  const data = resp.data || resp
  group.value = data.group || data.Group || {}
  summary.value = data.summary || data.Summary || summary.value
  hosts.value = (data.hosts || data.Hosts || []).map((h) => ({
    id: h.id || h.ID,
    name: h.name || h.Name || '',
    ip_addr: h.ip_addr || h.IPAddr || '',
    status: h.status ?? h.Status ?? 0,
    status_reason: h.Reason || h.reason || h.Error || h.error || h.ErrorMessage || h.error_message || h.LastError || h.last_error || '',
  }))
  buildStatusChart()
}

watch(() => summary.value, () => buildStatusChart(), { deep: true })

const onResize = () => {
  if (statusChart) statusChart.resize()
}

onMounted(() => {
  loadVisibleColumns()
  loadData()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (statusChart) statusChart.dispose()
})
</script>

<style scoped>
.detail-page {
  padding: 20px;
}
.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.subtitle {
  color: #6b7280;
  margin: 4px 0 0;
}
.stats-row {
  margin-bottom: 16px;
}
.stat-label {
  color: #6b7280;
  font-size: 12px;
}
.stat-value {
  font-size: 22px;
  font-weight: 600;
  margin-top: 6px;
}
.chart {
  width: 100%;
  height: 320px;
}
</style>
