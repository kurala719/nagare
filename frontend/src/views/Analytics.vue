<template>
  <div class="analytics-container">
    <!-- Glassmorphism Header -->
    <div class="glass-header animate__animated animate__fadeIn">
      <div class="header-content">
        <h1 class="page-title">{{ $t('analytics.title') || 'Trend Analysis' }}</h1>
        <p class="page-subtitle">{{ $t('analytics.subtitle') || 'Insights and statistical trends from your monitoring data' }}</p>
      </div>
      <div class="header-actions">
        <el-button 
          class="chaos-btn" 
          type="danger" 
          :loading="chaosLoading" 
          @click="triggerChaosStorm"
        >
          <el-icon class="btn-icon"><Warning /></el-icon>
          {{ $t('analytics.triggerChaos') || 'Trigger Alert Storm' }}
        </el-button>
      </div>
    </div>

    <!-- Stats Overview Cards -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="8" v-for="(stat, index) in summaryStats" :key="index">
        <div class="stat-card glass-card animate__animated animate__zoomIn" :style="{ animationDelay: (index * 0.1) + 's' }">
          <div class="stat-icon" :style="{ background: stat.color }">
            <el-icon><component :is="stat.icon" /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stat.value }}</div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
          <div class="stat-chart-mini" :id="'mini-chart-' + index"></div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :lg="16" :md="24">
        <el-card class="chart-card glass-card animate__animated animate__fadeInUp">
          <template #header>
            <div class="card-header">
              <span class="header-text">{{ $t('analytics.alertTrend') }}</span>
              <el-tag size="small" type="primary" effect="dark">14 Days</el-tag>
            </div>
          </template>
          <div ref="trendChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :lg="8" :md="24">
        <el-card class="chart-card glass-card animate__animated animate__fadeInUp" style="animation-delay: 0.1s">
          <template #header>
            <div class="card-header">
              <span class="header-text">{{ $t('analytics.severityDist') }}</span>
            </div>
          </template>
          <div ref="severityChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row">
      <el-col :lg="12" :md="24">
        <el-card class="chart-card glass-card animate__animated animate__fadeInUp" style="animation-delay: 0.2s">
          <template #header>
            <div class="card-header">
              <span class="header-text">{{ $t('analytics.topKeywords') }}</span>
            </div>
          </template>
          <div ref="wordChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :lg="12" :md="24">
        <el-card class="chart-card glass-card animate__animated animate__fadeInUp" style="animation-delay: 0.3s">
          <template #header>
            <div class="card-header">
              <span class="header-text">{{ $t('analytics.alertIntensity') }}</span>
            </div>
          </template>
          <div ref="heatmapChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="chart-row last-row">
      <el-col :span="24">
        <el-card class="chart-card glass-card animate__animated animate__fadeInUp" style="animation-delay: 0.4s">
          <template #header>
            <div class="card-header">
              <span class="header-text">{{ $t('analytics.topNoisyHosts') }}</span>
            </div>
          </template>
          <div ref="hostChart" class="chart noisy-host-chart"></div>
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
import { Warning, Histogram, Management, Connection } from '@element-plus/icons-vue'
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
const analyticsData = ref({
  summary: {
    totalAlerts: 0,
    systemHealth: 0,
    activeHosts: 0
  }
})

const summaryStats = computed(() => [
  { 
    label: t('analytics.totalAlerts'), 
    value: analyticsData.value.summary.totalAlerts, 
    icon: Histogram, 
    color: 'linear-gradient(135deg, #1890ff 0%, #36cfc9 100%)' 
  },
  { 
    label: t('analytics.systemHealth'), 
    value: analyticsData.value.summary.systemHealth + '%', 
    icon: Connection, 
    color: 'linear-gradient(135deg, #52c41a 0%, #b7eb8f 100%)' 
  },
  { 
    label: t('analytics.activeHosts'), 
    value: analyticsData.value.summary.activeHosts, 
    icon: Management, 
    color: 'linear-gradient(135deg, #722ed1 0%, #b37feb 100%)' 
  }
])

const fetchAnalytics = async () => {
  try {
    const res = await axios.get('/api/v1/analytics/alerts')
    if (res.success) {
      analyticsData.value = res.data
      updateCharts(res.data)
    }
  } catch (error) {
    console.error('Failed to fetch analytics:', error)
  }
}

const updateCharts = (data) => {
  if (data.trend) {
    trendChartInstance.setOption({
      xAxis: { data: data.trend.map(t => t.date) },
      series: [{ data: data.trend.map(t => t.count) }]
    })
  }

  if (data.severityDist) {
    const sevMap = { 
      0: t('alerts.severityNotClassified'), 
      1: t('alerts.severityInfo'), 
      2: t('alerts.severityWarning'), 
      3: t('alerts.severityAverage'), 
      4: t('alerts.severityHigh'), 
      5: t('alerts.severityDisaster') 
    }
    const sevColors = ['#909399', '#409EFF', '#E6A23C', '#F56C6C', '#CF4444', '#000000']
    severityChartInstance.setOption({
      series: [{
        data: data.severityDist.map(s => ({
          name: sevMap[s.severity] || `Level ${s.severity}`,
          value: s.count,
          itemStyle: { color: sevColors[s.severity] || '#909399' }
        }))
      }]
    })
  }

  if (data.wordCloud) {
    const topWords = data.wordCloud.sort((a, b) => b.value - a.value).slice(0, 10).reverse()
    wordChartInstance.setOption({
      yAxis: { data: topWords.map(w => w.name) },
      series: [{ data: topWords.map(w => w.value) }]
    })
  }

  if (data.heatmap) {
    heatmapChartInstance.setOption({
      series: [{ data: data.heatmap }]
    })
  }

  if (data.topHosts) {
    hostChartInstance.setOption({
      xAxis: { data: data.topHosts.map(h => h.name || `Host ${h.host_id}`) },
      series: [{ data: data.topHosts.map(h => h.count) }]
    })
  }
}

const initCharts = () => {
  trendChartInstance = echarts.init(trendChart.value)
  severityChartInstance = echarts.init(severityChart.value)
  wordChartInstance = echarts.init(wordChart.value)
  heatmapChartInstance = echarts.init(heatmapChart.value)
  hostChartInstance = echarts.init(hostChart.value)

  // Common config
  const textStyle = { color: '#909399', fontSize: 12 }

  trendChartInstance.setOption({
    tooltip: { trigger: 'axis', backgroundColor: 'rgba(255, 255, 255, 0.9)' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { 
      type: 'category', 
      boundaryGap: false, 
      data: [],
      axisLine: { lineStyle: { color: '#f0f0f0' } },
      axisLabel: textStyle
    },
    yAxis: { 
      type: 'value',
      splitLine: { lineStyle: { type: 'dashed' } },
      axisLabel: textStyle
    },
    series: [{
      name: t('analytics.alerts'),
      type: 'line',
      smooth: true,
      symbolSize: 8,
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(24, 144, 255, 0.5)' },
          { offset: 1, color: 'rgba(24, 144, 255, 0)' }
        ])
      },
      itemStyle: { color: '#1890ff', width: 3 },
      data: []
    }]
  })

  severityChartInstance.setOption({
    tooltip: { trigger: 'item' },
    legend: { top: '0%', left: 'center', itemWidth: 10, itemHeight: 10, textStyle: { fontSize: 11 } },
    series: [{
      type: 'pie',
      radius: ['50%', '80%'],
      center: ['50%', '60%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 8, borderColor: '#fff', borderWidth: 3 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: '16', fontWeight: 'bold' } },
      data: []
    }]
  })

  wordChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '5%', bottom: '3%', containLabel: true },
    xAxis: { type: 'value', splitLine: { show: false }, axisLabel: textStyle },
    yAxis: { type: 'category', data: [], axisLabel: textStyle },
    series: [{ 
      type: 'bar', 
      barWidth: '60%',
      data: [],
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
          { offset: 0, color: '#1890ff' },
          { offset: 1, color: '#36cfc9' }
        ]),
        borderRadius: [0, 4, 4, 0]
      }
    }]
  })

  heatmapChartInstance.setOption({
    tooltip: { position: 'top' },
    visualMap: {
      min: 0,
      max: 20,
      type: 'piecewise',
      orient: 'horizontal',
      left: 'center',
      top: 0,
      textStyle: { fontSize: 10 }
    },
    calendar: {
      top: 60,
      left: 30,
      right: 30,
      cellSize: ['auto', 13],
      range: [new Date(new Date().getTime() - 90 * 24 * 3600 * 1000), new Date()],
      itemStyle: { borderWidth: 0.5, borderColor: '#eee' },
      yearLabel: { show: false },
      dayLabel: { firstDay: 1, color: '#909399' }
    },
    series: [{
      type: 'heatmap',
      coordinateSystem: 'calendar',
      data: []
    }]
  })

  hostChartInstance.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '3%', bottom: '5%', containLabel: true },
    xAxis: { 
      type: 'category', 
      data: [], 
      axisLabel: { interval: 0, rotate: 25, fontSize: 10, color: '#909399' },
      axisLine: { lineStyle: { color: '#f0f0f0' } }
    },
    yAxis: { type: 'value', splitLine: { lineStyle: { type: 'dashed' } } },
    series: [{
      type: 'bar',
      barWidth: '35%',
      itemStyle: { 
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#ff4d4f' },
          { offset: 1, color: '#ffccc7' }
        ]),
        borderRadius: [4, 4, 0, 0]
      },
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
  padding: 24px;
  background-color: #f6f8fb;
  min-height: 100vh;
}

/* Glassmorphism Classes */
.glass-header {
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.07);
}

.glass-card {
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.4) !important;
  border-radius: 16px !important;
  box-shadow: 0 4px 16px 0 rgba(31, 38, 135, 0.05) !important;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.glass-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 12px 24px 0 rgba(31, 38, 135, 0.1) !important;
}

/* Header Styles */
.page-title {
  font-size: 28px;
  font-weight: 800;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, #1f2937 0%, #4b5563 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.page-subtitle {
  font-size: 15px;
  color: #6b7280;
  margin: 0;
}

.chaos-btn {
  border-radius: 12px;
  padding: 12px 24px;
  font-weight: 600;
  transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.chaos-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 0 20px rgba(245, 108, 108, 0.4);
}

.btn-icon {
  margin-right: 8px;
}

/* Stats Overview */
.stats-row {
  margin-bottom: 24px;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
  height: 100px;
  position: relative;
  overflow: hidden;
}

.stat-icon {
  width: 54px;
  height: 54px;
  border-radius: 14px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 24px;
  color: white;
  margin-right: 18px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-info {
  z-index: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 800;
  color: #1f2937;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #6b7280;
  margin-top: 2px;
}

/* Chart Cards */
.chart-card :deep(.el-card__header) {
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  padding: 16px 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-text {
  font-size: 16px;
  font-weight: 700;
  color: #374151;
}

.chart {
  height: 350px;
  width: 100%;
}

.noisy-host-chart {
  height: 300px;
}

.chart-row {
  margin-bottom: 24px;
}

.last-row {
  margin-bottom: 0;
}

/* Animations (assuming animate.css is available or added via CDN) */
@import url('https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css');

@media (max-width: 992px) {
  .chart-card {
    margin-bottom: 20px;
  }
}
</style>

