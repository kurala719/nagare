<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('providers.title') }}</h1>
      <p class="page-subtitle">{{ totalProviders }} {{ $t('dashboard.aiProviders') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('providers.search')" clearable style="width: 280px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('providers.filterStatus')" style="width: 140px">
          <el-option :label="$t('providers.filterAll')" value="all" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>

        <el-select v-model="typeFilter" :placeholder="$t('providers.filterType')" style="width: 160px">
          <el-option :label="$t('providers.filterAll')" value="all" />
          <el-option :label="$t('providers.typeGemini')" :value="1" />
          <el-option :label="$t('providers.typeOpenAI')" :value="2" />
        </el-select>

        <el-select v-model="sortKey" :placeholder="$t('common.sort')" style="width: 160px">
          <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
          <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
          <el-option :label="$t('common.sortStatusAsc')" value="status_asc" />
          <el-option :label="$t('common.sortStatusDesc')" value="status_desc" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button @click="loadProviders(true)" :loading="loading" :icon="Refresh" circle />
        <el-button type="primary" :icon="Plus" @click="openCreateDialog">
          {{ $t('providers.create') }}
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

            <!-- Loading State -->
            <div v-if="loading" class="providers-state">
                <el-icon class="is-loading" size="50" color="#2563eb">
                    <Loading />
                </el-icon>
                <p>{{ $t('providers.loading') }}</p>
            </div>

            <!-- Error State -->
            <el-alert
                v-if="error && !loading"
                :title="error"
                type="error"
                show-icon
                class="providers-alert"
                :closable="false"
            >
                <template #default>
                    <el-button size="small" @click="loadProviders">{{ $t('providers.retry') }}</el-button>
                </template>
            </el-alert>

            <!-- Empty State -->
            <el-empty
                v-if="!loading && !error && providers.length === 0"
                :description="$t('providers.noProviders')"
                class="providers-empty"
            />

            <el-empty
                v-if="!loading && !error && providers.length > 0 && filteredProviders.length === 0"
                :description="$t('providers.noResults')"
                class="providers-empty"
            />

  <div v-if="!loading && !error" class="providers-content">
    <el-row :gutter="24">
      <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="provider in pagedProviders" :key="provider.id" style="margin-bottom: 24px;">
        <el-card class="provider-card" :body-style="{ padding: '0px' }">
          <div class="provider-card-header">
            <div class="provider-icon-box">
              <el-icon :size="24"><Connection /></el-icon>
            </div>
            <div class="provider-title-area">
              <h3 class="provider-name">{{ provider.name }}</h3>
              <span class="provider-type-tag">{{ getTypeLabel(provider.type) }}</span>
            </div>
            <el-checkbox :model-value="isSelected(provider.id)" @change="toggleSelection(provider.id, $event)" class="provider-select" />
          </div>
          
          <div class="provider-card-body">
            <p class="provider-desc">{{ provider.description || '-' }}</p>
            <div class="provider-meta-row">
              <el-tag v-if="provider.default_model" size="small" effect="plain">{{ provider.default_model }}</el-tag>
              <el-tag :type="provider.enabled === 1 ? 'success' : 'info'" size="small">
                {{ provider.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
              </el-tag>
              <el-tooltip :content="provider.status_reason || getStatusInfo(provider.status).reason" placement="top">
                <el-tag :type="getStatusInfo(provider.status).type" size="small" effect="dark">
                  {{ getStatusInfo(provider.status).label }}
                </el-tag>
              </el-tooltip>
            </div>
          </div>

          <div class="provider-card-footer">
            <el-button-group>
              <el-button size="small" :icon="Edit" @click="openProperties(provider)">{{ $t('providers.properties') }}</el-button>
              <el-button size="small" type="danger" plain :icon="Delete" @click="onDelete(provider)">{{ $t('providers.delete') }}</el-button>
            </el-button-group>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>

            <div v-if="!loading && !error && totalProviders > 0" class="providers-pagination">
                <el-pagination
                    background
                    layout="sizes, prev, pager, next"
                    :page-sizes="[12, 24, 48]"
                    v-model:page-size="pageSize"
                    v-model:current-page="currentPage"
                    :total="totalProviders"
                />
            </div>

            <el-dialog v-model="dialogVisible" width="520px" align-center>
                <template #title>{{ dialogTitle }}</template>
                <el-form ref="providerForm" :model="dialogItem" :rules="formRules" label-width="120px" class="providers-form">
                    <el-form-item :label="$t('providers.providerName')" prop="name">
                        <el-input v-model="dialogItem.name" :placeholder="$t('providers.providerName')" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.type')" prop="type">
                        <el-select v-model="dialogItem.type" style="width: 100%;">
                            <el-option :label="$t('providers.typeGemini')" :value="1" />
                            <el-option :label="$t('providers.typeOpenAI')" :value="2" />
                            <el-option :label="$t('providers.typeOther')" :value="3" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('providers.url')">
                        <el-input v-model="dialogItem.url" :placeholder="$t('providers.url')" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.apiKey')" prop="api_key">
                        <el-input v-model="dialogItem.api_key" type="password" :placeholder="$t('providers.apiKey')" show-password />
                    </el-form-item>
                    <el-form-item :label="$t('providers.defaultModel')">
                        <el-input v-model="dialogItem.default_model" :placeholder="$t('providers.defaultModel')" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.description')">
                        <el-input v-model="dialogItem.description" :placeholder="$t('providers.description')" />
                    </el-form-item>
                    <el-form-item :label="$t('common.enabled')">
                        <el-switch v-model="dialogItem.enabled" :active-value="1" :inactive-value="0" />
                    </el-form-item>
                </el-form>
                <template #footer>
                    <el-button @click="dialogVisible = false">{{ $t('providers.cancel') }}</el-button>
                    <el-button v-if="editDialog" type="primary" @click="onEdit()">{{ $t('providers.edit') }}</el-button>
                    <el-button v-else type="primary" @click="onCreate()">{{ $t('providers.createBtn') }}</el-button>
                </template>
            </el-dialog>

            <el-dialog v-model="propertiesDialogVisible" :title="`${$t('providers.propertiesTitle')} - ${selectedProvider ? selectedProvider.name : ''}`" width="600px">
                <el-form ref="propertiesForm" :model="selectedProvider" :rules="formRules" label-width="120px" class="providers-form">
                    <el-form-item :label="$t('providers.name')" prop="name">
                        <el-input v-model="selectedProvider.name" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.type')" prop="type">
                        <el-select v-model="selectedProvider.type" style="width: 100%;">
                            <el-option :label="$t('providers.typeGemini')" :value="1" />
                            <el-option :label="$t('providers.typeOpenAI')" :value="2" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('providers.url')">
                        <el-input v-model="selectedProvider.url" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.apiKey')" prop="api_key">
                        <el-input v-model="selectedProvider.api_key" show-password />
                    </el-form-item>
                    <el-form-item :label="$t('providers.model')">
                        <el-input v-model="selectedProvider.default_model" />
                    </el-form-item>
                    <el-form-item :label="$t('providers.description')">
                        <el-input type="textarea" v-model="selectedProvider.description" />
                    </el-form-item>
                    <el-form-item :label="$t('common.enabled')">
                        <el-switch v-model="selectedProvider.enabled" :active-value="1" :inactive-value="0" />
                    </el-form-item>
                </el-form>
                <template #footer>
                    <el-button @click="cancelProperties">{{ $t('providers.cancel') }}</el-button>
                    <el-button type="primary" @click="saveProperties">{{ $t('providers.save') }}</el-button>
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
                </el-form>
                <template #footer>
                    <el-button @click="bulkDialogVisible = false">{{ $t('providers.cancel') }}</el-button>
                    <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
                </template>
            </el-dialog>

            <!-- Bulk Delete Confirmation Dialog -->
            <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
                <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
                <template #footer>
                    <el-button @click="bulkDeleteDialogVisible = false">{{ $t('providers.cancel') }}</el-button>
                    <el-button type="danger" @click="deleteSelectedProviders" :loading="bulkDeleting">{{ $t('providers.delete') }}</el-button>
                </template>
            </el-dialog>
  </div>
</template>

<script>
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Search, Plus, Refresh, Edit, Delete, Connection, ArrowDown } from '@element-plus/icons-vue';
import { fetchProviderData, addProvider, deleteProvider, updateProvider } from '@/api/providers';

const defaultProviderItem = () => ({
    id: 0,
    name: '',
    url: '',
    api_key: '',
    default_model: '',
    type: 2,
    description: '',
    enabled: 1,
    status: 1,
});

export default {
    name: 'Provider',
    components: { Loading, Search, Plus, Refresh, Edit, Delete, Connection, ArrowDown },
    data() {
        return {
            providers: [],
            dialogVisible: false,
            dialogTitle: '',
            dialogItem: defaultProviderItem(),
            editDialog: false,
            propertiesDialogVisible: false,
            selectedProvider: defaultProviderItem(),
            loading: false,
            error: null,
            search: '',
            statusFilter: 'all',
            typeFilter: 'all',
            sortKey: 'name_asc',
            pageSize: 12,
            currentPage: 1,
            bulkDialogVisible: false,
            bulkDeleteDialogVisible: false,
            bulkUpdating: false,
            bulkDeleting: false,
            selectedProviderIds: [],
            bulkForm: {
                enabled: 'nochange',
            },
            fetchLimit: 1000,
            // Icons for template usage
            Loading: markRaw(Loading),
            Search: markRaw(Search),
            Plus: markRaw(Plus),
            Refresh: markRaw(Refresh),
            Edit: markRaw(Edit),
            Delete: markRaw(Delete),
            Connection: markRaw(Connection),
            ArrowDown: markRaw(ArrowDown)
        };
    },
    computed: {
        formRules() {
            return {
                name: [{ required: true, message: this.$t('providers.nameRequired'), trigger: 'blur' }],
                api_key: [{ required: true, message: this.$t('providers.apiKeyRequired'), trigger: 'blur' }],
                type: [{ required: true, message: this.$t('providers.typeRequired'), trigger: 'change' }],
            };
        },
        filteredProviders() {
            const query = this.search.trim().toLowerCase();
            return this.providers.filter((provider) => {
                if (this.statusFilter !== 'all' && provider.status !== Number(this.statusFilter)) return false;
                if (this.typeFilter !== 'all' && provider.type !== Number(this.typeFilter)) return false;
                if (!query) return true;
                return [provider.name, provider.url, provider.description]
                    .filter(Boolean)
                    .some((field) => String(field).toLowerCase().includes(query));
            });
        },
        sortedProviders() {
            const items = [...this.filteredProviders];
            switch (this.sortKey) {
                case 'name_desc':
                    return items.sort((a, b) => b.name.localeCompare(a.name));
                case 'status_asc':
                    return items.sort((a, b) => a.status - b.status);
                case 'status_desc':
                    return items.sort((a, b) => b.status - a.status);
                case 'name_asc':
                default:
                    return items.sort((a, b) => a.name.localeCompare(b.name));
            }
        },
        pagedProviders() {
            const start = (this.currentPage - 1) * this.pageSize;
            return this.sortedProviders.slice(start, start + this.pageSize);
        },
        totalProviders() {
            return this.sortedProviders.length;
        },
        selectedCount() {
            return this.selectedProviderIds.length;
        },
    },
    created() {
        this.applyQueryParams();
        this.loadProviders(true);
    },
    watch: {
        search() {
            this.currentPage = 1;
            this.syncQueryParams();
        },
        statusFilter() {
            this.currentPage = 1;
            this.syncQueryParams();
        },
        typeFilter() {
            this.currentPage = 1;
            this.syncQueryParams();
        },
        sortKey() {
            this.currentPage = 1;
            this.syncQueryParams();
        },
        pageSize() {
            this.currentPage = 1;
            this.syncQueryParams();
        },
        currentPage() {
            this.syncQueryParams();
        },
        '$route.query': {
            handler() {
                this.applyQueryParams();
            },
        },
    },
    methods: {
        openCreateDialog() {
            this.dialogTitle = this.$t('providers.createTitle');
            this.editDialog = false;
            this.dialogItem = defaultProviderItem();
            this.dialogVisible = true;
        },
        isSelected(id) {
            return this.selectedProviderIds.includes(id);
        },
        toggleSelection(id, checked) {
            if (checked) {
                if (!this.selectedProviderIds.includes(id)) {
                    this.selectedProviderIds.push(id);
                }
            } else {
                this.selectedProviderIds = this.selectedProviderIds.filter((itemId) => itemId !== id);
            }
        },
        selectAll() {
            this.selectedProviderIds = this.providers.map(p => p.id);
        },
        clearSelection() {
            this.selectedProviderIds = [];
        },
        openBulkDeleteDialog() {
            if (this.selectedCount === 0) {
                ElMessage.warning(this.$t('common.selectAtLeastOne'));
                return;
            }
            this.bulkDeleteDialogVisible = true;
        },
        async deleteSelectedProviders() {
            if (this.selectedCount === 0) return;

            this.bulkDeleting = true;
            try {
                await Promise.all(this.selectedProviderIds.map((id) => deleteProvider(id)));
                ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
                this.bulkDeleteDialogVisible = false;
                this.selectedProviderIds = [];
                await this.loadProviders(true);
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
            };
            this.bulkDialogVisible = true;
        },
        async applyBulkUpdate() {
            if (this.selectedCount === 0) return;
            if (this.bulkForm.enabled === 'nochange') {
                ElMessage.warning(this.$t('common.bulkUpdateNoChanges'));
                return;
            }

            this.bulkUpdating = true;
            try {
                const enabledOverride = this.bulkForm.enabled;
                await Promise.all(this.providers.filter((p) => this.selectedProviderIds.includes(p.id)).map((provider) => {
                    const payload = {
                        name: provider.name,
                        url: provider.url,
                        api_key: provider.api_key,
                        default_model: provider.default_model,
                        type: provider.type,
                        description: provider.description,
                        enabled: enabledOverride === 'nochange' ? provider.enabled : (enabledOverride === 'enable' ? 1 : 0),
                    };
                    return updateProvider(provider.id, payload);
                }));
                ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
                this.bulkDialogVisible = false;
                this.selectedProviderIds = [];
                await this.loadProviders(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
            } finally {
                this.bulkUpdating = false;
            }
        },
        async loadProviders(reset = false) {
            this.loading = reset;
            this.error = null;
            try {
                const response = await fetchProviderData({
                    limit: this.fetchLimit,
                    offset: 0,
                });
                const data = Array.isArray(response) ? response : (response.data || response.providers || []);
                const mapped = data.map((p) => ({
                    id: p.ID || p.id,
                    name: p.Name || p.name || '',
                    url: p.URL || p.url || '',
                    api_key: p.APIKey || p.api_key || '',
                    default_model: p.DefaultModel || p.default_model || p.Model || p.model || '',
                    type: p.Type ?? p.type ?? 3,
                    description: p.Description || p.description || '',
                    enabled: p.Enabled ?? p.enabled ?? 1,
                    status: p.Status ?? p.status ?? 0,
                    status_reason: p.Reason || p.reason || p.Error || p.error || p.ErrorMessage || p.error_message || p.LastError || p.last_error || '',
                }));
                this.providers = mapped;
            } catch (err) {
                this.error = err.message || 'Failed to load providers';
                console.error('Error loading providers:', err);
            } finally {
                this.loading = false;
            }
        },
        async onCreate() {
            const form = this.$refs.providerForm;
            if (form) {
                const valid = await form.validate().catch(() => false);
                if (!valid) return;
            }

            try {
                const providerData = {
                    name: this.dialogItem.name,
                    url: this.dialogItem.url,
                    api_key: this.dialogItem.api_key,
                    default_model: this.dialogItem.default_model,
                    type: this.dialogItem.type,
                    description: this.dialogItem.description,
                    enabled: this.dialogItem.enabled,
                };

                await addProvider(providerData);
                await this.loadProviders(true);

                this.dialogItem = defaultProviderItem();
                this.dialogVisible = false;
                ElMessage({
                    type: 'success',
                    message: 'Provider created successfully!',
                });
            } catch (err) {
                ElMessage({
                    type: 'error',
                    message: 'Failed to create provider: ' + (err.message || 'Unknown error'),
                });
                console.error('Error creating provider:', err);
            }
        },
        async onEdit() {
            this.dialogTitle = this.$t('providers.editTitle');
            try {
                const updateData = {
                    name: this.dialogItem.name,
                    url: this.dialogItem.url,
                    api_key: this.dialogItem.api_key,
                    default_model: this.dialogItem.default_model,
                    type: this.dialogItem.type,
                    description: this.dialogItem.description,
                    enabled: this.dialogItem.enabled,
                };
                await updateProvider(this.dialogItem.id, updateData);
                await this.loadProviders(true);
                this.dialogVisible = false;
                ElMessage({
                    type: 'success',
                    message: 'Provider updated successfully!',
                });
            } catch (err) {
                ElMessage({
                    type: 'error',
                    message: 'Failed to update provider: ' + (err.message || 'Unknown error'),
                });
                console.error('Error updating provider:', err);
            }
        },
        onDelete(provider) {
            ElMessageBox.confirm(
                `Are you sure you want to delete ${provider.name}?`,
                'Warning',
                {
                    confirmButtonText: 'Yes',
                    cancelButtonText: 'No',
                    type: 'warning',
                }
            ).then(async () => {
                try {
                    await deleteProvider(provider.id);
                    await this.loadProviders(true);
                    ElMessage({
                        type: 'success',
                        message: 'Provider deleted successfully!',
                    });
                } catch (err) {
                    ElMessage({
                        type: 'error',
                        message: 'Failed to delete provider: ' + (err.message || 'Unknown error'),
                    });
                    console.error('Error deleting provider:', err);
                }
            }).catch(() => {
                ElMessage({
                    type: 'info',
                    message: 'Delete canceled',
                });
            });
        },
        async deleteProvider(provider) {
            try {
                await deleteProvider(provider.id);
                const index = this.providers.findIndex(p => p.id === provider.id);
                if (index !== -1) {
                    this.providers.splice(index, 1);
                }
            } catch (err) {
                ElMessage({
                    type: 'error',
                    message: 'Failed to delete provider: ' + (err.message || 'Unknown error'),
                });
                console.error('Error deleting provider:', err);
            }
        },
        openProperties(provider) {
            this.selectedProvider = { ...provider };
            this.propertiesDialogVisible = true;
        },
        cancelProperties() {
            this.propertiesDialogVisible = false;
        },
        async saveProperties() {
            const form = this.$refs.propertiesForm;
            if (form) {
                const valid = await form.validate().catch(() => false);
                if (!valid) return;
            }
            try {
                const updateData = {
                    name: this.selectedProvider.name,
                    url: this.selectedProvider.url,
                    api_key: this.selectedProvider.api_key,
                    default_model: this.selectedProvider.default_model,
                    type: this.selectedProvider.type,
                    description: this.selectedProvider.description,
                    enabled: this.selectedProvider.enabled,
                };
                await updateProvider(this.selectedProvider.id, updateData);
                await this.loadProviders(true);
                this.propertiesDialogVisible = false;
                ElMessage({
                    type: 'success',
                    message: 'Provider updated successfully!',
                });
            } catch (err) {
                ElMessage({
                    type: 'error',
                    message: 'Failed to update provider: ' + (err.message || 'Unknown error'),
                });
                console.error('Error updating provider:', err);
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
        getTypeLabel(type) {
            const map = {
                1: this.$t('providers.typeGemini'),
                2: this.$t('providers.typeOpenAI'),
                3: this.$t('providers.typeOther'),
            };
            return map[type] || this.$t('providers.typeOpenAI');
        },
        syncQueryParams() {
            const query = {};
            if (this.search) query.q = this.search;
            if (this.statusFilter !== 'all') query.status = String(this.statusFilter);
            if (this.typeFilter !== 'all') query.type = String(this.typeFilter);
            if (this.sortKey !== 'name_asc') query.sort = this.sortKey;
            if (this.currentPage !== 1) query.page = String(this.currentPage);
            if (this.pageSize !== 12) query.pageSize = String(this.pageSize);

            if (!this.isQueryEqual(query, this.$route.query)) {
                this.$router.replace({ query });
            }
        },
        applyQueryParams() {
            const query = this.$route.query || {};
            this.search = typeof query.q === 'string' ? query.q : '';
            this.statusFilter = typeof query.status === 'string' && query.status !== 'all'
                ? Number(query.status)
                : 'all';
            this.typeFilter = typeof query.type === 'string' && query.type !== 'all'
                ? Number(query.type)
                : 'all';
            this.sortKey = typeof query.sort === 'string' ? query.sort : 'name_asc';
            this.currentPage = typeof query.page === 'string' ? Number(query.page) : 1;
            this.pageSize = typeof query.pageSize === 'string' ? Number(query.pageSize) : 12;
        },
        isQueryEqual(nextQuery, currentQuery) {
            const current = {};
            Object.keys(currentQuery || {}).forEach((key) => {
                const value = currentQuery[key];
                if (typeof value === 'string') current[key] = value;
            });
            const nextKeys = Object.keys(nextQuery);
            const currentKeys = Object.keys(current);
            if (nextKeys.length !== currentKeys.length) return false;
            return nextKeys.every((key) => current[key] === nextQuery[key]);
        }
    }
};
</script>

<style scoped>
.providers-content {
  margin-top: 8px;
}

.provider-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.provider-card-header {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid var(--border-1);
  position: relative;
}

.provider-icon-box {
  width: 48px;
  height: 48px;
  background: var(--brand-50);
  color: var(--brand-600);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.provider-title-area {
  flex: 1;
  min-width: 0;
}

.provider-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0;
  color: var(--text-strong);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.provider-type-tag {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 600;
}

.provider-select {
  position: absolute;
  top: 12px;
  right: 12px;
}

.provider-card-body {
  padding: 20px;
  flex: 1;
}

.provider-desc {
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

.provider-meta-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.provider-card-footer {
  padding: 12px 20px;
  background: var(--surface-2);
  display: flex;
  justify-content: center;
  border-top: 1px solid var(--border-1);
}

.providers-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.providers-state {
  text-align: center;
  padding: 60px;
  color: var(--text-muted);
}
</style>
