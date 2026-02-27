<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="header-main">
        <h1 class="page-title">{{ $t('logs.title') }}</h1>
        <p class="page-subtitle">{{ totalLogs }} {{ $t('logs.title') }} recorded</p>
      </div>
      <div class="header-actions">
        <el-button-group>
          <el-button :icon="Download" @click="exportCSV">{{ $t('logs.export') }}</el-button>
          <el-button v-if="canViewSystem" type="danger" :icon="Delete" @click="handleClearLogs">{{ $t('logs.clear') }}</el-button>
        </el-button-group>
      </div>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('logs.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="severityFilter" :placeholder="$t('logs.filterLevel')" style="width: 140px">
          <el-option :label="$t('logs.filterAll')" value="all" />
          <el-option :label="$t('logs.levelFatal') || 'Fatal/Disaster'" :value="5" />
          <el-option :label="$t('logs.levelCritical') || 'Critical'" :value="4" />
          <el-option :label="$t('logs.levelAverage') || 'Average'" :value="3" />
          <el-option :label="$t('logs.levelWarn') || 'Warning'" :value="2" />
          <el-option :label="$t('logs.levelInfo') || 'Info'" :value="1" />
        </el-select>

        <el-divider direction="vertical" />

        <div class="auto-refresh-control">
          <span class="control-label">{{ $t('logs.autoRefresh') }}</span>
          <el-switch v-model="autoRefresh" size="small" @change="toggleAutoRefresh" />
        </div>
      </div>

      <div class="action-group">
        <el-button type="primary" :icon="Refresh" @click="loadLogs(true)">
          {{ $t('logs.refresh') }}
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="logs-tabs">
      <el-tab-pane :label="$t('logs.serviceTitle')" name="service" />
      <el-tab-pane v-if="canViewSystem" :label="$t('logs.systemTitle')" name="system" />
    </el-tabs>

    <div v-if="loading && logs.length === 0" class="loading-state">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p>{{ $t('logs.loading') }}</p>
    </div>

    <el-alert
      v-else-if="error && !loading"
      :title="error"
      type="error"
      show-icon
      style="margin: 20px;"
      :closable="false"
    >
      <template #default>
        <el-button size="small" @click="loadLogs(true)">{{ $t('logs.retry') }}</el-button>
      </template>
    </el-alert>

    <el-empty
      v-else-if="!loading && logs.length === 0"
      :description="$t('logs.noLogs')"
      style="margin: 40px;"
    />

    <div v-else class="logs-content">
      <el-table
        :data="logs"
        border
        stripe
        highlight-current-row
        @sort-change="onSortChange"
        @row-click="showLogDetail"
        class="animate-fade-in"
      >
        <el-table-column prop="created_at" :label="$t('logs.createdAt')" width="180" align="center" sortable="custom" />
        <el-table-column prop="severity" :label="$t('logs.level')" width="120" align="center" sortable="custom">
          <template #default="{ row }">
            <el-tag :type="severityTag(row.severity)" size="small" effect="dark">
              {{ severityLabel(row.severity) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" :label="$t('logs.message')" min-width="300" sortable="custom">
          <template #default="{ row }">
            <span class="log-message-text">{{ row.message }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ip" :label="$t('logs.ip')" width="140" align="center" sortable="custom" />
        <el-table-column prop="context" :label="$t('logs.context')" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <code class="context-preview">{{ row.context }}</code>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="!loading && totalLogs > 0" class="logs-pagination">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next"
        :page-sizes="[20, 50, 100, 200]"
        v-model:page-size="pageSize"
        v-model:current-page="currentPage"
        :total="totalLogs"
      />
    </div>

    <!-- Log Details Drawer -->
    <el-drawer
      v-model="detailVisible"
      :title="$t('logs.details')"
      size="50%"
      destroy-on-close
    >
      <div v-if="selectedLog" class="log-detail-view">
        <div class="detail-header">
          <el-tag :type="severityTag(selectedLog.severity)" effect="dark" size="large">
            {{ severityLabel(selectedLog.severity) }}
          </el-tag>
          <span class="detail-time">{{ selectedLog.created_at }}</span>
        </div>
        
        <div class="detail-section">
          <h4>{{ $t('logs.message') }}</h4>
          <div class="message-box">{{ selectedLog.message }}</div>
        </div>

        <div class="detail-section">
          <h4>{{ $t('logs.ip') }}</h4>
          <p>{{ selectedLog.ip || 'N/A' }}</p>
        </div>

        <div class="detail-section" v-if="selectedLog.context">
          <h4>{{ $t('logs.formattedContext') }}</h4>
          <div class="context-container">
            <pre class="pretty-json"><code>{{ formatJson(selectedLog.context) }}</code></pre>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import { defineComponent, markRaw } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, Search, Refresh, Download, Delete, View } from '@element-plus/icons-vue'
import { fetchSystemLogs, fetchServiceLogs } from '@/api/logs'
import { getUserPrivileges } from '@/utils/auth'
import request from '@/utils/request'

export default defineComponent({
  name: 'Log',
  components: {
    Loading,
    Search,
    Refresh
  },
  data() {
    return {
      logs: [],
      loading: false,
      error: null,
      search: '',
      severityFilter: 'all',
      activeTab: 'service',
      pageSize: 50,
      currentPage: 1,
      totalLogs: 0,
      sortBy: 'created_at',
      sortOrder: 'desc',
      autoRefresh: false,
      refreshTimer: null,
      detailVisible: false,
      selectedLog: null,
      // Icons
      Loading: markRaw(Loading),
      Search: markRaw(Search),
      Refresh: markRaw(Refresh),
      Download: markRaw(Download),
      Delete: markRaw(Delete),
      View: markRaw(View)
    }
  },
  computed: {
    canViewSystem() {
      return getUserPrivileges() >= 3
    }
  },
  watch: {
    activeTab() {
      this.currentPage = 1
      this.loadLogs(true)
    },
    search() {
      this.debouncedSearch()
    },
    severityFilter() {
      this.currentPage = 1
      this.loadLogs(true)
    },
    pageSize() {
      this.currentPage = 1
      this.loadLogs(true)
    },
    currentPage() {
      this.loadLogs()
    },
  },
  created() {
    if (!this.canViewSystem) {
      this.activeTab = 'service'
    }
    this.loadLogs(true)
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  },
  methods: {
    debouncedSearch() {
      clearTimeout(this._searchTimer)
      this._searchTimer = setTimeout(() => {
        this.currentPage = 1
        this.loadLogs(true)
      }, 500)
    },
    onSortChange({ prop, order }) {
      if (!prop || !order) {
        this.sortBy = 'created_at'
        this.sortOrder = 'desc'
      } else {
        this.sortBy = prop
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc'
      }
      this.currentPage = 1
      this.loadLogs(true)
    },
    async loadLogs(reset = false) {
      if (reset && !this.autoRefresh) {
        this.loading = true
      }
      this.error = null
      try {
        const params = {
          q: this.search || undefined,
          severity: this.severityFilter === 'all' ? undefined : this.severityFilter,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy,
          order: this.sortOrder,
          with_total: 1,
        }
        const response = this.activeTab === 'system'
          ? await fetchSystemLogs(params)
          : await fetchServiceLogs(params)
        
        const data = Array.isArray(response)
          ? response
          : (response.data?.items || response.items || response.data || response.logs || [])
        const total = response?.data?.total ?? response?.total ?? data.length
        
        this.logs = data.map((l) => ({
          id: l.ID || l.id || 0,
          type: l.Type || l.type || '',
          severity: this.coerceSeverity(l.Severity ?? l.severity),
          message: l.Message || l.message || '',
          context: l.Context || l.context || '',
          ip: l.IP || l.ip || '',
          created_at: l.CreatedAt || l.created_at || '',
        }))
        this.totalLogs = Number.isFinite(total) ? total : this.logs.length
      } catch (err) {
        if (!this.autoRefresh) {
          this.error = err.message || this.$t('logs.loadFailed')
          ElMessage.error(this.error)
        }
      } finally {
        this.loading = false
      }
    },
    toggleAutoRefresh(val) {
      if (val) {
        this.startAutoRefresh()
      } else {
        this.stopAutoRefresh()
      }
    },
    startAutoRefresh() {
      this.stopAutoRefresh()
      this.refreshTimer = setInterval(() => {
        this.loadLogs(false)
      }, 5000)
    },
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    },
    showLogDetail(row) {
      this.selectedLog = row
      this.detailVisible = true
    },
    formatJson(str) {
      try {
        if (!str) return ''
        const obj = JSON.parse(str)
        return JSON.stringify(obj, null, 2)
      } catch {
        return str
      }
    },
    severityTag(severity) {
      switch (severity) {
        case 5:
        case 4: return 'danger'
        case 3:
        case 2: return 'warning'
        case 1:
        case 0: return 'info'
        default: return 'info'
      }
    },
    severityLabel(severity) {
      switch (severity) {
        case 5: return this.$t('logs.levelFatal') || 'Fatal/Disaster'
        case 4: return this.$t('logs.levelCritical') || 'Critical'
        case 3: return this.$t('logs.levelAverage') || 'Average'
        case 2: return this.$t('logs.levelWarn') || 'Warning'
        case 1: return this.$t('logs.levelInfo') || 'Information'
        case 0: return this.$t('alerts.severityNotClassified') || 'Not Classified'
        default: return 'Unknown'
      }
    },
    coerceSeverity(value) {
      if (value === null || value === undefined || value === '') return 1
      if (typeof value === 'number') return value
      const text = String(value).toLowerCase()
      if (!Number.isNaN(Number(text))) return Number(text)
      if (text === 'disaster' || text === 'fatal') return 5
      if (text === 'critical' || text === 'high') return 4
      if (text === 'average') return 3
      if (text === 'warn' || text === 'warning') return 2
      if (text === 'info' || text === 'information') return 1
      if (text === 'not_classified' || text === 'none') return 0
      return 1
    },
    exportCSV() {
      const header = ['Time', 'Level', 'Message', 'IP', 'Context']
      const rows = this.logs.map((l) => [
        l.created_at,
        this.severityLabel(l.severity),
        l.message,
        l.ip,
        l.context
      ])
      
      const csvContent = [
        header.join(','),
        ...rows.map(r => r.map(cell => `"${String(cell).replace(/"/g, '""')}"`).join(','))
      ].join('\n')
      
      const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      const url = URL.createObjectURL(blob)
      link.setAttribute('href', url)
      link.setAttribute('download', `nagare_logs_${this.activeTab}_${new Date().toISOString()}.csv`)
      link.style.visibility = 'hidden'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    async handleClearLogs() {
      try {
        await ElMessageBox.confirm(
          'This will permanently delete ALL logs in this tab. Continue?',
          'Warning',
          { type: 'warning', confirmButtonText: 'Clear All', confirmButtonClass: 'el-button--danger' }
        )
        
        await request({
          url: `/system/logs/${this.activeTab}`,
          method: 'DELETE'
        })
        
        ElMessage.success('Logs cleared successfully')
        this.loadLogs(true)
      } catch (e) {
        // Canceled
      }
    }
  },
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.logs-content {
  margin-top: 8px;
}

.logs-tabs {
  margin-bottom: 16px;
}

.logs-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.loading-state {
  text-align: center;
  padding: 60px;
}

.auto-refresh-control {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px;
}

.control-label {
  font-size: 13px;
  color: var(--text-muted);
}

.context-preview {
  font-family: monospace;
  font-size: 12px;
  color: var(--text-muted);
  background: var(--surface-2);
  padding: 2px 6px;
  border-radius: 4px;
}

.log-message-text {
  font-weight: 500;
}

/* Detail Drawer Styles */
.log-detail-view {
  padding: 0 20px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.detail-time {
  color: var(--text-muted);
  font-family: monospace;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.message-box {
  font-size: 18px;
  font-weight: 600;
  line-height: 1.5;
  color: var(--text-strong);
}

.context-container {
  background: var(--surface-3);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid var(--border-1);
}

.pretty-json {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: var(--brand-600);
}

:deep(.el-tabs__nav-wrap::after) {
  display: none;
}

:deep(.el-tabs__item) {
  font-weight: 600;
  font-size: 15px;
}

:deep(.el-table__row) {
  cursor: pointer;
}
</style>
