<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('alarms.title') }}</h1>
      <p class="page-subtitle">{{ totalAlarms }} {{ $t('dashboard.alerts') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('alarms.search')" clearable style="width: 280px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('alarms.filterStatus')" style="width: 140px">
          <el-option :label="$t('alarms.filterAll')" value="all" />
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
        <el-button @click="loadAlarms" :loading="loading" :icon="Refresh" circle />
        <el-button type="primary" :icon="Plus" @click="createDialogVisible=true">
          {{ $t('alarms.create') }}
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

  <el-dialog v-model="createDialogVisible" :title="$t('alarms.createTitle')" width="500px" align-center>
    <el-form :model="newAlarm" label-width="120px">
      <el-form-item :label="$t('alarms.name')">
        <el-input v-model="newAlarm.name" :placeholder="$t('alarms.name')" />
      </el-form-item>
      <el-form-item :label="$t('alarms.url')">
        <el-input v-model="newAlarm.url" :placeholder="$t('alarms.url')" />
      </el-form-item>
      <el-form-item :label="$t('alarms.username')">
        <el-input v-model="newAlarm.username" :placeholder="$t('alarms.username')" />
      </el-form-item>
      <el-form-item :label="$t('alarms.password')">
        <el-input v-model="newAlarm.password" type="password" :placeholder="$t('alarms.password')" show-password />
      </el-form-item>
      <el-form-item :label="$t('alarms.type')">
        <el-select v-model="newAlarm.type" style="width: 100%;">
          <el-option label="Zabbix" :value="1" />
          <el-option label="Other" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('alarms.description')">
        <el-input v-model="newAlarm.description" type="textarea" :placeholder="$t('alarms.description')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newAlarm.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('alarms.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('alarms.createBtn') }}</el-button>
    </template>
  </el-dialog>

  <div v-if="loading" style="text-align: center; padding: 40px;">
    <el-icon class="is-loading" size="50" color="#409EFF">
      <Loading />
    </el-icon>
    <p style="margin-top: 16px; color: #909399;">{{ $t('alarms.loading') }}</p>
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
      <el-button size="small" @click="loadAlarms">{{ $t('alarms.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && alarms && alarms.length === 0"
    :description="$t('alarms.noAlarms')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && alarms && alarms.length > 0 && filteredAlarms.length === 0"
    :description="$t('alarms.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="alarms-content">
    <el-row :gutter="24">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="alarm in filteredAlarms" :key="alarm.id" style="margin-bottom: 24px;">
        <el-card class="alarm-card" :body-style="{ padding: '0px' }">
          <div class="alarm-card-header">
            <div class="alarm-icon-box">
              <el-icon :size="24"><Bell /></el-icon>
            </div>
            <div class="alarm-title-area">
              <h3 class="alarm-name">{{ alarm.name }}</h3>
              <span class="alarm-type-tag">{{ alarm.type === 1 ? 'Zabbix' : 'Other' }}</span>
            </div>
            <el-checkbox :model-value="isSelected(alarm.id)" @change="toggleSelection(alarm.id, $event)" class="alarm-select" />
          </div>
          
          <div class="alarm-card-body">
            <p class="alarm-desc">{{ alarm.description || $t('alarms.noDescription') }}</p>
            <div class="alarm-status-row">
              <el-tag :type="alarm.enabled === 1 ? 'success' : 'info'" size="small">
                {{ alarm.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
              <el-tooltip :content="alarm.status_reason || getStatusInfo(alarm.status).reason" placement="top">
                <el-tag :type="getStatusInfo(alarm.status).type" size="small" effect="dark">
                  {{ getStatusInfo(alarm.status).label }}
                </el-tag>
              </el-tooltip>
            </div>
          </div>

          <div class="alarm-card-footer">
            <el-button-group>
              <el-tooltip :content="$t('alarms.properties')" placement="bottom">
                <el-button size="small" :icon="Edit" @click="openProperties(alarm)" />
              </el-tooltip>
              <el-tooltip :content="alarm.auth_token ? $t('alarms.reLogin') : $t('alarms.login')" placement="bottom">
                <el-button size="small" :type="alarm.auth_token ? 'success' : 'warning'" plain :icon="alarm.auth_token ? SuccessFilled : CircleCloseFilled" @click="onLogin(alarm)" :loading="alarm.logging_in" />
              </el-tooltip>
              <el-tooltip content="Setup Media" placement="bottom">
                <el-button size="small" type="success" plain :icon="Message" @click="onSetupMedia(alarm)" :loading="alarm.setting_up_media" v-if="alarm.type === 1" />
              </el-tooltip>
              <el-tooltip :content="$t('alarms.delete')" placement="bottom">
                <el-button size="small" type="danger" plain :icon="Delete" @click="onDelete(alarm)" />
              </el-tooltip>
            </el-button-group>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>

  <div v-if="!loading && !error && totalAlarms > 0" class="alarms-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalAlarms"
    />
  </div>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('alarms.propertiesTitle')} - ${selectedAlarm ? selectedAlarm.name : ''}`" width="600px">
    <el-form :model="selectedAlarm" label-width="120px">
      <el-form-item :label="$t('alarms.name')">
        <el-input v-model="selectedAlarm.name" />
      </el-form-item>
      <el-form-item :label="$t('alarms.url')">
        <el-input v-model="selectedAlarm.url" />
      </el-form-item>
      <el-form-item :label="$t('alarms.username')">
        <el-input v-model="selectedAlarm.username" />
      </el-form-item>
      <el-form-item :label="$t('alarms.password')">
        <el-input v-model="selectedAlarm.password" show-password />
      </el-form-item>
      <el-form-item label="Event Token">
        <el-input v-model="selectedAlarm.event_token" readonly />
        <div style="display: flex; gap: 8px; margin-top: 8px;">
          <el-button size="small" @click="copyEventToken">Copy</el-button>
          <el-button size="small" type="warning" @click="regenerateEventToken">Regenerate</el-button>
        </div>
      </el-form-item>
      <el-form-item :label="$t('alarms.type')">
        <el-select v-model="selectedAlarm.type" style="width: 100%;">
          <el-option label="Zabbix" :value="1" />
          <el-option label="Other" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('alarms.description')">
        <el-input type="textarea" v-model="selectedAlarm.description" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedAlarm.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('alarms.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('alarms.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="bulkDialogVisible" :title="$t('common.bulkUpdateTitle')" width="460px">
    <el-form :model="bulkForm" label-width="140px">
      <el-form-item :label="$t('common.enabled')">
        <el-select v-model="bulkForm.enabled" style="width: 100%;">
          <el-option :label="$t('common.bulkUpdateNoChange')" value="nochange" />
          <el-option :label="$t('common.enabled')" value="enable" />
          <el-option :label="$t('common.disabled')" value="disable" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('alarms.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('alarms.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('alarms.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedAlarms" :loading="bulkDeleting">{{ $t('alarms.delete') }}</el-button>
    </template>
  </el-dialog>
  </div>
</template>

<script lang="ts">
import {
  Refresh,
  SuccessFilled,
  CircleCloseFilled,
  Loading,
  Plus,
  Search,
  Edit,
  Delete,
  ArrowDown,
  Bell,
  Message
} from '@element-plus/icons-vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { fetchAlarmData, addAlarm, deleteAlarm, updateAlarm, loginAlarm, regenerateAlarmEventToken, setupAlarmMedia } from '@/api/alarms';

interface Alarm {
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
  logging_in?: boolean;
  setting_up_media?: boolean;
}

export default {
  name: 'Alarm',
  components: {
    Refresh,
    SuccessFilled,
    CircleCloseFilled,
    Loading,
    Plus,
    Search,
    Edit,
    Delete,
    ArrowDown,
    Bell,
    Message
  },
  data() {
    return {
      alarms: [],
      createDialogVisible: false,
      propertiesDialogVisible: false,
      newAlarm: { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 },
      selectedAlarm: { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 },
      loading: false,
      error: null,
      search: '',
      searchField: 'all',
      statusFilter: 'all',
      refreshTimer: null,
      pageSize: 20,
      currentPage: 1,
      totalAlarms: 0,
      sortKey: 'updated_desc',
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedAlarmIds: [],
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
      // Icons for template usage
      Refresh: markRaw(Refresh),
      SuccessFilled: markRaw(SuccessFilled),
      CircleCloseFilled: markRaw(CircleCloseFilled),
      Loading: markRaw(Loading),
      Plus: markRaw(Plus),
      Search: markRaw(Search),
      Edit: markRaw(Edit),
      Delete: markRaw(Delete),
      ArrowDown: markRaw(ArrowDown),
      Bell: markRaw(Bell),
      Message: markRaw(Message)
    };
  },
  computed: {
    filteredAlarms() {
      return this.alarms;
    },
    selectedCount() {
      return this.selectedAlarmIds.length;
    },
  },
  watch: {
    '$route.query.q': function () {
      this.applySearchFromQuery();
    },
    search() {
      this.currentPage = 1;
      this.loadAlarms(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadAlarms(true);
    },
    sortKey() {
      this.currentPage = 1;
      this.loadAlarms(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadAlarms(true);
    },
    currentPage() {
      this.loadAlarms();
    },
  },
  created() {
    this.applySearchFromQuery();
    this.loadAlarms(true);
  },
  mounted() {
    this.refreshTimer = setInterval(() => {
      this.loadAlarms();
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
      return this.selectedAlarmIds.includes(id);
    },
    toggleSelection(id, checked) {
      if (checked) {
        if (!this.selectedAlarmIds.includes(id)) {
          this.selectedAlarmIds.push(id);
        }
      } else {
        this.selectedAlarmIds = this.selectedAlarmIds.filter((itemId) => itemId !== id);
      }
    },
    selectAll() {
      this.selectedAlarmIds = this.alarms.map(a => a.id);
    },
    clearSelection() {
      this.selectedAlarmIds = [];
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedAlarms() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedAlarmIds.map((id) => deleteAlarm(id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.selectedAlarmIds = [];
        await this.loadAlarms(true);
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
        await Promise.all(this.alarms.filter((a: Alarm) => this.selectedAlarmIds.includes(a.id)).map((alarm) => {
          const payload = {
            enabled: enabledOverride === 'nochange' ? alarm.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? alarm.status : statusOverride,
          };
          return updateAlarm(alarm.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.selectedAlarmIds = [];
        await this.loadAlarms(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    async loadAlarms(reset = false) {
      if (reset) {
        this.alarms = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
        const response = await fetchAlarmData({
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
          : (response.data?.items || response.items || response.data || response.alarms || []);
        const total = response?.data?.total ?? response?.total ?? data.length;
        const mapped = data.map((a: any) => ({
          id: a.ID || a.id || 0,
          name: a.Name || a.name || '',
          url: a.URL || a.url || '',
          username: a.Username || a.username || '',
          password: a.Password || a.password || '',
          auth_token: a.AuthToken || a.auth_token || '',
          event_token: a.EventToken || a.event_token || '',
          enabled: a.Enabled ?? a.enabled ?? 1,
          status: a.Status ?? a.status ?? 0,
          status_reason: a.Reason || a.reason || a.Error || a.error || a.ErrorMessage || a.error_message || a.LastError || a.last_error || '',
          description: a.Description || a.description || '',
          type: a.Type || a.type || '',
        }));
        this.alarms = mapped;
        this.totalAlarms = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || 'Failed to load alarms';
        console.error('Error loading alarms:', err);
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
    openProperties(alarm: Alarm) {
      this.selectedAlarm = Object.assign({}, alarm);
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        const updateData = {
          name: this.selectedAlarm.name,
          url: this.selectedAlarm.url,
          username: this.selectedAlarm.username,
          password: this.selectedAlarm.password,
          auth_token: this.selectedAlarm.auth_token,
          type: this.selectedAlarm.type,
          description: this.selectedAlarm.description,
          enabled: this.selectedAlarm.enabled,
          status: this.selectedAlarm.status,
        };
        await updateAlarm(this.selectedAlarm.id, updateData);
        const idx = this.alarms.findIndex((a: Alarm) => a.id === this.selectedAlarm.id);
        if (idx !== -1) {
          this.alarms.splice(idx, 1, Object.assign({}, this.selectedAlarm));
        }
        this.propertiesDialogVisible = false;
        ElMessage({
          type: 'success',
          message: 'Alarm updated successfully!',
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: 'Failed to update alarm: ' + (err.message || 'Unknown error'),
        });
        console.error('Error updating alarm:', err);
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
        const response = await regenerateAlarmEventToken(this.selectedAlarm.id);
        const payload = response?.data?.data || response?.data || response;
        const token = payload?.event_token || payload?.EventToken || '';
        if (token) {
          this.selectedAlarm.event_token = token;
          const idx = this.alarms.findIndex((a: Alarm) => a.id === this.selectedAlarm.id);
          if (idx !== -1) {
            this.alarms.splice(idx, 1, Object.assign({}, this.selectedAlarm));
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
      const token = this.selectedAlarm?.event_token || '';
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
    deleteAlarm(alarm: Alarm) {
      const index = this.alarms.findIndex((a: Alarm) => a.id === alarm.id);
      if (index !== -1) {
        this.alarms.splice(index, 1);
      }
    },
    async onDelete(alarm: Alarm) {
      ElMessageBox.confirm(
        `Are you sure you want to delete ${alarm.name}?`,
        'Warning',
        {
          confirmButtonText: 'Yes',
          cancelButtonText: 'No',
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteAlarm(alarm.id);
          const index = this.alarms.findIndex((a: Alarm) => a.id === alarm.id);
          if (index !== -1) {
            this.alarms.splice(index, 1);
          }
          ElMessage({
            type: 'success',
            message: 'Alarm deleted successfully!',
          });
        } catch (err) {
          ElMessage({
            type: 'error',
            message: 'Failed to delete alarm: ' + (err.message || 'Unknown error'),
          });
          console.error('Error deleting alarm:', err);
        }
      }).catch(() => {
        ElMessage({
          type: 'info',
          message: 'Delete canceled',
        });
      });
    },
    async onCreate() {
      if (!this.newAlarm.name) {
        ElMessage({
          type: 'warning',
          message: 'Please enter alarm name',
        });
        return;
      }

      try {
        const alarmData = {
          name: this.newAlarm.name,
          url: this.newAlarm.url,
          username: this.newAlarm.username,
          password: this.newAlarm.password,
          auth_token: this.newAlarm.auth_token,
          type: this.newAlarm.type,
          description: this.newAlarm.description,
          enabled: this.newAlarm.enabled,
          status: this.newAlarm.status,
        };

        const response = await addAlarm(alarmData);
        const createdAlarm = response.data || response;
        this.alarms.push({
          ...alarmData,
          id: createdAlarm.id,
          auth_token: createdAlarm.auth_token,
          event_token: createdAlarm.event_token,
        });

        this.newAlarm = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 };
        this.createDialogVisible = false;

        if (createdAlarm.auth_token) {
          ElMessage({
            type: 'success',
            message: 'Alarm created and logged in successfully!',
          });
        } else {
          ElMessage({
            type: 'success',
            message: 'Alarm created successfully!',
          });
        }
      } catch (err) {
        ElMessage({
          type: 'error',
          message: 'Failed to create alarm: ' + (err.message || 'Unknown error'),
        });
        console.error('Error creating alarm:', err);
      }
    },
    async onLogin(alarm: Alarm) {
      alarm.logging_in = true;
      try {
        const response = await loginAlarm(alarm.id);
        const updatedAlarm = response.data || response;
        const idx = this.alarms.findIndex((a: Alarm) => a.id === alarm.id);
        if (idx !== -1) {
          this.alarms.splice(idx, 1, {
            ...alarm,
            auth_token: updatedAlarm.auth_token,
            logging_in: false,
          });
        }

        ElMessage({
          type: 'success',
          message: 'Login successful!',
        });
      } catch (err) {
        alarm.logging_in = false;
        ElMessage({
          type: 'error',
          message: 'Login failed: ' + (err.response?.data?.error || err.message || 'Unknown error'),
        });
        console.error('Error logging in to alarm:', err);
      }
    },
    async onSetupMedia(alarm: Alarm) {
      alarm.setting_up_media = true;
      try {
        const response = await setupAlarmMedia(alarm.id);
        const result = response?.data || response;
        ElMessage({
          type: 'success',
          message: result?.message || 'Zabbix initialization completed (media/user/action bound)!',
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: 'Failed to initialize Zabbix alarm integration: ' + (err.response?.data?.error || err.message || 'Unknown error'),
        });
        console.error('Error initializing Zabbix alarm integration:', err);
      } finally {
        alarm.setting_up_media = false;
      }
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newAlarm = { id: 0, name: '', url: '', username: '', password: '', auth_token: '', event_token: '', enabled: 1, status: 1, description: '', type: 1 };
    },
    getStatusInfo(status: number) {
      const map: Record<number, { label: string; reason: string; type: string }> = {
        0: { label: this.$t('common.statusInactive'), reason: this.$t('common.reasonInactive'), type: 'info' },
        1: { label: this.$t('common.statusActive'), reason: this.$t('common.reasonActive'), type: 'success' },
        2: { label: this.$t('common.statusError'), reason: this.$t('common.reasonError'), type: 'danger' },
        3: { label: this.$t('common.statusSyncing'), reason: this.$t('common.reasonSyncing'), type: 'warning' },
      };
      return map[status] || map[0];
    },
  }
};
</script>

<style scoped>
.alarms-content {
  margin-top: 8px;
}

.alarm-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.alarm-card-header {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid var(--border-1);
  position: relative;
}

.alarm-icon-box {
  width: 48px;
  height: 48px;
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.alarm-title-area {
  flex: 1;
  min-width: 0;
}

.alarm-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: var(--text-strong);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.alarm-type-tag {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
}

.alarm-select {
  position: absolute;
  top: 12px;
  right: 12px;
}

.alarm-card-body {
  padding: 20px;
  flex: 1;
}

.alarm-desc {
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

.alarm-status-row {
  display: flex;
  gap: 8px;
}

.alarm-card-footer {
  padding: 12px 20px;
  background: var(--surface-2);
  display: flex;
  justify-content: center;
  border-top: 1px solid var(--border-1);
}

.alarms-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
