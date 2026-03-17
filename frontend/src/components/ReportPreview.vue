<template>
  <div class="report-preview-container" v-loading="loading">
    <div v-if="data" class="report-content">
      <div class="report-header">
        <h2 class="report-title">{{ title }}</h2>
        <p class="report-meta">Generated at {{ generatedAt }} | Nagare Intelligence</p>
      </div>

      <el-row :gutter="20" class="summary-cards">
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-label">Total Alerts</div>
            <div class="stat-value">{{ data.TotalAlerts }}</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-label">Avg Health Score</div>
            <div class="stat-value text-success">{{ avgHealthScoreDisplay }}%</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-label">Critical Issues</div>
            <div class="stat-value text-danger">{{ criticalIssues }}</div>
          </div>
        </el-col>
      </el-row>

      <div v-if="data.Summary" class="section-title">Executive Summary</div>
      <div v-if="data.Summary" class="summary-text">{{ data.Summary }}</div>

      <el-row :gutter="20" class="chart-section">
        <el-col :span="12">
          <div class="chart-container">
            <div class="chart-title">Infrastructure Status</div>
            <div ref="pieChartRef" class="chart-canvas"></div>
          </div>
        </el-col>
        <el-col :span="12">
          <div class="chart-container">
            <div class="chart-title">Alert Trends</div>
            <div ref="lineChartRef" class="chart-canvas"></div>
          </div>
        </el-col>
      </el-row>

      <div class="section-title">Top Alert Hosts (Top 3)</div>
      <el-table :data="mappedTopAlertHosts" border stripe size="small">
        <el-table-column prop="name" label="Asset Name" />
        <el-table-column prop="ip" label="IP Address" />
        <el-table-column prop="summary" label="Issue Summary" />
        <el-table-column prop="alertCount" label="Alert Count" width="150" />
      </el-table>
    </div>
    <el-empty v-else-if="!loading" description="No report data available" />
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  data: {
    type: Object,
    default: null
  },
  title: String,
  generatedAt: String,
  loading: Boolean
})

const pieChartRef = ref(null)
const lineChartRef = ref(null)
let pieChart = null
let lineChart = null

const mappedTopAlertHosts = computed(() => {
  const source = props.data?.TopAlertHosts || props.data?.LongestDowntimeHosts
  if (!source) return []
  return source.map(row => ({
    name: row[0],
    ip: row[1],
    summary: row[2],
    alertCount: row[3]
  }))
})

const getDistributionValue = (distribution, keys) => {
  for (const key of keys) {
    if (Object.prototype.hasOwnProperty.call(distribution, key)) {
      const n = Number(distribution[key])
      if (Number.isFinite(n)) return n
    }
  }
  return 0
}

const criticalIssues = computed(() => {
  if (!props.data) return 0

  const direct = Number(props.data.CriticalIssues)
  if (Number.isFinite(direct) && direct >= 0) {
    return Math.round(direct)
  }

  const distribution = props.data.StatusDistribution || {}
  const errorCount = getDistributionValue(distribution, ['Error', 'error', '错误'])
  return Math.round(errorCount)
})

const avgHealthScoreDisplay = computed(() => {
  if (!props.data) return '0.00'

  const raw = Number(props.data.AvgHealthScore)
  if (Number.isFinite(raw) && raw >= 0) {
    return raw.toFixed(2)
  }

  // Backward compatibility for previously generated report payloads.
  const legacy = Number(props.data.AvgUptime)
  if (Number.isFinite(legacy) && legacy >= 0) {
    return legacy.toFixed(2)
  }

  return '0.00'
})

const initCharts = () => {
  if (!props.data) return

  // Pie Chart
  if (pieChartRef.value) {
    pieChart = echarts.init(pieChartRef.value)
    const pieData = Object.entries(props.data.StatusDistribution || {}).map(([name, value]) => ({ name, value }))
    pieChart.setOption({
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: '70%',
        data: pieData,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }]
    }, { notMerge: true })
  }

  // Line Chart
  if (lineChartRef.value) {
    lineChart = echarts.init(lineChartRef.value)
    lineChart.setOption({
      xAxis: {
        type: 'category',
        data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
      },
      yAxis: { type: 'value' },
      series: [{
        data: props.data.AlertTrend || [],
        type: 'line',
        smooth: true,
        areaStyle: {}
      }]
    }, { notMerge: true })
  }
}

watch(() => props.data, () => {
  nextTick(() => {
    if (pieChart) pieChart.dispose()
    if (lineChart) lineChart.dispose()
    initCharts()
  })
}, { deep: true })

onMounted(() => {
  initCharts()
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (pieChart) pieChart.dispose()
  if (lineChart) lineChart.dispose()
})

const handleResize = () => {
  if (pieChart) pieChart.resize()
  if (lineChart) lineChart.resize()
}
</script>

<style scoped>
.report-preview-container {
  min-height: 400px;
  padding: 10px;
}

.report-header {
  text-align: center;
  margin-bottom: 30px;
  border-bottom: 2px solid var(--el-border-color-lighter);
  padding-bottom: 20px;
}

.report-title {
  margin: 0;
  color: var(--brand-600);
  font-size: 24px;
}

.report-meta {
  color: var(--text-muted);
  font-size: 14px;
  margin-top: 8px;
}

.summary-cards {
  margin-bottom: 30px;
}

.stat-card {
  background: var(--surface-2);
  padding: 20px;
  border-radius: var(--radius-md);
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: var(--text-muted);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  margin: 30px 0 15px;
  padding-left: 10px;
  border-left: 4px solid var(--brand-500);
}

.summary-text {
  line-height: 1.6;
  color: var(--text-strong);
  background: var(--brand-50);
  padding: 15px;
  border-radius: var(--radius-sm);
  font-style: italic;
}

.chart-section {
  margin-top: 20px;
}

.chart-container {
  background: var(--surface-1);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: var(--radius-md);
  padding: 15px;
}

.chart-title {
  text-align: center;
  font-weight: 600;
  margin-bottom: 10px;
}

.chart-canvas {
  height: 250px;
  width: 100%;
}

.text-success { color: #67C23A; }
.text-danger { color: #F56C6C; }
</style>
