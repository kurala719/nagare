<template>
  <div class="groups-toolbar">
    <div class="groups-filters">
      <span class="filter-label">{{ $t('common.columns') }}</span>
      <el-select v-model="selectedColumns" multiple collapse-tags :placeholder="$t('common.search')" class="groups-filter" style="min-width: 220px;">
        <el-option v-for="col in columnOptions" :key="col.key" :label="col.label" :value="col.key" />
      </el-select>
      <span class="filter-label">{{ $t('common.search') }}</span>
      <el-select v-model="searchField" :placeholder="$t('common.search')" class="groups-filter">
        <el-option :label="$t('groups.filterAll')" value="all" />
        <el-option v-for="col in searchableColumns" :key="col.key" :label="col.label" :value="col.key" />
      </el-select>
      <span class="filter-label">{{ $t('common.search') }}</span>
      <el-input v-model="search" :placeholder="$t('groups.search')" clearable class="groups-search" />
      <span class="filter-label">{{ $t('groups.filterStatus') }}</span>
      <el-select v-model="statusFilter" :placeholder="$t('groups.filterStatus')" class="groups-filter">
        <el-option :label="$t('groups.filterAll')" value="all" />
        <el-option :label="$t('common.statusInactive')" :value="0" />
        <el-option :label="$t('common.statusActive')" :value="1" />
        <el-option :label="$t('common.statusError')" :value="2" />
        <el-option :label="$t('common.statusSyncing')" :value="3" />
      </el-select>
      <span class="filter-label">{{ $t('common.sort') }}</span>
      <el-select v-model="sortKey" class="groups-filter">
        <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
        <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
        <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
        <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
        <el-option :label="$t('common.sortStatusAsc')" value="status_asc" />
        <el-option :label="$t('common.sortStatusDesc')" value="status_desc" />
      </el-select>
      <div class="groups-bulk-actions">
        <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
        <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
          {{ $t('common.bulkUpdate') }}
        </el-button>
        <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
          {{ $t('common.bulkDelete') }}
        </el-button>
      </div>
    </div>
    <el-button type="primary" @click="createDialogVisible = true">
      {{ $t('groups.create') }}
    </el-button>
  </div>

  <div v-if="loading" class="loading-state">
    <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
    <p>{{ $t('groups.loading') }}</p>
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
      <el-button size="small" @click="loadGroups">{{ $t('groups.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && groups.length === 0"
    :description="$t('groups.noGroups')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && groups.length > 0 && filteredGroups.length === 0"
    :description="$t('groups.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="groups-scroll">
  <el-table
    v-if="filteredGroups.length > 0"
    :data="filteredGroups"
    border
    style="margin: 20px; width: calc(100% - 40px);"
    ref="groupsTableRef"
    row-key="id"
    @selection-change="onSelectionChange"
  >
    <el-table-column type="selection" width="50" />
    <el-table-column v-if="isColumnVisible('name')" prop="name" :label="$t('groups.name')" min-width="160" />
    <el-table-column v-if="isColumnVisible('enabled')" :label="$t('common.enabled')" min-width="110">
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('status')" :label="$t('groups.status')" min-width="160">
      <template #default="{ row }">
        <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
          <el-tag :type="getStatusInfo(row.status).type">
            {{ getStatusInfo(row.status).label }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('description')" prop="description" :label="$t('groups.description')" min-width="200" show-overflow-tooltip />
    <el-table-column :label="$t('groups.actions')" min-width="240" fixed="right">
      <template #default="{ row }">
        <el-button size="small" type="primary" @click="openDetails(row)">{{ $t('groups.details') }}</el-button>
        <el-button size="small" @click="openProperties(row)">{{ $t('groups.properties') }}</el-button>
        <el-button size="small" type="danger" @click="onDelete(row)">{{ $t('groups.delete') }}</el-button>
      </template>
    </el-table-column>
  </el-table>
  </div>
  <div v-if="!loading && !error && totalGroups > 0" class="groups-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalGroups"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('groups.createTitle')" width="500px" align-center>
    <el-form :model="newGroup" label-width="120px">
      <el-form-item :label="$t('groups.name')">
        <el-input v-model="newGroup.name" :placeholder="$t('groups.name')" />
      </el-form-item>
      <el-form-item :label="$t('groups.description')">
        <el-input v-model="newGroup.description" type="textarea" :placeholder="$t('groups.description')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newGroup.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('groups.status')">
        <el-select v-model="newGroup.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('groups.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('groups.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('groups.properties')} - ${selectedGroup?.name || ''}`" width="600px">
    <el-form :model="selectedGroup" label-width="120px">
      <el-form-item :label="$t('groups.name')">
        <el-input v-model="selectedGroup.name" />
      </el-form-item>
      <el-form-item :label="$t('groups.description')">
        <el-input v-model="selectedGroup.description" type="textarea" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedGroup.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('groups.status')">
        <el-select v-model="selectedGroup.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('groups.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('groups.save') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Update Dialog -->
  <el-dialog v-model="bulkDialogVisible" :title="$t('common.bulkUpdateTitle')" width="460px">
    <el-form :model="bulkForm" label-width="140px">
      <el-form-item :label="$t('common.enabled')">
        <el-select v-model="bulkForm.enabled" style="width: 100%;">
          <el-option :label="$t('common.bulkUpdateNoChange')" value="nochange" />
          <el-option :label="$t('common.enabled')" value="enable" />
          <el-option :label="$t('common.disabled')" value="disable" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('groups.status')">
        <el-select v-model="bulkForm.status" style="width: 100%;">
          <el-option :label="$t('common.bulkUpdateNoChange')" value="nochange" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="bulkDialogVisible = false">{{ $t('groups.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('groups.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedGroups" :loading="bulkDeleting">{{ $t('groups.delete') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="detailDialogVisible" :title="$t('groups.detailTitle')" width="800px">
    <div v-if="detailLoading" style="text-align: center; padding: 30px;">
      <el-icon class="is-loading" size="40" color="#409EFF"><Loading /></el-icon>
    </div>
    <div v-else-if="groupDetail">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('groups.name')">{{ groupDetail.group.name }}</el-descriptions-item>
        <el-descriptions-item :label="$t('groups.status')">
          <el-tag :type="getStatusInfo(groupDetail.group.status).type">
            {{ getStatusInfo(groupDetail.group.status).label }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.enabled')">
          {{ groupDetail.group.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('groups.description')">{{ groupDetail.group.description }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">{{ $t('groups.summary') }}</el-divider>
      <el-row :gutter="12">
        <el-col :span="8"><el-card>{{ $t('groups.totalHosts') }}: {{ groupDetail.summary.total_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card>{{ $t('groups.activeHosts') }}: {{ groupDetail.summary.active_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card>{{ $t('groups.errorHosts') }}: {{ groupDetail.summary.error_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card>{{ $t('groups.syncingHosts') }}: {{ groupDetail.summary.syncing_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card>{{ $t('groups.totalItems') }}: {{ groupDetail.summary.total_items }}</el-card></el-col>
      </el-row>

      <el-divider content-position="left">{{ $t('groups.hosts') }}</el-divider>
      <el-table :data="groupDetail.hosts" border>
        <el-table-column prop="name" :label="$t('hosts.name')" min-width="160" />
        <el-table-column prop="ip_addr" :label="$t('hosts.ip')" min-width="140" />
        <el-table-column :label="$t('hosts.status')" min-width="160">
          <template #default="{ row }">
            <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
              <el-tag :type="getStatusInfo(row.status).type">{{ getStatusInfo(row.status).label }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div v-else class="empty-detail">
      {{ $t('groups.selectGroup') }}
    </div>
  </el-dialog>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { fetchGroupData, addGroup, updateGroup, deleteGroup } from '@/api/groups';

export default {
  name: 'Group',
  components: { Loading },
  data() {
    return {
      groups: [],
      pageSize: 20,
      currentPage: 1,
      totalGroups: 0,
      sortKey: 'updated_desc',
      loading: false,
      error: null,
      search: '',
      searchField: 'all',
      selectedColumns: ['name', 'enabled', 'status', 'description'],
      columnOptions: [
        { key: 'name', label: this.$t('groups.name') },
        { key: 'enabled', label: this.$t('common.enabled') },
        { key: 'status', label: this.$t('groups.status') },
        { key: 'description', label: this.$t('groups.description') },
      ],
      statusFilter: 'all',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      detailDialogVisible: false,
      detailLoading: false,
      groupDetail: null,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedGroupRows: [],
      newGroup: { name: '', description: '', enabled: 1, status: 1 },
      selectedGroup: { id: 0, name: '', description: '', enabled: 1, status: 1 },
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
    };
  },
  computed: {
    filteredGroups() {
      return this.groups;
    },
    searchableColumns() {
      return this.columnOptions;
    },
    selectedCount() {
      return this.selectedGroupRows.length;
    },
  },
  watch: {
    search() {
      this.currentPage = 1;
      this.loadGroups(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadGroups(true);
    },
    sortKey() {
      this.currentPage = 1;
      this.loadGroups(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadGroups(true);
    },
    currentPage() {
      this.loadGroups();
    },
  },
  created() {
    this.loadGroups(true);
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedGroupRows = selection || [];
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedGroups() {
      if (this.selectedCount === 0) return;
      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedGroupRows.map((group) => deleteGroup(group.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadGroups(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkDeleteFailed'));
      } finally {
        this.bulkDeleting = false;
      }
    },
    openBulkUpdateDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkForm = {
        enabled: 'nochange',
        status: 'nochange',
      };
      this.bulkDialogVisible = true;
    },
    async applyBulkUpdate() {
      if (this.selectedCount === 0) return;
      if (this.bulkForm.enabled === 'nochange' && this.bulkForm.status === 'nochange') {
        ElMessage.warning(this.$t('common.bulkUpdateNoChanges'));
        return;
      }

      this.bulkUpdating = true;
      try {
        const enabledOverride = this.bulkForm.enabled;
        const statusOverride = this.bulkForm.status;
        await Promise.all(this.selectedGroupRows.map((group) => {
          const payload = {
            name: group.name,
            description: group.description,
            enabled: enabledOverride === 'nochange' ? group.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? group.status : statusOverride,
          };
          return updateGroup(group.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadGroups(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.groupsTableRef && this.$refs.groupsTableRef.clearSelection) {
        this.$refs.groupsTableRef.clearSelection();
      }
      this.selectedGroupRows = [];
    },
    async loadGroups(reset = false) {
      if (reset) {
        this.groups = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
        const response = await fetchGroupData({
          q: this.search || undefined,
          status: this.statusFilter === 'all' ? undefined : this.statusFilter,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: sortBy,
          order: sortOrder,
          with_total: 1,
        });
        const data = Array.isArray(response)
          ? response
          : (response.data?.items || response.items || response.data || response.groups || []);
        const total = response?.data?.total ?? response?.total ?? data.length;
        const mapped = data.map((g) => ({
          id: g.ID || g.id || 0,
          name: g.Name || g.name || '',
          description: g.Description || g.description || '',
          enabled: g.Enabled ?? g.enabled ?? 1,
          status: g.Status ?? g.status ?? 0,
          status_reason: g.Reason || g.reason || g.Error || g.error || g.ErrorMessage || g.error_message || g.LastError || g.last_error || '',
        }));
        this.groups = mapped;
        this.totalGroups = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('groups.loadFailed');
      } finally {
        this.loading = false;
      }
    },
    parseSortKey(key) {
      switch (key) {
        case 'name_asc':
          return { sortBy: 'name', sortOrder: 'asc' };
        case 'name_desc':
          return { sortBy: 'name', sortOrder: 'desc' };
        case 'status_asc':
          return { sortBy: 'status', sortOrder: 'asc' };
        case 'status_desc':
          return { sortBy: 'status', sortOrder: 'desc' };
        case 'created_desc':
          return { sortBy: 'created_at', sortOrder: 'desc' };
        case 'updated_desc':
        default:
          return { sortBy: 'updated_at', sortOrder: 'desc' };
      }
    },
    async openDetails(group) {
      this.$router.push({ path: `/group/${group.id}/detail` });
    },
    openProperties(group) {
      this.selectedGroup = { ...group };
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        await updateGroup(this.selectedGroup.id, {
          name: this.selectedGroup.name,
          description: this.selectedGroup.description,
          enabled: this.selectedGroup.enabled,
          status: this.selectedGroup.status,
        });
        await this.loadGroups(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('groups.updated'));
      } catch (err) {
        ElMessage.error(this.$t('groups.updateFailed') + ': ' + (err.message || ''));
      }
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newGroup = { name: '', description: '', enabled: 1, status: 1 };
    },
    async onCreate() {
      if (!this.newGroup.name) {
        ElMessage.warning(this.$t('groups.name'));
        return;
      }
      try {
        await addGroup(this.newGroup);
        await this.loadGroups(true);
        this.createDialogVisible = false;
        this.newGroup = { name: '', description: '', enabled: 1, status: 1 };
        ElMessage.success(this.$t('groups.created'));
      } catch (err) {
        ElMessage.error(this.$t('groups.createFailed') + ': ' + (err.message || ''));
      }
    },
    onDelete(group) {
      ElMessageBox.confirm(
        `${this.$t('groups.delete')} ${group.name}?`,
        this.$t('groups.delete'),
        {
          confirmButtonText: this.$t('groups.delete'),
          cancelButtonText: this.$t('groups.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteGroup(group.id);
          await this.loadGroups();
          ElMessage.success(this.$t('groups.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('groups.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('groups.deleteCanceled'));
      });
    },
    getStatusInfo(status) {
      const map = {
        0: { label: this.$t('common.statusInactive'), reason: this.$t('common.reasonInactive'), type: 'info' },
        1: { label: this.$t('common.statusActive'), reason: this.$t('common.reasonActive'), type: 'success' },
        2: { label: this.$t('common.statusError'), reason: this.$t('common.reasonError'), type: 'danger' },
        3: { label: this.$t('common.statusSyncing'), reason: this.$t('common.reasonSyncing'), type: 'warning' },
      };
      return map[status] || map[0];
    },
    isColumnVisible(key) {
      return this.selectedColumns.includes(key);
    },
  },
};
</script>

<style scoped>
.groups-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 16px 20px 0;
}

.groups-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.groups-bulk-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.groups-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 16px;
}

.selected-count {
  color: #606266;
  font-size: 13px;
}

.groups-search {
  width: 240px;
}

.groups-filter {
  min-width: 160px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}

.empty-detail {
  text-align: center;
  padding: 20px;
  color: #909399;
}
</style>
