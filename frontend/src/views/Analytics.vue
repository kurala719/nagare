<template>
  <div class="analytics-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('analytics.title') || 'Trend Analysis' }}</h1>
      <p class="page-subtitle">{{ $t('analytics.subtitle') || 'Insights and statistical trends from your monitoring data' }}</p>
    </div>

    <el-row :gutter="20">
      <el-col :span="24">
        <el-card class="chaos-card">
          <div class="chaos-header">
            <div class="chaos-info">
              <h3>{{ $t('analytics.chaosTitle') || 'Chaos Simulator' }}</h3>
              <p>{{ $t('analytics.chaosDesc') || 'Trigger a simulated alert storm to test system resilience.' }}</p>
            </div>
            <el-button type="danger" :loading="chaosLoading" @click="triggerChaosStorm">
              <el-icon><Warning /></el-icon>
              {{ $t('analytics.triggerChaos') || 'Trigger Alert Storm' }}
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ $t('analytics.alertTrend') }}</span>
            </div>
          </template>
          <div ref="trendChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ $t('analytics.severityDist') }}</span>
            </div>
          </template>
          <div ref="severityChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ $t('analytics.topKeywords') }}</span>
            </div>
          </template>
          <div ref="wordChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ $t('analytics.alertIntensity') }}</span>
            </div>
          </template>
          <div ref="heatmapChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>{{ $t('analytics.topNoisyHosts') }}</span>
            </div>
          </template>
          <div ref="hostChart" class="chart" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from '@/utils/request'
import * as echarts from 'echarts'
import { Warning } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const trendChart = ref(null)
const severityChart = ref(null)
const wordChart = ref(null)
const heatmapChart = ref(null)
const hostChart = ref(null)

let trendChartInstance = null
let severityChartInstance = null
let wordChartInstance = null
let heatmapChartInstance = null
let hostChartInstance = null

const chaosLoading = ref(false)

const fetchAnalytics = async () => {
  try {
    const res = await axios.get('/api/v1/analytics/alerts')
    if (res.success) {
      updateCharts(res.data)
    }
  } catch (error) {
    console.error('Failed to fetch analytics:', error)
  }
}

const updateCharts = (data) => {
  // 1. Trend Chart
  if (data.trend) {
    trendChartInstance.setOption({
      xAxis: { data: data.trend.map(t => t.Date) },
      series: [{ data: data.trend.map(t => t.Count) }]
    })
  }

  // 2. Severity Chart
  if (data.severityDist) {
    const sevMap = { 0: 'Info', 1: 'Warning', 2: 'Average', 3: 'High', 4: 'Disaster' }
    const sevColors = ['#909399', '#E6A23C', '#F56C6C', '#CF4444', '#000000']
    severityChartInstance.setOption({
      series: [{
        data: data.severityDist.map(s => ({
          name: sevMap[s.Severity] || `Level ${s.Severity}`,
          value: s.Count
        }))
      }]
    })
  }

  // 3. Word Chart
  if (data.wordCloud) {
    const topWords = data.wordCloud.sort((a, b) => b.value - a.value).slice(0, 10).reverse()
    wordChartInstance.setOption({
      yAxis: { data: topWords.map(w => w.name) },
      series: [{ data: topWords.map(w => w.value) }]
    })
  }

  // 4. Heatmap
  if (data.heatmap) {
    heatmapChartInstance.setOption({
      series: [{ data: data.heatmap }]
    })
  }

  // 5. Host Chart
  if (data.topHosts) {
    hostChartInstance.setOption({
      xAxis: { data: data.topHosts.map(h => h.Name || `Host ${h.HostID}`) },
      series: [{ data: data.topHosts.map(h => h.Count) }]
    })
  }
}

const initCharts = () => {
  trendChartInstance = echarts.init(trendChart.value)
  severityChartInstance = echarts.init(severityChart.value)
  wordChartInstance = echarts.init(wordChart.value)
  heatmapChartInstance = echarts.init(heatmapChart.value)
  hostChartInstance = echarts.init(hostChart.value)

  // Trend Chart Config
  trendChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: [] },
    yAxis: { type: 'value' },
    series: [{
      name: t('analytics.alerts'),
      type: 'line',
      smooth: true,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(24, 144, 255, 0.3)' },
          { offset: 1, color: 'rgba(24, 144, 255, 0)' }
        ])
      },
      itemStyle: { color: '#1890ff' },
      data: []
    }]
  })

  // Severity Chart Config
  severityChartInstance.setOption({
    tooltip: { trigger: 'item' },
    legend: { bottom: '5%', left: 'center' },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: { borderRadius: 10, borderColor: '#fff', borderWidth: 2 },
      label: { show: false, position: 'center' },
      emphasis: { label: { show: true, fontSize: '20', fontWeight: 'bold' } },
      labelLine: { show: false },
      data: []
    }]
  })

  // Word Chart Config
  wordChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'value' },
    yAxis: { type: 'category', data: [] },
    series: [{ 
      type: 'bar', 
      data: [],
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: '#1890ff' },
          { offset: 1, color: '#36cfc9' }
        ])
      }
    }]
  })

  // Heatmap Config
  heatmapChartInstance.setOption({
    tooltip: { position: 'top' },
    visualMap: {
      min: 0,
      max: 50,
      type: 'piecewise',
      orient: 'horizontal',
      left: 'center',
      top: 0
    },
    calendar: {
      top: 60,
      left: 30,
      right: 30,
      cellSize: ['auto', 13],
      range: [new Date(new Date().getTime() - 90 * 24 * 3600 * 1000), new Date()],
      itemStyle: { borderWidth: 0.5 },
      yearLabel: { show: false }
    },
    series: {
      type: 'heatmap',
      coordinateSystem: 'calendar',
      data: []
    }
  })

  // Host Chart Config
  hostChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: [], axisLabel: { interval: 0, rotate: 30 } },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar',
      barWidth: '40%',
      itemStyle: { color: '#f56c6c' },
      data: []
    }]
  })
}

const triggerChaosStorm = async () => {
  chaosLoading.value = true
  try {
    const res = await axios.post('/api/v1/chaos/alert-storm')
    if (res.success) {
      ElMessage.success('Alert storm triggered! Check the dashboard or alert list.')
    }
  } catch (error) {
    ElMessage.error('Failed to trigger chaos storm')
  } finally {
    chaosLoading.value = false
  }
}

const handleResize = () => {
  trendChartInstance?.resize()
  severityChartInstance?.resize()
  wordChartInstance?.resize()
  heatmapChartInstance?.resize()
  hostChartInstance?.resize()
}

onMounted(() => {
  initCharts()
  fetchAnalytics()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  trendChartInstance?.dispose()
  severityChartInstance?.dispose()
  wordChartInstance?.dispose()
  heatmapChartInstance?.dispose()
  hostChartInstance?.dispose()
})
</script>

<style scoped>
.analytics-container {
  padding: 20px;
}

.chaos-card {
  margin-bottom: 20px;
  border-left: 5px solid var(--el-color-danger);
  background-color: var(--el-color-danger-light-9);
}

.chaos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chaos-info h3 {
  margin: 0 0 8px 0;
  color: var(--el-color-danger);
}

.chaos-info p {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.chart-row {
  margin-top: 20px;
}

.chart {
  height: 350px;
  width: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
}
</style>

