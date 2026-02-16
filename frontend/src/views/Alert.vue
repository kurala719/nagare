<template>
    <div class="databoard-layout">
        <el-container>
        <el-main>
            <!-- Header with Add Button -->
            <div class="alerts-header">
                <h2>{{ $t('alerts.title') }}</h2>
                <el-button type="primary" @click="openAddDialog">
                    <el-icon><Plus /></el-icon>
                    {{ $t('alerts.add') }}
                </el-button>
            </div>

            <div class="alerts-toolbar">
                <span class="filter-label">{{ $t('alerts.search') }}</span>
                <el-input v-model="search" :placeholder="$t('alerts.search')" clearable class="alerts-search" />
                <span class="filter-label">{{ $t('alerts.filterSeverity') }}</span>
                <el-select v-model="severityFilter" :placeholder="$t('alerts.filterSeverity')" class="alerts-filter">
                    <el-option :label="$t('alerts.filterAll')" value="all" />
                    <el-option :label="$t('alerts.severityCritical')" value="critical" />
                    <el-option :label="$t('alerts.severityHigh')" value="high" />
                    <el-option :label="$t('alerts.severityMedium')" value="medium" />
                    <el-option :label="$t('alerts.severityLow')" value="low" />
                    <el-option :label="$t('alerts.severityInfo')" value="info" />
                </el-select>
                <span class="filter-label">{{ $t('alerts.filterStatus') }}</span>
                <el-select v-model="statusFilter" :placeholder="$t('alerts.filterStatus')" class="alerts-filter">
                    <el-option :label="$t('alerts.filterAll')" value="all" />
                    <el-option :label="$t('alerts.statusOpen')" value="open" />
                    <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
                    <el-option :label="$t('alerts.statusResolved')" value="resolved" />
                    <el-option :label="$t('alerts.statusClosed')" value="closed" />
                </el-select>
                <span class="filter-label">{{ $t('common.sort') }}</span>
                <el-select v-model="sortKey" class="alerts-filter">
                    <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
                    <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
                    <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
                    <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
                </el-select>
                <div class="alerts-bulk-actions">
                    <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
                    <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
                        {{ $t('common.bulkUpdate') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
                        {{ $t('common.bulkDelete') }}
                    </el-button>
                </div>
            </div>

            <!-- Loading State -->
            <div v-if="loading" style="text-align: center; padding: 40px;">
                <el-icon class="is-loading" size="50" color="#409EFF">
                    <Loading />
                </el-icon>
                <p style="margin-top: 16px; color: #909399;">{{ $t('alerts.loading') }}</p>
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
                    <el-button size="small" @click="loadAlerts">{{ $t('alerts.retry') }}</el-button>
                </template>
            </el-alert>

            <!-- Empty State -->
            <el-empty 
                v-if="!loading && !error && alerts && alerts.length === 0"
                :description="$t('alerts.noAlerts')"
                style="margin: 40px;"
            />

            <el-empty
                v-if="!loading && !error && alerts && alerts.length > 0 && filteredAlerts.length === 0"
                :description="$t('alerts.noResults')"
                style="margin: 40px;"
            />

            <div v-if="!loading && !error" class="alerts-scroll">
            <!-- Alerts List -->
            <div v-if="alerts.length > 0" class="alerts-list">
                <el-card v-for="alert in alerts" :key="alert.id" class="alert-card" shadow="hover">
                    <div class="alert-card-header">
                        <el-checkbox :model-value="isSelected(alert.id)" @change="toggleSelection(alert.id, $event)" />
                        <span class="alert-card-title">{{ $t('alerts.alertLabel', { id: alert.id }) }}</span>
                    </div>
                    <el-descriptions :column="2" border>
                        <el-descriptions-item :label="$t('alerts.messageLabel')" :span="2">{{ alert.message }}</el-descriptions-item>
                        <el-descriptions-item :label="$t('alerts.severityLabel')">
                            <el-tag :type="getSeverityType(alert.severity)" size="small">{{ alert.severity }}</el-tag>
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('alerts.statusLabel')">
                            <el-tooltip :content="alert.status_reason || alert.message || alert.status" placement="top">
                                <el-tag :type="getStatusType(alert.status)" size="small">{{ alert.status }}</el-tag>
                            </el-tooltip>
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('alerts.createdAt')" :span="2">{{ alert.created_at }}</el-descriptions-item>
                    </el-descriptions>
                    <div style="margin-top: 16px; text-align: right;">
                        <el-button type="primary" @click="consultAI(alert)">{{ $t('alerts.consult') }}</el-button>
                        <el-button type="warning" @click="openEditDialog(alert)">{{ $t('alerts.edit') }}</el-button>
                        <el-button type="danger" @click="confirmDelete(alert)">{{ $t('alerts.remove') }}</el-button>
                    </div>
                </el-card>
            </div>
            </div>
            <div v-if="!loading && !error && totalAlerts > 0" class="alerts-pagination">
                <el-pagination
                    background
                    layout="sizes, prev, pager, next"
                    :page-sizes="[10, 20, 50, 100]"
                    v-model:page-size="pageSize"
                    v-model:current-page="currentPage"
                    :total="totalAlerts"
                />
            </div>
        </el-main>
        </el-container>

        <!-- Add/Edit Dialog -->
        <el-dialog 
            v-model="dialogVisible" 
            :title="isEditing ? $t('alerts.editTitle') : $t('alerts.addTitle')"
            width="500px"
        >
            <el-form :model="alertForm" label-width="100px" :rules="formRules" ref="alertFormRef">
                <el-form-item :label="$t('alerts.messageLabel')" prop="message">
                    <el-input v-model="alertForm.message" type="textarea" :rows="3" :placeholder="$t('alerts.enterMessage')" />
                </el-form-item>
                <el-form-item :label="$t('alerts.severityLabel')" prop="severity">
                    <el-select v-model="alertForm.severity" :placeholder="$t('alerts.selectSeverity')" style="width: 100%;">
                        <el-option :label="$t('alerts.severityCritical')" value="critical" />
                        <el-option :label="$t('alerts.severityHigh')" value="high" />
                        <el-option :label="$t('alerts.severityMedium')" value="medium" />
                        <el-option :label="$t('alerts.severityLow')" value="low" />
                        <el-option :label="$t('alerts.severityInfo')" value="info" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('alerts.statusLabel')" prop="status">
                    <el-select v-model="alertForm.status" :placeholder="$t('alerts.selectStatus')" style="width: 100%;">
                        <el-option :label="$t('alerts.statusOpen')" value="open" />
                        <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
                        <el-option :label="$t('alerts.statusResolved')" value="resolved" />
                        <el-option :label="$t('alerts.statusClosed')" value="closed" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">{{ $t('alerts.cancel') }}</el-button>
                <el-button type="primary" @click="saveAlert" :loading="saving">
                    {{ isEditing ? $t('alerts.update') : $t('alerts.create') }}
                </el-button>
            </template>
        </el-dialog>

        <!-- Delete Confirmation Dialog -->
        <el-dialog v-model="deleteDialogVisible" :title="$t('alerts.confirmDelete')" width="400px">
            <p>{{ $t('alerts.confirmDeleteText') }}</p>
            <p v-if="alertToDelete"><strong>{{ alertToDelete.message }}</strong></p>
            <template #footer>
                <el-button @click="deleteDialogVisible = false">{{ $t('alerts.cancel') }}</el-button>
                <el-button type="danger" @click="deleteAlertConfirmed" :loading="deleting">{{ $t('alerts.remove') }}</el-button>
            </template>
        </el-dialog>

        <!-- Bulk Update Dialog -->
        <el-dialog v-model="bulkDialogVisible" :title="$t('common.bulkUpdateTitle')" width="460px">
            <el-form :model="bulkForm" label-width="140px">
                <el-form-item :label="$t('alerts.statusLabel')">
                    <el-select v-model="bulkForm.status" style="width: 100%;">
                        <el-option :label="$t('common.bulkUpdateNoChange')" value="nochange" />
                        <el-option :label="$t('alerts.statusOpen')" value="open" />
                        <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
                        <el-option :label="$t('alerts.statusResolved')" value="resolved" />
                        <el-option :label="$t('alerts.statusClosed')" value="closed" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="bulkDialogVisible = false">{{ $t('alerts.cancel') }}</el-button>
                <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
            </template>
        </el-dialog>

        <!-- Bulk Delete Confirmation Dialog -->
        <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
            <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
            <template #footer>
                <el-button @click="bulkDeleteDialogVisible = false">{{ $t('alerts.cancel') }}</el-button>
                <el-button type="danger" @click="deleteSelectedAlerts" :loading="bulkDeleting">{{ $t('alerts.remove') }}</el-button>
            </template>
        </el-dialog>

        <!-- AI Response Dialog -->
        <el-dialog v-model="aiDialogVisible" :title="$t('alerts.aiTitle')" width="600px">
            <div v-if="consultingAI" style="text-align: center; padding: 40px;">
                <el-icon class="is-loading" size="40" color="#409EFF">
                    <Loading />
                </el-icon>
                <p style="margin-top: 16px; color: #909399;">{{ $t('alerts.aiLoading') }}</p>
            </div>
            <div v-else>
                <el-descriptions v-if="currentAlertForAI" :column="1" border style="margin-bottom: 16px;">
                    <el-descriptions-item :label="$t('alerts.title')">{{ currentAlertForAI.message }}</el-descriptions-item>
                </el-descriptions>
                <el-divider content-position="left">{{ $t('alerts.aiResponse') }}</el-divider>
                <div class="ai-response-content">
                    <p style="white-space: pre-wrap;">{{ aiResponse }}</p>
                </div>
            </div>
            <template #footer>
                <el-button @click="aiDialogVisible = false">{{ $t('alerts.close') }}</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script>
import { fetchAlertData, addAlert, updateAlert, deleteAlert, consultAlertAI } from '@/api/alerts';
import { ElMessage } from 'element-plus';
import { Loading, Plus } from '@element-plus/icons-vue';

export default {
    name: 'Alert',
    components: {
        Loading,
        Plus,
    },
    data() {
      return {
        alerts: [],
        pageSize: 20,
        currentPage: 1,
        totalAlerts: 0,
        sortKey: 'created_desc',
        loading: false,
                bulkUpdating: false,
                bulkDeleting: false,
        saving: false,
        deleting: false,
        consultingAI: false,
        error: null,
                bulkDialogVisible: false,
                bulkDeleteDialogVisible: false,
                search: '',
                severityFilter: 'all',
                statusFilter: 'all',
        dialogVisible: false,
                selectedAlertIds: [],
        deleteDialogVisible: false,
        aiDialogVisible: false,
        isEditing: false,
        editingId: null,
        alertToDelete: null,
        currentAlertForAI: null,
        aiResponse: '',
                bulkForm: {
                        status: 'nochange',
                },
        alertForm: {
            message: '',
            severity: '',
            status: 'open',
        },
        formRules: {},
      };
    },
    computed: {
        filteredAlerts() {
            return this.alerts;
        },
        selectedCount() {
            return this.selectedAlertIds.length;
        },
    },
    watch: {
        '$route.query.q': function () {
            this.applySearchFromQuery();
        },
        search() {
            this.currentPage = 1;
            this.loadAlerts(true);
        },
        severityFilter() {
            this.currentPage = 1;
            this.loadAlerts(true);
        },
        statusFilter() {
            this.currentPage = 1;
            this.loadAlerts(true);
        },
        sortKey() {
            this.currentPage = 1;
            this.loadAlerts(true);
        },
        pageSize() {
            this.currentPage = 1;
            this.loadAlerts(true);
        },
        currentPage() {
            this.loadAlerts();
        },
    },
    created() {
        this.formRules = {
            message: [{ required: true, message: this.$t('alerts.validationMessage'), trigger: 'blur' }],
            severity: [{ required: true, message: this.$t('alerts.validationSeverity'), trigger: 'change' }],
            status: [{ required: true, message: this.$t('alerts.validationStatus'), trigger: 'change' }],
        };
        this.applySearchFromQuery();
        this.loadAlerts(true);
    },
    methods: {
        applySearchFromQuery() {
            const queryValue = this.$route.query.q;
            const nextQuery = queryValue ? String(queryValue) : '';
            if (nextQuery !== this.search) {
                this.search = nextQuery;
            }
        },
        isSelected(id) {
            return this.selectedAlertIds.includes(id);
        },
        toggleSelection(id, checked) {
            if (checked) {
                if (!this.selectedAlertIds.includes(id)) {
                    this.selectedAlertIds.push(id);
                }
            } else {
                this.selectedAlertIds = this.selectedAlertIds.filter((itemId) => itemId !== id);
            }
        },
        openBulkDeleteDialog() {
            if (this.selectedCount === 0) {
                ElMessage.warning(this.$t('common.selectAtLeastOne'));
                return;
            }
            this.bulkDeleteDialogVisible = true;
        },
        async deleteSelectedAlerts() {
            if (this.selectedCount === 0) return;

            this.bulkDeleting = true;
            try {
                await Promise.all(this.selectedAlertIds.map((id) => deleteAlert(id)));
                ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
                this.bulkDeleteDialogVisible = false;
                this.selectedAlertIds = [];
                await this.loadAlerts(true);
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
            this.bulkForm = { status: 'nochange' };
            this.bulkDialogVisible = true;
        },
        async applyBulkUpdate() {
            if (this.selectedCount === 0) return;
            if (this.bulkForm.status === 'nochange') {
                ElMessage.warning(this.$t('common.bulkUpdateNoChanges'));
                return;
            }

            this.bulkUpdating = true;
            try {
                const statusOverride = this.bulkForm.status;
                await Promise.all(this.alerts.filter((alert) => this.selectedAlertIds.includes(alert.id)).map((alert) => {
                    const payload = {
                        message: alert.message,
                        severity: alert.severity,
                        status: statusOverride,
                    };
                    return updateAlert(alert.id, payload);
                }));
                ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
                this.bulkDialogVisible = false;
                this.selectedAlertIds = [];
                await this.loadAlerts(true);
            } catch (err) {
                ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
            } finally {
                this.bulkUpdating = false;
            }
        },
        async loadAlerts(reset = false) {
            if (reset) {
                this.alerts = [];
            }
            this.loading = reset;
            this.error = null;
            try {
                const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
                const response = await fetchAlertData({
                    q: this.search || undefined,
                    severity: this.severityFilterValue(),
                    status: this.statusFilterValue(),
                    limit: this.pageSize,
                    offset: (this.currentPage - 1) * this.pageSize,
                    sort: sortBy,
                    order: sortOrder,
                    with_total: 1,
                });
                const payload = Array.isArray(response)
                    ? response
                    : (response.data?.items || response.items || response.data || response.alerts || []);
                const total = response?.data?.total ?? response?.total ?? payload.length;
                const mapped = payload.map((a) => ({
                    id: a.ID || a.id,
                    message: a.Message || a.message || '',
                    severity: this.normalizeSeverity(a.Severity ?? a.severity ?? ''),
                    status: this.normalizeStatus(a.Status ?? a.status ?? ''),
                    status_reason: a.Reason || a.reason || a.Error || a.error || a.ErrorMessage || a.error_message || a.LastError || a.last_error || '',
                    created_at: a.CreatedAt || a.created_at || '',
                }));
                this.alerts = mapped;
                this.totalAlerts = Number.isFinite(total) ? total : mapped.length;
            } catch (err) {
                this.error = err.message || 'Failed to load alerts';
                console.error('Error loading alerts:', err);
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
                case 'updated_desc':
                    return { sortBy: 'updated_at', sortOrder: 'desc' };
                case 'created_desc':
                default:
                    return { sortBy: 'created_at', sortOrder: 'desc' };
            }
        },
        severityFilterValue() {
            if (this.severityFilter === 'all') return undefined;
            const map = { info: 0, low: 1, medium: 2, high: 3, critical: 4 };
            return map[this.severityFilter] ?? undefined;
        },
        statusFilterValue() {
            if (this.statusFilter === 'all') return undefined;
            const map = { open: 0, acknowledged: 1, resolved: 2, closed: 3 };
            return map[this.statusFilter] ?? undefined;
        },
        normalizeSeverity(value) {
            if (typeof value === 'number') {
                const map = { 0: 'info', 1: 'low', 2: 'medium', 3: 'high', 4: 'critical' };
                return map[value] || String(value);
            }
            return String(value || '');
        },
        normalizeStatus(value) {
            if (typeof value === 'number') {
                const map = { 0: 'open', 1: 'acknowledged', 2: 'resolved', 3: 'closed' };
                return map[value] || String(value);
            }
            return String(value || '');
        },
        openAddDialog() {
            this.isEditing = false;
            this.editingId = null;
            this.alertForm = {
                message: '',
                severity: '',
                status: 'open',
            };
            this.dialogVisible = true;
        },
        openEditDialog(alert) {
            this.isEditing = true;
            this.editingId = alert.id;
            this.alertForm = {
                message: alert.message,
                severity: alert.severity,
                status: alert.status,
            };
            this.dialogVisible = true;
        },
        async saveAlert() {
            try {
                await this.$refs.alertFormRef.validate();
            } catch {
                return;
            }

            this.saving = true;
            try {
                if (this.isEditing) {
                    await updateAlert(this.editingId, this.alertForm);
                    ElMessage.success('Alert updated successfully');
                } else {
                    await addAlert(this.alertForm);
                    ElMessage.success('Alert created successfully');
                }
                this.dialogVisible = false;
                await this.loadAlerts(true);
            } catch (err) {
                ElMessage.error(err.message || 'Failed to save alert');
                console.error('Error saving alert:', err);
            } finally {
                this.saving = false;
            }
        },
        confirmDelete(alert) {
            this.alertToDelete = alert;
            this.deleteDialogVisible = true;
        },
        async deleteAlertConfirmed() {
            if (!this.alertToDelete) return;
            
            this.deleting = true;
            try {
                await deleteAlert(this.alertToDelete.id);
                ElMessage.success('Alert deleted successfully');
                this.deleteDialogVisible = false;
                this.alertToDelete = null;
                await this.loadAlerts(true);
            } catch (err) {
                ElMessage.error(err.message || 'Failed to delete alert');
                console.error('Error deleting alert:', err);
            } finally {
                this.deleting = false;
            }
        },
        async consultAI(alert) {
            this.currentAlertForAI = alert;
            this.aiResponse = '';
            this.aiDialogVisible = true;
            this.consultingAI = true;
            
            try {
                const response = await consultAlertAI(alert.id);
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
                this.aiResponse = 'Error: ' + (err.message || 'Failed to consult AI');
                ElMessage.error(err.message || 'Failed to consult AI');
                console.error('Error consulting AI:', err);
            } finally {
                this.consultingAI = false;
            }
        },
        getSeverityType(severity) {
            const s = (severity || '').toLowerCase();
            if (s === 'critical' || s === 'high') return 'danger';
            if (s === 'medium' || s === 'warning') return 'warning';
            if (s === 'low') return 'info';
            return 'info';
        },
        getStatusType(status) {
            const s = (status || '').toLowerCase();
            if (s === 'open') return 'danger';
            if (s === 'acknowledged') return 'warning';
            if (s === 'resolved') return 'success';
            if (s === 'closed') return 'info';
            return 'info';
        },
    }
};
</script>

<style scoped>
.alerts-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid #e4e7ed;
}

.alerts-header h2 {
    margin: 0;
    color: #303133;
}

.alerts-toolbar {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    padding: 12px 16px 0;
    align-items: center;
}

.alerts-bulk-actions {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-left: auto;
}

.selected-count {
    color: #606266;
    font-size: 13px;
}

.alerts-search {
    width: 240px;
}

.alerts-filter {
    min-width: 160px;
}

.alerts-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 16px;
}

.alerts-pagination {
    display: flex;
    justify-content: flex-end;
    padding: 0 16px 16px;
}

.alert-card {
    width: 100%;
    margin: 0;
}

.alert-card-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
}

.alert-card-title {
    font-weight: 600;
    color: #303133;
}

.alert-card :deep(.el-card__body) {
    padding: 20px;
}

.alert-card :deep(.el-descriptions) {
    margin: 0;
}

.ai-response-content {
    background-color: #f5f7fa;
    border-radius: 8px;
    padding: 16px;
    max-height: 300px;
    overflow-y: auto;
    line-height: 1.6;
}

.ai-response-content p {
    margin: 0;
    color: #303133;
}
</style>