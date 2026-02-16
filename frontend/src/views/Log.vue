<template>
  <div class="logs-toolbar">
    <div class="logs-filters">
      <span class="filter-label">{{ $t('logs.search') }}</span>
      <el-input v-model="search" :placeholder="$t('logs.search')" clearable class="logs-search" />
      <span class="filter-label">{{ $t('logs.filterLevel') }}</span>
      <el-select v-model="severityFilter" :placeholder="$t('logs.filterLevel')" class="logs-filter">
        <el-option :label="$t('logs.filterAll')" value="all" />
        <el-option :label="$t('logs.levelInfo')" :value="0" />
        <el-option :label="$t('logs.levelWarn')" :value="1" />
        <el-option :label="$t('logs.levelError')" :value="2" />
      </el-select>
      <span class="filter-label">{{ $t('common.sort') }}</span>
      <el-select v-model="sortKey" class="logs-filter">
        <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
        <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
        <el-option :label="$t('common.sortNameAsc')" value="message_asc" />
        <el-option :label="$t('common.sortNameDesc')" value="message_desc" />
        <el-option :label="$t('common.sortStatusAsc')" value="severity_asc" />
        <el-option :label="$t('common.sortStatusDesc')" value="severity_desc" />
      </el-select>
    </div>
    <el-button type="primary" @click="loadLogs(true)">
      {{ $t('logs.refresh') }}
    </el-button>
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

  <div
    v-if="!loading && !error"
    class="logs-scroll"
  >
    <el-table
      v-if="logs.length > 0"
      :data="logs"
      border
      style="margin: 20px; width: calc(100% - 40px);"
    >
    <el-table-column prop="created_at" :label="$t('logs.createdAt')" min-width="180" />
    <el-table-column prop="severity" :label="$t('logs.level')" min-width="120">
      <template #default="{ row }">
        <el-tag :type="severityTag(row.severity)">{{ severityLabel(row.severity) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="message" :label="$t('logs.message')" min-width="260" />
    <el-table-column prop="ip" :label="$t('logs.ip')" min-width="140" />
    <el-table-column prop="context" :label="$t('logs.context')" min-width="240" show-overflow-tooltip />
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
</template>

<script lang="ts">
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { fetchSystemLogs, fetchServiceLogs } from '@/api/logs'
import { getUserPrivileges } from '@/utils/auth'

export default {
  name: 'Log',
  components: { Loading },
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
      sortKey: 'created_desc',
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
    sortKey() {
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
    async loadLogs(reset = false) {
      if (reset) {
        this.logs = []
      }
      this.loading = reset
      this.error = null
      try {
        const { sortBy, sortOrder } = this.parseSortKey(this.sortKey)
        const params = {
          q: this.search || undefined,
          severity: this.severityFilter === 'all' ? undefined : this.severityFilter,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: sortBy,
          order: sortOrder,
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
    parseSortKey(key) {
      switch (key) {
        case 'message_asc':
          return { sortBy: 'message', sortOrder: 'asc' }
        case 'message_desc':
          return { sortBy: 'message', sortOrder: 'desc' }
        case 'severity_asc':
          return { sortBy: 'severity', sortOrder: 'asc' }
        case 'severity_desc':
          return { sortBy: 'severity', sortOrder: 'desc' }
        case 'updated_desc':
          return { sortBy: 'updated_at', sortOrder: 'desc' }
        case 'created_desc':
        default:
          return { sortBy: 'created_at', sortOrder: 'desc' }
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
.logs-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 16px 20px 0;
}

.logs-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.logs-search {
  width: 240px;
}

.logs-filter {
  min-width: 160px;
}

.logs-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 16px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}

.logs-tabs {
  margin: 0 20px;
}
</style>
