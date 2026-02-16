<template>
  <div class="dashboard-container">
    <div class="dashboard-header">
      <el-text size="large" tag="b">{{ $t('dashboard.title') }}</el-text>
      <span v-if="lastUpdated" class="dashboard-updated">
        {{ $t('dashboard.summaryLastUpdated') }}: {{ lastUpdated }}
      </span>
      <el-button size="small" @click="refreshAll" :loading="loading" class="refresh-btn">
        {{ $t('common.refresh') }}
      </el-button>
    </div>

    <div v-if="loading && !lastUpdated" class="loading-container">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p>{{ $t('dashboard.loading') }}</p>
    </div>

    <div v-else class="dashboard-content">
      <!-- Top Row: Summary Cards -->
      <SummaryCards :summary="summary" />

      <!-- Health Overview Section -->
      <div class="section-container">
        <HealthStats :score="healthScore" :loading="healthLoading" />
        
        <el-row :gutter="20">
          <el-col :xs="24" :lg="14">
            <HealthTrendChart ref="trendChart" />
          </el-col>
          <el-col :xs="24" :lg="10">
            <MetricsTable ref="metricsTable" />
          </el-col>
        </el-row>
      </div>

      <!-- Topology Section -->
      <div class="section-container">
        <TopologyChart ref="topologyChart" />
      </div>

      <!-- Experimental Features Row -->
      <el-row :gutter="20" class="section-container">
        <el-col :xs="24" :md="12">
          <VoiceControl />
        </el-col>
        <el-col :xs="24" :md="12">
          <MatrixStream />
        </el-col>
      </el-row>

      <!-- Recent Data Section -->
      <el-row :gutter="20" class="section-container">
        <el-col :xs="24" :lg="8">
          <RecentAlerts :alerts="recentAlerts" :loading="loading" />
        </el-col>
        <el-col :xs="24" :lg="8">
          <RecentHosts :hosts="recentHosts" :loading="loading" />
        </el-col>
        <el-col :xs="24" :lg="8">
          <RecentProviders :providers="recentProviders" :loading="loading" />
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { fetchAlertData } from '@/api/alerts'
import { fetchHostData } from '@/api/hosts'
import { fetchProviderData } from '@/api/providers'
import { fetchMonitorData } from '@/api/monitors'
import { authFetch } from '@/utils/authFetch'
import { useI18n } from 'vue-i18n'

import SummaryCards from './components/SummaryCards.vue'
import HealthStats from './components/HealthStats.vue'
import HealthTrendChart from './components/HealthTrendChart.vue'
import MetricsTable from './components/MetricsTable.vue'
import TopologyChart from './components/TopologyChart.vue'
import VoiceControl from './components/VoiceControl.vue'
import MatrixStream from './components/MatrixStream.vue'
import RecentAlerts from './components/RecentAlerts.vue'
import RecentHosts from './components/RecentHosts.vue'
import RecentProviders from './components/RecentProviders.vue'

export default defineComponent({
  name: 'Dashboard',
  components: {
    Loading,
    SummaryCards,
    HealthStats,
    HealthTrendChart,
    MetricsTable,
    TopologyChart,
    VoiceControl,
    MatrixStream,
    RecentAlerts,
    RecentHosts,
    RecentProviders
  },
  setup() {
    const { t } = useI18n()
    const loading = ref(false)
    const lastUpdated = ref('')
    
    // Data refs
    const summary = ref({
      alerts: { total: 0, critical: 0 },
      hosts: { total: 0, online: 0 },
      monitors: { total: 0, active: 0 },
      providers: { total: 0, active: 0 }
    })
    
    const recentAlerts = ref([])
    const recentHosts = ref([])
    const recentProviders = ref([])
    
    const healthLoading = ref(false)
    const healthScore = ref({})

    // Child refs for manual refresh
    const trendChart = ref(null)
    const metricsTable = ref(null)
    const topologyChart = ref(null)

    const loadDashboardData = async () => {
      loading.value = true
      try {
        await Promise.all([
          loadAlerts(),
          loadHosts(),
          loadMonitors(),
          loadProviders(),
          loadHealthScore()
        ])
        lastUpdated.value = new Date().toLocaleString()
      } catch (err) {
        console.error('Dashboard load failed:', err)
      } finally {
        loading.value = false
      }
    }

    const loadAlerts = async () => {
      const res = await fetchAlertData()
      const data = Array.isArray(res?.data || res) ? (res?.data || res) : []
      summary.value.alerts.total = data.length
      summary.value.alerts.critical = data.filter(a => 
        String(a.severity).toLowerCase() === 'critical' || String(a.severity).toLowerCase() === 'high'
      ).length
      recentAlerts.value = data.slice(0, 5)
    }

    const loadHosts = async () => {
      const res = await fetchHostData()
      const data = Array.isArray(res?.data || res) ? (res?.data || res) : []
      summary.value.hosts.total = data.length
      summary.value.hosts.online = data.filter(h => h.status === 1).length
      recentHosts.value = data.slice(0, 5)
    }

    const loadMonitors = async () => {
      const res = await fetchMonitorData()
      const data = Array.isArray(res?.data || res) ? (res?.data || res) : []
      summary.value.monitors.total = data.length
      summary.value.monitors.active = data.filter(m => m.status === 1).length
    }

    const loadProviders = async () => {
      const res = await fetchProviderData()
      const data = Array.isArray(res?.data || res) ? (res?.data || res) : []
      summary.value.providers.total = data.length
      summary.value.providers.active = data.filter(p => p.status === 1).length
      recentProviders.value = data.slice(0, 5)
    }

    const loadHealthScore = async () => {
      healthLoading.value = true
      try {
        const res = await authFetch('/api/v1/system/health')
        const json = await res.json()
        if (json.success) {
          healthScore.value = json.data
        }
      } catch (e) {
        console.error('Health load failed', e)
      } finally {
        healthLoading.value = false
      }
    }

    const refreshAll = () => {
      loadDashboardData()
      if (trendChart.value?.handleRefresh) trendChart.value.handleRefresh()
      if (metricsTable.value?.loadMetrics) metricsTable.value.loadMetrics()
      if (topologyChart.value?.handleRefresh) topologyChart.value.handleRefresh()
    }

    onMounted(() => {
      loadDashboardData()
    })

    return {
      loading,
      lastUpdated,
      summary,
      recentAlerts,
      recentHosts,
      recentProviders,
      healthLoading,
      healthScore,
      refreshAll,
      trendChart,
      metricsTable,
      topologyChart
    }
  }
})
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
}
.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.dashboard-updated {
  font-size: 13px;
  color: #909399;
  margin-left: auto;
  margin-right: 16px;
}
.loading-container {
  text-align: center;
  padding: 60px;
  color: #909399;
}
.section-container {
  margin-bottom: 20px;
}
</style>
