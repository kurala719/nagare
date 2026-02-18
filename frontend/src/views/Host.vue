<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('hosts.search') }}</h1>
      <p class="page-subtitle">{{ $t('hosts.loading') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-select v-model="selectedColumns" multiple collapse-tags :placeholder="$t('common.columns')" style="width: 180px">
          <el-option v-for="col in columnOptions" :key="col.key" :label="col.label" :value="col.key" />
        </el-select>

        <el-input v-model="search" :placeholder="$t('hosts.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('hosts.filterStatus')" style="width: 120px">
          <el-option :label="$t('hosts.filterAll')" value="all" />
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
        <el-button type="primary" :icon="Plus" @click="createDialogVisible=true">
          {{ $t('hosts.create') }}
        </el-button>
        <el-button type="warning" :icon="Download" :disabled="(!syncMonitorId && !monitorFilter && selectedCount === 0) || pullingHosts" :loading="pullingHosts" @click="pullHosts">
          {{ $t('hosts.pullHosts') }}
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

  <!-- Create Dialog -->
  <el-dialog v-model="createDialogVisible" :title="$t('hosts.createTitle')" width="500px" align-center>
    <el-form :model="newHost" label-width="120px">
      <el-form-item :label="$t('hosts.name')">
        <el-input v-model="newHost.name" :placeholder="$t('hosts.name')" />
      </el-form-item>
      <el-form-item :label="$t('hosts.ip')">
        <el-input v-model="newHost.ip_addr" :placeholder="$t('hosts.ip')" />
      </el-form-item>
      <el-form-item :label="$t('hosts.hostId')">
        <el-input v-model="newHost.hostid" :placeholder="$t('hosts.hostId')" />
      </el-form-item>
      <el-form-item :label="$t('groups.title')">
          <el-select v-model="newHost.group_id" style="width: 100%;" clearable>
          <el-option :label="$t('hosts.filterAll')" :value="0" />
            <el-option v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('hosts.description')">
        <el-input v-model="newHost.description" type="textarea" :placeholder="$t('hosts.description')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newHost.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('hosts.status')">
        <el-select v-model="newHost.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('hosts.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('hosts.createBtn') }}</el-button>
    </template>  
  </el-dialog>

  <!-- Loading State -->
  <div v-if="loading" style="text-align: center; padding: 40px;">
    <el-icon class="is-loading" size="50" color="#409EFF">
      <Loading />
    </el-icon>
    <p style="margin-top: 16px; color: #909399;">{{ $t('hosts.loading') }}</p>
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
      <el-button size="small" @click="loadHosts">{{ $t('hosts.retry') }}</el-button>
    </template>
  </el-alert>

  <!-- Empty State -->
  <el-empty 
    v-if="!loading && !error && hosts && hosts.length === 0"
    :description="$t('hosts.noHosts')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && hosts && hosts.length > 0 && filteredHosts.length === 0"
    :description="$t('hosts.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="hosts-scroll">
  <!-- Host Table -->
  <el-table
    v-if="filteredHosts.length > 0"
    :data="filteredHosts"
    border
    style="margin: 20px; width: calc(100% - 40px);"
    ref="hostsTableRef"
    row-key="id"
    @selection-change="onSelectionChange"
    @sort-change="onSortChange"
  >
    <el-table-column type="selection" width="50" />
    <el-table-column v-if="isColumnVisible('name')" prop="name" :label="$t('hosts.name')" min-width="150" sortable="custom" />
    <el-table-column v-if="isColumnVisible('monitor')" :label="$t('hosts.monitor')" min-width="150" prop="m_id" sortable="custom">
      <template #default="{ row }">
        {{ getMonitorName(row) }}
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('group')" :label="$t('groups.title')" min-width="140" prop="group_id" sortable="custom">
      <template #default="{ row }">
          {{ getGroupName(row.group_id) }}
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('ip_addr')" prop="ip_addr" :label="$t('hosts.ip')" min-width="140" sortable="custom" />
    <el-table-column v-if="isColumnVisible('hostid')" prop="hostid" :label="$t('hosts.hostId')" min-width="120" sortable="custom" />
    <el-table-column v-if="isColumnVisible('enabled')" :label="$t('common.enabled')" min-width="110" prop="enabled" sortable="custom">
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('status')" :label="$t('hosts.status')" min-width="160" prop="status" sortable="custom">
      <template #default="{ row }">
        <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
          <el-tag :type="getStatusInfo(row.status).type">
            {{ getStatusInfo(row.status).label }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('description')" :label="$t('hosts.description')" min-width="200" show-overflow-tooltip prop="description" />
    <el-table-column :label="$t('hosts.actions')" min-width="300" fixed="right" align="center">
      <template #default="{ row }">
        <el-button-group>
          <el-tooltip :content="$t('hosts.ai')" placement="top">
            <el-button size="small" type="success" :icon="Search" @click="consultAI(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.details')" placement="top">
            <el-button size="small" type="primary" :icon="Document" @click="openDetails(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.items')" placement="top">
            <el-button size="small" type="info" :icon="Setting" @click="viewItems(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.pullItems')" placement="top">
            <el-button size="small" type="warning" plain :icon="Download" @click="pullHostItems(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.pushItems')" placement="top">
            <el-button size="small" type="success" plain :icon="Upload" @click="pushHostItems(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.properties')" placement="top">
            <el-button size="small" :icon="Edit" @click="openProperties(row)" />
          </el-tooltip>
          <el-tooltip :content="$t('hosts.delete')" placement="top">
            <el-button size="small" type="danger" :icon="Delete" @click="onDelete(row)" />
          </el-tooltip>
        </el-button-group>
      </template>
    </el-table-column>
  </el-table>
  </div>
  <div v-if="!loading && !error && totalHosts > 0" class="hosts-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalHosts"
    />
  </div>

  <!-- Properties Dialog -->
  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('hosts.propertiesTitle')} - ${selectedHost ? selectedHost.name : ''}`" width="600px">
    <el-form :model="selectedHost" label-width="120px">
      <el-form-item :label="$t('hosts.name')">
        <el-input v-model="selectedHost.name" />
      </el-form-item>
      <el-form-item :label="$t('hosts.ip')">
        <el-input v-model="selectedHost.ip_addr" />
      </el-form-item>
      <el-form-item :label="$t('hosts.hostId')">
        <el-input v-model="selectedHost.hostid" />
      </el-form-item>
      <el-form-item :label="$t('groups.title')">
        <el-select v-model="selectedHost.group_id" style="width: 100%;" clearable>
          <el-option :label="$t('hosts.filterAll')" :value="0" />
          <el-option v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('hosts.description')">
        <el-input type="textarea" v-model="selectedHost.description" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedHost.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('hosts.status')">
        <el-select v-model="selectedHost.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('hosts.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('hosts.save') }}</el-button>
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
      <el-form-item :label="$t('hosts.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('hosts.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('hosts.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedHosts" :loading="bulkDeleting">{{ $t('hosts.delete') }}</el-button>
    </template>
  </el-dialog>

  <!-- AI Response Dialog -->
  <el-dialog v-model="aiDialogVisible" :title="$t('hosts.aiTitle')" width="600px">
    <div v-if="consultingAI" style="text-align: center; padding: 40px;">
      <el-icon class="is-loading" size="40" color="#409EFF">
        <Loading />
      </el-icon>
      <p style="margin-top: 16px; color: #909399;">{{ $t('hosts.aiLoading') }}</p>
    </div>
    <div v-else>
      <el-descriptions v-if="currentHostForAI" :column="1" border style="margin-bottom: 16px;">
        <el-descriptions-item :label="$t('hosts.name')">{{ currentHostForAI.name }}</el-descriptions-item>
        <el-descriptions-item :label="$t('hosts.ip')">{{ currentHostForAI.ip_addr }}</el-descriptions-item>
      </el-descriptions>
      <el-divider content-position="left">{{ $t('hosts.aiResponse') }}</el-divider>
      <div class="ai-response-content">
        <p style="white-space: pre-wrap;">{{ aiResponse }}</p>
      </div>
    </div>
    <template #footer>
      <el-button @click="aiDialogVisible = false">{{ $t('hosts.close') }}</el-button>
    </template>
  </el-dialog>
  </div>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { fetchHostData, addHost, updateHost, deleteHost, consultHostAI, syncHostsFromMonitor, pushHostsToMonitor, pullHostFromMonitor, pushHostToMonitor } from '@/api/hosts';
import { fetchGroupData } from '@/api/groups';
import { fetchMonitorData } from '@/api/monitors';
import { pullItemsFromHost, pushItemsToHost } from '@/api/items';
import { Loading, Plus, Delete, Edit, Download, Upload, Search, Refresh, Document, Setting, ArrowDown } from '@element-plus/icons-vue';

interface Host {
  id: number;
  name: string;
  mid: number;
  group_id: number;
  hostid: string;
  description: string;
  enabled: number;
  status: number;
  ip_addr: string;
  status_reason?: string;
  monitor_name?: string;
}

export default {
  name: 'Host',
  components: {
    Loading,
    Search,
    Plus,
    Delete,
    Edit,
    Download,
    Upload,
    Refresh,
    Document,
    Setting,
    ArrowDown
  },
  data() {
    return {
      hosts: [],
      pageSize: 20,
      currentPage: 1,
      totalHosts: 0,
      sortBy: '',
      sortOrder: '',
      monitors: [],
      groups: [],
      createDialogVisible: false,
      propertiesDialogVisible: false,
      aiDialogVisible: false,
      consultingAI: false,
      currentHostForAI: null as Host | null,
      aiResponse: '',
      selectedHosts: [],
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      pullingHosts: false,
      pushingHosts: false,
      newHost: { id: 0, name: '', ip_addr: '', hostid: '', group_id: 0, description: '', enabled: 1, status: 1, mid: 0 },
      selectedHost: { id: 0, name: '', ip_addr: '', hostid: '', group_id: 0, description: '', enabled: 1, status: 1, mid: 0 },
      loading: false,
      error: null,
      search: '',
      searchField: 'all',
      selectedColumns: ['name', 'monitor', 'group', 'ip_addr', 'hostid', 'enabled', 'status', 'description'],
      statusFilter: 'all',
      monitorFilter: 0,
      groupFilter: 0,
      syncMonitorId: 0,
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
      // Icons
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
    filteredHosts() {
      return this.hosts;
    },
    columnOptions() {
      return [
        { key: 'name', label: this.$t('hosts.name') },
        { key: 'monitor', label: this.$t('hosts.monitor') },
        { key: 'group', label: this.$t('groups.title') },
        { key: 'ip_addr', label: this.$t('hosts.ip') },
        { key: 'hostid', label: this.$t('hosts.hostId') },
        { key: 'enabled', label: this.$t('common.enabled') },
        { key: 'status', label: this.$t('hosts.status') },
        { key: 'description', label: this.$t('hosts.description') },
      ];
    },
    searchableColumns() {
      return this.columnOptions.filter((col) => ['name', 'ip_addr', 'hostid', 'description', 'monitor', 'group', 'status', 'enabled'].includes(col.key));
    },
    selectedCount() {
      return this.selectedHosts.length;
    },
  },
  watch: {
    '$route.query.q': function () {
      this.applySearchFromQuery();
    },
    search() {
      this.currentPage = 1;
      this.loadHosts(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadHosts(true);
    },
    monitorFilter() {
      this.currentPage = 1;
      this.loadHosts(true);
    },
    groupFilter() {
      this.currentPage = 1;
      this.loadHosts(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadHosts(true);
    },
    currentPage() {
      this.loadHosts();
    },
  },
  created() {
    this.applySearchFromQuery();
    this.loadHosts(true);
    this.loadMonitors();
    this.loadGroups();
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
    async loadMonitors() {
      try {
        console.log('Host page loading monitors...')
        const response = await fetchMonitorData();
        console.log('Host page monitors response:', response)
        // Backend returns {success: true, data: [...]}
        let data = []
        if (response?.success && response?.data !== undefined) {
          data = Array.isArray(response.data) ? response.data : 
                 (Array.isArray(response.data.items) ? response.data.items : [])
        } else if (Array.isArray(response)) {
          data = response
        }
        const list = data;
        this.monitors = list.map((m: any) => ({
          id: Number(
            m.ID || m.Id || m.id || m.MID || m.Mid || m.mid ||
            m.MonitorID || m.MonitorId || m.monitor_id || m.monitorId || m.monitorID || 0
          ),
          name: m.Name || m.name || m.MonitorName || m.monitor_name || m.monitorName || m.Monitor?.Name || m.Monitor?.name || '',
        }));
        console.log('Host page monitors loaded:', this.monitors.length)
      } catch (err) {
        console.error('Error loading monitors:', err);
      }
    },
    async loadGroups() {
      try {
        console.log('Host page loading groups...')
        const response = await fetchGroupData();
        console.log('Host page groups response:', response)
        // Backend returns {success: true, data: [...]}
        let data = []
        if (response?.success && response?.data !== undefined) {
          data = Array.isArray(response.data) ? response.data : 
                 (Array.isArray(response.data.items) ? response.data.items : [])
        } else if (Array.isArray(response)) {
          data = response
        }
        this.groups = data.map((g: any) => ({
          id: g.ID || g.id || 0,
          name: g.Name || g.name || '',
        }));
        console.log('Host page groups loaded:', this.groups.length)
      } catch (err) {
        console.error('Error loading groups:', err);
      }
    },
    getMonitorName(target: Host | number) {
      if (typeof target !== 'number' && target?.monitor_name) return target.monitor_name;
      const mid = typeof target === 'number' ? Number(target || 0) : Number(target?.mid || 0);
      const monitor = this.monitors.find((m: any) => m.id === mid);
      if (monitor) return monitor.name;
      return mid ? `${this.$t('hosts.unknown')} (#${mid})` : this.$t('hosts.unknown');
    },
    getGroupName(groupId: number) {
      if (!groupId) return this.$t('hosts.unknown');
      const group = this.groups.find((g: any) => g.id === groupId);
      return group ? group.name : this.$t('hosts.unknown');
    },
    async loadHosts(reset = false) {
      if (reset) {
        this.hosts = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const response = await fetchHostData({
          q: this.search || undefined,
          status: this.statusFilter === 'all' ? undefined : this.statusFilter,
          m_id: this.monitorFilter || undefined,
          group_id: this.groupFilter || undefined,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
          with_total: 1,
        });
        console.log('Host page hosts response:', response)
        // Backend returns {success: true, data: [...]} or {success: true, data: {items: [...], total: N}}
        let payload = []
        let total = 0
        if (response?.success && response?.data !== undefined) {
          const data = response.data
          if (Array.isArray(data)) {
            payload = data
            total = data.length
          } else if (data.items && Array.isArray(data.items)) {
            payload = data.items
            total = data.total ?? data.items.length
          }
        } else if (Array.isArray(response)) {
          payload = response
          total = response.length
        }
        console.log('Host page extracted payload:', { count: payload.length, total })
        const mapped = payload.map((h: any) => {
          const monitorId = h.m_id || h.MID || h.Mid || h.mid || h.MonitorID || h.MonitorId || h.monitorId || h.monitor_id || h.Monitorid || h.monitorID || h.Monitor?.ID || h.Monitor?.Id || h.monitor?.id || h.monitor?.ID || 0;
          return {
            id: Number(h.ID || h.id || 0),
            name: h.Name || h.name || '',
            ip_addr: h.IPAddr || h.ip_addr || h.ip || '',
            hostid: h.HostID || h.hostid || '',
            group_id: Number(h.GroupID || h.group_id || 0),
            description: h.Description || h.description || '',
            enabled: this.normalizeEnabled(h.Enabled ?? h.enabled ?? h.ENABLED),
            status: this.normalizeStatus(h.Status ?? h.status ?? h.STATUS),
            status_reason: h.Reason || h.reason || h.Error || h.error || h.ErrorMessage || h.error_message || h.LastError || h.last_error || '',
            mid: Number(monitorId || 0),
            monitor_name: h.MonitorName || h.monitor_name || h.monitorName || h.Monitor?.Name || h.Monitor?.name || h.monitor?.Name || h.monitor?.name || '',
          };
        });
        this.hosts = mapped;
        this.totalHosts = Number.isFinite(total) ? total : mapped.length;
        console.log('Host page hosts loaded:', this.hosts.length)
      } catch (err) {
        this.error = err.message || this.$t('hosts.loadFailed') || 'Failed to load hosts';
        console.error('Error loading hosts:', err);
      } finally {
        this.loading = false;
      }
    },
    openProperties(host: Host) {
      this.selectedHost = { ...host };
      this.propertiesDialogVisible = true;
    },
    viewItems(host: Host) {
      this.$router.push({ path: '/item', query: { hostId: host.id } });
    },
    openDetails(host: Host) {
      this.$router.push({ path: `/host/${host.id}/detail` });
    },
    async pullHosts() {
      this.pullingHosts = true;
      try {
        if (this.selectedCount > 0) {
          const results = await this.batchSyncSelectedHosts('pull');
          ElMessage({
            type: results.success > 0 ? 'success' : 'warning',
            message: this.$t('hosts.pullHostsSuccess') + ` (${results.success}/${results.total}${results.skipped ? `, ${this.$t('common.skipped') || 'skipped'}: ${results.skipped}` : ''})`,
          });
        } else {
          const monitorId = this.syncMonitorId || this.monitorFilter;
          if (!monitorId) {
            ElMessage.warning(this.$t('hosts.selectMonitorFirst') || this.$t('common.selectAtLeastOne'));
            return;
          }
          const result = await syncHostsFromMonitor(monitorId);
          ElMessage({
            type: 'success',
            message: this.$t('hosts.pullHostsSuccess') + (result?.data?.added || result?.added ? ` (${result?.data?.added || result?.added})` : ''),
          });
        }
        await this.loadHosts(true);
        this.clearSelection();
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.pullHostsFailed') + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
      } finally {
        this.pullingHosts = false;
      }
    },
    async pushHosts() {
      this.pushingHosts = true;
      try {
        if (this.selectedCount > 0) {
          const results = await this.batchSyncSelectedHosts('push');
          ElMessage({
            type: results.success > 0 ? 'success' : 'warning',
            message: this.$t('hosts.pushHostsSuccess') + ` (${results.success}/${results.total}${results.skipped ? `, ${this.$t('common.skipped') || 'skipped'}: ${results.skipped}` : ''})`,
          });
        } else {
          const monitorId = this.syncMonitorId || this.monitorFilter;
          if (!monitorId) {
            ElMessage.warning(this.$t('hosts.selectMonitorFirst') || this.$t('common.selectAtLeastOne'));
            return;
          }
          const result = await pushHostsToMonitor(monitorId);
          ElMessage({
            type: 'success',
            message: this.$t('hosts.pushHostsSuccess') + (result?.data?.added || result?.added ? ` (${result?.data?.added || result?.added})` : ''),
          });
        }
        await this.loadHosts(true);
        this.clearSelection();
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.pushHostsFailed') + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
      } finally {
        this.pushingHosts = false;
      }
    },
    async batchSyncSelectedHosts(action: 'pull' | 'push') {
      const targets: Host[] = this.selectedHosts || [];
      const tasks: Array<Promise<any>> = [];
      let skipped = 0;
      targets.forEach((host) => {
        const monitorId = Number(host.mid || this.syncMonitorId || this.monitorFilter || 0);
        if (!monitorId) {
          skipped += 1;
          return;
        }
        tasks.push(action === 'pull'
          ? pullHostFromMonitor(monitorId, host.id)
          : pushHostToMonitor(monitorId, host.id));
      });
      const results = await Promise.allSettled(tasks);
      const success = results.filter((result) => result.status === 'fulfilled').length;
      return { total: tasks.length + skipped, success, skipped };
    },
    async pullHostItems(host: Host) {
      try {
        if (!host.mid) {
          ElMessage({
            type: 'warning',
            message: this.$t('hosts.pullItemsFailed', { name: host.name }),
          });
          return;
        }
        await pullItemsFromHost(host.mid, host.id);
        ElMessage({
          type: 'success',
          message: this.$t('hosts.pullItemsSuccess', { name: host.name }),
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.pullItemsFailed', { name: host.name }) + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
      }
    },
    async pushHostItems(host: Host) {
      try {
        if (!host.mid) {
          ElMessage({
            type: 'warning',
            message: this.$t('hosts.pushItemsFailed', { name: host.name }),
          });
          return;
        }
        await pushItemsToHost(host.mid, host.id);
        ElMessage({
          type: 'success',
          message: this.$t('hosts.pushItemsSuccess', { name: host.name }),
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.pushItemsFailed', { name: host.name }) + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
      }
    },
    onSelectionChange(selection) {
      this.selectedHosts = selection || [];
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
      this.loadHosts(true);
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedHosts() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedHosts.map((host: Host) => deleteHost(host.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadHosts(true);
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
        await Promise.all(this.selectedHosts.map((host: Host) => {
          const payload = {
            name: host.name,
            ip_addr: host.ip_addr,
            hostid: host.hostid,
            group_id: host.group_id,
            description: host.description,
            enabled: enabledOverride === 'nochange' ? host.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? host.status : statusOverride,
          };
          return updateHost(host.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadHosts(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.hostsTableRef && this.$refs.hostsTableRef.clearSelection) {
        this.$refs.hostsTableRef.clearSelection();
      }
      this.selectedHosts = [];
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        const updateData = {
          name: this.selectedHost.name,
          ip_addr: this.selectedHost.ip_addr,
          hostid: this.selectedHost.hostid,
          group_id: this.selectedHost.group_id,
          description: this.selectedHost.description,
          enabled: this.selectedHost.enabled,
          status: this.selectedHost.status,
        };
        await updateHost(this.selectedHost.id, updateData);
        const idx = this.hosts.findIndex((h: Host) => h.id === this.selectedHost.id);
        if (idx !== -1) {
          this.hosts.splice(idx, 1, { ...this.selectedHost });
        }
        this.propertiesDialogVisible = false;
        ElMessage({
          type: 'success',
          message: this.$t('hosts.updated'),
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.updateFailed') + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
        console.error('Error updating host:', err);
      }
    },
    onDelete(host: Host) {
      ElMessageBox.confirm(
        this.$t('hosts.deleteConfirmText', { name: host.name }),
        this.$t('hosts.deleteConfirmTitle'),
        {
          confirmButtonText: this.$t('hosts.confirm'),
          cancelButtonText: this.$t('hosts.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteHost(host.id);
          const index = this.hosts.findIndex((h: Host) => h.id === host.id);
          if (index !== -1) {
            this.hosts.splice(index, 1);
          }
          ElMessage({
            type: 'success',
            message: this.$t('hosts.deleted'),
          });
        } catch (err) {
          ElMessage({
            type: 'error',
            message: this.$t('hosts.deleteFailed') + ': ' + (err.message || this.$t('hosts.unknownError')),
          });
          console.error('Error deleting host:', err);
        }
      }).catch(() => {
        ElMessage({
          type: 'info',
          message: this.$t('hosts.deleteCanceled'),
        });
      });
    },
    async onCreate() {
      if (!this.newHost.name) {
        ElMessage({
          type: 'warning',
          message: this.$t('hosts.validationName'),
        });
        return;
      }
      
      try {
        const hostData = {
          name: this.newHost.name,
          ip_addr: this.newHost.ip_addr,
          hostid: this.newHost.hostid,
          group_id: this.newHost.group_id,
          description: this.newHost.description,
          enabled: this.newHost.enabled,
          status: this.newHost.status,
        };
        
        // Call API to add host to database
        await addHost(hostData);
        
        // Reload hosts from database to get the updated list
        await this.loadHosts(true);
        
        this.newHost = { id: 0, name: '', ip_addr: '', hostid: '', group_id: 0, description: '', enabled: 1, status: 1, mid: 0 };
        this.createDialogVisible = false;
        ElMessage({
          type: 'success',
          message: this.$t('hosts.created'),
        });
      } catch (err) {
        ElMessage({
          type: 'error',
          message: this.$t('hosts.createFailed') + ': ' + (err.message || this.$t('hosts.unknownError')),
        });
        console.error('Error creating host:', err);
      }
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newHost = { id: 0, name: '', ip_addr: '', hostid: '', group_id: 0, description: '', enabled: 1, status: 1, mid: 0 };
    },
    async consultAI(host: Host) {
      this.currentHostForAI = host;
      this.aiResponse = '';
      this.aiDialogVisible = true;
      this.consultingAI = true;
      
      try {
        const response = await consultHostAI(host.id);
        console.log('AI response:', response);
        // Handle different response formats
        if (typeof response === 'string') {
          this.aiResponse = response;
        } else if (response.message) {
          this.aiResponse = response.message;
        } else if (response.content) {
          this.aiResponse = response.content;
        } else if (response.data) {
          if (typeof response.data === 'string') {
            this.aiResponse = response.data;
          } else if (response.data.message) {
            this.aiResponse = response.data.message;
          } else if (response.data.content) {
            this.aiResponse = response.data.content;
          } else if (response.data.Content) {
            this.aiResponse = response.data.Content;
          } else {
            this.aiResponse = JSON.stringify(response.data, null, 2);
          }
        } else {
          this.aiResponse = JSON.stringify(response, null, 2);
        }
      } catch (err) {
        this.aiResponse = this.$t('hosts.aiError') + ': ' + (err.message || this.$t('hosts.unknownError'));
        ElMessage({
          type: 'error',
          message: err.message || this.$t('hosts.aiFailed'),
        });
        console.error('Error consulting AI:', err);
      } finally {
        this.consultingAI = false;
      }
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
    normalizeStatus(value: any) {
      if (value === null || value === undefined || value === '') return 0;
      if (typeof value === 'boolean') return value ? 1 : 0;
      const num = Number(value);
      return Number.isNaN(num) ? 0 : num;
    },
    normalizeEnabled(value: any) {
      if (value === null || value === undefined || value === '') return 1;
      if (typeof value === 'boolean') return value ? 1 : 0;
      const num = Number(value);
      return Number.isNaN(num) ? 1 : num;
    },
    isColumnVisible(key: string) {
      return this.selectedColumns.includes(key);
    }
  }
};
</script>

<style scoped>
.hosts-scroll {
  margin-top: 8px;
}

.hosts-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.ai-response-content {
  background: var(--surface-2);
  border-radius: var(--radius-md);
  padding: 20px;
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.7;
  border: 1px solid var(--border-1);
}

.ai-response-content p {
  margin: 0;
  color: var(--text-strong);
}

:deep(.el-table__row) {
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  background-color: var(--brand-50) !important;
}
</style>
