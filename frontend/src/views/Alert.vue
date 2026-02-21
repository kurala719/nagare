<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('alerts.title') }}</h1>
      <p class="page-subtitle">{{ totalAlerts }} {{ $t('dashboard.alerts') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('alerts.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="severityFilter" :placeholder="$t('alerts.filterSeverity')" style="width: 140px">
          <el-option :label="$t('alerts.filterAll')" value="all" />
          <el-option :label="$t('alerts.severityCritical')" value="critical" />
          <el-option :label="$t('alerts.severityHigh')" value="high" />
          <el-option :label="$t('alerts.severityMedium')" value="medium" />
          <el-option :label="$t('alerts.severityLow')" value="low" />
          <el-option :label="$t('alerts.severityInfo')" value="info" />
        </el-select>

        <el-select v-model="statusFilter" :placeholder="$t('alerts.statusLabel')" style="width: 140px">
          <el-option :label="$t('alerts.filterAll')" value="all" />
          <el-option :label="$t('alerts.statusOpen')" value="open" />
          <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
          <el-option :label="$t('alerts.statusResolved')" value="resolved" />
          <el-option :label="$t('alerts.statusClosed')" value="closed" />
        </el-select>

        <el-select v-model="hostFilter" :placeholder="$t('hosts.monitor')" clearable filterable style="width: 180px">
          <el-option v-for="h in allHosts" :key="h.id" :label="h.name" :value="h.id" />
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
        <el-button type="primary" :icon="Plus" @click="openAddDialog">
          {{ $t('alerts.add') }}
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
                  <div class="header-left">
                    <el-checkbox :model-value="isSelected(alert.id)" @change="toggleSelection(alert.id, $event)" />
                    <span class="alert-card-title">{{ $t('alerts.alertLabel', { id: alert.id }) }}</span>
                  </div>
                  <div class="header-right">
                    <el-tag :type="getSeverityType(alert.severity)" effect="dark" size="small">{{ alert.severity }}</el-tag>
                    <el-tag :type="getStatusType(alert.status)" style="margin-left: 8px" size="small">{{ alert.status }}</el-tag>
                  </div>
              </div>
              
                              <div class="alert-content">
                              <div class="alert-message">{{ alert.message }}</div>
                              
                              <div v-if="alert.status_reason" class="alert-comment">
                                <div class="comment-label">{{ $t('alerts.commentLabel') || 'Comment' }}:</div>
                                <div class="comment-text">{{ alert.status_reason }}</div>
                              </div>
              
                              <el-row :gutter="20" class="alert-meta">
                                <el-col :span="8">
                                  <div class="meta-item">
                                    <el-icon><Monitor /></el-icon>
                                    <span class="meta-label">Host:</span>
                                    <router-link v-if="alert.host_name" :to="'/hosts/' + alert.host_id" class="meta-value link">
                                      {{ alert.host_name }}
                                    </router-link>
                                    <span v-else-if="alert.host_id" class="meta-value">{{ 'Host #' + alert.host_id }}</span>
                                    <span v-else class="meta-value">N/A</span>
                                  </div>
                                </el-col>
                                <el-col :span="8">
                                  <div class="meta-item">
                                    <el-icon><Document /></el-icon>
                                    <span class="meta-label">Metric:</span>
                                    <router-link v-if="alert.item_name" :to="'/items/' + alert.item_id" class="meta-value link">
                                      {{ alert.item_name }}
                                    </router-link>
                                    <span v-else-if="alert.item_id" class="meta-value">{{ 'Item #' + alert.item_id }}</span>
                                    <span v-else class="meta-value">N/A</span>
                                  </div>
                                </el-col>
                                <el-col :span="8">
                                  <div class="meta-item">
                                    <el-icon><Bell /></el-icon>
                                    <span class="meta-label">Source:</span>
                                    <span class="meta-value">{{ alert.alarm_name || 'System' }}</span>
                                  </div>
                                </el-col>
                              </el-row>
              
                              <div class="alert-footer">
                                <div class="alert-time">
                                  <el-icon><Clock /></el-icon>
                                  {{ alert.created_at ? new Date(alert.created_at).toLocaleString() : 'N/A' }}
                                </div>
                                <div class="alert-actions">
                                  <el-button type="primary" link :icon="ChatLineRound" @click="consultAI(alert)">{{ $t('alerts.consult') }}</el-button>
                                  <el-button type="warning" link :icon="Edit" @click="openEditDialog(alert)">{{ $t('alerts.edit') }}</el-button>
                                  <el-button type="danger" link :icon="Delete" @click="confirmDelete(alert)">{{ $t('alerts.remove') }}</el-button>
                                </div>
                              </div>
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
              
                  <!-- Add/Edit Dialog -->
                  <el-dialog 
                      v-model="dialogVisible" 
                      :title="isEditing ? $t('alerts.editTitle') : $t('alerts.addTitle')"
                      width="550px"
                  >
                      <el-form :model="alertForm" label-width="100px" :rules="formRules" ref="alertFormRef">
                          <el-form-item :label="$t('alerts.messageLabel')" prop="message">
                              <el-input v-model="alertForm.message" type="textarea" :rows="3" :placeholder="$t('alerts.enterMessage')" />
                          </el-form-item>
                          <el-row :gutter="20">
                              <el-col :span="12">
                                  <el-form-item :label="$t('alerts.severityLabel')" prop="severity">
                                      <el-select v-model="alertForm.severity" :placeholder="$t('alerts.selectSeverity')" style="width: 100%;">
                                          <el-option :label="$t('alerts.severityCritical')" :value="4" />
                                          <el-option :label="$t('alerts.severityHigh')" :value="3" />
                                          <el-option :label="$t('alerts.severityMedium')" :value="2" />
                                          <el-option :label="$t('alerts.severityLow')" :value="1" />
                                          <el-option :label="$t('alerts.severityInfo')" :value="0" />
                                      </el-select>
                                  </el-form-item>
                              </el-col>
                              <el-col :span="12">
                                  <el-form-item :label="$t('alerts.statusLabel')" prop="status">
                                      <el-select v-model="alertForm.status" :placeholder="$t('alerts.selectStatus')" style="width: 100%;">
                                          <el-option :label="$t('alerts.statusOpen')" value="open" />
                                          <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
                                          <el-option :label="$t('alerts.statusResolved')" value="resolved" />
                                          <el-option :label="$t('alerts.statusClosed')" value="closed" />
                                      </el-select>
                                  </el-form-item>
                              </el-col>
                          </el-row>
                          <el-form-item :label="$t('hosts.monitor')" prop="host_id">
                              <el-select v-model="alertForm.host_id" clearable filterable @change="onHostChange" style="width: 100%;">
                                  <el-option v-for="h in allHosts" :key="h.id" :label="h.name" :value="h.id" />
                              </el-select>
                          </el-form-item>
                          <el-form-item :label="$t('menu.item')" prop="item_id">
                              <el-select v-model="alertForm.item_id" clearable filterable :disabled="!alertForm.host_id" style="width: 100%;">
                                  <el-option v-for="it in filteredItems" :key="it.id" :label="it.name" :value="it.id" />
                              </el-select>
                          </el-form-item>
                          <el-form-item :label="$t('alerts.commentLabel') || 'Comment'" prop="comment">
                              <el-input v-model="alertForm.comment" type="textarea" :rows="3" :placeholder="$t('alerts.enterComment') || 'Add details or resolution...'" />
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
                  <el-dialog v-model="aiDialogVisible" :title="$t('alerts.aiTitle')" width="700px">
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
                          
                          <div class="ai-consult-tabs">
                              <el-divider content-position="left">{{ $t('alerts.aiResponse') }}</el-divider>
                              <div class="ai-response-content">
                                  <p style="white-space: pre-wrap;">{{ aiResponse }}</p>
                              </div>
              
                              <div class="ai-action-panel" style="margin-top: 24px; padding: 16px; background: var(--el-fill-color-extra-light); border-radius: 8px;">
                                  <h4 style="margin-top: 0; margin-bottom: 16px; display: flex; align-items: center; gap: 8px;">
                                      <el-icon color="#409EFF"><Edit /></el-icon>
                                      {{ $t('alerts.adoptTitle') || 'Adopt AI Suggestions' }}
                                  </h4>
                                  
                                  <el-form label-position="top">
                                      <el-form-item :label="$t('alerts.commentLabel') || 'Comment'">
                                          <el-input v-model="aiComment" type="textarea" :rows="4" :placeholder="$t('alerts.aiCommentPlaceholder') || 'Summary of AI analysis...'" />
                                      </el-form-item>
                                      
                                      <el-row :gutter="20">
                                          <el-col :span="12">
                                              <el-form-item :label="$t('alerts.statusLabel')">
                                                  <el-select v-model="aiStatus" style="width: 100%;">
                                                      <el-option :label="$t('alerts.statusOpen')" value="open" />
                                                      <el-option :label="$t('alerts.statusAcknowledged')" value="acknowledged" />
                                                      <el-option :label="$t('alerts.statusResolved')" value="resolved" />
                                                  </el-select>
                                              </el-form-item>
                                          </el-col>
                                          <el-col :span="12" style="display: flex; align-items: flex-end; padding-bottom: 18px;">
                                              <el-button type="primary" @click="applyAISuggestions" :loading="saving" style="width: 100%;">
                                                  {{ $t('common.applyAndSave') || 'Apply & Save' }}
                                              </el-button>
                                          </el-col>
                                      </el-row>
                                  </el-form>
                              </div>
                          </div>
                      </div>
                      <template #footer>
                          <el-button @click="aiDialogVisible = false">{{ $t('alerts.close') }}</el-button>
                      </template>
                  </el-dialog>  </div>
</template>

<script>
import { fetchAlertData, addAlert, updateAlert, deleteAlert, consultAlertAI } from '@/api/alerts';
import { fetchHostData } from '@/api/hosts';
import { fetchItemData } from '@/api/items';
import { ElMessage } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Plus, Search, Edit, Delete, ArrowDown, Document, Monitor, Bell, Clock, ChatLineRound } from '@element-plus/icons-vue';

export default {
    name: 'Alert',
    components: {
        Loading,
        Plus,
        Search,
        Edit,
        Delete,
        ArrowDown,
        Document,
        Monitor,
        Bell,
        Clock,
        ChatLineRound
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
                hostFilter: '',
                allHosts: [],
                filteredItems: [],
        dialogVisible: false,
                selectedAlertIds: [],
        deleteDialogVisible: false,
        aiDialogVisible: false,
        isEditing: false,
        editingId: null,
        alertToDelete: null,
        currentAlertForAI: null,
        aiResponse: '',
        aiComment: '',
        aiStatus: 'acknowledged',
                bulkForm: {
                        status: 'nochange',
                },
        alertForm: {
            message: '',
            severity: 0,
            status: 'open',
            host_id: null,
            item_id: null,
            comment: '',
        },
        formRules: {},
        // Icons for template usage
        Plus: markRaw(Plus),
        Search: markRaw(Search),
        Edit: markRaw(Edit),
        Delete: markRaw(Delete),
        ArrowDown: markRaw(ArrowDown),
        Document: markRaw(Document),
        Monitor: markRaw(Monitor),
        Bell: markRaw(Bell),
        Clock: markRaw(Clock),
        ChatLineRound: markRaw(ChatLineRound),
        Loading: markRaw(Loading)
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
        hostFilter() {
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
        this.loadAllHosts();
        this.loadAlerts(true);
    },
    methods: {
        async loadAllHosts() {
            try {
                const res = await fetchHostData({ limit: 1000 });
                this.allHosts = res.data?.items || res.items || res.data || [];
            } catch (err) {
                console.error('Failed to load hosts for filter', err);
            }
        },
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
        selectAll() {
            this.selectedAlertIds = this.alerts.map(a => a.id);
        },
        clearSelection() {
            this.selectedAlertIds = [];
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
        statusStringToInt(statusStr) {
            const statusMap = {
                'open': 0,
                'acknowledged': 1,
                'resolved': 2,
                'closed': 2
            };
            return statusMap[statusStr] !== undefined ? statusMap[statusStr] : 0;
        },
        async applyBulkUpdate() {
            if (this.selectedCount === 0) return;
            if (this.bulkForm.status === 'nochange') {
                ElMessage.warning(this.$t('common.bulkUpdateNoChanges'));
                return;
            }

            this.bulkUpdating = true;
            try {
                const statusOverride = this.statusStringToInt(this.bulkForm.status);
                await Promise.all(this.alerts.filter((alert) => this.selectedAlertIds.includes(alert.id)).map((alert) => {
                    const payload = {
                        message: alert.message,
                        severity: parseInt(alert.severity) || 0,
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
                    host_id: this.hostFilter || undefined,
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
                    id: a.id,
                    message: a.message || '',
                    severity: this.normalizeSeverity(a.severity ?? ''),
                    status: this.normalizeStatus(a.status ?? ''),
                    status_reason: a.comment || '',
                    created_at: a.created_at || '',
                    host_id: a.host_id,
                    host_name: a.host_name || '',
                    item_id: a.item_id,
                    item_name: a.item_name || '',
                    alarm_id: a.alarm_id,
                    alarm_name: a.alarm_name || ''
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
        async onHostChange(hostId) {
            this.alertForm.item_id = null;
            this.filteredItems = [];
            if (!hostId) return;
            
            try {
                const res = await fetchItemData({ host_id: hostId, limit: 1000 });
                this.filteredItems = res.data?.items || res.items || res.data || [];
            } catch (err) {
                console.error('Failed to load items for host', err);
            }
        },
        openAddDialog() {
            this.isEditing = false;
            this.editingId = null;
            this.alertForm = {
                message: '',
                severity: 0,
                status: 'open',
                host_id: null,
                item_id: null,
                comment: '',
            };
            this.filteredItems = [];
            this.dialogVisible = true;
        },
        openEditDialog(alert) {
            this.isEditing = true;
            this.editingId = alert.id;
            this.alertForm = {
                message: alert.message,
                severity: this.severityLabelToInt(alert.severity),
                status: alert.status,
                host_id: alert.host_id,
                item_id: alert.item_id,
                comment: alert.status_reason || '',
            };
            if (alert.host_id) {
                this.onHostChange(alert.host_id).then(() => {
                    this.alertForm.item_id = alert.item_id;
                });
            }
            this.dialogVisible = true;
        },
        severityLabelToInt(label) {
            const map = { 'info': 0, 'low': 1, 'medium': 2, 'high': 3, 'critical': 4 };
            return map[label.toLowerCase()] ?? 0;
        },
        async saveAlert() {
            try {
                await this.$refs.alertFormRef.validate();
            } catch {
                return;
            }

            this.saving = true;
            try {
                // Convert status string to integer and ensure severity is integer for API
                const payload = {
                    ...this.alertForm,
                    severity: parseInt(this.alertForm.severity) || 0,
                    status: this.statusStringToInt(this.alertForm.status)
                };
                
                if (this.isEditing) {
                    await updateAlert(this.editingId, payload);
                    ElMessage.success('Alert updated successfully');
                } else {
                    await addAlert(payload);
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
            this.aiComment = '';
            this.aiStatus = alert.status === 'resolved' ? 'resolved' : 'acknowledged';
            this.aiDialogVisible = true;
            this.consultingAI = true;
            
            try {
                const response = await consultAlertAI(alert.id);
                let content = '';
                
                // Handle different response formats
                if (typeof response === 'string') {
                    content = response;
                } else if (response.message) {
                    content = response.message;
                } else if (response.content) {
                    content = response.content;
                } else if (response.data) {
                    if (typeof response.data === 'string') {
                        content = response.data;
                    } else if (response.data.message) {
                        content = response.data.message;
                    } else if (response.data.content) {
                        content = response.data.content;
                    } else if (response.data.Content) {
                        content = response.data.Content;
                    } else {
                        content = JSON.stringify(response.data, null, 2);
                    }
                } else {
                    content = JSON.stringify(response, null, 2);
                }
                
                this.aiResponse = content;
                this.aiComment = content; // Pre-fill with AI response
            } catch (err) {
                this.aiResponse = 'Error: ' + (err.message || 'Failed to consult AI');
                ElMessage.error(err.message || 'Failed to consult AI');
                console.error('Error consulting AI:', err);
            } finally {
                this.consultingAI = false;
            }
        },
        async applyAISuggestions() {
            if (!this.currentAlertForAI) return;
            
            this.saving = true;
            try {
                const payload = {
                    message: this.currentAlertForAI.message,
                    severity: this.severityLabelToInt(this.currentAlertForAI.severity),
                    status: this.statusStringToInt(this.aiStatus),
                    host_id: this.currentAlertForAI.host_id,
                    item_id: this.currentAlertForAI.item_id,
                    comment: this.aiComment
                };
                
                await updateAlert(this.currentAlertForAI.id, payload);
                ElMessage.success('AI suggestions applied and alert updated');
                this.aiDialogVisible = false;
                await this.loadAlerts(true);
            } catch (err) {
                ElMessage.error(err.message || 'Failed to apply AI suggestions');
                console.error('Error applying AI suggestions:', err);
            } finally {
                this.saving = false;
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
.alerts-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 8px;
}

.alert-card {
  border: 1px solid var(--el-border-color-lighter);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.alert-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
}

.alert-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.alert-card-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.alert-message {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 12px;
  line-height: 1.5;
}

.alert-comment {
  background: var(--el-color-info-light-9);
  border-left: 4px solid var(--el-color-info);
  padding: 12px;
  margin-bottom: 16px;
  border-radius: 0 4px 4px 0;
}

.comment-label {
  font-size: 12px;
  font-weight: bold;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
  text-transform: uppercase;
}

.comment-text {
  font-size: 14px;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  line-height: 1.6;
}

.alert-meta {
  background: var(--el-fill-color-blank);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-extra-light);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.meta-item .el-icon {
  color: var(--el-text-color-secondary);
}

.meta-label {
  color: var(--el-text-color-secondary);
  margin-right: 4px;
}

.meta-value {
  color: var(--el-text-color-primary);
  font-weight: 500;
}

.meta-value.link {
  color: var(--el-color-primary);
  text-decoration: none;
}

.meta-value.link:hover {
  text-decoration: underline;
}

.alert-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px dashed var(--el-border-color-extra-light);
}

.alert-time {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.alert-actions {
  display: flex;
  gap: 8px;
}

.alerts-pagination {
  margin-top: 32px;
  display: flex;
  justify-content: flex-end;
}

.ai-response-content {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  padding: 20px;
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.7;
  border: 1px solid var(--el-border-color-lighter);
}

.ai-response-content p {
  margin: 0;
  color: var(--el-text-color-primary);
}
</style>
