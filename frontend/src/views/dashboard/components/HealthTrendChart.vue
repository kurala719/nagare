<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.healthTrendTitle') }}</span>
        <div class="card-actions">
          <el-switch
            v-model="compareMode"
            size="small"
            :active-text="$t('common.comparePrevious')"
            style="margin-right: 8px;"
            @change="handleRefresh"
          />
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            :shortcuts="shortcuts"
            :start-placeholder="$t('common.startTime')"
            :end-placeholder="$t('common.endTime')"
            size="small"
            class="trend-range"
            @change="handleRefresh"
          />
          <el-button size="small" @click="handleRefresh" :loading="loading">
            {{ $t('common.refresh') }}
          </el-button>
        </div>
      </div>
    </template>

    <el-skeleton v-if="loading" animated :rows="6" />
    <el-alert
      v-else-if="error"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="trend-alert"
    />
    <el-empty v-else-if="empty" :description="$t('common.noHistoryData')" />
    <div v-else ref="chartRef" class="trend-chart"></div>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { useI18n } from 'vue-i18n'
import { fetchNetworkStatusHistory } from '@/api/system'
import { getToken } from '@/utils/auth'

export default defineComponent({
  name: 'HealthTrendChart',
  setup() {
    const { t } = useI18n()
    const chartRef = ref(null)
    const chartInstance = ref(null)
    const loading = ref(false)
    const error = ref(null)
    const empty = ref(false)
    const compareMode = ref(false)
    const dateRange = ref([])

    const shortcuts = [
      { text: '1h', value: () => [new Date(Date.now() - 3600 * 1000), new Date()] },
      { text: '6h', value: () => [new Date(Date.now() - 6 * 3600 * 1000), new Date()] },
      { text: '24h', value: () => [new Date(Date.now() - 24 * 3600 * 1000), new Date()] },
      { text: '7d', value: () => [new Date(Date.now() - 7 * 24 * 3600 * 1000), new Date()] },
      { text: '30d', value: () => [new Date(Date.now() - 30 * 24 * 3600 * 1000), new Date()] },
    ]

    const resolveTrendWindow = () => {
      if (Array.isArray(dateRange.value) && dateRange.value.length === 2 && dateRange.value[0] && dateRange.value[1]) {
        return [new Date(dateRange.value[0]), new Date(dateRange.value[1])]
      }
      const end = new Date()
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000)
      return [start, end]
    }

    const buildChart = (series, prevSeries = []) => {
      if (!chartRef.value) return
      if (!chartInstance.value) {
        chartInstance.value = echarts.init(chartRef.value)
      }
      
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

      chartInstance.value.setOption({
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            let tip = ''
            params.forEach((point) => {
              const value = point.data?.[1]
              const time = new Date(point.data?.[0])
              const timeLabel = Number.isNaN(time.getTime()) ? '-' : time.toLocaleString()
              tip += `<div style="margin-bottom:4px;"><strong>${point.seriesName}</strong><br/>${timeLabel}<br/>Score: ${value ?? '-'}</div>`
            })
            return tip
          },
        },
        legend: {
          show: prevSeries.length > 0,
          data: prevSeries.length > 0 ? [t('common.currentPeriod'), t('common.previousPeriod')] : [],
        },
        grid: { left: 40, right: 20, top: prevSeries.length > 0 ? 40 : 20, bottom: 40 },
        xAxis: { type: 'time' },
        yAxis: { type: 'value', min: 0, max: 100 },
        series: chartSeries,
      }, { notMerge: true })
    }

    const loadData = async () => {
      if (!getToken()) return
      
      loading.value = true
      error.value = null
      empty.value = false
      
      // Dispose old instance to prevent stale DOM references when v-if toggles
      if (chartInstance.value) {
        chartInstance.value.dispose()
        chartInstance.value = null
      }
      
      try {
        const [start, end] = resolveTrendWindow()
        const from = Math.floor(start.getTime() / 1000)
        const to = Math.floor(end.getTime() / 1000)
        
        
        
        const response = await fetchNetworkStatusHistory({ from, to, limit: 500 })
        
        
        
        // Backend returns {success: true, data: [...]}
        const extractData = (res) => {
          if (!res) return []
          if (res.success && res.data !== undefined) {
            return Array.isArray(res.data) ? res.data : []
          }
          if (Array.isArray(res)) return res
          return []
        }
        const rows = extractData(response)
        
        
        
        if (rows.length === 0) {
          empty.value = true
          if (chartInstance.value) chartInstance.value.dispose(); chartInstance.value = null;
          return
        }

        const series = rows.map(row => {
          const t = new Date(row.sampled_at || row.SampledAt).getTime()
          return [t, Number(row.score ?? row.Score)]
        }).filter(p => !Number.isNaN(p[0]))

        let prevSeries = []
        if (compareMode.value) {
          const duration = end.getTime() - start.getTime()
          const prevStart = new Date(start.getTime() - duration)
          const prevEnd = new Date(start.getTime())
          
          try {
            const prevResp = await fetchNetworkStatusHistory({
              from: Math.floor(prevStart.getTime() / 1000),
              to: Math.floor(prevEnd.getTime() / 1000),
              limit: 500
            })
            const prevRows = Array.isArray(prevResp?.data || prevResp) ? (prevResp?.data || prevResp) : []
            
            prevSeries = prevRows.map(row => {
              const t = new Date(row.sampled_at || row.SampledAt).getTime()
              return [t + duration, Number(row.score ?? row.Score)] // Shift time
            }).filter(p => !Number.isNaN(p[0]))
          } catch (e) {
            console.warn('Failed to load previous period data', e)
          }
        }
        
        
        
        empty.value = series.length === 0
        loading.value = false
        if (!empty.value) {
          await nextTick()
          buildChart(series, prevSeries)
        }
      } catch (err) {
        console.error('HealthTrendChart load error:', err)
        const msg = err.message || t('common.historyLoadFailed')
        if (err.message && (err.message.includes('401') || err.message.includes('403') || err.message.includes('Unauthorized') || err.message.includes('Forbidden'))) {
          error.value = t('common.sessionExpired') || 'Please log in to view health trend'
        } else {
          error.value = msg
        }
      } finally {
        loading.value = false
      }
    }

    const handleRefresh = () => {
      loadData()
    }

    const onResize = () => {
      if (chartInstance.value) chartInstance.value.resize()
    }

    onMounted(() => {
      // Set default range if empty
      if (dateRange.value.length === 0) {
        const end = new Date()
        const start = new Date(end.getTime() - 24 * 3600 * 1000)
        dateRange.value = [start, end]
      }
      loadData()
      window.addEventListener('resize', onResize)
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', onResize)
      if (chartInstance.value) {
        chartInstance.value.dispose()
        chartInstance.value = null
      }
    })

    return {
      chartRef,
      loading,
      error,
      empty,
      compareMode,
      dateRange,
      shortcuts,
      handleRefresh
    }
  }
})
</script>

<style scoped>
.trend-range {
  width: 260px;
}
.trend-chart {
  width: 100%;
  height: 260px;
}
.trend-alert {
  margin-bottom: 12px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
