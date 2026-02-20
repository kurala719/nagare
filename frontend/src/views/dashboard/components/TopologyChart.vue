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
import { getToken } from '@/utils/auth'

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

    const safeId = (val) => {
      if (val === null || val === undefined || val === '') return null
      const n = Number(val)
      return isNaN(n) ? null : n
    }

    const getProp = (obj, ...keys) => {
      if (!obj) return null
      for (const k of keys) {
        if (obj[k] !== undefined && obj[k] !== null) return obj[k]
      }
      return null
    }

    const buildGraph = (groups, hosts, monitors) => {
      const nodes = []
      const links = []
      const nodeSet = new Set()
      
      const impactedGroupIds = new Set()
      const impactedMonitorIds = new Set()
      const impactedHostIds = new Set()

      // 1. Pre-process impacts
      hosts.forEach(h => {
        const s = h.status ?? h.Status ?? 0
        if (s === 2) {
          const hid = safeId(h.id || h.ID)
          const gid = safeId(getProp(h, 'group_id', 'GroupID', 'gid'))
          const mid = safeId(getProp(h, 'm_id', 'monitor_id', 'MonitorID'))
          if (hid !== null) impactedHostIds.add(hid)
          if (gid !== null) impactedGroupIds.add(gid)
          if (mid !== null) impactedMonitorIds.add(mid)
        }
      })

      // 2. Build Monitor Nodes - Category Index 0
      monitors.forEach(m => {
        const id = safeId(m.id || m.ID)
        if (id === null) return
        const nodeId = `m-${id}`
        nodeSet.add(nodeId)
        const status = m.status ?? m.Status ?? 0
        const isImpacted = impactedMonitorIds.has(id)
        nodes.push({
          name: nodeId,
          displayName: m.name || m.Name || `Source ${id}`,
          category: 0,
          symbolSize: 60,
          value: { status, type: 'source', impacted: isImpacted },
          itemStyle: { 
            color: getStatusColor(status),
            borderColor: isImpacted ? '#f56c6c' : undefined,
            borderWidth: isImpacted ? 3 : 0
          },
          label: { show: true, fontWeight: 'bold' }
        })
      })

      // 3. Build Group Nodes - Category Index 1
      groups.forEach(g => {
        const id = safeId(g.id || g.ID)
        if (id === null) return
        const nodeId = `g-${id}`
        nodeSet.add(nodeId)
        
        const status = g.status ?? g.Status ?? 0
        const mid = safeId(getProp(g, 'monitor_id', 'MonitorID', 'm_id', 'MID', 'monitorId'))
        const isImpacted = impactedGroupIds.has(id)
        
        nodes.push({
          name: nodeId,
          displayName: g.name || g.Name || `Group ${id}`,
          category: 1,
          symbolSize: 45,
          value: { status, type: 'group', impacted: isImpacted },
          itemStyle: { 
            color: getStatusColor(status),
            borderColor: isImpacted ? '#f56c6c' : undefined,
            borderWidth: isImpacted ? 2 : 0
          },
          label: { show: true }
        })

        // Link Group -> Monitor
        if (mid !== null && nodeSet.has(`m-${mid}`)) {
          links.push({ source: `m-${mid}`, target: nodeId })
        }
      })

      // 4. Build Host Nodes - Category Index 2
      const groupStats = new Map()
      hosts.forEach(h => {
        const id = safeId(h.id || h.ID)
        if (id === null) return
        const nodeId = `h-${id}`
        nodeSet.add(nodeId)
        
        const status = h.status ?? h.Status ?? 0
        const gid = safeId(getProp(h, 'group_id', 'GroupID', 'gid'))
        const mid = safeId(getProp(h, 'm_id', 'monitor_id', 'MonitorID'))
        const isImpacted = impactedHostIds.has(id)

        nodes.push({
          name: nodeId,
          displayName: h.name || h.Name || `Asset ${id}`,
          category: 2,
          symbolSize: 25,
          value: { 
            status, type: 'asset', impacted: isImpacted,
            ip: h.ip_addr || h.IPAddr || ''
          },
          itemStyle: { color: getStatusColor(status) },
          label: { show: true, fontSize: 10 }
        })

        if (gid !== null && nodeSet.has(`g-${gid}`)) {
          links.push({ source: `g-${gid}`, target: nodeId })
          groupStats.set(gid, (groupStats.get(gid) || 0) + 1)
        } else if (mid !== null && nodeSet.has(`m-${mid}`)) {
          links.push({ source: `m-${mid}`, target: nodeId })
        }
      })

      // Rescaling
      nodes.forEach(n => {
        if (n.value.type === 'group') {
          const id = Number(n.name.replace('g-', ''))
          const count = groupStats.get(id) || 0
          n.symbolSize = Math.min(80, 45 + Math.sqrt(count) * 6)
        }
      })

      return { nodes, links }
    }

    const initChart = (nodes, links) => {
      if (!chartRef.value) return
      
      // Ensure we have a valid instance
      if (!chartInstance.value) {
        chartInstance.value = echarts.init(chartRef.value)
      } else {
        chartInstance.value.clear()
      }
      
      const catNames = [
        t('dashboard.topologyMonitor'),
        t('dashboard.topologySite'),
        t('dashboard.topologyHost')
      ]

      

      // Final validation of links to prevent ECharts crashes
      const validLinks = links.filter(l => {
        const sourceExists = nodes.some(n => n.name === l.source)
        const targetExists = nodes.some(n => n.name === l.target)
        return sourceExists && targetExists
      })

      chartInstance.value.setOption({
        tooltip: {
          formatter: (params) => {
            if (params.dataType !== 'node') return ''
            const v = params.data?.value || {}
            const catIdx = typeof params.data?.category === 'number' ? params.data.category : 0
            const typeName = catNames[catIdx] || ''
            const statusLabel = getStatusInfo(v.status ?? 0).label
            const ipLine = v.ip ? `<br/>IP: ${v.ip}` : ''
            const impactLine = v.impacted ? `<br/><span style="color:#f56c6c">${t('dashboard.impactedLabel')}</span>` : ''
            return `<div style="padding:4px"><strong>${params.data.displayName || '-'}</strong> (${typeName})<br/>Status: ${statusLabel}${ipLine}${impactLine}</div>`
          }
        },
        legend: {
          data: catNames,
          top: 8,
          textStyle: { color: 'var(--text-strong)' }
        },
        series: [{
          type: 'graph',
          layout: 'force',
          roam: true,
          draggable: true,
          data: nodes.map(n => ({
            ...n,
            category: typeof n.category === 'number' ? n.category : 0
          })),
          links: validLinks,
          categories: catNames.map(name => ({ name })),
          edgeSymbol: ['none', 'arrow'],
          edgeSymbolSize: [0, 8],
          force: {
            repulsion: 450,
            edgeLength: 150,
            gravity: 0.05,
            layoutAnimation: true
          },
          label: {
            position: 'right',
            formatter: (params) => params.data.displayName,
            fontSize: 11,
            color: 'var(--text-strong)'
          },
          lineStyle: {
            color: 'source',
            width: 1.5,
            opacity: 0.4,
            curveness: 0.1
          },
          emphasis: {
            focus: 'adjacency',
            lineStyle: { width: 4 }
          }
        }]
      }, { notMerge: true })
    }

    const loadData = async () => {
      if (!getToken()) return
      
      loading.value = true
      error.value = null
      empty.value = false
      
      if (chartInstance.value) {
        chartInstance.value.dispose()
        chartInstance.value = null
      }
      
      try {
        const [groupRes, hostRes, monitorRes] = await Promise.all([
          fetchGroupData({ limit: 500 }),
          fetchHostData({ limit: 1000 }),
          fetchMonitorData({ limit: 200 }),
        ])

        const extractData = (res) => {
          if (!res) return []
          if (res.success && res.data !== undefined) {
            const data = res.data
            if (Array.isArray(data)) return data
            if (Array.isArray(data.items)) return data.items
            return []
          }
          if (Array.isArray(res)) return res
          return []
        }

        const groups = extractData(groupRes)
        const hosts = extractData(hostRes)
        const monitors = extractData(monitorRes)

        const { nodes, links } = buildGraph(groups, hosts, monitors)
        
        if (nodes.length === 0) {
          empty.value = true
        } else {
          loading.value = false
          await nextTick()
          initChart(nodes, links)
        }
      } catch (err) {
        console.error('TopologyChart error:', err)
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
