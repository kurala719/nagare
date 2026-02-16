<template>
  <div class="register-app-page">
    <div class="register-app-toolbar">
      <div class="register-app-filters">
        <span class="filter-label">{{ $t('registerApplications.search') }}</span>
        <el-input v-model="search" :placeholder="$t('registerApplications.search')" clearable class="register-app-search" />
        <span class="filter-label">{{ $t('registerApplications.filterStatus') }}</span>
        <el-select v-model="statusFilter" :placeholder="$t('registerApplications.filterStatus')" class="register-app-filter">
          <el-option :label="$t('registerApplications.filterAll')" value="all" />
          <el-option :label="$t('registerApplications.statusPending')" :value="0" />
          <el-option :label="$t('registerApplications.statusApproved')" :value="1" />
          <el-option :label="$t('registerApplications.statusRejected')" :value="2" />
        </el-select>
        <div class="register-app-bulk-actions">
          <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
          <el-button type="success" plain :disabled="selectedCount === 0" :loading="bulkApproving" @click="approveSelected">
            {{ $t('registerApplications.approveSelected') }}
          </el-button>
          <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkReject">
            {{ $t('registerApplications.rejectSelected') }}
          </el-button>
        </div>
      </div>
      <el-button type="primary" @click="loadApplications(true)">
        {{ $t('registerApplications.refresh') }}
      </el-button>
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

    <div
      v-if="!loading && !error"
      class="register-app-scroll"
      v-infinite-scroll="loadMoreApplications"
      :infinite-scroll-disabled="loadingMore || !hasMore"
      :infinite-scroll-distance="120"
    >
      <el-table
        v-if="filteredApplications.length > 0"
        :data="filteredApplications"
        border
        class="register-app-table"
        ref="applicationsTableRef"
        row-key="id"
        @selection-change="onSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="id" :label="$t('registerApplications.id')" width="90" />
        <el-table-column prop="username" :label="$t('registerApplications.username')" min-width="160" />
        <el-table-column prop="status" :label="$t('registerApplications.status')" width="140">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" :label="$t('registerApplications.reason')" min-width="220" show-overflow-tooltip />
        <el-table-column prop="approved_by" :label="$t('registerApplications.approvedBy')" width="140" />
        <el-table-column prop="created_at" :label="$t('registerApplications.createdAt')" min-width="180" />
        <el-table-column :label="$t('registerApplications.actions')" width="220" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" :disabled="row.status !== 0" @click="approve(row)">
              {{ $t('registerApplications.approve') }}
            </el-button>
            <el-button size="small" type="danger" :disabled="row.status !== 0" @click="openReject(row)">
              {{ $t('registerApplications.reject') }}
            </el-button>
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
</template>

<script>
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { searchRegisterApplications, approveRegisterApplication, rejectRegisterApplication } from '@/api/users'

export default {
  name: 'RegisterApplication',
  components: { Loading },
  data() {
    return {
      applications: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 100,
      pageOffset: 0,
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
    onSelectionChange(selection) {
      this.selectedApplicationRows = selection || []
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
        }
        const response = await searchRegisterApplications(params)
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
        await approveRegisterApplication(row.id)
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
        await Promise.all(this.selectedApplicationRows.map((app) => approveRegisterApplication(app.id)))
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
        await Promise.all(this.rejectForm.ids.map((id) => rejectRegisterApplication(id, { reason: this.rejectForm.reason })))
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
.register-app-page {
  padding: 16px;
}

.register-app-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  gap: 12px;
}

.register-app-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.register-app-bulk-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.selected-count {
  color: #606266;
  font-size: 13px;
}

.register-app-search {
  max-width: 260px;
}

.register-app-filter {
  min-width: 180px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}

.register-app-error {
  margin: 16px 0;
}

.register-app-empty {
  margin: 40px 0;
}

.register-app-table {
  width: 100%;
}
</style>
