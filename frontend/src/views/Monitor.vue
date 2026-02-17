<template>
  <div class="monitors-toolbar">
    <div class="monitors-filters">
      <span class="filter-label">{{ $t('monitors.searchField') }}</span>
      <el-select v-model="searchField" :placeholder="$t('monitors.searchField')" class="search-field-selector" style="width: 130px;">
        <el-option :label="$t('monitors.searchAll')" value="all" />
        <el-option :label="$t('monitors.name')" value="name" />
        <el-option :label="$t('monitors.url')" value="url" />
        <el-option :label="$t('monitors.type')" value="type" />
        <el-option :label="$t('monitors.description')" value="description" />
      </el-select>
      <span class="filter-label">{{ $t('monitors.search') }}</span>
      <el-input v-model="search" :placeholder="$t('monitors.search')" clearable class="monitors-search" />
      <span class="filter-label">{{ $t('monitors.filterStatus') }}</span>
      <el-select v-model="statusFilter" :placeholder="$t('monitors.filterStatus')" class="monitors-filter">
        <el-option :label="$t('monitors.filterAll')" value="all" />
        <el-option :label="$t('common.statusInactive')" :value="0" />
        <el-option :label="$t('common.statusActive')" :value="1" />
        <el-option :label="$t('common.statusError')" :value="2" />
        <el-option :label="$t('common.statusSyncing')" :value="3" />
      </el-select>
      <span class="filter-label">{{ $t('common.sort') }}</span>
      <el-select v-model="sortKey" class="monitors-filter">
        <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
        <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
        <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
        <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
        <el-option :label="$t('common.sortStatusAsc')" value="status_asc" />
        <el-option :label="$t('common.sortStatusDesc')" value="status_desc" />
      </el-select>
      <div class="monitors-bulk-actions">
        <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
        <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
          {{ $t('common.bulkUpdate') }}
        </el-button>
        <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
          {{ $t('common.bulkDelete') }}
        </el-button>
      </div>
    </div>
    <div style="display: flex; gap: 8px;">
      <el-button @click="loadMonitors" :loading="loading">
        <el-icon><Refresh /></el-icon>
      </el-button>
      <el-button type="primary" @click="createDialogVisible=true">
        {{ $t('monitors.create') }}
      </el-button>
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
          <el-option label="Zabbix" :value="1" />
          <el-option label="Prometheus" :value="2" />
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

  <div
    v-if="!loading && !error"
    class="monitors-scroll"
  >
  <!-- Monitor Cards -->
  <el-row :gutter="20" style="margin: 20px" v-if="filteredMonitors.length > 0">
    <el-col :span="6" v-for="monitor in filteredMonitors" :key="monitor.id" style="margin-bottom: 20px;">
      <el-card style="height: 300px;">
        <template #header>
          <div class="card-header" style="display: flex; flex-direction: column; gap: 12px;">
            <div style="display: flex; align-items: center; gap: 8px;">
              <el-checkbox :model-value="isSelected(monitor.id)" @change="toggleSelection(monitor.id, $event)" />
              <span style="font-size: 32px; margin: 0; flex-shrink: 0;">{{ monitor.name }}</span>
              <span>
                <el-icon size="large" color="green" v-if="monitor.auth_token"><SuccessFilled /></el-icon>
                <span v-else>
                  <el-icon size="large" color="red"><CircleCloseFilled /></el-icon>
                </span>
              </span>
            </div>
            <div style="display: flex; gap: 8px; flex-wrap: wrap;">
              <el-button size="small" @click="openProperties(monitor)">{{ $t('monitors.properties') }}</el-button>
              <el-button size="small" type="primary" plain @click="onSyncGroups(monitor)" :loading="monitor.syncing_groups">
                {{ $t('monitors.syncGroups') }}
              </el-button>
              <el-button size="small" @click="onLogin(monitor)" :loading="monitor.logging_in">
                {{ monitor.auth_token ? $t('monitors.reLogin') : $t('monitors.login') }}
              </el-button>
              <el-button size="small" @click="onDelete(monitor)">{{ $t('monitors.delete') }}</el-button>
            </div>
          </div>
        </template>
        <div class="card-body" style="margin-top: 20px;">
          <p>{{ monitor.description }}</p>
          <div style="display: flex; gap: 8px; flex-wrap: wrap; margin-top: 12px;">
            <el-tag :type="monitor.enabled === 1 ? 'success' : 'info'">
              {{ monitor.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
            </el-tag>
            <el-tooltip :content="monitor.status_reason || getStatusInfo(monitor.status).reason" placement="top">
              <el-tag :type="getStatusInfo(monitor.status).type">
                {{ getStatusInfo(monitor.status).label }}
              </el-tag>
            </el-tooltip>
          </div>
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
          <el-option label="Zabbix" :value="1" />
          <el-option label="Prometheus" :value="2" />
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

  
</template>

<script lang="ts">
import {
  Check,
  Delete,
  Edit,
  Message,
  Search,
  Star,
  Refresh, // Add Refresh icon
  SuccessFilled,
  CircleCloseFilled,
  Loading
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { ElMessageBox } from 'element-plus'
import { fetchMonitorData, addMonitor, deleteMonitor, updateMonitor, loginMonitor, regenerateMonitorEventToken, syncGroupsFromMonitor } from '@/api/monitors'

interface Monitor {
  id: number;
  name: string;
  url: string;
  username: string;
  password: string;
  auth_token: string;
  event_token?: string;
  enabled: number;
  status: number;
  status_reason?: string;
  description: string;
  type: number;
}

export default {
    name: 'Monitor',
    data() {
      return {
        monitors: [],
        createDialogVisible: false,
        propertiesDialogVisible: false,
        newMonitor: { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 },
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
          await Promise.all(this.monitors.filter((m: Monitor) => this.selectedMonitorIds.includes(m.id)).map((monitor) => {
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
          const mapped = data.map((m: any) => ({
            id: m.ID || m.id || 0,
            name: m.Name || m.name || '',
            url: m.URL || m.url || '',
            username: m.Username || m.username || '',
            password: m.Password || m.password || '',
            auth_token: m.AuthToken || m.auth_token || '',
            event_token: m.EventToken || m.event_token || '',
            enabled: m.Enabled ?? m.enabled ?? 1,
            status: m.Status ?? m.status ?? 0,
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
      openProperties(monitor: Monitor) {
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
          const idx = this.monitors.findIndex((m: Monitor) => m.id === this.selectedMonitor.id);
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
            const idx = this.monitors.findIndex((m: Monitor) => m.id === this.selectedMonitor.id);
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
      deleteMonitor(monitor: Monitor) {
        const index = this.monitors.findIndex((m: Monitor) => m.id === monitor.id);
        if (index !== -1) {
          this.monitors.splice(index, 1);
        }
      },
      async onDelete(monitor: Monitor) {
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
            const index = this.monitors.findIndex((m: Monitor) => m.id === monitor.id);
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
          
          this.newMonitor = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: '' };
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
      async onLogin(monitor: Monitor) {
        // Set loading state for this specific monitor
        monitor.logging_in = true;
        try {
          const response = await loginMonitor(monitor.id);
          
          // Update monitor with new auth token
          const updatedMonitor = response.data || response;
          const idx = this.monitors.findIndex((m: Monitor) => m.id === monitor.id);
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
      async onSyncGroups(monitor: Monitor) {
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
        this.newMonitor = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: '' };
      },
      getStatusInfo(status: number) {
        const map: Record<number, { label: string; reason: string; type: string }> = {
          0: { label: this.$t('common.statusInactive'), reason: this.$t('common.reasonInactive'), type: 'info' },
          1: { label: this.$t('common.statusActive'), reason: this.$t('common.reasonActive'), type: 'success' },
          2: { label: this.$t('common.statusError'), reason: this.$t('common.reasonError'), type: 'danger' },
          3: { label: this.$t('common.statusSyncing'), reason: this.$t('common.reasonSyncing'), type: 'warning' },
        };
        return map[status] || map[0];
      }
    }
  };
</script>

<style scoped>
.monitors-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 16px 20px 0;
}

.monitors-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.monitors-bulk-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.monitors-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 16px;
}

.selected-count {
  color: #606266;
  font-size: 13px;
}

.monitors-search {
  width: 240px;
}

.monitors-filter {
  min-width: 160px;
}

.el-row {
  margin-bottom: 20px;
}
.el-row:last-child {
  margin-bottom: 0;
}
.el-col {
  border-radius: 4px;
}

.grid-content {
  border-radius: 4px;
  min-height: 36px;
}
</style>