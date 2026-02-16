<template>
  <el-card class="detail-card topology-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.topologyTitle') }}</span>
        <div class="card-actions">
          <el-button size="small" @click="handleRefresh" :loading="loading">
            {{ $t('dashboard.refresh') }}
          </el-button>
        </div>
      </div>
    </template>

    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading" size="40" color="#409EFF">
        <Loading />
      </el-icon>
      <p>{{ $t('dashboard.loadingTopology') }}</p>
    </div>

    <el-alert
      v-else-if="error"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="topology-alert"
    />

    <el-empty v-else-if="empty" :description="$t('dashboard.noTopology')" />
    <div v-else ref="chartRef" class="topology-chart"></div>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { useI18n } from 'vue-i18n'
import { Loading } from '@element-plus/icons-vue'
import { fetchSiteData } from '@/api/sites'
import { fetchHostData } from '@/api/hosts'
import { fetchMonitorData } from '@/api/monitors'

export default defineComponent({
  name: 'TopologyChart',
  components: { Loading },
  setup() {
    const { t } = useI18n()
    const chartRef = ref(null)
    const chartInstance = ref(null)
    const loading = ref(false)
    const error = ref(null)
    const empty = ref(false)

    const getStatusInfo = (status) => {
      const map = {
        0: { label: t('common.statusInactive'), type: 'info' },
        1: { label: t('common.statusActive'), type: 'success' },
        2: { label: t('common.statusError'), type: 'danger' },
        3: { label: t('common.statusSyncing'), type: 'warning' },
      }
      return map[status] || map[0]
    }

    const getStatusColor = (status) => {
      const palette = {
        0: '#909399',
        1: '#67c23a',
        2: '#f56c6c',
        3: '#e6a23c',
      }
      return palette[status] || palette[0]
    }

    const buildGraph = (sites, hosts, monitors) => {
      const nodes = []
      const links = []
      const siteMap = new Map()
      const monitorMap = new Map()

      monitors.forEach((monitor) => {
        const id = Number(monitor.ID || monitor.id)
        if (!id) return
        const name = monitor.Name || monitor.name || `Monitor ${id}`
        const status = monitor.Status ?? monitor.status ?? 0
        const nodeId = `monitor-${id}`
        monitorMap.set(id, nodeId)
        nodes.push({
          id: nodeId,
          name,
          category: 0,
          symbolSize: 64,
          value: { status },
          itemStyle: { color: getStatusColor(status) },
          label: { show: true },
        })
      })

      sites.forEach((site) => {
        const id = Number(site.ID || site.id)
        if (!id) return
        const name = site.Name || site.name || `Site ${id}`
        const status = site.Status ?? site.status ?? 0
        const nodeId = `site-${id}`
        siteMap.set(id, nodeId)
        nodes.push({
          id: nodeId,
          name,
          category: 1,
          symbolSize: 48,
          value: { status },
          itemStyle: { color: getStatusColor(status) },
          label: { show: true },
        })
      })

      const siteHostCount = new Map()
      hosts.forEach((host) => {
        const hostId = Number(host.ID || host.id)
        if (!hostId) return
        const siteId = Number(host.SiteID || host.site_id || 0)
        const monitorId = Number(host.MonitorID || host.monitor_id || 0)
        const name = host.Name || host.name || `Host ${hostId}`
        const status = host.Status ?? host.status ?? 0
        const nodeId = `host-${hostId}`
        
        nodes.push({
          id: nodeId,
          name,
          category: 2,
          symbolSize: 30,
          value: { status, ip: host.IPAddr || host.ip_addr || '' },
          itemStyle: { color: getStatusColor(status) },
          label: { show: true },
        })

        if (siteId && siteMap.has(siteId)) {
          links.push({ source: siteMap.get(siteId), target: nodeId })
          siteHostCount.set(siteId, (siteHostCount.get(siteId) || 0) + 1)
          if (monitorId && monitorMap.has(monitorId)) {
            links.push({ source: monitorMap.get(monitorId), target: siteMap.get(siteId) })
          }
        } else if (monitorId && monitorMap.has(monitorId)) {
          links.push({ source: monitorMap.get(monitorId), target: nodeId })
        }
      })

      // Adjust site node size based on host count
      nodes.forEach((node) => {
        if (!node.id.startsWith('site-')) return
        const siteId = Number(node.id.replace('site-', ''))
        const count = siteHostCount.get(siteId) || 0
        node.symbolSize = Math.min(72, 40 + count * 3)
      })

      return { nodes, links }
    }

    const initChart = (nodes, links) => {
      if (!chartRef.value) return
      if (!chartInstance.value) {
        chartInstance.value = echarts.init(chartRef.value)
      }
      
      chartInstance.value.setOption({
        tooltip: {
          formatter: (params) => {
            if (params.dataType !== 'node') return ''
            const status = params.data?.value?.status ?? 0
            const statusLabel = getStatusInfo(status).label
            const ip = params.data?.value?.ip
            const ipLine = ip ? `<br/>IP: ${ip}` : ''
            return `<strong>${params.data?.name || '-'}</strong><br/>${statusLabel}${ipLine}`
          }
        },
        legend: [{
          data: [t('dashboard.topologyMonitor'), t('dashboard.topologySite'), t('dashboard.topologyHost')],
          top: 8,
        }],
        series: [{
          type: 'graph',
          layout: 'force',
          roam: true,
          data: nodes,
          links,
          categories: [
            { name: t('dashboard.topologyMonitor') },
            { name: t('dashboard.topologySite') },
            { name: t('dashboard.topologyHost') },
          ],
          force: {
            repulsion: 160,
            edgeLength: 120,
            gravity: 0.2,
          },
          label: {
            position: 'right',
            formatter: '{b}',
            fontSize: 12,
          },
          lineStyle: {
            color: 'source',
            width: 1.5,
            opacity: 0.7,
            curveness: 0.2,
          },
        }],
      })
    }

    const loadData = async () => {
      loading.value = true
      error.value = null
      empty.value = false
      
      try {
        const [siteRes, hostRes, monitorRes] = await Promise.all([
          fetchSiteData({ limit: 200 }),
          fetchHostData({ limit: 500 }),
          fetchMonitorData({ limit: 200 }),
        ])
        
        const sites = Array.isArray(siteRes?.data || siteRes) ? (siteRes?.data || siteRes) : []
        const hosts = Array.isArray(hostRes?.data || hostRes) ? (hostRes?.data || hostRes) : []
        const monitors = Array.isArray(monitorRes?.data || monitorRes) ? (monitorRes?.data || monitorRes) : []
        
        const { nodes, links } = buildGraph(sites, hosts, monitors)
        
        if (nodes.length === 0) {
          empty.value = true
          if (chartInstance.value) {
            chartInstance.value.dispose()
            chartInstance.value = null
          }
        } else {
          await nextTick()
          initChart(nodes, links)
        }
      } catch (err) {
        error.value = err.message || t('dashboard.topologyLoadFailed')
      } finally {
        loading.value = false
      }
    }

    const handleRefresh = () => {
      loadData()
    }

    const onResize = () => {
      if (chartInstance.value) chartInstance.value.resize()
    }

    onMounted(() => {
      loadData()
      window.addEventListener('resize', onResize)
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', onResize)
      if (chartInstance.value) {
        chartInstance.value.dispose()
        chartInstance.value = null
      }
    })

    return {
      chartRef,
      loading,
      error,
      empty,
      handleRefresh
    }
  }
})
</script>

<style scoped>
.topology-card {
  margin-top: 20px;
}
.topology-chart {
  width: 100%;
  height: 420px;
}
.topology-alert {
  margin-bottom: 12px;
}
.loading-container {
  text-align: center;
  padding: 60px;
  color: #909399;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
