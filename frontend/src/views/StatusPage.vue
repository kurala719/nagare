<template>
  <div class="status-page">
    <div class="status-container">
      <!-- Header -->
      <div class="status-header">
        <div class="logo-area">
          <el-icon :size="32" class="logo-icon"><Monitor /></el-icon>
          <h1 class="logo-text">Nagare Status</h1>
        </div>
        <div class="header-actions">
          <el-button type="primary" link @click="$router.push('/login')">Login</el-button>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-state">
        <el-icon class="is-loading" :size="40"><Loading /></el-icon>
        <p>Checking system status...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <el-result icon="error" title="Failed to load status" :sub-title="error">
          <template #extra>
            <el-button type="primary" @click="fetchStatus">Retry</el-button>
          </template>
        </el-result>
      </div>

      <!-- Content -->
      <div v-else class="status-content">
        <!-- Overall Status Banner -->
        <div :class="['status-banner', overallStatusClass]">
          <el-icon :size="24" class="banner-icon">
            <CircleCheckFilled v-if="statusData.overall_status === 'operational'" />
            <WarningFilled v-else-if="statusData.overall_status === 'degraded'" />
            <CircleCloseFilled v-else />
          </el-icon>
          <span class="banner-text">{{ statusData.overall_message }}</span>
        </div>

        <!-- Active Incidents -->
        <div v-if="statusData.active_incidents && statusData.active_incidents.length > 0" class="section incidents-section">
          <h2 class="section-title">Active Incidents</h2>
          <div v-for="incident in statusData.active_incidents" :key="incident.id" class="incident-card">
            <div class="incident-header">
              <span class="incident-title">{{ incident.message }}</span>
              <el-tag type="danger" effect="dark" size="small">Critical</el-tag>
            </div>
            <div class="incident-body">
              <p class="incident-time">Reported at {{ new Date(incident.created_at).toLocaleString() }}</p>
            </div>
          </div>
        </div>

        <!-- Services List -->
        <div class="section services-section">
          <h2 class="section-title">System Metrics</h2>
          <div class="services-list">
            <div v-for="service in statusData.services" :key="service.id" class="service-item">
              <div class="service-info">
                <span class="service-name">{{ service.name }}</span>
                <!-- <span class="service-uptime" v-if="service.uptime">{{ service.uptime }} uptime</span> -->
              </div>
              <div class="service-status">
                <el-icon :class="['status-icon', getServiceStatusClass(service.status)]">
                  <CircleCheckFilled v-if="service.status === 'operational'" />
                  <WarningFilled v-else-if="service.status === 'degraded'" />
                  <CircleCloseFilled v-else />
                </el-icon>
                <span :class="['status-text', getServiceStatusClass(service.status)]">
                  {{ formatStatus(service.status) }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="status-footer">
        <p>Powered by Nagare Monitoring</p>
      </div>
    </div>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted, computed } from 'vue'
import { Monitor, Loading, CircleCheckFilled, WarningFilled, CircleCloseFilled } from '@element-plus/icons-vue'
import request from '@/utils/request'

export default defineComponent({
  name: 'StatusPage',
  components: {
    Monitor,
    Loading,
    CircleCheckFilled,
    WarningFilled,
    CircleCloseFilled
  },
  setup() {
    const loading = ref(true)
    const error = ref(null)
    const statusData = ref({
      overall_status: 'operational', // operational, degraded, outage
      overall_message: 'Loading...',
      active_incidents: [],
      services: []
    })

    const overallStatusClass = computed(() => {
      switch (statusData.value.overall_status) {
        case 'operational': return 'status-operational'
        case 'degraded': return 'status-degraded'
        case 'outage': return 'status-outage'
        default: return 'status-unknown'
      }
    })

    const fetchStatus = async () => {
      loading.value = true
      error.value = null
      try {
        const response = await request.get('/public/status')
        if (response.data.success) {
          statusData.value = response.data.data
        } else {
          error.value = response.data.error || 'Failed to fetch status'
        }
      } catch (err) {
        error.value = err.message || 'Network error'
      } finally {
        loading.value = false
      }
    }

    const getServiceStatusClass = (status) => {
      switch (status) {
        case 'operational': return 'text-success'
        case 'degraded': return 'text-warning'
        case 'outage': return 'text-danger'
        default: return 'text-info'
      }
    }

    const formatStatus = (status) => {
      if (!status) return 'Unknown'
      return status.charAt(0).toUpperCase() + status.slice(1)
    }

    onMounted(() => {
      fetchStatus()
    })

    return {
      loading,
      error,
      statusData,
      overallStatusClass,
      fetchStatus,
      getServiceStatusClass,
      formatStatus
    }
  }
})
</script>

<style scoped>
.status-page {
  min-height: 100vh;
  background-color: #f8f9fa;
  display: flex;
  justify-content: center;
  padding: 40px 20px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
}

.status-container {
  width: 100%;
  max-width: 800px;
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #2c3e50;
}

.logo-icon {
  color: #409eff;
}

.logo-text {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
}

.loading-state, .error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #606266;
}

.status-banner {
  display: flex;
  align-items: center;
  padding: 24px;
  border-radius: 8px;
  margin-bottom: 40px;
  color: white;
  font-size: 20px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.banner-icon {
  margin-right: 16px;
}

.status-operational {
  background-color: #67c23a;
}

.status-degraded {
  background-color: #e6a23c;
}

.status-outage {
  background-color: #f56c6c;
}

.section {
  margin-bottom: 40px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
  overflow: hidden;
  border: 1px solid #ebeef5;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
  padding: 20px 24px;
  margin: 0;
  border-bottom: 1px solid #ebeef5;
  background-color: #fcfcfc;
}

.incident-card {
  padding: 24px;
  border-bottom: 1px solid #ebeef5;
}

.incident-card:last-child {
  border-bottom: none;
}

.incident-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.incident-title {
  font-weight: 600;
  font-size: 16px;
  color: #303133;
}

.incident-time {
  font-size: 13px;
  color: #909399;
  margin: 0;
}

.services-list {
  padding: 0;
}

.service-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #ebeef5;
}

.service-item:last-child {
  border-bottom: none;
}

.service-name {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
}

.service-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
.text-info { color: #909399; }

.status-footer {
  text-align: center;
  color: #909399;
  font-size: 13px;
  margin-top: 60px;
  border-top: 1px solid #e0e0e0;
  padding-top: 20px;
}
</style>
