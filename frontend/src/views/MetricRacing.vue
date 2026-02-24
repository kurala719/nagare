<template>
  <div class="metric-racing-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('metricRacing.title') }}</h1>
      <p class="page-subtitle">{{ $t('metricRacing.subtitle') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-radio-group v-model="metricType" size="large" @change="handleTypeChange">
          <el-radio-button label="cpu">CPU</el-radio-button>
          <el-radio-button label="memory">Memory</el-radio-button>
          <el-radio-button label="network">Network</el-radio-button>
        </el-radio-group>
        
        <el-input-number v-model="refreshInterval" :min="2" :max="60" label="Refresh (s)" @change="handleIntervalChange" />
        <span class="filter-label">Refresh (s)</span>
      </div>
      
      <div class="action-group">
        <el-button :type="isRunning ? 'danger' : 'primary'" @click="toggleRacing">
          {{ isRunning ? $t('metricRacing.stop') : $t('metricRacing.start') }}
        </el-button>
      </div>
    </div>

    <el-card shadow="hover" class="racing-card">
      <div ref="racingChartRef" class="racing-chart"></div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts'
import axios from '@/utils/request'

const { t } = useI18n()
const racingChartRef = ref(null)
let racingChart = null
let timer = null

const metricType = ref('cpu')
const refreshInterval = ref(3)
const isRunning = ref(true)
const hostData = ref({}) // map of hostName -> value

const getTitle = () => {
  if (metricType.value === 'cpu') return t('metricRacing.cpuTitle')
  if (metricType.value === 'memory') return t('metricRacing.itemTitle') + ' (Memory)'
  return t('metricRacing.netTitle')
}

const fetchMetrics = async () => {
  try {
    const res = await axios.get('/api/v1/system/metrics')
    if (res.success && Array.isArray(res.data)) {
      const newMap = {}
      res.data.forEach(m => {
        const name = m.host_name || `Host ${m.host_id}`
        const itemLower = m.item_name.toLowerCase()
        
        // Basic filtering based on metricType
        let match = false
        if (metricType.value === 'cpu') {
          match = (itemLower.includes('cpu') || itemLower.includes('processor') || itemLower.includes('load')) && 
                  (m.units === '%' || itemLower.includes('usage') || itemLower.includes('util'))
        } else if (metricType.value === 'memory') {
          match = (itemLower.includes('mem') || itemLower.includes('ram')) && !itemLower.includes('percent')
        } else if (metricType.value === 'network') {
          match = itemLower.includes('bps') || itemLower.includes('traffic') || itemLower.includes('ifout') || itemLower.includes('ifin')
        }

        if (match) {
          const val = parseFloat(m.value)
          if (!isNaN(val)) {
            // Keep the max value for that host if multiple items match
            if (!newMap[name] || val > newMap[name]) {
              newMap[name] = val
            }
          }
        }
      })
      
      if (Object.keys(newMap).length > 0) {
        updateChart(newMap)
      }
    }
  } catch (error) {
    console.error('Failed to fetch racing metrics:', error)
  }
}

const updateChart = (data) => {
  if (!racingChart) return

  const sortedData = Object.entries(data)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 15) // Top 15 hosts

  const names = sortedData.map(d => d[0])
  const values = sortedData.map(d => d[1])

  racingChart.setOption({
    title: {
      text: getTitle()
    },
    yAxis: {
      data: names
    },
    series: [
      {
        data: values
      }
    ]
  })
}

const initChart = () => {
  if (!racingChartRef.value) return
  racingChart = echarts.init(racingChartRef.value)

  const option = {
    title: {
      text: getTitle(),
      left: 'center',
      textStyle: {
        fontSize: 20,
        fontWeight: 'bold'
      }
    },
    grid: {
      top: 60,
      bottom: 30,
      left: 150,
      right: 80
    },
    xAxis: {
      type: 'value',
      max: 'dataMax',
      splitLine: { show: true }
    },
    yAxis: {
      type: 'category',
      data: [],
      inverse: true,
      animationDuration: 300,
      animationDurationUpdate: 300,
      axisLabel: {
        show: true,
        fontSize: 14,
        fontWeight: 'bold'
      }
    },
    series: [
      {
        realtimeSort: true,
        name: 'Value',
        type: 'bar',
        data: [],
        label: {
          show: true,
          position: 'right',
          valueAnimation: true,
          fontWeight: 'bold',
          fontSize: 16
        },
        itemStyle: {
          color: function (param) {
            const colors = ['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc'];
            return colors[param.dataIndex % colors.length];
          }
        }
      }
    ],
    animationDuration: 0,
    animationDurationUpdate: 2000,
    animationEasing: 'linear',
    animationEasingUpdate: 'linear'
  }

  racingChart.setOption(option)
}

const toggleRacing = () => {
  isRunning.value = !isRunning.value
  if (isRunning.value) {
    startTimer()
  } else {
    stopTimer()
  }
}

const startTimer = () => {
  stopTimer()
  fetchMetrics()
  timer = setInterval(fetchMetrics, refreshInterval.value * 1000)
}

const stopTimer = () => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

const handleTypeChange = () => {
  if (racingChart) {
    racingChart.setOption({
      title: { text: getTitle() },
      yAxis: { data: [] },
      series: [{ data: [] }]
    })
  }
  fetchMetrics()
}

const handleIntervalChange = () => {
  if (isRunning.value) startTimer()
}

const handleResize = () => {
  racingChart?.resize()
}

onMounted(() => {
  initChart()
  startTimer()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  stopTimer()
  window.removeEventListener('resize', handleResize)
  racingChart?.dispose()
})
</script>

<style scoped>
.metric-racing-container {
  padding: 24px;
}

.racing-card {
  margin-top: 8px;
}

.racing-chart {
  width: 100%;
  height: calc(100vh - 280px);
  min-height: 500px;
}

.filter-label {
  margin-left: 8px;
  font-size: 13px;
  color: var(--text-muted);
}
</style>
