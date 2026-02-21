<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('monitors.title') }}</h1>
      <p class="page-subtitle">{{ totalMonitors }} {{ $t('dashboard.monitors') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('monitors.search')" clearable style="width: 280px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('monitors.filterStatus')" style="width: 140px">
          <el-option :label="$t('monitors.filterAll')" value="all" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>

        <el-select v-model="sortKey" :placeholder="$t('common.sort')" style="width: 160px">
          <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
          <el-option :label="$t('common.sortCreatedAsc')" value="created_asc" />
          <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
          <el-option :label="$t('common.sortUpdatedAsc')" value="updated_asc" />
          <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
          <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button @click="loadMonitors" :loading="loading" :icon="Refresh" circle />
        <el-button type="primary" :icon="Plus" @click="createDialogVisible=true">
          {{ $t('monitors.create') }}
        </el-button>
        <el-dropdown trigger="click" v-if="selectedCount > 0">
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
  <el-dialog v-model="createDialogVisible" :title="$t('monitors.createTitle')" width="500px" align-center>
    <el-form :model="newMonitor" label-width="120px">
      <el-form-item :label="$t('monitors.name')">
        <el-input v-model="newMonitor.name" :placeholder="$t('monitors.name')" />
      </el-form-item>
      <el-form-item :label="$t('monitors.url')">
        <el-input v-model="newMonitor.url" :placeholder="$t('monitors.url')" />
      </el-form-item>
      <el-form-item :label="$t('monitors.username')">
        <el-input v-model="newMonitor.username" :placeholder="$t('monitors.username')" />
      </el-form-item>
      <el-form-item :label="$t('monitors.password')">
        <el-input v-model="newMonitor.password" type="password" :placeholder="$t('monitors.password')" show-password />
      </el-form-item>
      <el-form-item :label="$t('monitors.type')">
        <el-select v-model="newMonitor.type" style="width: 100%;">
          <el-option label="Zabbix" :value="2" />
          <el-option label="Other" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('monitors.description')">
        <el-input v-model="newMonitor.description" type="textarea" :placeholder="$t('monitors.description')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newMonitor.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('monitors.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('monitors.createBtn') }}</el-button>
    </template>  
  </el-dialog>

  <!-- Loading State -->
  <div v-if="loading" style="text-align: center; padding: 40px;">
    <el-icon class="is-loading" size="50" color="#409EFF">
      <Loading />
    </el-icon>
    <p style="margin-top: 16px; color: #909399;">{{ $t('monitors.loading') }}</p>
  </div>

  <!-- Error State -->
  <el-alert 
    v-if="error && !loading" 
    :title="error" 
    type="error" 
    show-icon
    style="margin: 20px;"
    :closable="false"
  >
    <template #default>
      <el-button size="small" @click="loadMonitors">{{ $t('monitors.retry') }}</el-button>
    </template>
  </el-alert>

  <!-- Empty State -->
  <el-empty 
    v-if="!loading && !error && monitors && monitors.length === 0"
    :description="$t('monitors.noMonitors')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && monitors && monitors.length > 0 && filteredMonitors.length === 0"
    :description="$t('monitors.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="monitors-content">
    <el-row :gutter="24">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="monitor in filteredMonitors" :key="monitor.id" style="margin-bottom: 24px;">
        <el-card class="monitor-card" :body-style="{ padding: '0px' }">
          <div class="monitor-card-header">
            <div class="monitor-icon-box">
              <el-icon :size="24"><MonitorIcon /></el-icon>
            </div>
            <div class="monitor-title-area">
              <h3 class="monitor-name">{{ monitor.name }}</h3>
              <span class="monitor-type-tag">{{ monitor.type === 1 ? 'SNMP' : monitor.type === 2 ? 'Zabbix' : 'Other' }}</span>
            </div>
            <el-checkbox :model-value="isSelected(monitor.id)" @change="toggleSelection(monitor.id, $event)" class="monitor-select" />
          </div>
          
          <div class="monitor-card-body">
            <p class="monitor-desc">{{ monitor.description || $t('monitors.noDescription') }}</p>
            <div class="monitor-status-row">
              <el-tag :type="(monitor.id === 1 || monitor.enabled === 1) ? 'success' : 'info'" size="small">
                {{ (monitor.id === 1 || monitor.enabled === 1) ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
              <el-tooltip :content="monitor.id === 1 ? $t('common.reasonActive') : (monitor.status_reason || getStatusInfo(monitor.status).reason)" placement="top">
                <el-tag :type="monitor.id === 1 ? 'success' : getStatusInfo(monitor.status).type" size="small" effect="dark">
                  {{ monitor.id === 1 ? $t('common.statusActive') : getStatusInfo(monitor.status).label }}
                </el-tag>
              </el-tooltip>
            </div>
            <div style="margin-top: 12px">
              <span style="font-size: 12px; color: var(--text-muted)">Health</span>
              <el-progress :percentage="monitor.id === 1 ? 100 : monitor.health_score" :status="getHealthStatus(monitor.id === 1 ? 100 : monitor.health_score)" :stroke-width="4" />
            </div>
          </div>

          <div class="monitor-card-footer">
            <el-button-group>
              <el-tooltip :content="$t('monitors.properties')" placement="bottom">
                <el-button size="small" :icon="Edit" @click="openProperties(monitor)" :disabled="monitor.id === 1" />
              </el-tooltip>
              <el-tooltip v-if="monitor.type !== 4" :content="monitor.auth_token ? $t('monitors.reLogin') : $t('monitors.login')" placement="bottom">
                <el-button size="small" :type="monitor.auth_token ? 'success' : 'warning'" plain :icon="monitor.auth_token ? SuccessFilled : CircleCloseFilled" @click="onLogin(monitor)" :loading="monitor.logging_in" :disabled="monitor.id === 1" />
              </el-tooltip>
              <el-tooltip :content="$t('monitors.delete')" placement="bottom">
                <el-button size="small" type="danger" plain :icon="Delete" @click="onDelete(monitor)" :disabled="monitor.id === 1" />
              </el-tooltip>
            </el-button-group>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
  <div v-if="!loading && !error && totalMonitors > 0" class="monitors-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalMonitors"
    />
  </div>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('monitors.propertiesTitle')} - ${selectedMonitor ? selectedMonitor.name : ''}`" width="600px">
    <el-form :model="selectedMonitor" label-width="120px">
      <el-form-item :label="$t('monitors.name')">
        <el-input v-model="selectedMonitor.name" />
      </el-form-item>
      <el-form-item :label="$t('monitors.url')">
        <el-input v-model="selectedMonitor.url" />
      </el-form-item>
      <el-form-item :label="$t('monitors.username')">
        <el-input v-model="selectedMonitor.username" />
      </el-form-item>
      <el-form-item :label="$t('monitors.password')">
        <el-input v-model="selectedMonitor.password" show-password />
      </el-form-item>
      <el-form-item label="Event Token">
        <el-input v-model="selectedMonitor.event_token" readonly />
        <div style="display: flex; gap: 8px; margin-top: 8px;">
          <el-button size="small" @click="copyEventToken">Copy</el-button>
          <el-button size="small" type="warning" @click="regenerateEventToken">Regenerate</el-button>
        </div>
      </el-form-item>
      <el-form-item :label="$t('monitors.type')">
        <el-select v-model="selectedMonitor.type" style="width: 100%;">
          <el-option v-if="selectedMonitor.id === 1" label="SNMP" :value="1" />
          <el-option label="Zabbix" :value="2" />
          <el-option label="Other" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('monitors.description')">
        <el-input type="textarea" v-model="selectedMonitor.description" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedMonitor.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('monitors.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('monitors.save') }}</el-button>
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
      <el-form-item :label="$t('monitors.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('monitors.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
        <template #footer>
          <el-button @click="bulkDeleteDialogVisible = false">{{ $t('monitors.cancel') }}</el-button>
          <el-button type="danger" @click="deleteSelectedMonitors" :loading="bulkDeleting">{{ $t('monitors.delete') }}</el-button>
        </template>
      </el-dialog>
      </div>
    </template>
    

<script>
import {
  Check,
  Delete,
  Edit,
  Message,
  Search,
  Star,
  Refresh,
  SuccessFilled,
  CircleCloseFilled,
  Loading,
  Plus,
  Monitor as MonitorIcon,
  ArrowDown
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { ElMessageBox } from 'element-plus'
import { markRaw } from 'vue'
import { fetchMonitorData, addMonitor, deleteMonitor, updateMonitor, loginMonitor, regenerateMonitorEventToken, syncGroupsFromMonitor } from '@/api/monitors'

export default {
    name: 'Monitor',
    components: {
      Check,
      Delete,
      Edit,
      Message,
      Search,
      Star,
      Refresh,
      SuccessFilled,
      CircleCloseFilled,
      Loading,
      Plus,
      MonitorIcon,
      ArrowDown
    },
    data() {
      return {
        monitors: [],
        createDialogVisible: false,
        propertiesDialogVisible: false,
        newMonitor: { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 2 },
        selectedMonitor: { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 },
        loading: false,
        error: null,
        search: '',
        searchField: 'all',
        statusFilter: 'all',
        refreshTimer: null,
          pageSize: 20,
          currentPage: 1,
          totalMonitors: 0,
          sortKey: 'updated_desc',
          bulkDialogVisible: false,
          bulkDeleteDialogVisible: false,
          bulkUpdating: false,
          bulkDeleting: false,
          selectedMonitorIds: [],
                      bulkForm: {
                        enabled: 'nochange',
                        status: 'nochange',
                      },
                      // Icons for template usage
                      Check: markRaw(Check),
                      Delete: markRaw(Delete),
                      Edit: markRaw(Edit),
                      Message: markRaw(Message),
                      Search: markRaw(Search),
                      Star: markRaw(Star),
                      Refresh: markRaw(Refresh),
                      SuccessFilled: markRaw(SuccessFilled),
                      CircleCloseFilled: markRaw(CircleCloseFilled),
                      Loading: markRaw(Loading),
                      Plus: markRaw(Plus),
                      Monitor: markRaw(MonitorIcon),
                      ArrowDown: markRaw(ArrowDown)
                  };
              },
          
    computed: {
      filteredMonitors() {
        return this.monitors;
      },
      selectedCount() {
        return this.selectedMonitorIds.length;
      },
    },
    watch: {
      '$route.query.q': function () {
        this.applySearchFromQuery();
      },
      search() {
        this.currentPage = 1;
        this.loadMonitors(true);
      },
      statusFilter() {
        this.currentPage = 1;
        this.loadMonitors(true);
      },
      sortKey() {
        this.currentPage = 1;
        this.loadMonitors(true);
      },
      pageSize() {
        this.currentPage = 1;
        this.loadMonitors(true);
      },
      currentPage() {
        this.loadMonitors();
      },
    },
    created() {
      this.applySearchFromQuery();
      this.loadMonitors(true);
    },
    mounted() {
      // Auto-refresh every 30 seconds
      this.refreshTimer = setInterval(() => {
        this.loadMonitors();
      }, 30000);
    },
    beforeUnmount() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
      }
    },
    methods: {
      applySearchFromQuery() {
        const queryValue = this.$route.query.q;
        const nextQuery = queryValue ? String(queryValue) : '';
        if (nextQuery !== this.search) {
          this.search = nextQuery;
          if (nextQuery) {
            this.searchField = 'all';
          }
        }
      },
      isSelected(id) {
        return this.selectedMonitorIds.includes(id);
      },
      toggleSelection(id, checked) {
        if (checked) {
          if (!this.selectedMonitorIds.includes(id)) {
            this.selectedMonitorIds.push(id);
          }
        } else {
          this.selectedMonitorIds = this.selectedMonitorIds.filter((itemId) => itemId !== id);
        }
      },
      selectAll() {
        this.selectedMonitorIds = this.monitors.map(m => m.id);
      },
      clearSelection() {
        this.selectedMonitorIds = [];
      },
      openBulkDeleteDialog() {
        if (this.selectedCount === 0) {
          ElMessage.warning(this.$t('common.selectAtLeastOne'));
          return;
        }
        this.bulkDeleteDialogVisible = true;
      },
      async deleteSelectedMonitors() {
        if (this.selectedCount === 0) return;

        this.bulkDeleting = true;
        try {
          await Promise.all(this.selectedMonitorIds.map((id) => deleteMonitor(id)));
          ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
          this.bulkDeleteDialogVisible = false;
          this.selectedMonitorIds = [];
          await this.loadMonitors(true);
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
          await Promise.all(this.monitors.filter((m) => this.selectedMonitorIds.includes(m.id)).map((monitor) => {
            const payload = {
              enabled: enabledOverride === 'nochange' ? monitor.enabled : (enabledOverride === 'enable' ? 1 : 0),
              status: statusOverride === 'nochange' ? monitor.status : statusOverride,
            };
            return updateMonitor(monitor.id, payload);
          }));
          ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
          this.bulkDialogVisible = false;
          this.selectedMonitorIds = [];
          await this.loadMonitors(true);
        } catch (err) {
          ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
        } finally {
          this.bulkUpdating = false;
        }
      },
      async loadMonitors(reset = false) {
        if (reset) {
          this.monitors = [];
        }
        this.loading = reset;
        this.error = null;
        try {
          const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
          const response = await fetchMonitorData({
            q: this.search || undefined,
            status: this.statusFilter === 'all' ? undefined : this.statusFilter,
            limit: this.pageSize,
            offset: (this.currentPage - 1) * this.pageSize,
            sort: sortBy,
            order: sortOrder,
            with_total: 1,
          });
          // Handle different response formats
          const data = Array.isArray(response)
            ? response
            : (response.data?.items || response.items || response.data || response.monitors || []);
          const total = response?.data?.total ?? response?.total ?? data.length;
          const mapped = data.map((m) => ({
            id: m.ID || m.id || 0,
            name: m.Name || m.name || '',
            url: m.URL || m.url || '',
            username: m.Username || m.username || '',
            password: m.Password || m.password || '',
            auth_token: m.AuthToken || m.auth_token || '',
            event_token: m.EventToken || m.event_token || '',
            enabled: m.Enabled ?? m.enabled ?? 1,
            status: m.Status ?? m.status ?? 0,
            health_score: m.health_score ?? m.HealthScore ?? 100,
            status_reason: m.Reason || m.reason || m.Error || m.error || m.ErrorMessage || m.error_message || m.LastError || m.last_error || '',
            description: m.Description || m.description || '',
            type: m.Type || m.type || '',
          }));
          this.monitors = mapped;
          this.totalMonitors = Number.isFinite(total) ? total : mapped.length;
        } catch (err) {
          this.error = err.message || 'Failed to load monitors';
          console.error('Error loading monitors:', err);
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
      openProperties(monitor) {
        this.selectedMonitor = Object.assign({}, monitor);
        this.propertiesDialogVisible = true;
      },
      cancelProperties() {
        this.propertiesDialogVisible = false;
      },
      async saveProperties() {
        try {
          const updateData = {
            name: this.selectedMonitor.name,
            url: this.selectedMonitor.url,
            username: this.selectedMonitor.username,
            password: this.selectedMonitor.password,
            auth_token: this.selectedMonitor.auth_token,
            type: this.selectedMonitor.type,
            description: this.selectedMonitor.description,
            enabled: this.selectedMonitor.enabled,
            status: this.selectedMonitor.status,
          };
          await updateMonitor(this.selectedMonitor.id, updateData);
          const idx = this.monitors.findIndex((m) => m.id === this.selectedMonitor.id);
          if (idx !== -1) {
            this.monitors.splice(idx, 1, Object.assign({}, this.selectedMonitor));
          }
          this.propertiesDialogVisible = false;
          ElMessage({
            type: 'success',
            message: 'Monitor updated successfully!',
          });
        } catch (err) {
          ElMessage({
            type: 'error',
            message: 'Failed to update monitor: ' + (err.message || 'Unknown error'),
          });
          console.error('Error updating monitor:', err);
        }
      },
      async regenerateEventToken() {
        try {
          await ElMessageBox.confirm(
            'Regenerate event token? This will invalidate the previous token.',
            'Confirm',
            {
              confirmButtonText: 'Regenerate',
              cancelButtonText: 'Cancel',
              type: 'warning',
            }
          );
          const response = await regenerateMonitorEventToken(this.selectedMonitor.id);
          const payload = response?.data?.data || response?.data || response;
          const token = payload?.event_token || payload?.EventToken || '';
          if (token) {
            this.selectedMonitor.event_token = token;
            const idx = this.monitors.findIndex((m) => m.id === this.selectedMonitor.id);
            if (idx !== -1) {
              this.monitors.splice(idx, 1, Object.assign({}, this.selectedMonitor));
            }
            ElMessage.success('Event token regenerated.');
          } else {
            ElMessage.warning('Token regenerated but not returned by server.');
          }
        } catch (err) {
          if (err !== 'cancel') {
            ElMessage.error(err.message || 'Failed to regenerate event token.');
          }
        }
      },
      async copyEventToken() {
        const token = this.selectedMonitor?.event_token || '';
        if (!token) {
          ElMessage.warning('No event token to copy.');
          return;
        }
        try {
          if (navigator?.clipboard?.writeText) {
            await navigator.clipboard.writeText(token);
            ElMessage.success('Event token copied.');
            return;
          }
        } catch (err) {
          console.warn('Clipboard write failed:', err);
        }
        ElMessage.info('Copy not available in this browser.');
      },
      deleteMonitor(monitor) {
        const index = this.monitors.findIndex((m) => m.id === monitor.id);
        if (index !== -1) {
          this.monitors.splice(index, 1);
        }
      },
      async onDelete(monitor) {
        ElMessageBox.confirm(
          `Are you sure you want to delete ${monitor.name}?`,
          'Warning',
          {
            confirmButtonText: 'Yes',
            cancelButtonText: 'No',
            type: 'warning',
          }
        ).then(async () => {
          try {
            await deleteMonitor(monitor.id);
            const index = this.monitors.findIndex((m) => m.id === monitor.id);
            if (index !== -1) {
              this.monitors.splice(index, 1);
            }
            ElMessage({
              type: 'success',
              message: 'Monitor deleted successfully!',
            });
          } catch (err) {
            ElMessage({
              type: 'error',
              message: 'Failed to delete monitor: ' + (err.message || 'Unknown error'),
            });
            console.error('Error deleting monitor:', err);
          }
        }).catch(() => {
          ElMessage({
            type: 'info',
            message: 'Delete canceled',
          });
        });
      },
      async onCreate() {
        if (!this.newMonitor.name) {
          ElMessage({
            type: 'warning',
            message: 'Please enter monitor name',
          });
          return;
        }
        
        try {
          const monitorData = {
            name: this.newMonitor.name,
            url: this.newMonitor.url,
            username: this.newMonitor.username,
            password: this.newMonitor.password,
            auth_token: this.newMonitor.auth_token,
            type: this.newMonitor.type,
            description: this.newMonitor.description,
            enabled: this.newMonitor.enabled,
            status: this.newMonitor.status,
          };
          
          // Call API to add monitor (it will auto-login if credentials provided)
          const response = await addMonitor(monitorData);
          
          // Add the created monitor to the list (includes auth_token if login succeeded)
          const createdMonitor = response.data || response;
          this.monitors.push({
            ...monitorData,
            id: createdMonitor.id,
            auth_token: createdMonitor.auth_token,
            event_token: createdMonitor.event_token,
          });
          
          this.newMonitor = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 2 };
          this.createDialogVisible = false;
          
          if (createdMonitor.auth_token) {
            ElMessage({
              type: 'success',
              message: 'Monitor created and logged in successfully!',
            });
          } else {
            ElMessage({
              type: 'success',
              message: 'Monitor created successfully!',
            });
          }
        } catch (err) {
          ElMessage({
            type: 'error',
            message: 'Failed to create monitor: ' + (err.message || 'Unknown error'),
          });
          console.error('Error creating monitor:', err);
        }
      },
      async onLogin(monitor) {
        // Set loading state for this specific monitor
        monitor.logging_in = true;
        try {
          const response = await loginMonitor(monitor.id);
          
          // Update monitor with new auth token
          const updatedMonitor = response.data || response;
          const idx = this.monitors.findIndex((m) => m.id === monitor.id);
          if (idx !== -1) {
            this.monitors.splice(idx, 1, {
              ...monitor,
              auth_token: updatedMonitor.auth_token,
              logging_in: false,
            });
          }
          
          ElMessage({
            type: 'success',
            message: 'Login successful!',
          });
        } catch (err) {
          monitor.logging_in = false;
          ElMessage({
            type: 'error',
            message: 'Login failed: ' + (err.response?.data?.error || err.message || 'Unknown error'),
          });
          console.error('Error logging in to monitor:', err);
        }
      },
      async onSyncGroups(monitor) {
        monitor.syncing_groups = true;
        try {
          const response = await syncGroupsFromMonitor(monitor.id);
          // Axios returns response object, response.data is body (APIResponse), response.data.data is SyncResult
          const apiData = response.data || response;
          const result = apiData.data || {}; 
          ElMessage.success(`Groups synced: ${result.added || 0} added, ${result.updated || 0} updated, ${result.failed || 0} failed.`);
        } catch (err) {
          ElMessage.error('Sync groups failed: ' + (err.message || 'Unknown error'));
        } finally {
          monitor.syncing_groups = false;
        }
      },
      cancelCreate() {
        this.createDialogVisible = false;
        this.newMonitor = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 2 };
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
      getHealthStatus(score) {
        if (score >= 90) return 'success';
        if (score >= 70) return 'warning';
        return 'exception';
      }
    }
  };
</script>

<style scoped>
.monitors-content {
  margin-top: 8px;
}

.monitor-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.monitor-card-header {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid var(--border-1);
  position: relative;
}

.monitor-icon-box {
  width: 48px;
  height: 48px;
  background: var(--brand-50);
  color: var(--brand-600);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.monitor-title-area {
  flex: 1;
  min-width: 0;
}

.monitor-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: var(--text-strong);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.monitor-type-tag {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
}

.monitor-select {
  position: absolute;
  top: 12px;
  right: 12px;
}

.monitor-card-body {
  padding: 20px;
  flex: 1;
}

.monitor-desc {
  font-size: 14px;
  color: var(--text-muted);
  margin: 0 0 16px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.5;
  height: 3em;
}

.monitor-status-row {
  display: flex;
  gap: 8px;
}

.monitor-card-footer {
  padding: 12px 20px;
  background: var(--surface-2);
  display: flex;
  justify-content: center;
  border-top: 1px solid var(--border-1);
}

.monitors-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
