<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="header-main">
        <h1 class="page-title">{{ $t('auditLog.title') || 'Security Audit Log' }}</h1>
        <p class="page-subtitle">{{ total }} operations tracked across the system</p>
      </div>
      <div class="header-actions">
        <el-button-group>
          <el-button :icon="Download" @click="exportCSV">Export</el-button>
          <el-button :icon="Refresh" @click="fetchAuditLogs" :loading="loading">Refresh</el-button>
        </el-button-group>
      </div>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input
          v-model="searchQuery"
          :placeholder="$t('common.search') || 'Search operations...'"
          clearable
          style="width: 300px"
          @input="handleSearch"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>
    </div>

    <div class="timeline-wrapper">
      <el-empty v-if="!loading && auditLogs.length === 0" description="No security records found" />
      
      <div v-if="loading && auditLogs.length === 0" class="loading-state">
        <el-icon class="is-loading" size="40"><Loading /></el-icon>
      </div>

      <el-timeline v-else class="audit-timeline">
        <el-timeline-item
          v-for="log in auditLogs"
          :key="log.id"
          :timestamp="formatTime(log.created_at)"
          :type="getLogType(log.method)"
          placement="top"
          class="animate-slide-up"
        >
          <el-card class="audit-card" shadow="hover">
            <div class="log-item">
              <div class="log-header">
                <el-tag :type="getLogType(log.method)" size="small" effect="dark" class="method-tag">
                  {{ log.method }}
                </el-tag>
                <span class="log-action">{{ log.action }}</span>
                <div class="log-user">
                  <el-avatar :size="24" class="user-avatar">{{ (log.username || 'S')[0].toUpperCase() }}</el-avatar>
                  <span class="username">{{ log.username || 'System' }}</span>
                </div>
              </div>
              
              <div class="log-details-grid">
                <div class="detail-cell">
                  <span class="cell-label">Endpoint</span>
                  <code class="cell-value">{{ log.path }}</code>
                </div>
                <div class="detail-cell">
                  <span class="cell-label">Client IP</span>
                  <span class="cell-value">{{ log.ip }}</span>
                </div>
                <div class="detail-cell">
                  <span class="cell-label">Outcome</span>
                  <el-tag :type="log.status >= 400 ? 'danger' : 'success'" size="small" class="status-tag">
                    HTTP {{ log.status }}
                  </el-tag>
                </div>
                <div class="detail-cell">
                  <span class="cell-label">Latency</span>
                  <span class="cell-value">{{ (log.latency / 1000).toFixed(2) }}ms</span>
                </div>
              </div>
            </div>
          </el-card>
        </el-timeline-item>
      </el-timeline>

      <div class="pagination-footer">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          background
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { Search, Refresh, User, Download, Loading } from '@element-plus/icons-vue'

const auditLogs = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')
const loading = ref(false)

const fetchAuditLogs = async () => {
  loading.value = true
  try {
    const res = await request.get('/api/v1/audit-logs', {
      params: {
        limit: pageSize.value,
        offset: (currentPage.value - 1) * pageSize.value,
        q: searchQuery.value
      }
    })
    if (res.success) {
      auditLogs.value = res.data.items || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
  } finally {
    loading.value = false
  }
}

const getLogType = (method) => {
  switch (method) {
    case 'POST': return 'success'
    case 'PUT':
    case 'PATCH': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString()
}

const handleSearch = () => {
  currentPage.value = 1
  fetchAuditLogs()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  fetchAuditLogs()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchAuditLogs()
}

const exportCSV = () => {
  const header = ['Time', 'User', 'Action', 'Method', 'Path', 'IP', 'Status', 'Latency']
  const rows = auditLogs.value.map(log => [
    log.created_at,
    log.username || 'System',
    log.action,
    log.method,
    log.path,
    log.ip,
    log.status,
    `${(log.latency / 1000).toFixed(2)}ms`
  ])
  
  const csvContent = [
    header.join(','),
    ...rows.map(r => r.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
  ].join('\n')
  
  const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `nagare_audit_logs_${new Date().toISOString()}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onMounted(() => {
  fetchAuditLogs()
})
</script>

<style scoped>
.header-main {
  flex: 1;
}

.timeline-wrapper {
  padding: 8px 0;
}

.audit-timeline {
  max-width: 1000px;
  margin: 0 auto;
}

.audit-card {
  border-radius: 16px;
  border: 1px solid var(--border-1);
}

.log-item {
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
  border-bottom: 1px solid var(--border-1);
  padding-bottom: 12px;
}

.method-tag {
  min-width: 60px;
  text-align: center;
  font-weight: 800;
}

.log-action {
  font-weight: 700;
  font-size: 16px;
  color: var(--text-strong);
}

.log-user {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-avatar {
  background: var(--brand-500);
  font-size: 12px;
  font-weight: bold;
}

.username {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-muted);
}

.log-details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.detail-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cell-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-muted);
  font-weight: 700;
}

.cell-value {
  font-size: 14px;
  color: var(--text-strong);
  word-break: break-all;
}

code.cell-value {
  background: var(--surface-2);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
}

.pagination-footer {
  margin-top: 40px;
  display: flex;
  justify-content: center;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 100px 0;
}

:deep(.el-timeline-item__node) {
  box-shadow: 0 0 0 4px var(--surface-1);
}
</style>
