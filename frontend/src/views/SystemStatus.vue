<template>
  <div class="system-status-container">
    <el-row :gutter="20">
      <el-col :span="6" v-for="card in summaryCards" :key="card.title">
        <el-card shadow="hover" class="summary-card">
          <div class="summary-content">
            <el-icon :size="32" :color="card.color"><component :is="card.icon" /></el-icon>
            <div class="summary-text">
              <div class="summary-title">{{ card.title }}</div>
              <div class="summary-value">{{ card.value }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card header="Memory Usage (MB)">
          <div ref="memoryChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="Goroutines">
          <div ref="goroutineChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <el-card header="System Information">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="Go Version">{{ status.go_version }}</el-descriptions-item>
            <el-descriptions-item label="CPUs">{{ status.num_cpu }}</el-descriptions-item>
            <el-descriptions-item label="Uptime">{{ formatUptime(status.uptime) }}</el-descriptions-item>
            <el-descriptions-item label="Total Alloc">{{ formatBytes(status.memory_total) }}</el-descriptions-item>
            <el-descriptions-item label="System Memory">{{ formatBytes(status.memory_sys) }}</el-descriptions-item>
            <el-descriptions-item label="GC Count">{{ status.num_gc }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import axios from '@/utils/request'
import * as echarts from 'echarts'
import { Timer, Connection, PieChart, Cpu } from '@element-plus/icons-vue'

const status = ref({})
const memoryChart = ref(null)
const goroutineChart = ref(null)
let memoryChartInstance = null
let goroutineChartInstance = null
let refreshTimer = null

const memoryHistory = ref([])
const goroutineHistory = ref([])
const timeLabels = ref([])

const summaryCards = computed(() => [
  { title: 'Uptime', value: formatUptime(status.value.uptime), icon: Timer, color: '#409eff' },
  { title: 'Goroutines', value: status.value.goroutines, icon: Connection, color: '#67c23a' },
  { title: 'Allocated Mem', value: formatBytes(status.value.memory_alloc), icon: PieChart, color: '#e6a23c' },
  { title: 'CPUs', value: status.value.num_cpu, icon: Cpu, color: '#f56c6c' }
])

const fetchStatus = async () => {
  try {
    const res = await axios.get('/api/v1/system/status')
    if (res.success) {
      status.value = res.data
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
  if (!seconds) return '0s'
  const d = Math.floor(seconds / (3600 * 24))
  const h = Math.floor((seconds % (3600 * 24)) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  return `${d}d ${h}h ${m}m ${s}s`
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
