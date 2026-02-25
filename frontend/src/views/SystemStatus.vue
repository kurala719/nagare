<template>
  <div class="system-status-container">
    <div class="page-header system-status-header">
      <h1 class="page-title">{{ $t('systemStatus.title') }}</h1>
      <div class="refresh-info" v-if="lastUpdated">
        <el-tag size="small" type="info" effect="plain" class="refresh-tag">
          <el-icon class="is-loading"><Refresh /></el-icon>
          {{ $t('dashboard.summaryLastUpdated') }}: {{ lastUpdated }}
        </el-tag>
      </div>
    </div>

    <el-row :gutter="20">
      <el-col :span="6" v-for="card in summaryCards" :key="card.key">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <el-icon :size="32" :color="card.color"><component :is="card.icon" /></el-icon>
            <div class="summary-text">
              <div class="summary-title">{{ $t('systemStatus.' + card.key) }}</div>
              <div class="summary-value">{{ card.value }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card :header="$t('systemStatus.memoryUsage')">
          <div ref="memoryChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card :header="$t('systemStatus.goroutines')">
          <div ref="goroutineChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <el-card :header="$t('systemStatus.systemInfo')">
          <el-descriptions :column="3" border>
            <el-descriptions-item :label="$t('systemStatus.goVersion')">{{ status.go_version }}</el-descriptions-item>
            <el-descriptions-item :label="$t('systemStatus.cpus')">{{ status.num_cpu }}</el-descriptions-item>
            <el-descriptions-item :label="$t('systemStatus.uptime')">{{ formatUptime(status.uptime) }}</el-descriptions-item>
            <el-descriptions-item :label="$t('systemStatus.totalAlloc')">{{ formatBytes(status.memory_total) }}</el-descriptions-item>
            <el-descriptions-item :label="$t('systemStatus.systemMemory')">{{ formatBytes(status.memory_sys) }}</el-descriptions-item>
            <el-descriptions-item :label="$t('systemStatus.gcCount')">{{ status.num_gc }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from '@/utils/request'
import * as echarts from 'echarts'
import { Timer, Connection, PieChart, Cpu, Refresh } from '@element-plus/icons-vue'

const { t } = useI18n()
const status = ref({})
const lastUpdated = ref('')
const memoryChart = ref(null)
const goroutineChart = ref(null)
let memoryChartInstance = null
let goroutineChartInstance = null
let refreshTimer = null

const memoryHistory = ref([])
const goroutineHistory = ref([])
const timeLabels = ref([])

const summaryCards = computed(() => [
  { key: 'uptime', value: formatUptime(status.value.uptime), icon: Timer, color: '#409eff' },
  { key: 'goroutines', value: status.value.goroutines, icon: Connection, color: '#67c23a' },
  { key: 'allocatedMem', value: formatBytes(status.value.memory_alloc), icon: PieChart, color: '#e6a23c' },
  { key: 'cpus', value: status.value.num_cpu, icon: Cpu, color: '#f56c6c' }
])

const fetchStatus = async () => {
  try {
    const res = await axios.get('/api/v1/system/status')
    if (res.success) {
      status.value = res.data
      lastUpdated.value = new Date().toLocaleString()
      updateCharts(res.data)
    }
  } catch (error) {
    console.error('Failed to fetch system status:', error)
  }
}

const updateCharts = (data) => {
  const now = new Date().toLocaleTimeString()
  timeLabels.value.push(now)
  memoryHistory.value.push((data.memory_alloc / 1024 / 1024).toFixed(2))
  goroutineHistory.value.push(data.goroutines)

  if (timeLabels.value.length > 20) {
    timeLabels.value.shift()
    memoryHistory.value.shift()
    goroutineHistory.value.shift()
  }

  memoryChartInstance.setOption({
    xAxis: { data: timeLabels.value },
    series: [{ data: memoryHistory.value }]
  })

  goroutineChartInstance.setOption({
    xAxis: { data: timeLabels.value },
    series: [{ data: goroutineHistory.value }]
  })
}

const initCharts = () => {
  memoryChartInstance = echarts.init(memoryChart.value)
  goroutineChartInstance = echarts.init(goroutineChart.value)

  const baseOption = {
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', boundaryGap: false, data: [] },
    yAxis: { type: 'value' },
    series: [{ type: 'line', smooth: true, areaStyle: {} }]
  }

  memoryChartInstance.setOption({ ...baseOption, title: { show: false } })
  goroutineChartInstance.setOption({ ...baseOption, series: [{ ...baseOption.series[0], itemStyle: { color: '#67c23a' } }] })
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatUptime = (seconds) => {
  if (!seconds) return '0' + t('common.secondUnit')
  const d = Math.floor(seconds / (3600 * 24))
  const h = Math.floor((seconds % (3600 * 24)) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  
  let res = ''
  if (d > 0) res += d + t('common.dayUnit') + ' '
  if (h > 0) res += h + t('common.hourUnit') + ' '
  if (m > 0) res += m + t('common.minuteUnit') + ' '
  if (s > 0 || res === '') res += s + t('common.secondUnit')
  return res.trim()
}

onMounted(() => {
  initCharts()
  fetchStatus()
  refreshTimer = setInterval(fetchStatus, 3000)
  window.addEventListener('resize', () => {
    memoryChartInstance.resize()
    goroutineChartInstance.resize()
  })
})

onBeforeUnmount(() => {
  if (refreshTimer) clearInterval(refreshTimer)
  memoryChartInstance.dispose()
  goroutineChartInstance.dispose()
})
</script>

<style scoped>
.system-status-container {
  padding: 20px;
}

.system-status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.refresh-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.summary-card {
  height: 100px;
  display: flex;
  align-items: center;
}

.summary-content {
  display: flex;
  align-items: center;
  gap: 15px;
}

.summary-title {
  font-size: 14px;
  color: #909399;
}

.summary-value {
  font-size: 20px;
  font-weight: bold;
  color: #303133;
}

.chart-row {
  margin-top: 20px;
}

.chart {
  height: 300px;
  width: 100%;
}
</style>
