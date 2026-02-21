<template>
  <div class="nagare-container animate-fade-in">
    <div class="detail-header">
      <div v-if="item.name">
        <h1 class="page-title">{{ item.name }}</h1>
        <p class="subtitle">{{ item.value || '-' }} {{ item.units || '' }}</p>
      </div>
      <div v-else class="page-header">
        <h1 class="page-title">{{ $t('items.detailTitle') }}</h1>
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
        <el-col :xs="12" :md="6">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('items.status') }}</div>
            <el-tooltip :content="item.status_reason || statusLabel(item.status)" placement="top">
              <div class="stat-value">
                <el-tag :type="statusTag(item.status)" effect="dark">{{ statusLabel(item.status) }}</el-tag>
              </div>
            </el-tooltip>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="6">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('common.enabled') }}</div>
            <div class="stat-value">{{ item.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}</div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="6">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('items.host') }}</div>
            <div class="stat-value">
              <router-link :to="'/host/' + item.hid + '/detail'" class="detail-link">
                {{ item.host_name || '#' + item.hid }}
              </router-link>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="12" :md="6">
          <el-card shadow="never">
            <div class="stat-label">{{ $t('items.comment') }}</div>
            <div class="stat-value">{{ item.comment || '-' }}</div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>{{ $t('items.statusChart') }}</span>
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
            <div style="position: relative; min-height: 400px;">
              <div ref="statusChartRef" class="chart"></div>
              <div v-if="historyLoading || historyError || historyEmpty" style="position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: white; display: flex; align-items: center; justify-content: center; z-index: 1;">
                <el-skeleton v-if="historyLoading" animated :rows="8" style="width: 100%; padding: 20px;" />
                <el-alert
                  v-else-if="historyError"
                  :title="historyError"
                  type="error"
                  show-icon
                  :closable="false"
                  style="width: calc(100% - 40px);"
                />
                <el-empty v-else-if="historyEmpty" :description="$t('common.noHistoryData')" />
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :lg="12">
          <el-card>
            <template #header>
              <div style="display: flex; justify-content: space-between; align-items: center; width: 100%;">
                <span>{{ $t('items.details') }}</span>
                <el-button link size="small" @click="showColumnDialog = true">{{ $t('common.columns') }}</el-button>
              </div>
            </template>
            <el-descriptions :column="1" border>
              <el-descriptions-item v-if="visibleColumns.includes('name')" :label="$t('items.name')">{{ item.name || '-' }}</el-descriptions-item>
              <el-descriptions-item v-if="visibleColumns.includes('value')" :label="$t('items.value')">{{ item.value || '-' }} {{ item.units || '' }}</el-descriptions-item>
              <el-descriptions-item v-if="visibleColumns.includes('status')" :label="$t('items.status')">{{ statusLabel(item.status) }}</el-descriptions-item>
              <el-descriptions-item v-if="visibleColumns.includes('comment')" :label="$t('items.comment')">{{ item.comment || '-' }}</el-descriptions-item>
            </el-descriptions>
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
</template>

<script setup>
import { onMounted, ref, onBeforeUnmount, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import { fetchItemHistory, getItemById } from '@/api/items'

const route = useRoute()
const { t } = useI18n()
const item = ref({})
const statusChartRef = ref(null)
const showColumnDialog = ref(false)
const historyRange = ref([])
const historyLoading = ref(false)
const historyError = ref(null)
const historyEmpty = ref(false)
const compareMode = ref(false)
const loading = ref(false)
const error = ref(null)
const isMounted = ref(false)
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

// Column configuration
const availableColumns = [
  { key: 'name', label: t('items.name') },
  { key: 'value', label: t('items.value') },
  { key: 'status', label: t('items.status') },
  { key: 'comment', label: t('items.comment') },
]

const visibleColumns = ref(['name', 'value', 'status', 'comment'])

const loadVisibleColumns = () => {
  const saved = localStorage.getItem('itemDetailColumns')
  if (saved) {
    try {
      visibleColumns.value = JSON.parse(saved)
    } catch (e) {
      visibleColumns.value = ['name', 'value', 'status', 'comment']
    }
  }
}

const saveVisibleColumns = () => {
  localStorage.setItem('itemDetailColumns', JSON.stringify(visibleColumns.value))
}

watch(visibleColumns, saveVisibleColumns, { deep: true })

let statusChart

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

const parseNumericValue = (value) => {
  const num = Number.parseFloat(String(value))
  return Number.isFinite(num) ? num : null
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

const buildStatusChart = (series, units, prevSeries = []) => {
  console.log('[ItemDetail] buildStatusChart called, chartRef exists:', !!statusChartRef.value, 'series length:', series.length)
  
  if (!statusChartRef.value) {
    console.warn('[ItemDetail] Chart ref is null, cannot build chart')
    return
  }

  try {
    if (!statusChart) {
      console.log('[ItemDetail] Initializing EChart...')
      statusChart = echarts.init(statusChartRef.value)
    }
    const safeUnits = units || ''
    const chartSeries = [
      {
        name: t('common.currentPeriod'),
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: series,
        itemStyle: { color: '#409EFF' },
        areaStyle: { opacity: 0.1 },
      }
    ]
    if (prevSeries.length > 0) {
      chartSeries.push({
        name: t('common.previousPeriod'),
        type: 'line',
        smooth: true,
        showSymbol: false,
        data: prevSeries,
        itemStyle: { color: '#909399' },
        lineStyle: { type: 'dashed' },
        areaStyle: { opacity: 0.05 },
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
            tip += `<div style="margin-bottom:4px;"><strong>${point.seriesName}</strong><br/>${timeLabel}<br/>${t('items.value')}: ${value ?? '-'} ${safeUnits}</div>`
          })
          return tip
        },
      },
      legend: {
        show: prevSeries.length > 0,
        data: prevSeries.length > 0 ? [t('common.currentPeriod'), t('common.previousPeriod')] : [],
      },
      grid: { left: 60, right: 20, top: prevSeries.length > 0 ? 40 : 20, bottom: 40 },
      xAxis: { type: 'time' },
      yAxis: { 
        type: 'value', 
        min: 'dataMin',
        name: safeUnits,
        nameLocation: 'end',
        nameGap: 10,
        nameTextStyle: {
          fontSize: 12,
          color: '#606266'
        }
      },
      series: chartSeries
    })
  } catch (err) {
    console.error('[ItemDetail] Error building chart:', err)
  }
}

const loadHistory = async () => {
  const itemId = Number(route.params.id)
  if (!itemId) return
  
  // Don't load history until component is mounted
  if (!isMounted.value) {
    console.log('[ItemDetail] Component not yet mounted, deferring loadHistory')
    return
  }
  
  historyLoading.value = true
  historyError.value = null
  historyEmpty.value = false
  try {
    const [start, end] = resolveRangeWindow()
    const from = Math.floor(start.getTime() / 1000)
    const to = Math.floor(end.getTime() / 1000)
    
    console.log('[ItemDetail] Loading history for item:', itemId, { from, to })
    
    // Fetch current period
    const resp = await fetchItemHistory(itemId, {
      from,
      to,
      limit: 500,
    })
    console.log('[ItemDetail] History response:', resp)
    
    const payload = resp?.data || resp || []
    let rows = Array.isArray(payload) ? payload : []
    console.log('[ItemDetail] Parsed rows:', rows)
    
    if (rows.length === 0) {
      const fallbackResp = await fetchItemHistory(itemId, { limit: 500 })
      const fallbackPayload = fallbackResp?.data || fallbackResp || []
      rows = Array.isArray(fallbackPayload) ? fallbackPayload : []
      if (rows.length === 0) {
        console.log('[ItemDetail] No history data found, showing empty state')
        historyEmpty.value = true
        await nextTick()
        buildStatusChart([], '', [])
        return
      }
      console.log('[ItemDetail] Using fallback data (no time filter), got:', rows.length, 'rows')
    }
    
    const series = []
    let units = ''
    rows.forEach((row) => {
      const sampledAt = row.sampled_at || row.SampledAt
      const value = row.value ?? row.Value
      const parsed = parseNumericValue(value)
      if (parsed === null || !sampledAt) return
      const time = new Date(sampledAt).getTime()
      if (Number.isNaN(time)) return
      units = units || row.units || row.Units || ''
      series.push([time, parsed])
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
        const prevResp = await fetchItemHistory(itemId, {
          from: prevFrom,
          to: prevTo,
          limit: 500,
        })
        const prevPayload = prevResp?.data || prevResp || []
        const prevRows = Array.isArray(prevPayload) ? prevPayload : []
        
        prevRows.forEach((row) => {
          const sampledAt = row.sampled_at || row.SampledAt
          const value = row.value ?? row.Value
          const parsed = parseNumericValue(value)
          if (parsed === null || !sampledAt) return
          const time = new Date(sampledAt).getTime()
          if (Number.isNaN(time)) return
          // Shift timestamps to align with current period
          const shiftedTime = time + duration
          prevSeries.push([shiftedTime, parsed])
        })
      } catch (err) {
        console.warn('Failed to load previous period:', err)
      }
    }
    
    historyEmpty.value = series.length === 0
    await nextTick()
    buildStatusChart(series, units, prevSeries)
  } catch (err) {
    historyError.value = err?.message || t('common.historyLoadFailed')
    await nextTick()
    buildStatusChart([], '', [])
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
  const itemId = Number(route.params.id)
  if (!itemId) return
  
  loading.value = true
  error.value = null
  
  try {
    const resp = await getItemById(itemId)
    const data = resp.data || resp
    item.value = {
      id: data.id || data.ID,
      name: data.name || data.Name || '',
      value: data.value || data.Value || '',
      units: data.units || data.Units || '',
      status: data.status ?? data.Status ?? 0,
      enabled: data.enabled ?? data.Enabled ?? 1,
      hid: data.hid || data.HID,
      comment: data.comment || data.Comment || '',
      status_reason: data.Reason || data.reason || data.Error || data.error || data.ErrorMessage || data.error_message || data.LastError || data.last_error || '',
    }
  } catch (err) {
    console.error('Failed to load item detail data', err)
    error.value = err.message || 'Failed to load item data'
  } finally {
    loading.value = false
  }
  
  // Load history after loading is complete and DOM has updated
  if (!error.value) {
    await nextTick()
    await loadHistory()
  }
}

const onResize = () => {
  if (statusChart) statusChart.resize()
}

let resizeObserver

onMounted(async () => {
  loadVisibleColumns()
  setDefaultHistoryRange()
  isMounted.value = true  
  
  // Wait for DOM to be ready
  await nextTick()
  
  // Use ResizeObserver for more reliable resizing than window resize event
  if (statusChartRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (statusChart) statusChart.resize()
    })
    resizeObserver.observe(statusChartRef.value)
  }
  
  await loadData()
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  if (statusChart) {
    statusChart.dispose()
    statusChart = null
  }
})

watch(historyRange, () => {
  loadHistory()
})

watch(compareMode, () => {
  loadHistory()
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
  font-size: 16px;
  font-weight: 600;
  margin-top: 6px;
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
</style>
