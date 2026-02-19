<template>
  <div class="analytics-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card class="chaos-card" bg-color="var(--el-color-danger-light-9)">
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
      <el-col :span="12">
        <el-card header="Top Alert Keywords">
          <div ref="wordChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="Alert Intensity (Heatmap)">
          <div ref="heatmapChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import axios from '@/utils/request'
import * as echarts from 'echarts'
import { Warning } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const wordChart = ref(null)
const heatmapChart = ref(null)
let wordChartInstance = null
let heatmapChartInstance = null
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
  // Update Bar Chart (Top Words)
  const topWords = data.wordCloud.sort((a, b) => b.value - a.value).slice(0, 10).reverse()
  wordChartInstance.setOption({
    yAxis: { data: topWords.map(w => w.name) },
    series: [{ data: topWords.map(w => w.value) }]
  })

  // Update Heatmap
  heatmapChartInstance.setOption({
    series: [{ data: data.heatmap }]
  })
}

const initCharts = () => {
  wordChartInstance = echarts.init(wordChart.value)
  heatmapChartInstance = echarts.init(heatmapChart.value)

  wordChartInstance.setOption({
    tooltip: { trigger: 'axis' },
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

onMounted(() => {
  initCharts()
  fetchAnalytics()
  window.addEventListener('resize', () => {
    wordChartInstance.resize()
    heatmapChartInstance.resize()
  })
})

onBeforeUnmount(() => {
  wordChartInstance.dispose()
  heatmapChartInstance.dispose()
})
</script>

<style scoped>
.analytics-container {
  padding: 20px;
}

.chaos-card {
  margin-bottom: 20px;
  border-left: 5px solid var(--el-color-danger);
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
  height: 400px;
  width: 100%;
}
</style>
