<template>
  <div class="nagare-container animate-fade-in">
    <div class="detail-header">
      <div v-if="host.name">
        <h1 class="page-title">{{ host.name }}</h1>
        <p class="page-subtitle">{{ host.ip_addr || host.hostid || '-' }}</p>
      </div>
      <div v-else class="page-header">
        <h1 class="page-title">{{ $t('hosts.detailTitle') }}</h1>
      </div>
      <div class="detail-actions">
        <el-button type="info" @click="$router.push(`/host/${host.id}/terminal`)">
          Terminal
        </el-button>
        <el-button type="primary" :loading="reportGenerating" @click="generateReport">
          {{ $t('hosts.reportGenerate') }}
        </el-button>
        <el-button @click="$router.back()">{{ $t('common.back') }}</el-button>
      </div>
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
              <el-progress type="circle" :percentage="host.health_score || 0" :width="40" :stroke-width="4" :status="getHealthStatus(host.health_score)" />
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">Monitor</div>
            <div class="stat-value" style="font-size: 14px;">{{ host.monitor_name || '-' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">Group</div>
            <div class="stat-value" style="font-size: 14px;">{{ host.group_name || '-' }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('items.total') }}</div>
            <div class="stat-value">{{ stats.totalItems }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('common.statusActive') }}</div>
            <div class="stat-value" style="color: var(--el-color-success)">{{ stats.active }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="4">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('common.statusError') }}</div>
            <div class="stat-value" style="color: var(--el-color-danger)">{{ stats.error }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16" class="metrics-row">
        <el-col :xs="24" :lg="8">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>{{ $t('hosts.cpuUsage') || 'CPU' }}</span>
                <el-tag size="small" type="primary">{{ currentCpu }}%</el-tag>
              </div>
            </template>
            <div ref="cpuChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="8">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>{{ $t('hosts.memoryUsage') || 'Memory' }}</span>
                <el-tag size="small" type="success">{{ currentMem }}%</el-tag>
              </div>
            </template>
            <div ref="memChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="8">
          <el-card shadow="hover">
            <template #header>
              <div class="card-header">
                <span>{{ $t('hosts.networkTraffic') || 'Network' }}</span>
                <el-tag size="small" type="warning">{{ currentNet }} KB/s</el-tag>
              </div>
            </template>
            <div ref="netChartRef" class="mini-chart"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>{{ $t('hosts.itemStatusChart') }}</span>
                <div class="card-actions">
                  <el-switch
                    v-model="compareMode"
                    size="small"
                    :active-text="$t('common.comparePrevious')"
                    style="margin-right: 8px;"
                  />
                  <el-date-picker
                    v-model="historyRange"
                    type="datetimerange"
                    :shortcuts="historyShortcuts"
                    :start-placeholder="$t('common.startTime')"
                    :end-placeholder="$t('common.endTime')"
                    size="small"
                    class="range-picker"
                  />
                  <el-button size="small" @click="loadHistory" :loading="historyLoading">{{ $t('common.refresh') }}</el-button>
                </div>
              </div>
            </template>
            <el-skeleton v-if="historyLoading" animated :rows="8" />
            <el-alert
              v-else-if="historyError"
              :title="historyError"
              type="error"
              show-icon
              :closable="false"
              class="chart-alert"
            />
            <el-empty v-else-if="historyEmpty" :description="$t('common.noHistoryData')" />
            <div v-else>
              <div ref="statusChartRef" class="chart"></div>
              <div class="status-legend">
                <span class="legend-title">{{ $t('hosts.statusLegendTitle') }}:</span>
                <span class="legend-item">
                  <span class="legend-dot" style="background: #909399;"></span>
                  0 - {{ $t('common.statusInactive') }}
                </span>
                <span class="legend-item">
                  <span class="legend-dot" style="background: #67C23A;"></span>
                  1 - {{ $t('common.statusActive') }}
                </span>
                <span class="legend-item">
                  <span class="legend-dot" style="background: #F56C6C;"></span>
                  2 - {{ $t('common.statusError') }}
                </span>
                <span class="legend-item">
                  <span class="legend-dot" style="background: #E6A23C;"></span>
                  3 - {{ $t('common.statusSyncing') }}
                </span>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
                <span>{{ $t('hosts.itemList') }}</span>
                <div style="display: flex; align-items: center; gap: 10px;">
                  <span class="filter-label">{{ $t('hosts.importantOnly') }}</span>
                  <el-switch v-model="showImportantOnly" :active-value="true" :inactive-value="false" />
                  <el-button link size="small" @click="showColumnDialog = true">{{ $t('common.columns') }}</el-button>
                </div>
              </div>
            </template>
            <el-table :data="displayItems" height="320" border>
              <el-table-column v-if="visibleColumns.includes('name')" prop="name" :label="$t('items.name')" min-width="160" sortable />
              <el-table-column v-if="visibleColumns.includes('value')" prop="value" :label="$t('items.value')" min-width="140" sortable />
              <el-table-column v-if="visibleColumns.includes('units')" prop="units" :label="$t('items.units')" min-width="100" sortable />
              <el-table-column v-if="visibleColumns.includes('status')" prop="status" :label="$t('items.status')" min-width="120" sortable>
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
  </div>

  <!-- Columns Dialog -->
  <el-dialog v-model="showColumnDialog" :title="$t('common.columns')" width="400px">
    <el-transfer v-model="visibleColumns" :data="availableColumns" />
    <template #footer>
      <el-button @click="showColumnDialog = false">{{ $t('common.close') }}</el-button>
    </template>
  </el-dialog>

  <div v-if="reportSnapshot" class="report-canvas" ref="reportCanvasRef" aria-hidden="true">
    <div class="report-header">
      <div>
        <h2>{{ $t('hosts.reportTitle') }}</h2>
        <p class="report-subtitle">{{ reportSnapshot.generatedAt }}</p>
      </div>
      <div class="report-badge">Nagare</div>
    </div>
    <div class="report-section">
      <h3>{{ $t('hosts.reportSummary') }}</h3>
      <div class="report-grid">
        <div><strong>{{ $t('hosts.name') }}:</strong> {{ reportSnapshot.hostName }}</div>
        <div><strong>{{ $t('hosts.ip') }}:</strong> {{ reportSnapshot.hostIp }}</div>
        <div><strong>{{ $t('common.status') }}:</strong> {{ reportSnapshot.hostStatus }}</div>
        <div><strong>{{ $t('items.total') }}:</strong> {{ reportSnapshot.totalItems }}</div>
        <div><strong>{{ $t('common.statusActive') }}:</strong> {{ reportSnapshot.activeItems }}</div>
        <div><strong>{{ $t('common.statusError') }}:</strong> {{ reportSnapshot.errorItems }}</div>
      </div>
    </div>
    <div class="report-section">
      <h3>{{ $t('hosts.reportKeyMetrics') }}</h3>
      <ul>
        <li v-for="item in reportSnapshot.keyItems" :key="item.id">
          {{ item.name }}: {{ item.value || '--' }} ({{ statusLabel(item.status) }})
        </li>
      </ul>
    </div>
    <div class="report-section">
      <h3>{{ $t('hosts.reportChatTitle') }}</h3>
      <div v-if="reportSnapshot.chatHighlights.length === 0" class="report-muted">
        {{ $t('hosts.reportChatEmpty') }}
      </div>
      <div v-else class="report-chat">
        <div v-for="(msg, idx) in reportSnapshot.chatHighlights" :key="`${msg.role}-${idx}`" class="report-chat-line">
          <span class="report-chat-role">{{ msg.role }}:</span>
          <span>{{ msg.content }}</span>
        </div>
      </div>
    </div>
    <div class="report-section report-disclaimer">
      {{ $t('hosts.reportDisclaimer') }}
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, watch, onBeforeUnmount, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { fetchHostHistory, getHostById } from '@/api/hosts'
import { fetchItemsByHost, fetchItemHistory } from '@/api/items'

const route = useRoute()
const { t } = useI18n()
const host = ref({})
const items = ref([])
const statusChartRef = ref(null)
const cpuChartRef = ref(null)
const memChartRef = ref(null)
const netChartRef = ref(null)
const showColumnDialog = ref(false)
const showImportantOnly = ref(true)
const historyRange = ref([])
const historyLoading = ref(false)
const historyError = ref(null)
const historyEmpty = ref(false)
const compareMode = ref(false)
const reportGenerating = ref(false)
const reportSnapshot = ref(null)
const reportCanvasRef = ref(null)
const loading = ref(false)
const error = ref(null)

const currentCpu = ref(0)
const currentMem = ref(0)
const currentNet = ref(0)

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
  {
    text: '30d',
    value: () => {
      const end = new Date()
      const start = new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000)
      return [start, end]
    },
  },
]
let statusChart
let cpuChart
let memChart
let netChart

// Column configuration for items table
const availableColumns = [
  { key: 'name', label: t('items.name') },
  { key: 'value', label: t('items.value') },
  { key: 'units', label: t('items.units') },
  { key: 'status', label: t('items.status') },
]

const visibleColumns = ref(['name', 'value', 'units', 'status'])

const loadVisibleColumns = () => {
  const saved = localStorage.getItem('hostDetailColumns')
  if (saved) {
    try {
      visibleColumns.value = JSON.parse(saved)
    } catch (e) {
      visibleColumns.value = ['name', 'value', 'status']
    }
  }
}

const saveVisibleColumns = () => {
  localStorage.setItem('hostDetailColumns', JSON.stringify(visibleColumns.value))
}

watch(visibleColumns, saveVisibleColumns, { deep: true })

const stats = computed(() => {
  const totals = { totalItems: items.value.length, active: 0, error: 0, syncing: 0, inactive: 0 }
  items.value.forEach((item) => {
    switch (item.status) {
      case 1:
        totals.active += 1
        break
      case 2:
        totals.error += 1
        break
      case 3:
        totals.syncing += 1
        break
      default:
        totals.inactive += 1
    }
  })
  return totals
})

const importantKeywords = ['cpu', 'memory', 'mem', 'ram', 'disk', 'storage', 'network', 'net', 'swap', 'load']

const isImportantItem = (name) => {
  const label = (name || '').toLowerCase()
  return importantKeywords.some((key) => label.includes(key))
}

const displayItems = computed(() => {
  if (!showImportantOnly.value) return items.value
  const filtered = items.value.filter((item) => isImportantItem(item.name))
  return filtered.length > 0 ? filtered : items.value
})

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

const getHealthStatus = (score) => {
  if (score >= 90) return 'success'
  if (score >= 70) return 'warning'
  return 'exception'
}

const resolveRangeWindow = () => {
  const range = historyRange.value
  if (Array.isArray(range) && range.length === 2 && range[0] && range[1]) {
    return [new Date(range[0]), new Date(range[1])]
  }
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  return [start, end]
}

const initMetricCharts = () => {
  if (cpuChartRef.value && !cpuChart) {
    cpuChart = echarts.init(cpuChartRef.value)
  }
  if (memChartRef.value && !memChart) {
    memChart = echarts.init(memChartRef.value)
  }
  if (netChartRef.value && !netChart) {
    netChart = echarts.init(netChartRef.value)
  }
}

const setMetricChartOption = (chart, name, data, color, unit = '%') => {
  if (!chart) return
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        const p = params[0]
        const time = new Date(p.data[0]).toLocaleString()
        return `${time}<br/>${p.seriesName}: ${p.data[1]}${unit}`
      }
    },
    grid: { left: 40, right: 10, top: 10, bottom: 20 },
    xAxis: { type: 'time', splitLine: { show: false } },
    yAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed' } } },
    series: [{
      name,
      type: 'line',
      smooth: true,
      showSymbol: false,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: color + '80' },
          { offset: 1, color: color + '00' }
        ])
      },
      itemStyle: { color },
      data
    }]
  })
}

const loadMetricHistory = async () => {
  initMetricCharts()
  const [start, end] = resolveRangeWindow()
  const from = Math.floor(start.getTime() / 1000)
  const to = Math.floor(end.getTime() / 1000)

  // Categorize items
  const cpuItems = items.value.filter(i => {
    const n = i.name.toLowerCase()
    return (n.includes('cpu') || n.includes('load')) && !n.includes('temp')
  })
  const memItems = items.value.filter(i => {
    const n = i.name.toLowerCase()
    return n.includes('mem') || n.includes('ram') || n.includes('memory')
  })
  const netItems = items.value.filter(i => {
    const n = i.name.toLowerCase()
    return n.includes('net') || n.includes('eth') || n.includes('if') || n.includes('traffic') || n.includes('bps')
  })

  // Helper to fetch and plot
  const fetchAndPlot = async (itemList, chart, name, color, unit, currentRef) => {
    if (itemList.length === 0 || !chart) return
    // Pick the most representative item (e.g., the first one for now)
    const item = itemList[0]
    currentRef.value = item.value || 0
    try {
      const resp = await fetchItemHistory(item.id, { from, to, limit: 100 })
      const rows = Array.isArray(resp.data) ? resp.data : (Array.isArray(resp) ? resp : [])
      const data = rows.map(r => [new Date(r.sampled_at || r.SampledAt).getTime(), parseFloat(r.value || r.Value || 0)])
      setMetricChartOption(chart, name, data, color, unit)
    } catch (e) {
      console.warn(`Failed to fetch history for ${item.name}`, e)
    }
  }

  await Promise.all([
    fetchAndPlot(cpuItems, cpuChart, t('hosts.cpuUsage'), '#409EFF', '%', currentCpu),
    fetchAndPlot(memItems, memChart, t('hosts.memoryUsage'), '#67C23A', '%', currentMem),
    fetchAndPlot(netItems, netChart, t('hosts.networkTraffic'), '#E6A23C', ' KB/s', currentNet)
  ])
}

const buildStatusChart = (series, prevSeries = []) => {
  if (!statusChartRef.value) return
  if (!statusChart) {
    statusChart = echarts.init(statusChartRef.value)
  }
  const chartSeries = [
    {
      name: t('common.currentPeriod'),
      type: 'line',
      step: 'end',
      showSymbol: false,
      data: series,
      lineStyle: { width: 2 },
    }
  ]
  if (prevSeries.length > 0) {
    chartSeries.push({
      name: t('common.previousPeriod'),
      type: 'line',
      step: 'end',
      showSymbol: false,
      data: prevSeries,
      itemStyle: { color: '#909399' },
      lineStyle: { type: 'dashed', width: 2 },
    })
  }
  statusChart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        let tip = ''
        params.forEach((point) => {
          const value = point.data?.[1]
          const time = new Date(point.data?.[0])
          const timeLabel = Number.isNaN(time.getTime()) ? '-' : time.toLocaleString()
          tip += `<div style="margin-bottom:4px;"><strong>${point.seriesName}</strong><br/>${timeLabel}<br/>${t('items.status')}: ${statusLabel(value)}</div>`
        })
        return tip
      },
    },
    legend: {
      show: prevSeries.length > 0,
      data: prevSeries.length > 0 ? [t('common.currentPeriod'), t('common.previousPeriod')] : [],
    },
    visualMap: {
      show: false,
      dimension: 1,
      pieces: [
        { value: 0, color: '#909399' },  // Inactive - gray
        { value: 1, color: '#67C23A' },  // Active - green
        { value: 2, color: '#F56C6C' },  // Error - red
        { value: 3, color: '#E6A23C' },  // Syncing - orange
      ],
      seriesIndex: 0,
    },
    grid: { left: 80, right: 20, top: prevSeries.length > 0 ? 40 : 20, bottom: 40 },
    xAxis: { type: 'time' },
    yAxis: {
      type: 'value',
      min: 0,
      max: 3,
      interval: 1,
      axisLabel: {
        formatter: (value) => statusLabel(value),
      },
    },
    series: chartSeries
  })
}

const loadHistory = async () => {
  const hostId = Number(route.params.id)
  if (!hostId) return
  historyLoading.value = true
  historyError.value = null
  historyEmpty.value = false
  try {
    const [start, end] = resolveRangeWindow()
    const from = Math.floor(start.getTime() / 1000)
    const to = Math.floor(end.getTime() / 1000)
    
    // Fetch current period
    const resp = await fetchHostHistory(hostId, {
      from,
      to,
      limit: 500,
    })
    const payload = resp?.data || resp || []
    let rows = Array.isArray(payload) ? payload : []
    
    if (rows.length === 0) {
      const fallbackResp = await fetchHostHistory(hostId, { limit: 500 })
      const fallbackPayload = fallbackResp?.data || fallbackResp || []
      rows = Array.isArray(fallbackPayload) ? fallbackPayload : []
      if (rows.length === 0) {
        historyEmpty.value = true
        buildStatusChart([], [])
        return
      }
    }
    
    const series = []
    rows.forEach((row) => {
      const sampledAt = row.sampled_at || row.SampledAt
      const status = row.status ?? row.Status
      if (sampledAt === undefined || status === undefined) return
      const time = new Date(sampledAt).getTime()
      if (Number.isNaN(time)) return
      series.push([time, Number(status)])
    })
    
    // Fetch previous period if compare mode is enabled
    let prevSeries = []
    if (compareMode.value) {
      const duration = end.getTime() - start.getTime()
      const prevStart = new Date(start.getTime() - duration)
      const prevEnd = new Date(start.getTime())
      const prevFrom = Math.floor(prevStart.getTime() / 1000)
      const prevTo = Math.floor(prevEnd.getTime() / 1000)
      
      try {
        const prevResp = await fetchHostHistory(hostId, {
          from: prevFrom,
          to: prevTo,
          limit: 500,
        })
        const prevPayload = prevResp?.data || prevResp || []
        const prevRows = Array.isArray(prevPayload) ? prevPayload : []
        
        prevRows.forEach((row) => {
          const sampledAt = row.sampled_at || row.SampledAt
          const status = row.status ?? row.Status
          if (sampledAt === undefined || status === undefined) return
          const time = new Date(sampledAt).getTime()
          if (Number.isNaN(time)) return
          // Shift timestamps to align with current period
          const shiftedTime = time + duration
          prevSeries.push([shiftedTime, Number(status)])
        })
      } catch (err) {
        console.warn('Failed to load previous period:', err)
      }
    }
    
    historyEmpty.value = series.length === 0
    buildStatusChart(series, prevSeries)
  } catch (err) {
    historyError.value = err?.message || t('common.historyLoadFailed')
    buildStatusChart([], [])
  } finally {
    historyLoading.value = false
  }
}

const setDefaultHistoryRange = () => {
  if (Array.isArray(historyRange.value) && historyRange.value.length === 2) return
  const end = new Date()
  const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
  historyRange.value = [start, end]
}

const loadData = async () => {
  const hostId = Number(route.params.id)
  if (!hostId) return
  
  loading.value = true
  error.value = null
  
  try {
    const hostResp = await getHostById(hostId)
    const hostData = hostResp.data || hostResp
    host.value = {
      id: hostData.id || hostData.ID || hostId,
      name: hostData.name || hostData.Name || '',
      ip_addr: hostData.ip_addr || hostData.IPAddr || '',
      hostid: hostData.hostid || hostData.Hostid || '',
      group_name: hostData.group_name || hostData.GroupName || '',
      monitor_name: hostData.monitor_name || hostData.MonitorName || '',
      health_score: hostData.health_score || hostData.HealthScore || 100,
    }
    
    const itemsResp = await fetchItemsByHost(hostId)
    const itemsData = Array.isArray(itemsResp) ? itemsResp : (itemsResp.data || itemsResp.items || [])
    items.value = itemsData.map((i) => ({
      id: i.id || i.ID,
      name: i.name || i.Name || '',
      value: i.value || i.Value || '',
      status: i.status ?? i.Status ?? 0,
      status_reason: i.Reason || i.reason || i.Error || i.error || i.ErrorMessage || i.error_message || i.LastError || i.last_error || i.Comment || i.comment || '',
    }))
    
    await Promise.allSettled([
      loadHistory(),
      loadMetricHistory()
    ])
  } catch (err) {
    console.error('Failed to load host detail data', err)
    error.value = err.message || 'Failed to load host data'
  } finally {
    loading.value = false
  }
}

const onResize = () => {
  if (statusChart) statusChart.resize()
  if (cpuChart) cpuChart.resize()
  if (memChart) memChart.resize()
  if (netChart) netChart.resize()
}

onMounted(() => {
  loadVisibleColumns()
  setDefaultHistoryRange()
  loadData()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (statusChart) statusChart.dispose()
  if (cpuChart) cpuChart.dispose()
  if (memChart) memChart.dispose()
  if (netChart) netChart.dispose()
})

watch(historyRange, () => {
  loadHistory()
  loadMetricHistory()
})

watch(compareMode, () => {
  loadHistory()
  loadMetricHistory()
})

async function generateReport() {
  reportGenerating.value = true
  try {
    const keyItems = items.value.filter((item) => isImportantItem(item.name)).slice(0, 10)
    reportSnapshot.value = {
      generatedAt: new Date().toLocaleString(),
      hostName: host.value.name || '--',
      hostIp: host.value.ip_addr || '--',
      hostStatus: statusLabel(host.value.status ?? 0),
      totalItems: items.value.length,
      activeItems: stats.value.active,
      errorItems: stats.value.error,
      keyItems,
      chatHighlights: [],
    }
    await nextTick();

    const { default: html2canvas } = await import('html2canvas');
    const { default: jsPDF } = await import('jspdf');

    const canvasEl = reportCanvasRef.value;
    if (!canvasEl) throw new Error('Report element not found');

    const canvas = await html2canvas(canvasEl, {
      scale: 2,
      useCORS: true,
      backgroundColor: '#ffffff',
    });
    const imgData = canvas.toDataURL('image/png');
    const pdf = new jsPDF({
      orientation: 'portrait',
      unit: 'mm',
      format: 'a4',
    });

    const pdfWidth = pdf.internal.pageSize.getWidth();
    const pdfHeight = (canvas.height * pdfWidth) / canvas.width;
    pdf.addImage(imgData, 'PNG', 0, 0, pdfWidth, pdfHeight);
    pdf.save(`report-${host.value.name || 'host'}-${Date.now()}.pdf`);

    reportSnapshot.value = null;
    ElMessage.success(t('hosts.reportSuccess'));
  } catch (err) {
    console.error('Report generation error:', err);
    ElMessage.error(t('hosts.reportFailed'));
  } finally {
    reportGenerating.value = false
  }
}
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
.metrics-row {
  margin-bottom: 16px;
}
.mini-chart {
  width: 100%;
  height: 180px;
}
.chart {
  width: 100%;
  height: 320px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.range-picker {
  width: 260px;
}
.chart-alert {
  margin-bottom: 12px;
}
.status-legend {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-top: 8px;
  font-size: 13px;
  flex-wrap: wrap;
}
.legend-title {
  font-weight: 600;
  color: #303133;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #606266;
}
.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}

.report-canvas {
  position: fixed;
  top: -9999px;
  left: -9999px;
  width: 800px;
  background: #ffffff;
  padding: 40px;
  font-family: sans-serif;
  z-index: -1;
  color: #111827;
}

.report-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 3px solid #3b82f6;
  padding-bottom: 24px;
  margin-bottom: 32px;
}

.report-header h2 {
  margin: 0;
  font-size: 26px;
  color: #111827;
}

.report-subtitle {
  margin: 4px 0 0;
  color: #6b7280;
  font-size: 14px;
}

.report-badge {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #ffffff;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 18px;
}

.report-section {
  margin-bottom: 24px;
}

.report-section h3 {
  font-size: 18px;
  margin-bottom: 12px;
  color: #1f2937;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 6px;
}

.report-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
  font-size: 14px;
  color: #374151;
  line-height: 1.6;
}

.report-muted {
  color: #6b7280;
  font-size: 13px;
  font-style: italic;
}

.report-chat-line {
  margin-bottom: 10px;
  font-size: 13px;
  line-height: 1.5;
}

.report-chat-role {
  font-weight: 600;
  color: #374151;
}

.report-disclaimer {
  background: #f9fafb;
  padding: 16px;
  border-left: 4px solid #3b82f6;
  font-size: 12px;
  color: #4b5563;
  line-height: 1.6;
}
</style>
