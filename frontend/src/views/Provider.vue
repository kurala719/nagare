<template>
    <div class="providers-page">
            <div class="providers-toolbar">
                <div class="providers-filters">
                    <span class="filter-label">{{ $t('providers.search') }}</span>
                    <el-input v-model="search" :placeholder="$t('providers.search')" clearable class="providers-search" />
                    <span class="filter-label">{{ $t('providers.filterStatus') }}</span>
                    <el-select v-model="statusFilter" :placeholder="$t('providers.filterStatus')" class="providers-filter">
                        <el-option :label="$t('providers.filterAll')" value="all" />
                        <el-option :label="$t('common.statusInactive')" :value="0" />
                        <el-option :label="$t('common.statusActive')" :value="1" />
                        <el-option :label="$t('common.statusError')" :value="2" />
                        <el-option :label="$t('common.statusSyncing')" :value="3" />
                    </el-select>
                    <span class="filter-label">{{ $t('providers.filterType') }}</span>
                    <el-select v-model="typeFilter" :placeholder="$t('providers.filterType')" class="providers-filter">
                        <el-option :label="$t('providers.filterAll')" value="all" />
                        <el-option :label="$t('providers.typeGemini')" :value="1" />
                        <el-option :label="$t('providers.typeOpenAI')" :value="2" />
                        <el-option :label="$t('providers.typeOllama')" :value="3" />
                        <el-option :label="$t('providers.typeOtherOpenAI')" :value="4" />
                        <el-option :label="$t('providers.typeOther')" :value="5" />
                    </el-select>
                    <span class="filter-label">{{ $t('providers.sort') }}</span>
                    <el-select v-model="sortKey" class="providers-filter">
                        <el-option :label="$t('providers.sortNameAsc')" value="name_asc" />
                        <el-option :label="$t('providers.sortNameDesc')" value="name_desc" />
                        <el-option :label="$t('providers.sortStatusAsc')" value="status_asc" />
                        <el-option :label="$t('providers.sortStatusDesc')" value="status_desc" />
                    </el-select>
                    <div class="providers-bulk-actions">
                        <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
                        <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
                            {{ $t('common.bulkUpdate') }}
                        </el-button>
                        <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
                            {{ $t('common.bulkDelete') }}
                        </el-button>
                    </div>
                </div>
                <div class="providers-actions">
                    <el-button type="primary" @click="openCreateDialog">
                        {{ $t('providers.create') }}
                    </el-button>
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

            <div v-if="!loading && !error" class="providers-list">
                <!-- Provider Cards -->
                <el-row :gutter="20" class="providers-grid" v-if="pagedProviders.length > 0">
                    <el-col :span="6" v-for="provider in pagedProviders" :key="provider.id" class="provider-col">
                        <el-card class="provider-card" shadow="hover">
                            <template #header>
                                <div class="card-header">
                                    <div class="card-title">
                                        <el-checkbox :model-value="isSelected(provider.id)" @change="toggleSelection(provider.id, $event)" />
                                        <span class="provider-name">{{ provider.name }}</span>
                                        <el-tooltip placement="top-start" :content="provider.status_reason || getStatusInfo(provider.status).reason">
                                            <el-tag :type="getStatusInfo(provider.status).type" size="small">
                                                {{ getStatusInfo(provider.status).label }}
                                            </el-tag>
                                        </el-tooltip>
                                    </div>
                                    <div class="card-actions">
                                        <el-button size="small" @click="openProperties(provider)">{{ $t('providers.properties') }}</el-button>
                                        <el-button size="small" @click="onDelete(provider)">{{ $t('providers.delete') }}</el-button>
                                    </div>
                                </div>
                            </template>
                            <div class="card-body">
                                <p class="provider-desc">{{ provider.description || '-' }}</p>
                                <div class="provider-meta">
                                    <el-tag effect="light" type="info">{{ getTypeLabel(provider.type) }}</el-tag>
                                    <el-tag v-if="provider.default_model" effect="light" type="success">{{ provider.default_model }}</el-tag>
                                    <el-tag :type="provider.enabled === 1 ? 'success' : 'info'">
                                        {{ provider.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
                                    </el-tag>
                                </div>
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
                            <el-option :label="$t('providers.typeOllama')" :value="3" />
                            <el-option :label="$t('providers.typeOtherOpenAI')" :value="4" />
                            <el-option :label="$t('providers.typeOther')" :value="5" />
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

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { fetchProviderData, addProvider, deleteProvider, updateProvider } from '@/api/providers';

interface Provider {
    id: number;
    name: string;
    url: string;
    api_key: string;
    default_model: string;
    type: number;
    description: string;
    enabled: number;
    status: number;
    status_reason?: string;
}

const defaultProviderItem = (): Provider => ({
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
    components: { Loading },
    data() {
        return {
            providers: [] as Provider[],
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
                const mapped = data.map((p: any) => ({
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
            const form = this.$refs.providerForm as any;
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
            const form = this.$refs.propertiesForm as any;
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
        getStatusInfo(status: number) {
            const map: Record<number, { label: string; reason: string; type: string }> = {
                0: { label: this.$t('common.statusInactive'), reason: this.$t('common.reasonInactive'), type: 'info' },
                1: { label: this.$t('common.statusActive'), reason: this.$t('common.reasonActive'), type: 'success' },
                2: { label: this.$t('common.statusError'), reason: this.$t('common.reasonError'), type: 'danger' },
                3: { label: this.$t('common.statusSyncing'), reason: this.$t('common.reasonSyncing'), type: 'warning' },
            };
            return map[status] || map[0];
        },
        getTypeLabel(type: number) {
            const map: Record<number, string> = {
                1: this.$t('providers.typeGemini'),
                2: this.$t('providers.typeOpenAI'),
                3: this.$t('providers.typeOllama'),
                4: this.$t('providers.typeOtherOpenAI'),
                5: this.$t('providers.typeOther'),
            };
            return map[type] || this.$t('providers.typeOther');
        },
        syncQueryParams() {
            const query: Record<string, string> = {};
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
        isQueryEqual(nextQuery: Record<string, string>, currentQuery: Record<string, any>) {
            const current: Record<string, string> = {};
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
.providers-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.providers-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin: 8px 0 0;
    padding: 14px 16px;
    border-radius: 16px;
    background: var(--surface-1);
    border: 1px solid var(--border-1);
    box-shadow: var(--shadow-soft);
}

.providers-filters {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    align-items: center;
}

.providers-actions {
    display: flex;
    align-items: center;
    gap: 10px;
}

.providers-bulk-actions {
    display: flex;
    gap: 8px;
    align-items: center;
}

.selected-count {
    color: var(--text-muted);
    font-size: 12px;
}

.providers-search {
    width: 240px;
}

.providers-filter {
    min-width: 160px;
}

.providers-state {
    text-align: center;
    padding: 32px 0;
    color: var(--text-muted);
}

.providers-alert {
    margin: 8px 0;
}

.providers-empty {
    margin: 24px 0;
}

.providers-list {
    padding: 0 4px;
}

.providers-grid {
    margin: 0;
}

.provider-col {
    margin-bottom: 20px;
}

.provider-card {
    min-height: 300px;
    border-radius: 18px;
    border: 1px solid var(--border-1);
    box-shadow: var(--shadow-soft);
}

.card-header {
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.card-title {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
}

.provider-name {
    font-size: 20px;
    font-weight: 600;
    color: var(--text-strong);
}

.card-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
}

.card-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    color: var(--text-muted);
}

.provider-desc {
    min-height: 40px;
}

.provider-meta {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
}

.providers-pagination {
    display: flex;
    justify-content: flex-end;
    padding-bottom: 12px;
}

.providers-form {
    margin-top: 12px;
}
</style>