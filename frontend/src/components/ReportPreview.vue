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
            <div class="stat-label">Avg Uptime</div>
            <div class="stat-value text-success">{{ data.AvgUptime }}%</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-card">
            <div class="stat-label">Critical Issues</div>
            <div class="stat-value text-danger">{{ data.StatusDistribution?.Error || 0 }}</div>
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

      <div class="section-title">Top Resource Consumers (CPU)</div>
      <el-table :data="mappedCPUHosts" border stripe size="small">
        <el-table-column prop="name" label="Asset Name" />
        <el-table-column prop="ip" label="IP Address" />
        <el-table-column prop="usage" label="Avg Usage" width="100" />
        <el-table-column prop="units" label="Units" width="80" />
        <el-table-column prop="status" label="Status" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusTag(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="section-title">Stability Issues (Downtime)</div>
      <el-table :data="mappedDowntimeHosts" border stripe size="small">
        <el-table-column prop="name" label="Asset Name" />
        <el-table-column prop="ip" label="IP Address" />
        <el-table-column prop="downtime" label="Total Downtime" />
        <el-table-column prop="frequency" label="Frequency" width="150" />
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

const mappedCPUHosts = computed(() => {
  if (!props.data?.TopCPUHosts) return []
  return props.data.TopCPUHosts.map(row => ({
    name: row[0],
    ip: row[1],
    usage: row[2],
    units: row[3],
    status: row[4]
  }))
})

const mappedDowntimeHosts = computed(() => {
  if (!props.data?.LongestDowntimeHosts) return []
  return props.data.LongestDowntimeHosts.map(row => ({
    name: row[0],
    ip: row[1],
    downtime: row[2],
    frequency: row[3]
  }))
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
    })
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
    })
  }
}

const getStatusTag = (status) => {
  if (status === 'Active') return 'success'
  if (status === 'Warning') return 'warning'
  if (status === 'Error') return 'danger'
  return 'info'
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
