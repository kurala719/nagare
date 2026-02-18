<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('logs.title') }}</h1>
      <p class="page-subtitle">{{ totalLogs }} {{ $t('logs.title') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('logs.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="severityFilter" :placeholder="$t('logs.filterLevel')" style="width: 140px">
          <el-option :label="$t('logs.filterAll')" value="all" />
          <el-option :label="$t('logs.levelInfo')" :value="0" />
          <el-option :label="$t('logs.levelWarn')" :value="1" />
          <el-option :label="$t('logs.levelError')" :value="2" />
        </el-select>
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

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p>{{ $t('logs.loading') }}</p>
    </div>

  <el-alert
    v-if="error && !loading"
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
    v-if="!loading && !error && logs.length === 0"
    :description="$t('logs.noLogs')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="logs-content">
    <el-table
      v-if="logs.length > 0"
      :data="logs"
      border
      @sort-change="onSortChange"
    >
      <el-table-column prop="created_at" :label="$t('logs.createdAt')" width="180" align="center" sortable="custom" />
      <el-table-column prop="severity" :label="$t('logs.level')" width="120" align="center" sortable="custom">
        <template #default="{ row }">
          <el-tag :type="severityTag(row.severity)" size="small" effect="dark">{{ severityLabel(row.severity) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" :label="$t('logs.message')" min-width="260" sortable="custom" />
      <el-table-column prop="ip" :label="$t('logs.ip')" width="140" align="center" sortable="custom" />
      <el-table-column prop="context" :label="$t('logs.context')" min-width="240" show-overflow-tooltip sortable="custom" />
    </el-table>
  </div>
  <div v-if="!loading && !error && totalLogs > 0" class="logs-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalLogs"
    />
  </div>
  </div>
</template>

<script lang="ts">
import { ElMessage } from 'element-plus'
import { markRaw } from 'vue'
import { Loading, Search, Refresh } from '@element-plus/icons-vue'
import { fetchSystemLogs, fetchServiceLogs } from '@/api/logs'
import { getUserPrivileges } from '@/utils/auth'

export default {
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
      pageSize: 20,
      currentPage: 1,
      totalLogs: 0,
      sortBy: '',
      sortOrder: '',
      // Icons for template usage
      Loading: markRaw(Loading),
      Search: markRaw(Search),
      Refresh: markRaw(Refresh)
    }
  },
  computed: {
    canViewSystem() {
      return getUserPrivileges() >= 3
    },
  },
  watch: {
    activeTab() {
      this.currentPage = 1
      this.loadLogs(true)
    },
    search() {
      this.currentPage = 1
      this.loadLogs(true)
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
  methods: {
    onSortChange({ prop, order }) {
      if (!prop || !order) {
        this.sortBy = ''
        this.sortOrder = ''
      } else {
        this.sortBy = prop
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc'
      }
      this.currentPage = 1
      this.loadLogs(true)
    },
    async loadLogs(reset = false) {
      if (reset) {
        this.logs = []
      }
      this.loading = reset
      this.error = null
      try {
        const params = {
          q: this.search || undefined,
          severity: this.severityFilter === 'all' ? undefined : this.severityFilter,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
          with_total: 1,
        }
        const response = this.activeTab === 'system'
          ? await fetchSystemLogs(params)
          : await fetchServiceLogs(params)
        const data = Array.isArray(response)
          ? response
          : (response.data?.items || response.items || response.data || response.logs || [])
        const total = response?.data?.total ?? response?.total ?? data.length
        const mapped = data.map((l) => ({
          id: l.ID || l.id || 0,
          type: l.Type || l.type || '',
          severity: this.coerceSeverity(l.Severity ?? l.severity),
          message: l.Message || l.message || '',
          context: l.Context || l.context || '',
          ip: l.IP || l.ip || '',
          created_at: l.CreatedAt || l.created_at || '',
        }))
        this.logs = mapped
        this.totalLogs = Number.isFinite(total) ? total : mapped.length
      } catch (err) {
        this.error = err.message || this.$t('logs.loadFailed')
        ElMessage.error(this.error)
      } finally {
        this.loading = false
      }
    },
    severityTag(severity) {
      switch (this.coerceSeverity(severity)) {
        case 2:
          return 'danger'
        case 1:
          return 'warning'
        default:
          return 'info'
      }
    },
    severityLabel(severity) {
      switch (this.coerceSeverity(severity)) {
        case 2:
          return this.$t('logs.levelError')
        case 1:
          return this.$t('logs.levelWarn')
        default:
          return this.$t('logs.levelInfo')
      }
    },
    coerceSeverity(value) {
      if (value === null || value === undefined || value === '') return 0
      if (typeof value === 'number') return value
      const text = String(value).toLowerCase()
      if (!Number.isNaN(Number(text))) return Number(text)
      if (text === 'error') return 2
      if (text === 'warn' || text === 'warning') return 1
      return 0
    },
  },
}
</script>

<style scoped>
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

:deep(.el-tabs__nav-wrap::after) {
  display: none;
}

:deep(.el-tabs__item) {
  font-weight: 600;
  font-size: 15px;
}
</style>
