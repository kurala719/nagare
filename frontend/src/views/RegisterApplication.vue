<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('registerApplications.title') }}</h1>
      <p class="page-subtitle">{{ applications.length }} {{ $t('registerApplications.title') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange" style="margin-right: 20px">
          <el-tab-pane label="Registration" name="register" />
          <el-tab-pane label="Password Reset" name="reset" />
        </el-tabs>
        
        <el-input v-model="search" :placeholder="$t('registerApplications.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('registerApplications.filterStatus')" style="width: 140px">
          <el-option :label="$t('registerApplications.filterAll')" value="all" />
          <el-option :label="$t('registerApplications.statusPending')" :value="0" />
          <el-option :label="$t('registerApplications.statusApproved')" :value="1" />
          <el-option :label="$t('registerApplications.statusRejected')" :value="2" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button type="primary" :icon="Refresh" @click="loadApplications(true)">
          {{ $t('registerApplications.refresh') }}
        </el-button>
        <el-dropdown trigger="click" v-if="selectedCount > 0" style="margin-left: 8px">
          <el-button>
            {{ $t('common.selectedCount', { count: selectedCount }) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="Check" @click="approveSelected" style="color: var(--el-color-success)">{{ $t('registerApplications.approveSelected') }}</el-dropdown-item>
              <el-dropdown-item :icon="Close" @click="openBulkReject" style="color: var(--el-color-danger)">{{ $t('registerApplications.rejectSelected') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p>{{ $t('registerApplications.loading') }}</p>
    </div>

    <el-alert
      v-if="error && !loading"
      :title="error"
      type="error"
      show-icon
      class="register-app-error"
      :closable="false"
    >
      <template #default>
        <el-button size="small" @click="loadApplications(true)">{{ $t('registerApplications.retry') }}</el-button>
      </template>
    </el-alert>

    <el-empty
      v-if="!loading && !error && filteredApplications.length === 0"
      :description="$t('registerApplications.noApplications')"
      class="register-app-empty"
    />

  <div v-if="!loading && !error" class="register-app-content">
    <div
      class="register-app-scroll"
      v-infinite-scroll="loadMoreApplications"
      :infinite-scroll-disabled="loadingMore || !hasMore"
      :infinite-scroll-distance="120"
    >
      <el-table
        v-if="filteredApplications.length > 0"
        :data="filteredApplications"
        border
        ref="applicationsTableRef"
        row-key="id"
        @selection-change="onSelectionChange"
        @sort-change="onSortChange"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column prop="id" :label="$t('registerApplications.id')" width="90" align="center" sortable="custom" />
        <el-table-column prop="username" :label="$t('registerApplications.username')" min-width="160" sortable="custom" />
        <el-table-column prop="status" :label="$t('registerApplications.status')" width="140" align="center" sortable="custom">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small" effect="dark">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" :label="$t('registerApplications.reason')" min-width="220" show-overflow-tooltip sortable="custom" />
        <el-table-column prop="approved_by" :label="$t('registerApplications.approvedBy')" width="140" align="center" sortable="custom" />
        <el-table-column prop="created_at" :label="$t('registerApplications.createdAt')" width="180" align="center" sortable="custom" />
        <el-table-column :label="$t('registerApplications.actions')" width="220" fixed="right" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" type="success" :icon="Check" :disabled="row.status !== 0" @click="approve(row)">
                {{ $t('registerApplications.approve') }}
              </el-button>
              <el-button size="small" type="danger" :icon="Close" :disabled="row.status !== 0" @click="openReject(row)">
                {{ $t('registerApplications.reject') }}
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="loadingMore" class="load-more">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>{{ $t('registerApplications.loading') }}</span>
      </div>
      <div v-else-if="!hasMore && applications.length > 0" class="load-more done">
        <span>{{ $t('common.noMore') }}</span>
      </div>
    </div>

    <el-dialog v-model="rejectDialogVisible" :title="$t('registerApplications.rejectTitle')" width="480px">
      <el-form :model="rejectForm" label-width="120px">
        <el-form-item :label="$t('registerApplications.reason')">
          <el-input v-model="rejectForm.reason" type="textarea" :rows="3" :placeholder="$t('registerApplications.reasonPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">{{ $t('registerApplications.cancel') }}</el-button>
        <el-button type="danger" :loading="submitting" @click="confirmReject">
          {{ $t('registerApplications.reject') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus'
import { markRaw } from 'vue'
import { Loading, Search, Refresh, Check, Close, ArrowDown } from '@element-plus/icons-vue'
import { 
  searchRegisterApplications, approveRegisterApplication, rejectRegisterApplication,
  searchResetApplications, approveResetApplication, rejectResetApplication 
} from '@/api/users'

export default {
  name: 'RegisterApplication',
  components: {
    Loading,
    Search,
    Refresh,
    Check,
    Close,
    ArrowDown
  },
  data() {
    return {
      activeTab: 'register',
      applications: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 100,
      pageOffset: 0,
      sortBy: '',
      sortOrder: '',
      hasMore: true,
      loadingMore: false,
      rejectDialogVisible: false,
      submitting: false,
      bulkApproving: false,
      bulkRejectMode: false,
      selectedApplicationRows: [],
      rejectForm: {
        ids: [],
        reason: '',
      },
      // Icons for template usage
      Search: markRaw(Search),
      Refresh: markRaw(Refresh),
      Check: markRaw(Check),
      Close: markRaw(Close),
      ArrowDown: markRaw(ArrowDown),
      Loading: markRaw(Loading)
    }
  },
  computed: {
    filteredApplications() {
      const q = this.search.trim().toLowerCase()
      return this.applications.filter((app) => {
        if (this.statusFilter !== 'all' && app.status !== this.statusFilter) return false
        if (!q) return true
        return (app.username || '').toLowerCase().includes(q)
      })
    },
    selectedCount() {
      return this.selectedApplicationRows.length
    },
  },
  created() {
    this.loadApplications(true)
  },
  methods: {
    handleTabChange() {
      this.loadApplications(true)
    },
    onSelectionChange(selection) {
      this.selectedApplicationRows = selection || []
    },
    selectAll() {
      if (this.$refs.applicationsTableRef) {
        this.applications.forEach((row) => {
          this.$refs.applicationsTableRef.toggleRowSelection(row, true)
        })
      }
    },
    onSortChange({ prop, order }) {
      if (!prop || !order) {
        this.sortBy = ''
        this.sortOrder = ''
      } else {
        this.sortBy = prop
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc'
      }
      this.loadApplications(true)
    },
    async loadApplications(reset = false) {
      if (reset) {
        this.pageOffset = 0
        this.hasMore = true
        this.applications = []
      }
      if (!this.hasMore || this.loadingMore) return

      this.loadingMore = true
      this.loading = reset
      this.error = null
      try {
        const params = {
          q: this.search || undefined,
          status: this.statusFilter === 'all' ? undefined : this.statusFilter,
          limit: this.pageSize,
          offset: this.pageOffset,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
          with_total: false
        }
        
        const fetchFn = this.activeTab === 'reset' ? searchResetApplications : searchRegisterApplications
        const response = await fetchFn(params)
        
        const data = Array.isArray(response.data) ? response.data : (response.data?.data || [])
        const mapped = data.map((a) => ({
          id: a.ID || a.id || 0,
          username: a.Username || a.username || '',
          status: a.Status ?? a.status ?? 0,
          reason: a.Reason || a.reason || '',
          approved_by: a.ApprovedBy || a.approved_by || null,
          created_at: a.CreatedAt || a.created_at || '',
        }))
        this.applications = reset ? mapped : this.applications.concat(mapped)
        this.pageOffset += mapped.length
        if (mapped.length < this.pageSize) {
          this.hasMore = false
        }
      } catch (err) {
        this.error = err?.response?.data?.error || err.message || this.$t('registerApplications.loadFailed')
      } finally {
        this.loading = false
        this.loadingMore = false
      }
    },
    async loadMoreApplications() {
      await this.loadApplications(false)
    },
    statusLabel(status) {
      switch (status) {
        case 1:
          return this.$t('registerApplications.statusApproved')
        case 2:
          return this.$t('registerApplications.statusRejected')
        default:
          return this.$t('registerApplications.statusPending')
      }
    },
    statusTagType(status) {
      switch (status) {
        case 1:
          return 'success'
        case 2:
          return 'danger'
        default:
          return 'warning'
      }
    },
    async approve(row) {
      this.submitting = true
      try {
        if (this.activeTab === 'reset') {
          await approveResetApplication(row.id)
        } else {
          await approveRegisterApplication(row.id)
        }
        ElMessage.success(this.$t('registerApplications.approved'))
        await this.loadApplications(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('registerApplications.approveFailed'))
      } finally {
        this.submitting = false
      }
    },
    async approveSelected() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'))
        return
      }
      this.bulkApproving = true
      try {
        const approveFn = this.activeTab === 'reset' ? approveResetApplication : approveRegisterApplication
        await Promise.all(this.selectedApplicationRows.map((app) => approveFn(app.id)))
        ElMessage.success(this.$t('registerApplications.approved'))
        this.clearSelection()
        await this.loadApplications(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('registerApplications.approveFailed'))
      } finally {
        this.bulkApproving = false
      }
    },
    openReject(row) {
      this.rejectForm = { ids: [row.id], reason: '' }
      this.bulkRejectMode = false
      this.rejectDialogVisible = true
    },
    openBulkReject() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'))
        return
      }
      this.rejectForm = { ids: this.selectedApplicationRows.map((app) => app.id), reason: '' }
      this.bulkRejectMode = true
      this.rejectDialogVisible = true
    },
    async confirmReject() {
      if (!this.rejectForm.ids.length) return
      this.submitting = true
      try {
        const rejectFn = this.activeTab === 'reset' ? rejectResetApplication : rejectRegisterApplication
        await Promise.all(this.rejectForm.ids.map((id) => rejectFn(id, { reason: this.rejectForm.reason })))
        ElMessage.success(this.$t('registerApplications.rejected'))
        this.rejectDialogVisible = false
        this.clearSelection()
        await this.loadApplications(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('registerApplications.rejectFailed'))
      } finally {
        this.submitting = false
      }
    },
    clearSelection() {
      if (this.$refs.applicationsTableRef && this.$refs.applicationsTableRef.clearSelection) {
        this.$refs.applicationsTableRef.clearSelection()
      }
      this.selectedApplicationRows = []
    },
  },
}
</script>

<style scoped>
.register-app-content {
  margin-top: 8px;
}

.register-app-scroll {
  max-height: calc(100vh - 280px);
  overflow-y: auto;
}

.loading-state, .load-more {
  text-align: center;
  padding: 20px;
  color: var(--text-muted);
}

.load-more.done {
  padding: 40px;
}

:deep(.el-table__row) {
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  background-color: var(--brand-50) !important;
}
</style>
