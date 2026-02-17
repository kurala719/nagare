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
import { fetchGroupData } from '@/api/groups'
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

    const buildGraph = (groups, hosts, monitors) => {
      const nodes = []
      const links = []
      const groupMap = new Map()
      const monitorMap = new Map()
      const impactedGroupIds = new Set()
      const impactedHostIds = new Set()

      const buildLink = (source, target, impacted) => ({
        source,
        target,
        lineStyle: impacted
          ? { color: '#f56c6c', width: 2.5, opacity: 0.9 }
          : undefined,
      })
      const linkIndex = new Map()
      const pushLink = (source, target, impacted) => {
        const key = `${source}|${target}`
        const existing = linkIndex.get(key)
        if (existing !== undefined) {
          if (impacted) {
            links[existing].lineStyle = { color: '#f56c6c', width: 2.5, opacity: 0.9 }
          }
          return
        }
        links.push(buildLink(source, target, impacted))
        linkIndex.set(key, links.length - 1)
      }

      hosts.forEach((host) => {
        const hostId = Number(host.ID || host.id)
        if (!hostId) return
        const groupId = Number(host.GroupID || host.group_id || 0)
        const status = host.Status ?? host.status ?? 0
        if (status === 2) {
          impactedHostIds.add(hostId)
          if (groupId) impactedGroupIds.add(groupId)
        }
      })

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

      groups.forEach((group) => {
        const id = Number(group.ID || group.id)
        if (!id) return
        const name = group.Name || group.name || `Group ${id}`
        const status = group.Status ?? group.status ?? 0
        const nodeId = `group-${id}`
        groupMap.set(id, nodeId)
        const impacted = impactedGroupIds.has(id)
        nodes.push({
          id: nodeId,
          name,
          category: 1,
          symbolSize: 48,
          value: { status, impacted },
          itemStyle: {
            color: getStatusColor(status),
            borderColor: impacted ? '#f56c6c' : undefined,
            borderWidth: impacted ? 3 : 0,
            shadowBlur: impacted ? 8 : 0,
            shadowColor: impacted ? '#f56c6c' : undefined,
          },
          label: { show: true },
        })
      })

      const groupHostCount = new Map()
      hosts.forEach((host) => {
        const hostId = Number(host.ID || host.id)
        if (!hostId) return
        const groupId = Number(host.GroupID || host.group_id || 0)
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

        if (groupId && groupMap.has(groupId)) {
          pushLink(groupMap.get(groupId), nodeId, impactedHostIds.has(hostId))
          groupHostCount.set(groupId, (groupHostCount.get(groupId) || 0) + 1)
          if (monitorId && monitorMap.has(monitorId)) {
            pushLink(monitorMap.get(monitorId), groupMap.get(groupId), impactedGroupIds.has(groupId))
          }
        } else if (monitorId && monitorMap.has(monitorId)) {
          pushLink(monitorMap.get(monitorId), nodeId, impactedHostIds.has(hostId))
        }
      })

      // Adjust group node size based on host count
      nodes.forEach((node) => {
        if (!node.id.startsWith('group-')) return
        const groupId = Number(node.id.replace('group-', ''))
        const count = groupHostCount.get(groupId) || 0
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
            const impacted = params.data?.value?.impacted
            const impactedLine = impacted ? `<br/>${t('dashboard.impactedLabel')}` : ''
            return `<strong>${params.data?.name || '-'}</strong><br/>${statusLabel}${ipLine}${impactedLine}`
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
        const [groupRes, hostRes, monitorRes] = await Promise.all([
          fetchGroupData({ limit: 200 }),
          fetchHostData({ limit: 500 }),
          fetchMonitorData({ limit: 200 }),
        ])

        console.log('TopologyChart raw responses:', { groupRes, hostRes, monitorRes })

        // Backend always returns {success: true, data: ...}
        // For arrays: data is array directly
        // For paginated: data is {items: [...], total: N}
        const extractData = (res) => {
          if (!res) return []
          // Check if response has success/data structure
          if (res.success && res.data !== undefined) {
            const data = res.data
            // If data is array, return it
            if (Array.isArray(data)) return data
            // If data.items is array (paginated), return items
            if (Array.isArray(data.items)) return data.items
            return []
          }
          // Fallback: if response itself is array
          if (Array.isArray(res)) return res
          return []
        }

        const groups = extractData(groupRes)
        const hosts = extractData(hostRes)
        const monitors = extractData(monitorRes)

        console.log('TopologyChart extracted data:', { groups, hosts, monitors })

        const { nodes, links } = buildGraph(groups, hosts, monitors)
        
        console.log('TopologyChart graph:', { nodeCount: nodes.length, linkCount: links.length })
        
        if (nodes.length === 0) {
          empty.value = true
          if (chartInstance.value) {
            chartInstance.value.dispose()
            chartInstance.value = null
          }
        } else {
          loading.value = false
          await nextTick()
          initChart(nodes, links)
        }
      } catch (err) {
        console.error('TopologyChart load error:', err)
        const msg = err.message || t('dashboard.topologyLoadFailed')
        if (err.message && (err.message.includes('401') || err.message.includes('403') || err.message.includes('Unauthorized') || err.message.includes('Forbidden'))) {
          error.value = t('common.sessionExpired') || 'Please log in to view topology'
        } else {
          error.value = msg
        }
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
