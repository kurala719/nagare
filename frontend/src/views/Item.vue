<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="header-main">
        <h1 class="page-title">{{ titleLabel }}</h1>
        <div class="header-info">
          <p class="page-subtitle">{{ totalItems }} {{ $t('items.total') }}</p>
          <div class="refresh-info" v-if="lastUpdated">
            <span class="last-updated">{{ $t('dashboard.summaryLastUpdated') }}: {{ lastUpdated }}</span>
            <el-tag v-if="autoRefreshEnabled" size="small" type="success" effect="plain" class="auto-refresh-tag">
              <el-icon class="is-loading"><Refresh /></el-icon>
              Auto-refreshing (30s)
            </el-tag>
          </div>
        </div>
      </div>
      <div class="header-actions">
        <el-switch
          v-model="autoRefreshEnabled"
          style="margin-right: 16px"
          :active-text="$t('common.autoRefresh') || 'Auto-refresh'"
          @change="handleAutoRefreshChange"
        />
        <el-button type="primary" @click="loadItems(true)" :loading="loading" :icon="Refresh">
          {{ $t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-select v-model="selectedColumns" multiple collapse-tags :placeholder="$t('common.columns')" style="width: 180px">
          <el-option v-for="col in columnOptions" :key="col.key" :label="col.label" :value="col.key" />
        </el-select>

        <el-input v-model="search" :placeholder="$t('items.search')" clearable style="width: 320px" class="search-with-select">
          <template #prepend>
            <el-select v-model="searchField" style="width: 110px">
              <el-option :label="$t('monitors.searchAll') || 'All'" value="all" />
              <el-option :label="$t('items.name')" value="name" />
              <el-option :label="$t('items.value')" value="value" />
              <el-option :label="$t('items.comment')" value="comment" />
            </el-select>
          </template>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="hostFilter" :placeholder="$t('items.filterHost')" style="width: 160px" clearable>
          <el-option :label="$t('items.filterAll')" :value="0" />
          <el-option v-for="host in hostOptions" :key="host.id" :label="host.name" :value="host.id" />
        </el-select>

        <el-select v-model="statusFilter" :placeholder="$t('items.filterStatus')" style="width: 120px">
          <el-option :label="$t('items.filterAll')" value="all" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button type="primary" :icon="Plus" @click="openAddDialog">
          {{ $t('items.add') }}
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

    <!-- Loading State -->
    <div v-if="loading" style="text-align: center; padding: 40px;">
        <el-icon class="is-loading" size="50" color="#409EFF">
            <Loading />
        </el-icon>
        <p style="margin-top: 16px; color: #909399;">{{ $t('items.loading') }}</p>
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
            <el-button size="small" @click="loadItems">{{ $t('items.retry') }}</el-button>
        </template>
    </el-alert>

    <!-- Empty State -->
    <el-empty 
        v-if="!loading && !error && items && items.length === 0"
        :description="$t('items.noItems')"
        style="margin: 40px;"
    />

    <el-empty
        v-if="!loading && !error && items && items.length > 0 && filteredItems.length === 0"
        :description="$t('items.noResults')"
        style="margin: 40px;"
    />

    <div v-if="!loading && !error" class="items-scroll">
      <el-table
        v-if="filteredItems.length > 0"
        ref="itemsTableRef"
        :data="filteredItems"
        border
        style="width: 100%; border-radius: 4px; overflow: hidden; box-shadow: 0 1px 4px rgba(0,0,0,0.05);"
        row-key="id"
        @selection-change="onSelectionChange"
        @sort-change="onSortChange"
        header-cell-class-name="table-header"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column v-if="isColumnVisible('id')" prop="id" :label="$t('items.id')" width="80" sortable="custom" />
        <el-table-column v-if="isColumnVisible('name')" prop="name" :label="$t('items.name')" min-width="150" show-overflow-tooltip sortable="custom" />
        <el-table-column v-if="isColumnVisible('value')" prop="value" :label="$t('items.value')" min-width="150" show-overflow-tooltip sortable="custom" />
        <el-table-column v-if="isColumnVisible('units')" prop="units" :label="$t('items.units')" width="100" show-overflow-tooltip sortable="custom" />
        <el-table-column v-if="isColumnVisible('enabled')" :label="$t('common.enabled')" width="110" align="center" prop="enabled" sortable="custom">
            <template #default="{ row }">
                <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small" effect="light">
                    {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
                </el-tag>
            </template>
        </el-table-column>
        <el-table-column v-if="isColumnVisible('status')" prop="status" :label="$t('items.status')" width="160" align="center" sortable="custom">
            <template #default="{ row }">
                <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
                    <el-tag :type="getStatusInfo(row.status).type" size="small" effect="dark">
                        {{ getStatusInfo(row.status).label }}
                    </el-tag>
                </el-tooltip>
            </template>
        </el-table-column>
        <el-table-column v-if="isColumnVisible('lastSync')" :label="$t('hosts.lastSync')" min-width="180" prop="last_sync_at" sortable="custom">
            <template #default="{ row }">
                {{ row.last_sync_at ? new Date(row.last_sync_at).toLocaleString() : '-' }}
            </template>
        </el-table-column>
        <el-table-column v-if="isColumnVisible('externalSource')" :label="$t('hosts.externalSource')" min-width="140" prop="external_source" sortable="custom" />
        <el-table-column v-if="isColumnVisible('description')" prop="description" :label="$t('items.description')" min-width="200" show-overflow-tooltip />
        <el-table-column :label="$t('items.actions')" width="260" fixed="right" align="center">
            <template #default="{ row }">
              <el-button-group>
                <el-tooltip :content="$t('items.ai')" placement="top">
                  <el-button size="small" type="success" :icon="Search" @click="consultAI(row)" />
                </el-tooltip>
                <el-tooltip :content="$t('items.details')" placement="top">
                  <el-button size="small" type="primary" :icon="Document" @click="openDetails(row)" />
                </el-tooltip>
                <el-tooltip :content="$t('items.edit')" placement="top">
                  <el-button size="small" type="primary" :icon="Edit" @click="openEditDialog(row)" />
                </el-tooltip>
                <el-tooltip :content="$t('items.delete')" placement="top">
                  <el-button size="small" type="danger" :icon="Delete" @click="confirmDelete(row)" />
                </el-tooltip>
              </el-button-group>
            </template>
        </el-table-column>
      </el-table>
    </div>
    
    <div v-if="!loading && !error && totalItems > 0" class="items-pagination">
        <el-pagination
            background
            layout="sizes, prev, pager, next"
            :page-sizes="[10, 20, 50, 100]"
            v-model:page-size="pageSize"
            v-model:current-page="currentPage"
            :total="totalItems"
        />
    </div>

        <el-dialog 
            v-model="dialogVisible" 
            :title="isEditing ? $t('items.editTitle') : $t('items.addTitle')"
            width="550px"
        >
            <el-form :model="itemForm" label-width="100px" :rules="formRules" ref="itemFormRef">
                <el-form-item :label="$t('items.host')" prop="host_id">
                    <el-select v-model="itemForm.host_id" style="width: 100%;" placeholder="Select Host">
                        <el-option v-for="host in hostOptions" :key="host.id" :label="host.name" :value="host.id" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('items.name')" prop="name">
                    <el-input v-model="itemForm.name" :placeholder="$t('items.enterName')" />
                </el-form-item>
                <el-form-item :label="$t('items.value')" prop="value">
                    <el-input v-model="itemForm.value" :placeholder="$t('items.enterValue')" />
                </el-form-item>
                <el-form-item :label="$t('common.enabled')" prop="enabled">
                    <el-switch v-model="itemForm.enabled" :active-value="1" :inactive-value="0" />
                </el-form-item>
                <el-form-item :label="$t('items.description')" prop="description">
                    <el-input v-model="itemForm.description" type="textarea" :rows="3" :placeholder="$t('items.enterDescription')" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">{{ $t('items.cancel') }}</el-button>
                <el-button @click="saveItem(false)">{{ $t('common.saveLocally') }}</el-button>
                <el-button type="primary" @click="saveItem(true)" :loading="saving">
                    {{ $t('common.saveAndPush') }}
                </el-button>
            </template>
        </el-dialog>

        <!-- Delete Confirmation Dialog -->
        <el-dialog v-model="deleteDialogVisible" :title="$t('items.confirmDelete')" width="420px">
            <p>{{ $t('items.confirmDeleteText') }}</p>
            <p v-if="itemToDelete"><strong>{{ itemToDelete.name }}</strong></p>
            <template #footer>
                <el-button @click="deleteDialogVisible = false">{{ $t('items.cancel') }}</el-button>
                <el-button type="danger" plain @click="deleteItemConfirmed(false)" :loading="deleting">{{ $t('common.saveLocally') || 'Delete Locally' }}</el-button>
                <el-button type="danger" @click="deleteItemConfirmed(true)" :loading="deleting">{{ $t('common.saveAndPush') || 'Delete and Push' }}</el-button>
            </template>
        </el-dialog>

        <!-- Bulk Update Dialog -->
        <el-dialog v-model="bulkDialogVisible" :title="$t('items.bulkUpdateTitle')" width="460px">
            <el-form :model="bulkForm" label-width="140px">
                <el-form-item :label="$t('items.bulkUpdateEnabled')">
                    <el-select v-model="bulkForm.enabled" style="width: 100%;">
                        <el-option :label="$t('items.bulkUpdateNoChange')" value="nochange" />
                        <el-option :label="$t('common.enabled')" value="enable" />
                        <el-option :label="$t('common.disabled')" value="disable" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="bulkDialogVisible = false">{{ $t('items.cancel') }}</el-button>
                <el-button @click="applyBulkUpdate(false)" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
                <el-button type="primary" @click="applyBulkUpdate(true)" :loading="bulkUpdating">{{ $t('common.saveAndPush') }}</el-button>
            </template>
        </el-dialog>

        <!-- Bulk Delete Confirmation Dialog -->
        <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('items.bulkDeleteConfirmTitle')" width="420px">
            <p>{{ $t('items.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
            <template #footer>
                <el-button @click="bulkDeleteDialogVisible = false">{{ $t('items.cancel') }}</el-button>
                <el-button type="danger" plain @click="deleteSelectedItems(false)" :loading="bulkDeleting">{{ $t('common.saveLocally') || 'Delete Locally' }}</el-button>
                <el-button type="danger" @click="deleteSelectedItems(true)" :loading="bulkDeleting">{{ $t('common.saveAndPush') || 'Delete and Push' }}</el-button>
            </template>
        </el-dialog>

        <!-- AI Response Dialog -->
        <el-dialog v-model="aiDialogVisible" :title="$t('items.aiTitle')" width="600px">
            <div v-if="consultingAI && !aiResponse" style="text-align: center; padding: 40px;">
                <el-icon class="is-loading" size="40" color="#409EFF">
                    <Loading />
                </el-icon>
                <p style="margin-top: 16px; color: #909399;">{{ $t('items.aiLoading') }}</p>
            </div>
            <div v-else>
                <el-descriptions v-if="currentItemForAI" :column="1" border style="margin-bottom: 16px;">
                    <el-descriptions-item :label="$t('items.name')">{{ currentItemForAI.name }}</el-descriptions-item>
                    <el-descriptions-item :label="$t('items.value')">{{ currentItemForAI.value }}</el-descriptions-item>
                </el-descriptions>

                <div class="provider-selector-row" style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px; padding: 12px; background: var(--el-fill-color-extra-light); border-radius: 8px;">
                  <span style="font-weight: 600; font-size: 13px;">{{ $t('system.aiProviderId') || 'AI Provider' }}:</span>
                  <el-select v-model="selectedProviderId" @change="onProviderChange" style="width: 160px;" size="small">
                    <el-option v-for="p in aiProviders" :key="p.id" :label="p.name" :value="p.id" />
                  </el-select>
                  <el-select v-if="availableModels.length > 0" v-model="selectedModel" style="width: 160px;" size="small" placeholder="Model">
                    <el-option v-for="m in availableModels" :key="m" :label="m" :value="m" />
                  </el-select>
                  <el-button type="primary" size="small" :icon="Search" @click="consultAI(currentItemForAI, true)" :loading="consultingAI">
                    {{ $t('common.retry') || 'Re-analyze' }}
                  </el-button>
                </div>

                <el-divider content-position="left">{{ $t('items.aiResponse') }}</el-divider>
                <div class="ai-response-content" v-if="aiResponse">
                    <p style="white-space: pre-wrap;">{{ aiResponse }}</p>
                </div>
            </div>
        </el-dialog>
    </div>
</template>

<script>
import { fetchItemData, addItem, updateItem, deleteItem, consultItemAI, pullItemsFromHost, pushItemsToHost } from '@/api/items';
import { fetchHostData } from '@/api/hosts';
import { fetchProviderData } from '@/api/providers';
import { getMainConfig } from '@/api/config';
import { ElMessage } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Plus, Delete, Edit, Download, Upload, Search, Refresh, Document, Setting, ArrowDown } from '@element-plus/icons-vue';

export default {
    name: 'Item',
    components: {
        Loading,
        Plus,
        Search,
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
                items: [],
                hosts: [],
            pageSize: 20,
            currentPage: 1,
            totalItems: 0,
            sortBy: '',
            sortOrder: '',
        loading: false,
        saving: false,
        deleting: false,
        bulkDeleting: false,
        pushing: false,
        consultingAI: false,
        error: null,
        dialogVisible: false,
        deleteDialogVisible: false,
        bulkDialogVisible: false,
        bulkDeleteDialogVisible: false,
        aiDialogVisible: false,
        isEditing: false,
        editingId: null,
        itemToDelete: null,
        selectedItems: [],
        bulkUpdating: false,
        currentItemForAI: null,
        aiResponse: '',
        aiProviders: [],
        availableModels: [],
        selectedProviderId: 1,
        selectedModel: '',
        itemForm: {
            name: '',
            value: '',
            type: '',
            interval: 60,
            enabled: 1,
            description: '',
            host_id: null,
        },
        formRules: {},
        search: '',
        searchField: 'all',
        selectedColumns: ['id', 'name', 'value', 'units', 'enabled', 'status', 'lastSync', 'externalSource', 'description'],
        columnOptions: [
            { key: 'id', label: this.$t('items.id') },
            { key: 'name', label: this.$t('items.name') },
            { key: 'value', label: this.$t('items.value') },
            { key: 'units', label: this.$t('items.units') },
            { key: 'enabled', label: this.$t('common.enabled') },
            { key: 'status', label: this.$t('items.status') },
            { key: 'lastSync', label: this.$t('hosts.lastSync') },
            { key: 'externalSource', label: this.$t('hosts.externalSource') },
            { key: 'description', label: this.$t('items.description') },
        ],
        hostFilter: 0,
        statusFilter: 'all',
                lastUpdated: '',
                autoRefreshEnabled: true,
                refreshInterval: null,
                bulkForm: {
                        enabled: 'nochange',
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
        filteredItems() {
            return this.items;
        },
        searchableColumns() {
            return this.columnOptions;
        },
        hostOptions() {
            return this.hosts;
        },
        titleLabel() {
            if (this.hostFilter) {
                const host = this.hosts.find((h) => h.id === this.hostFilter);
                return host ? this.$t('items.titleHostName', { name: host.name }) : this.$t('items.titleHost', { id: this.hostFilter });
            }
            return this.$t('items.titleAll');
        },
        selectedCount() {
            return this.selectedItems.length;
        },
    },
    created() {
        this.formRules = {
            name: [{ required: true, message: this.$t('items.validationName'), trigger: 'blur' }],
            host_id: [{ required: true, message: 'Host is required', trigger: 'change' }],
        };
        this.applySearchFromQuery();
        this.applyHostFromQuery();
        this.loadHosts();
        this.loadItems(true);
        this.loadAIProviders();
        if (this.autoRefreshEnabled) {
            this.startAutoRefresh();
        }
    },
    beforeUnmount() {
        this.stopAutoRefresh();
    },
    watch: {
        '$route.query.q': function () {
            this.applySearchFromQuery();
        },
        '$route.query.hostId': function () {
            this.applyHostFromQuery();
        },
        hostFilter(newVal) {
            const next = newVal ? String(newVal) : undefined;
            if (this.$route.query.hostId !== next) {
                const nextQuery = { ...this.$route.query };
                if (next) {
                    nextQuery.hostId = next;
                } else {
                    delete nextQuery.hostId;
                }
                this.$router.replace({ query: nextQuery });
            }
            this.currentPage = 1;
            this.loadItems(true);
        },
        search() {
            this.currentPage = 1;
            this.loadItems(true);
        },
        searchField() {
            this.currentPage = 1;
            this.loadItems(true);
        },
        statusFilter() {
            this.currentPage = 1;
            this.loadItems(true);
        },
        pageSize() {
            this.currentPage = 1;
            this.loadItems(true);
        },
        currentPage() {
            this.loadItems();
        },
    },
    methods: {
        startAutoRefresh() {
            this.stopAutoRefresh();
            this.refreshInterval = setInterval(() => {
                if (!this.loading) {
                    this.loadItems();
                }
            }, 30000);
        },
        stopAutoRefresh() {
            if (this.refreshInterval) {
                clearInterval(this.refreshInterval);
                this.refreshInterval = null;
            }
        },
        handleAutoRefreshChange(val) {
            if (val) {
                this.startAutoRefresh();
            } else {
                this.stopAutoRefresh();
            }
        },
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
        applyHostFromQuery() {
            const queryHostId = this.toNumber(this.$route.query.hostId, 0);
            this.hostFilter = queryHostId;
        },
        async loadHosts() {
            try {
                const response = await fetchHostData();
                const data = Array.isArray(response) ? response : (response.data || response.hosts || []);
                this.hosts = data.map((h) => ({
                    id: h.ID || h.id || 0,
                    name: h.Name || h.name || '',
                    mid: h.m_id || h.MID || h.Mid || h.mid || h.MonitorID || h.monitorId || h.monitor_id || 0,
                }));
            } catch (err) {
                console.error('Error loading hosts:', err);
            }
        },
        async loadItems(reset = false) {
            if (reset) {
                this.items = [];
            }
            this.loading = reset;
            this.error = null;
            try {
                const response = await fetchItemData({
                    q: this.search || undefined,
                    search_field: this.searchField !== 'all' ? this.searchField : undefined,
                    hid: this.hostFilter || undefined,
                    status: this.statusFilter === 'all' ? undefined : this.statusFilter,
                    limit: this.pageSize,
                    offset: (this.currentPage - 1) * this.pageSize,
                    sort: this.sortBy || undefined,
                    order: this.sortOrder || undefined,
                    with_total: 1,
                });
                const data = Array.isArray(response)
                    ? response
                    : (response.data?.items || response.items || response.data || response.items || []);
                const total = response?.data?.total ?? response?.total ?? data.length;
                if (!Array.isArray(data)) {
                    this.items = [];
                    return;
                }
                const mapped = data.map((item) => ({
                    id: item.ID || item.id,
                    name: item.Name || item.name || '',
                    value: item.Value || item.value || item.LastValue || item.last_value || '',
                    units: item.units || item.Units || '',
                    enabled: this.normalizeEnabled(item.Enabled ?? item.enabled ?? item.ENABLED),
                    status: this.normalizeStatus(item.Status ?? item.status ?? item.STATUS),
                    status_reason: item.Reason || item.reason || item.Error || item.error || item.ErrorMessage || item.error_message || item.LastError || item.last_error || item.Comment || item.comment || '',
                    description: item.Comment || item.comment || item.Description || item.description || '',
                    host_id: item.HID || item.hid || item.HostID || item.host_id || item.hostId || null,
                    last_sync_at: item.last_sync_at,
                    external_source: item.external_source || '',
                }));
                this.items = mapped;
                this.totalItems = Number.isFinite(total) ? total : mapped.length;
                this.lastUpdated = new Date().toLocaleString();
            } catch (err) {
                this.error = err.message || this.$t('items.loadFailed');
                console.error('Error loading items:', err);
            } finally {
                this.loading = false;
            }
        },
        openAddDialog() {
            this.isEditing = false;
            this.editingId = null;
            this.itemForm = {
                name: '',
                value: '',
                enabled: 1,
                description: '',
                host_id: this.hostFilter || null,
            };
            this.dialogVisible = true;
        },
        openEditDialog(item) {
            this.isEditing = true;
            this.editingId = item.id;
            this.itemForm = {
                name: item.name,
                value: item.value,
                enabled: item.enabled,
                description: item.description,
                host_id: item.host_id ?? null,
            };
            this.dialogVisible = true;
        },
        async saveItem(pushToMonitor = false) {
            try {
                await this.$refs.itemFormRef.validate();
            } catch {
                return;
            }

            this.saving = true;
            try {
                const payload = {
                    name: this.itemForm.name,
                    value: this.itemForm.value,
                    enabled: this.itemForm.enabled,
                    comment: this.itemForm.description,
                    hid: this.itemForm.host_id,
                    push_to_monitor: pushToMonitor
                };
                if (this.isEditing) {
                    await updateItem(this.editingId, payload);
                    ElMessage.success(this.$t('items.updated'));
                } else {
                    await addItem(payload);
                    ElMessage.success(this.$t('items.created'));
                }
                this.dialogVisible = false;
                await this.loadItems(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('items.saveFailed'));
                console.error('Error saving item:', err);
            } finally {
                this.saving = false;
            }
        },
        confirmDelete(item) {
            this.itemToDelete = item;
            this.deleteDialogVisible = true;
        },
        async deleteItemConfirmed(push = false) {
            if (!this.itemToDelete) return;
            
            this.deleting = true;
            try {
                await deleteItem(this.itemToDelete.id, push);
                ElMessage.success(this.$t('items.deleted'));
                this.deleteDialogVisible = false;
                this.itemToDelete = null;
                await this.loadItems(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('items.deleteFailed'));
                console.error('Error deleting item:', err);
            } finally {
                this.deleting = false;
            }
        },
        async pushItems() {
            this.pushing = true;
            try {
                if (this.selectedCount > 0) {
                    const results = await this.batchSyncSelectedItems('push');
                    ElMessage({
                        type: results.success > 0 ? 'success' : 'warning',
                        message: this.$t('items.pushSuccess') + ` (${results.success}/${results.total}${results.skipped ? `, ${this.$t('common.skipped') || 'skipped'}: ${results.skipped}` : ''})`,
                    });
                } else {
                    if (!this.hostFilter) {
                        ElMessage.warning(this.$t('items.pushSelectHost'));
                        return;
                    }
                    const host = this.hosts.find((h) => h.id === this.hostFilter);
                    if (!host || !host.mid) {
                        ElMessage.error(this.$t('items.pushFailed'));
                        return;
                    }
                    await pushItemsToHost(host.mid, this.hostFilter);
                    ElMessage.success(this.$t('items.pushSuccess'));
                }
                await this.loadItems(true);
                this.clearSelection();
            } catch (err) {
                ElMessage.error(err.message || this.$t('items.pushFailed'));
            } finally {
                this.pushing = false;
            }
        },
        async batchSyncSelectedItems(action) {
            const hostMap = new Map(this.hosts.map((host) => [this.toNumber(host.id, 0), host]));
            const hostIds = Array.from(new Set(
                this.selectedItems
                    .map((item) => this.toNumber(item.host_id, 0))
                    .filter((id) => id)
            ));
            const tasks = [];
            let skipped = 0;
            hostIds.forEach((hostId) => {
                const host = hostMap.get(hostId);
                const monitorId = this.toNumber(host?.mid, 0);
                if (!hostId || !monitorId) {
                    skipped += 1;
                    return;
                }
                tasks.push(action === 'pull'
                    ? pullItemsFromHost(monitorId, hostId)
                    : pushItemsToHost(monitorId, hostId));
            });
            const results = await Promise.allSettled(tasks);
            const success = results.filter((result) => result.status === 'fulfilled').length;
            return { total: tasks.length + skipped, success, skipped };
        },
        onSelectionChange(selection) {
            this.selectedItems = selection || [];
        },
        selectAll() {
            if (this.$refs.itemsTableRef) {
                this.items.forEach((row) => {
                    this.$refs.itemsTableRef.toggleRowSelection(row, true);
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
            this.loadItems(true);
        },
        openBulkDeleteDialog() {
            if (this.selectedCount === 0) {
                ElMessage.warning(this.$t('items.selectAtLeastOne'));
                return;
            }
            this.bulkDeleteDialogVisible = true;
        },
        async deleteSelectedItems(push = false) {
            if (this.selectedCount === 0) return;

            this.bulkDeleting = true;
            try {
                await Promise.all(this.selectedItems.map((item) => deleteItem(item.id, push)));
                ElMessage.success(this.$t('items.bulkDeleteSuccess', { count: this.selectedCount }));
                this.bulkDeleteDialogVisible = false;
                this.clearSelection();
                await this.loadItems(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('items.bulkDeleteFailed'));
            } finally {
                this.bulkDeleting = false;
            }
        },
        openBulkUpdateDialog() {
            if (this.selectedCount === 0) {
                ElMessage.warning(this.$t('items.selectAtLeastOne'));
                return;
            }
            this.bulkForm = {
                enabled: 'nochange',
                status: 'nochange',
            };
            this.bulkDialogVisible = true;
        },
            async applyBulkUpdate(pushToMonitor = false) {
            if (this.selectedCount === 0) return;
            if (this.bulkForm.enabled === 'nochange') {
                ElMessage.warning(this.$t('items.bulkUpdateNoChanges'));
                return;
            }

            this.bulkUpdating = true;
            try {
                const enabledOverride = this.bulkForm.enabled;
                await Promise.all(this.selectedItems.map((item) => {
                    const payload = {
                        name: item.name,
                        value: item.value,
                        enabled: enabledOverride === 'nochange' ? item.enabled : (enabledOverride === 'enable' ? 1 : 0),
                        comment: item.description,
                        hid: item.host_id,
                        push_to_monitor: pushToMonitor
                    };
                    return updateItem(item.id, payload);
                }));
                ElMessage.success(this.$t('items.bulkUpdateSuccess', { count: this.selectedCount }));
                this.bulkDialogVisible = false;
                this.clearSelection();
                await this.loadItems(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('items.bulkUpdateFailed'));
            } finally {
                this.bulkUpdating = false;
            }
        },
        clearSelection() {
            if (this.$refs.itemsTableRef && this.$refs.itemsTableRef.clearSelection) {
                this.$refs.itemsTableRef.clearSelection();
            }
            this.selectedItems = [];
        },
        openDetails(item) {
            this.$router.push({ path: `/item/${item.id}/detail` });
        },
        async loadAIProviders() {
            try {
                const configRes = await getMainConfig();
                const config = configRes.data?.data || configRes.data || configRes;
                if (config.ai?.provider_id) {
                    this.selectedProviderId = config.ai.provider_id;
                    this.selectedModel = config.ai.model || '';
                }

                const res = await fetchProviderData({ enabled: 1 });
                const list = res.data?.items || res.items || res.data || [];
                this.aiProviders = list.map(p => ({
                    id: p.ID || p.id,
                    name: p.name || p.Name,
                    models: p.models || p.Models || []
                }));

                this.onProviderChange(this.selectedProviderId);
            } catch (err) {
                console.error('Failed to load AI providers', err);
            }
        },
        onProviderChange(providerId) {
            const provider = this.aiProviders.find(p => p.id === providerId);
            if (provider) {
                this.availableModels = provider.models || [];
                if (this.selectedModel && !this.availableModels.includes(this.selectedModel)) {
                    this.selectedModel = this.availableModels.length > 0 ? this.availableModels[0] : '';
                } else if (!this.selectedModel && this.availableModels.length > 0) {
                    this.selectedModel = this.availableModels[0];
                }
            } else {
                this.availableModels = [];
            }
        },
        async consultAI(item, reanalyze = false) {
            this.currentItemForAI = item;
            if (!reanalyze) {
                this.aiResponse = '';
            }
            this.aiDialogVisible = true;
            this.consultingAI = true;
            
            try {
                const params = {
                    provider_id: this.selectedProviderId,
                    model: this.selectedModel
                };
                const response = await consultItemAI(item.id, params);
                
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
                this.aiResponse = this.$t('items.aiError') + ': ' + (err.message || this.$t('items.unknownError'));
                ElMessage.error(err.message || this.$t('items.aiFailed'));
                console.error('Error consulting AI:', err);
            } finally {
                this.consultingAI = false;
            }
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
        normalizeStatus(value) {
            if (value === null || value === undefined || value === '') return 0;
            if (typeof value === 'boolean') return value ? 1 : 0;
            const num = Number(value);
            return Number.isNaN(num) ? 0 : num;
        },
        normalizeEnabled(value) {
            if (value === null || value === undefined || value === '') return 1;
            if (typeof value === 'boolean') return value ? 1 : 0;
            const num = Number(value);
            return Number.isNaN(num) ? 1 : num;
        },
        toNumber(value, fallback) {
            if (value === null || value === undefined || value === '') return fallback;
            const num = Number(value);
            return Number.isNaN(num) ? fallback : num;
        },
        isColumnVisible(key) {
            return this.selectedColumns.includes(key);
        }
    }
};
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-main {
  display: flex;
  flex-direction: column;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 4px;
}

.refresh-info {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.auto-refresh-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-actions {
  display: flex;
  align-items: center;
}

.items-scroll {
  margin-top: 8px;
}

.items-pagination {
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
