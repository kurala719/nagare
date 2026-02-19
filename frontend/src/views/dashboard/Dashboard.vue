<template>
  <div class="nagare-container">
    <div class="page-header dashboard-header">
      <div>
        <h1 class="page-title">{{ $t('dashboard.title') }}</h1>
        <p class="page-subtitle" v-if="lastUpdated">
          {{ $t('dashboard.summaryLastUpdated') }}: {{ lastUpdated }}
        </p>
      </div>
      <el-button type="primary" @click="refreshAll" :loading="loading" :icon="Refresh">
        {{ $t('common.refresh') }}
      </el-button>
    </div>

    <div v-if="loading && !lastUpdated" class="dashboard-content skeleton-container">
      <el-skeleton :rows="5" animated />
      <el-row :gutter="20" style="margin-top: 20px">
        <el-col :span="16"><el-skeleton style="width: 100%" :rows="10" animated /></el-col>
        <el-col :span="8"><el-skeleton style="width: 100%" :rows="10" animated /></el-col>
      </el-row>
    </div>

    <div v-else class="dashboard-content">
      <!-- Health Overview Section -->
      <div class="section-container">
        <HealthStats 
          :score="healthScore" 
          :alerts="summary.alerts"
          :providers="summary.providers"
          :monitors="summary.monitors"
          :loading="healthLoading || loading" 
        />
        
        <el-row :gutter="20">
          <el-col :xs="24" :lg="16">
            <HealthTrendChart ref="trendChart" />
          </el-col>
          <el-col :xs="24" :lg="8">
            <VoiceControl class="voice-control-card" />
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
          <MetricsTable ref="metricsTable" />
        </el-col>
        <el-col :xs="24" :md="12">
          <MatrixStream />
        </el-col>
      </el-row>

      <!-- Recent Data Section -->
      <el-row :gutter="20" class="section-container">
        <el-col :xs="24" :lg="12">
          <RecentAlerts :alerts="recentAlerts" :loading="loading" />
        </el-col>
        <el-col :xs="24" :lg="12">
          <RecentHosts :hosts="recentHosts" :loading="loading" />
        </el-col>
      </el-row>
      <el-row :gutter="20" class="section-container">
        <el-col :xs="24" :lg="12">
          <RecentMonitors :monitors="recentMonitors" :loading="loading" />
        </el-col>
        <el-col :xs="24" :lg="12">
          <RecentProviders :providers="recentProviders" :loading="loading" />
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { Loading, Refresh } from '@element-plus/icons-vue'
import { fetchAlertData } from '@/api/alerts'
import { fetchHostData } from '@/api/hosts'
import { fetchProviderData } from '@/api/providers'
import { fetchMonitorData } from '@/api/monitors'
import { authFetch } from '@/utils/authFetch'
import { useI18n } from 'vue-i18n'

import HealthStats from './components/HealthStats.vue'
import HealthTrendChart from './components/HealthTrendChart.vue'
import MetricsTable from './components/MetricsTable.vue'
import TopologyChart from './components/TopologyChart.vue'
import VoiceControl from './components/VoiceControl.vue'
import MatrixStream from './components/MatrixStream.vue'
import RecentAlerts from './components/RecentAlerts.vue'
import RecentHosts from './components/RecentHosts.vue'
import RecentProviders from './components/RecentProviders.vue'
import RecentMonitors from './components/RecentMonitors.vue'

export default defineComponent({
  name: 'Dashboard',
  components: {
    Loading,
    HealthStats,
    HealthTrendChart,
    MetricsTable,
    TopologyChart,
    VoiceControl,
    MatrixStream,
    RecentAlerts,
    RecentHosts,
    RecentProviders,
    RecentMonitors
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
    const recentMonitors = ref([])
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
      recentMonitors.value = data.slice(0, 5)
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
      recentMonitors,
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
.dashboard-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: 32px;
}

.loading-container {
  text-align: center;
  padding: 80px;
  color: var(--text-muted);
}

.section-container {
  margin-bottom: 24px;
}

.voice-control-card {
  height: 100%;
  min-height: 360px;
}

:deep(.el-card) {
  height: 100%;
}
</style>
