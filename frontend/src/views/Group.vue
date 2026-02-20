<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('groups.title') }}</h1>
      <p class="page-subtitle">{{ $t('groups.loading') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-select v-model="selectedColumns" multiple collapse-tags :placeholder="$t('common.columns')" style="width: 180px">
          <el-option v-for="col in columnOptions" :key="col.key" :label="col.label" :value="col.key" />
        </el-select>
        
        <el-input v-model="search" :placeholder="$t('groups.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('groups.filterStatus')" style="width: 120px">
          <el-option :label="$t('groups.filterAll')" value="all" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>

        <el-select v-model="monitorFilter" :placeholder="$t('hosts.filterMonitor')" style="width: 140px" clearable>
          <el-option :label="$t('hosts.filterAll')" :value="0" />
          <el-option v-for="monitor in monitors" :key="monitor.id" :label="monitor.name" :value="monitor.id" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button type="primary" :icon="Plus" @click="createDialogVisible = true">
          {{ $t('groups.create') }}
        </el-button>
        <el-dropdown trigger="click" v-if="selectedCount > 0" style="margin-left: 8px">
          <el-button>
            {{ $t('common.selectedCount', { count: selectedCount }) }}<el-icon class="el-icon--right"><arrow-down /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="Edit" @click="openBulkUpdateDialog">{{ $t('common.bulkUpdate') }}</el-dropdown-item>
              <el-dropdown-item :icon="Delete" @click="openBulkDeleteDialog" style="color: var(--el-color-danger)">{{ $t('common.bulkDelete') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
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
    style="width: 100%; border-radius: 4px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,0.05);"
    ref="groupsTableRef"
    row-key="id"
    @selection-change="onSelectionChange"
    @sort-change="onSortChange"
    header-cell-class-name="table-header"
  >
    <el-table-column type="selection" width="50" align="center" />
    <el-table-column v-if="isColumnVisible('name')" prop="name" :label="$t('groups.name')" min-width="160" show-overflow-tooltip sortable="custom" />
    <el-table-column v-if="isColumnVisible('monitor')" :label="$t('hosts.monitor')" min-width="150" show-overflow-tooltip prop="monitor_id" sortable="custom">
      <template #default="{ row }">
        <el-tag effect="plain" type="info" size="small">{{ getMonitorName(row.monitor_id) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('enabled')" :label="$t('common.enabled')" width="100" align="center" prop="enabled" sortable="custom">
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small" effect="light">
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('status')" :label="$t('groups.status')" width="120" align="center" prop="status" sortable="custom">
      <template #default="{ row }">
        <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
          <el-tag :type="getStatusInfo(row.status).type" size="small" effect="dark">
            {{ getStatusInfo(row.status).label }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('health_score')" label="Health" width="100" prop="health_score" sortable="custom">
      <template #default="{ row }">
        <el-progress :percentage="row.health_score" :status="getHealthStatus(row.health_score)" />
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('lastSync')" :label="$t('hosts.lastSync')" min-width="180" prop="last_sync_at" sortable="custom">
      <template #default="{ row }">
        {{ row.last_sync_at ? new Date(row.last_sync_at).toLocaleString() : '-' }}
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('externalSource')" :label="$t('hosts.externalSource')" min-width="140" prop="external_source" sortable="custom" />
    <el-table-column v-if="isColumnVisible('description')" prop="description" :label="$t('groups.description')" min-width="200" show-overflow-tooltip />
    <el-table-column :label="$t('groups.actions')" width="300" fixed="right" align="center">
      <template #default="{ row }">
        <el-button-group>
          <el-tooltip :content="$t('groups.details')" placement="top">
            <el-button size="small" :icon="Document" @click="openDetails(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('groups.properties')" placement="top">
            <el-button size="small" :icon="Setting" @click="openProperties(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('groups.delete')" placement="top">
            <el-button size="small" type="danger" :icon="Delete" @click="onDelete(row)" />
          </el-tooltip>
        </el-button-group>
      </template>
    </el-table-column>
  </el-table>
  </div>
  </div>
  <div v-if="!loading && !error && totalGroups > 0" class="groups-pagination">
    <el-pagination
      background
      layout="total, sizes, prev, pager, next, jumper"
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
      <el-form-item :label="$t('hosts.monitor')">
        <el-select v-model="newGroup.monitor_id" style="width: 100%;" clearable>
          <el-option :label="$t('hosts.filterAll')" :value="0" />
          <el-option v-for="monitor in monitors" :key="monitor.id" :label="monitor.name" :value="monitor.id" />
        </el-select>
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
      <el-form-item :label="$t('hosts.monitor')">
        <el-select v-model="selectedGroup.monitor_id" style="width: 100%;" clearable>
          <el-option :label="$t('hosts.filterAll')" :value="0" />
          <el-option v-for="monitor in monitors" :key="monitor.id" :label="monitor.name" :value="monitor.id" />
        </el-select>
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
        <el-col :span="8"><el-card shadow="hover">{{ $t('groups.totalHosts') }}: {{ groupDetail.summary.total_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card shadow="hover">{{ $t('groups.activeHosts') }}: {{ groupDetail.summary.active_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card shadow="hover">{{ $t('groups.errorHosts') }}: {{ groupDetail.summary.error_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card shadow="hover">{{ $t('groups.syncingHosts') }}: {{ groupDetail.summary.syncing_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card shadow="hover">{{ $t('groups.totalItems') }}: {{ groupDetail.summary.total_items }}</el-card></el-col>
      </el-row>

      <el-divider content-position="left">{{ $t('groups.hosts') }}</el-divider>
      <el-table :data="groupDetail.hosts" border style="width: 100%">
        <el-table-column prop="name" :label="$t('hosts.name')" min-width="160" show-overflow-tooltip />
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
import { markRaw } from 'vue';
import { Loading, Plus, Delete, Edit, Download, Upload, Search, Refresh, Document, Setting, ArrowDown } from '@element-plus/icons-vue';
import { fetchGroupData, addGroup, updateGroup, deleteGroup, pullGroup, pushGroup } from '@/api/groups';
import { fetchMonitorData, syncGroupsFromMonitor } from '@/api/monitors';

export default {
  name: 'Group',
  components: { Loading, Plus, Delete, Edit, Download, Upload, Search, Refresh, Document, Setting, ArrowDown },
  data() {
    return {
      groups: [],
      monitors: [],
      pageSize: 20,
      currentPage: 1,
      totalGroups: 0,
      sortBy: '',
      sortOrder: '',
      loading: false,
      error: null,
      search: '',
      searchField: 'all',
      selectedColumns: ['name', 'monitor', 'enabled', 'status', 'health_score', 'lastSync', 'externalSource', 'description'],
      statusFilter: 'all',
      monitorFilter: 0,
      syncMonitorId: 0,
      createDialogVisible: false,
      propertiesDialogVisible: false,
      detailDialogVisible: false,
      detailLoading: false,
      groupDetail: null,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      pullingGroups: false,
      pushingGroups: false,
      selectedGroupRows: [],
      newGroup: { name: '', description: '', enabled: 1, status: 1, monitor_id: 0 },
      selectedGroup: { id: 0, name: '', description: '', enabled: 1, status: 1, monitor_id: 0 },
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
      // Icons for template usage
      Plus: markRaw(Plus),
      Delete: markRaw(Delete),
      Edit: markRaw(Edit),
      Download: markRaw(Download),
      Upload: markRaw(Upload),
      Search: markRaw(Search),
      Refresh: markRaw(Refresh),
      Document: markRaw(Document),
      Setting: markRaw(Setting),
      ArrowDown: markRaw(ArrowDown),
      Loading: markRaw(Loading)
    };
  },
  computed: {
    filteredGroups() {
      return this.groups;
    },
    columnOptions() {
      return [
        { key: 'name', label: this.$t('groups.name') },
        { key: 'monitor', label: this.$t('hosts.monitor') },
        { key: 'enabled', label: this.$t('common.enabled') },
        { key: 'status', label: this.$t('groups.status') },
        { key: 'health_score', label: 'Health' },
        { key: 'lastSync', label: this.$t('hosts.lastSync') },
        { key: 'externalSource', label: this.$t('hosts.externalSource') },
        { key: 'description', label: this.$t('groups.description') },
      ];
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
    monitorFilter() {
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
    this.loadMonitors();
    this.loadGroups(true);
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedGroupRows = selection || [];
    },
    selectAll() {
      if (this.$refs.groupsTableRef) {
        this.groups.forEach((row) => {
          this.$refs.groupsTableRef.toggleRowSelection(row, true);
        });
      }
    },
    onSortChange({ prop, order }) {
      if (!prop || !order) {
        this.sortBy = '';
        this.sortOrder = '';
      } else {
        this.sortBy = prop;
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc';
      }
      this.currentPage = 1;
      this.loadGroups(true);
    },
    async loadMonitors() {
      try {
        const response = await fetchMonitorData();
        let data = [];
        if (response?.success && response?.data !== undefined) {
          data = Array.isArray(response.data) ? response.data : 
                 (Array.isArray(response.data.items) ? response.data.items : []);
        } else if (Array.isArray(response)) {
          data = response;
        }
        this.monitors = data.map((m: any) => ({
          id: Number(m.ID || m.id || 0),
          name: m.Name || m.name || '',
        }));
      } catch (err) {
        console.error('Error loading monitors:', err);
      }
    },
    getMonitorName(monitorId) {
      if (!monitorId) return this.$t('hosts.unknown');
      const monitor = this.monitors.find((m: any) => m.id === monitorId);
      return monitor ? monitor.name : `${this.$t('hosts.unknown')} (#${monitorId})`;
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
            monitor_id: group.monitor_id,
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
        const response = await fetchGroupData({
          q: this.search || undefined,
          status: this.statusFilter === 'all' ? undefined : this.statusFilter,
          monitor_id: this.monitorFilter || undefined,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
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
          health_score: g.health_score ?? g.HealthScore ?? 100,
          monitor_id: g.MonitorID || g.monitor_id || g.monitorId || 0,
          status_reason: g.Reason || g.reason || g.Error || g.error || g.ErrorMessage || g.error_message || g.LastError || g.last_error || '',
          last_sync_at: g.last_sync_at,
          external_source: g.external_source || '',
        }));
        this.groups = mapped;
        this.totalGroups = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('groups.loadFailed');
      } finally {
        this.loading = false;
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
          monitor_id: this.selectedGroup.monitor_id,
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
      this.newGroup = { name: '', description: '', enabled: 1, status: 1, monitor_id: 0 };
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
        this.newGroup = { name: '', description: '', enabled: 1, status: 1, monitor_id: 0 };
        ElMessage.success(this.$t('groups.created'));
      } catch (err) {
        ElMessage.error(this.$t('groups.createFailed') + ': ' + (err.message || ''));
      }
    },
    async onPull(group) {
      try {
        await pullGroup(group.id);
        ElMessage.success(this.$t('groups.pullSuccess') || 'Group pulled successfully');
        await this.loadGroups();
      } catch (err) {
        ElMessage.error(this.$t('groups.pullFailed') + ': ' + (err.message || ''));
      }
    },
    async onPush(group) {
      try {
        await pushGroup(group.id);
        ElMessage.success(this.$t('groups.pushSuccess') || 'Group pushed successfully');
        await this.loadGroups();
      } catch (err) {
        ElMessage.error(this.$t('groups.pushFailed') + ': ' + (err.message || ''));
      }
    },
    async pullGroups() {
      this.pullingGroups = true;
      try {
        if (this.selectedCount > 0) {
          const results = await this.batchSyncSelectedGroups('pull');
          ElMessage({
            type: results.success > 0 ? 'success' : 'warning',
            message: this.$t('groups.pullSuccess') + ` (${results.success}/${results.total}${results.skipped ? `, ${this.$t('common.skipped') || 'skipped'}: ${results.skipped}` : ''})`,
          });
        } else {
          const monitorId = this.syncMonitorId || this.monitorFilter;
          if (!monitorId) {
            ElMessage.warning(this.$t('hosts.selectMonitorFirst') || this.$t('common.selectAtLeastOne'));
            return;
          }
          await syncGroupsFromMonitor(monitorId);
          ElMessage.success(this.$t('groups.pullSuccess'));
        }
        await this.loadGroups(true);
        this.clearSelection();
      } catch (err) {
        ElMessage.error(err.message || this.$t('groups.pullFailed'));
      } finally {
        this.pullingGroups = false;
      }
    },
    async pushGroups() {
      this.pushingGroups = true;
      try {
        if (this.selectedCount > 0) {
          const results = await this.batchSyncSelectedGroups('push');
          ElMessage({
            type: results.success > 0 ? 'success' : 'warning',
            message: this.$t('groups.pushSuccess') + ` (${results.success}/${results.total}${results.skipped ? `, ${this.$t('common.skipped') || 'skipped'}: ${results.skipped}` : ''})`,
          });
        } else {
           // For now, only selected push is supported as "Push All" endpoint might not exist or be safe
           ElMessage.warning(this.$t('common.selectAtLeastOne'));
        }
        await this.loadGroups(true);
        this.clearSelection();
      } catch (err) {
        ElMessage.error(err.message || this.$t('groups.pushFailed'));
      } finally {
        this.pushingGroups = false;
      }
    },
    async batchSyncSelectedGroups(action) {
      const targets = this.selectedGroupRows || [];
      const tasks = [];
      let skipped = 0;
      targets.forEach((group) => {
        // Group sync usually relies on associated monitor_id in backend
        if (action === 'pull') {
           tasks.push(pullGroup(group.id));
        } else if (action === 'push') {
           tasks.push(pushGroup(group.id));
        }
      });
      const results = await Promise.allSettled(tasks);
      const success = results.filter((result) => result.status === 'fulfilled').length;
      return { total: tasks.length + skipped, success, skipped };
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
    getHealthStatus(score: number) {
      if (score >= 90) return 'success';
      if (score >= 70) return 'warning';
      return 'exception';
    },
    isColumnVisible(key) {
      return this.selectedColumns.includes(key);
    },
  },
};
</script>

<style scoped>
.groups-scroll {
  margin-top: 8px;
}

.groups-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.loading-state {
  text-align: center;
  padding: 60px;
}

:deep(.el-table__row) {
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  background-color: var(--brand-50) !important;
}
</style>
