<template>
  <div class="nagare-container animate-fade-in">
    <div class="detail-header">
      <div v-if="monitor.name">
        <h1 class="page-title">{{ monitor.name }}</h1>
        <p class="subtitle">{{ monitor.url || '-' }}</p>
      </div>
      <div v-else class="page-header">
        <h1 class="page-title">Monitor Detail</h1>
      </div>
      <el-button @click="$router.back()">{{ $t('common.back') }}</el-button>
    </div>

    <div v-if="loading" style="text-align: center; padding: 100px;">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p style="margin-top: 16px; color: #909399;">{{ $t('common.loading') }}</p>
    </div>

    <div v-else-if="error" style="padding: 40px;">
      <el-alert :title="error" type="error" show-icon :closable="false">
        <template #default>
          <el-button size="small" @click="loadData">{{ $t('common.refresh') }}</el-button>
        </template>
      </el-alert>
    </div>

    <div v-else>
      <el-row :gutter="16" class="stats-row">
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">Health</div>
            <div class="stat-value">
              <el-progress type="circle" :percentage="monitor.health_score || 0" :width="40" :stroke-width="4" :status="getHealthStatus(monitor.health_score)" />
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('groups.title') }}</div>
            <div class="stat-value">{{ summary.group_total }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('hosts.title') }}</div>
            <div class="stat-value">{{ summary.host_total }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('common.statusActive') }}</div>
            <div class="stat-value" style="color: var(--el-color-success)">{{ summary.host_active }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>History</span>
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-date-picker
                    v-model="historyRange"
                    type="datetimerange"
                    :shortcuts="historyShortcuts"
                    :start-placeholder="$t('common.startTime')"
                    :end-placeholder="$t('common.endTime')"
                    size="small"
                  />
                  <el-button size="small" @click="loadHistory" :loading="historyLoading">{{ $t('common.refresh') }}</el-button>
                </div>
              </div>
            </template>
            <div ref="historyChartRef" class="chart"></div>
          </el-card>
        </el-col>

        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <span>{{ $t('groups.title') }}</span>
            </template>
            <el-table :data="groups" height="320" border>
              <el-table-column prop="name" :label="$t('groups.name')" min-width="160" />
              <el-table-column prop="status" :label="$t('groups.status')" min-width="120">
                <template #default="{ row }">
                  <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="health_score" label="Health" min-width="120">
                <template #default="{ row }">
                  <el-progress :percentage="row.health_score || 0" :status="getHealthStatus(row.health_score)" />
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch, nextTick, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import * as echarts from 'echarts'
import { getMonitorById, fetchMonitorHistory } from '@/api/monitors'
import { fetchGroupData } from '@/api/groups'

const route = useRoute()
const monitor = ref({})
const groups = ref([])
const summary = ref({
  group_total: 0,
  host_total: 0,
  host_active: 0,
})
const loading = ref(false)
const error = ref(null)
const historyLoading = ref(false)
const historyRange = ref([])
const historyChartRef = ref(null)
let historyChart

const historyShortcuts = [
  {
    text: '1h',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 60 * 60 * 1000)
      return [start, end]
    },
  },
  {
    text: '6h',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 6 * 60 * 60 * 1000)
      return [start, end]
    },
  },
  {
    text: '24h',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      return [start, end]
    },
  },
  {
    text: '7d',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000)
      return [start, end]
    },
  },
]

const setDefaultHistoryRange = () => {
  if (Array.isArray(historyRange.value) && historyRange.value.length === 2) return
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  historyRange.value = [start, end]
}

const statusLabel = (status) => {
  switch (status) {
    case 1:
      return 'Active'
    case 2:
      return 'Error'
    case 3:
      return 'Syncing'
    default:
      return 'Inactive'
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

const getHealthStatus = (score) => {
  if ((score || 0) >= 90) return 'success'
  if ((score || 0) >= 70) return 'warning'
  return 'exception'
}

const buildHistoryChart = (rows) => {
  if (!historyChartRef.value) return
  if (!historyChart) {
    historyChart = echarts.init(historyChartRef.value)
  }

  const points = Array.isArray(rows) ? rows : []
  const groupActive = points.map((h) => [new Date(h.sampled_at || h.SampledAt).getTime(), h.group_active ?? h.GroupActive ?? 0])
  const hostActive = points.map((h) => [new Date(h.sampled_at || h.SampledAt).getTime(), h.host_active ?? h.HostActive ?? 0])

  historyChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['Group Active', 'Host Active'] },
    xAxis: { type: 'time' },
    yAxis: { type: 'value', minInterval: 1 },
    series: [
      { name: 'Group Active', type: 'line', smooth: true, data: groupActive },
      { name: 'Host Active', type: 'line', smooth: true, data: hostActive },
    ],
  }, { notMerge: true })
}

const loadHistory = async () => {
  const monitorID = Number(route.params.id)
  if (!monitorID) return

  historyLoading.value = true
  try {
    const [fromDate, toDate] = Array.isArray(historyRange.value) ? historyRange.value : []
    const from = fromDate ? Math.floor(new Date(fromDate).getTime() / 1000) : undefined
    const to = toDate ? Math.floor(new Date(toDate).getTime() / 1000) : undefined

    const resp = await fetchMonitorHistory(monitorID, { from, to, limit: 500 })
    const rows = Array.isArray(resp) ? resp : (resp.data || [])
    buildHistoryChart(rows)

    if (rows.length > 0) {
      const last = rows[rows.length - 1]
      summary.value = {
        group_total: last.group_total ?? last.GroupTotal ?? summary.value.group_total,
        host_total: last.host_total ?? last.HostTotal ?? summary.value.host_total,
        host_active: last.host_active ?? last.HostActive ?? summary.value.host_active,
      }
    }
  } finally {
    historyLoading.value = false
  }
}

const loadData = async () => {
  const monitorID = Number(route.params.id)
  if (!monitorID) return

  loading.value = true
  error.value = null
  try {
    const [monitorResp, groupsResp] = await Promise.all([
      getMonitorById(monitorID),
      fetchGroupData({ monitor_id: monitorID, limit: 500, offset: 0 }),
    ])

    const monitorData = monitorResp?.data || monitorResp || {}
    monitor.value = {
      id: monitorData.id || monitorData.ID,
      name: monitorData.name || monitorData.Name || '',
      url: monitorData.url || monitorData.URL || '',
      status: monitorData.status ?? monitorData.Status ?? 0,
      health_score: monitorData.health_score ?? monitorData.HealthScore ?? 0,
    }

    const groupRows = Array.isArray(groupsResp)
      ? groupsResp
      : (groupsResp?.data?.items || groupsResp?.items || groupsResp?.data || [])

    groups.value = groupRows.map((g) => ({
      id: g.id || g.ID,
      name: g.name || g.Name,
      status: g.status ?? g.Status ?? 0,
      health_score: g.health_score ?? g.HealthScore ?? 0,
    }))

    summary.value.group_total = groups.value.length
  } catch (err) {
    error.value = err.message || 'Failed to load monitor detail'
  } finally {
    loading.value = false
  }

  await nextTick()
  await loadHistory()
}

watch(historyRange, () => {
  loadHistory()
})

const onResize = () => {
  if (historyChart) historyChart.resize()
}

onMounted(() => {
  setDefaultHistoryRange()
  loadData()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (historyChart) {
    historyChart.dispose()
  }
})
</script>

<style scoped>
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
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
</style>
