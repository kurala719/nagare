<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('actions.title') }}</h1>
      <p class="page-subtitle">{{ totalActions }} {{ $t('actions.title') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('actions.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('actions.filterStatus')" style="width: 120px">
          <el-option :label="$t('actions.filterAll')" value="all" />
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </div>

      <div class="action-group">
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ $t('actions.create') }}
        </el-button>
        <el-dropdown trigger="click" v-if="selectedCount > 0" style="margin-left: 8px">
          <el-button>
            {{ $t('common.selectedCount', { count: selectedCount }) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
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
      <p>{{ $t('actions.loading') }}</p>
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
      <el-button size="small" @click="loadActions">{{ $t('actions.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && actions.length === 0"
    :description="$t('actions.noActions')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && actions.length > 0 && filteredActions.length === 0"
    :description="$t('actions.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="actions-content">
    <el-table
      v-if="filteredActions.length > 0"
      :data="filteredActions"
      border
      ref="actionsTableRef"
      row-key="id"
      @selection-change="onSelectionChange"
      @sort-change="onSortChange"
    >
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column prop="name" :label="$t('actions.name')" min-width="160" sortable="custom" />
      <el-table-column :label="$t('actions.media')" min-width="160" prop="media_id" sortable="custom">
        <template #default="{ row }">
          {{ mediaName(row.media_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="template" :label="$t('actions.template')" min-width="220" show-overflow-tooltip sortable="custom" />
      <el-table-column :label="$t('common.enabled')" width="110" align="center" prop="enabled" sortable="custom">
        <template #default="{ row }">
          <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
            {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('actions.status')" width="160" align="center" prop="status" sortable="custom">
        <template #default="{ row }">
          <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column :label="$t('actions.actions')" width="200" fixed="right" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" :icon="Setting" @click="openProperties(row)">{{ $t('actions.properties') }}</el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="onDelete(row)">{{ $t('actions.delete') }}</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <div v-if="!loading && !error && totalActions > 0" class="actions-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalActions"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('actions.createTitle')" width="560px" align-center>
    <el-form :model="newAction" label-width="130px">
      <el-form-item :label="$t('actions.name')">
        <el-input v-model="newAction.name" :placeholder="$t('actions.name')" />
      </el-form-item>
      <el-form-item :label="$t('actions.media')">
        <el-select v-model="newAction.media_id" style="width: 100%;">
          <el-option v-for="media in mediaOptions" :key="media.id" :label="media.name" :value="media.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('actions.template')">
        <el-input v-model="newAction.template" type="textarea" :placeholder="$t('actions.templateHint')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newAction.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('actions.description')">
        <el-input v-model="newAction.description" type="textarea" :placeholder="$t('actions.description')" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('actions.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('actions.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('actions.properties')} - ${selectedAction?.name || ''}`" width="600px">
    <el-form :model="selectedAction" label-width="130px">
      <el-form-item :label="$t('actions.name')">
        <el-input v-model="selectedAction.name" />
      </el-form-item>
      <el-form-item :label="$t('actions.media')">
        <el-select v-model="selectedAction.media_id" style="width: 100%;">
          <el-option v-for="media in mediaOptions" :key="media.id" :label="media.name" :value="media.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('actions.template')">
        <el-input v-model="selectedAction.template" type="textarea" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedAction.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('actions.description')">
        <el-input type="textarea" v-model="selectedAction.description" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('actions.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('actions.save') }}</el-button>
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
      <el-form-item :label="$t('actions.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('actions.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('actions.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedActions" :loading="bulkDeleting">{{ $t('actions.delete') }}</el-button>
    </template>
  </el-dialog>
  </div>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Search, Plus, Edit, Delete, ArrowDown, Setting } from '@element-plus/icons-vue';
import { fetchActionData, addAction, updateAction, deleteAction } from '@/api/actions';
import { fetchMediaData } from '@/api/media';

export default {
  name: 'Action',
  components: {
    Loading,
    Search,
    Plus,
    Edit,
    Delete,
    ArrowDown,
    Setting
  },
  data() {
    return {
      actions: [],
      mediaOptions: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 20,
      currentPage: 1,
      totalActions: 0,
      sortBy: '',
      sortOrder: '',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedActions: [],
      newAction: { name: '', media_id: 0, template: '', enabled: 1, status: 1, description: '' },
      selectedAction: { id: 0, name: '', media_id: 0, template: '', enabled: 1, status: 1, description: '' },
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
      // Icons for template usage
      Plus: markRaw(Plus),
      Search: markRaw(Search),
      Edit: markRaw(Edit),
      Delete: markRaw(Delete),
      ArrowDown: markRaw(ArrowDown),
      Setting: markRaw(Setting),
      Loading: markRaw(Loading)
    };
  },
  computed: {
    filteredActions() {
      return this.actions;
    },
    selectedCount() {
      return this.selectedActions.length;
    },
  },
  created() {
    this.loadActions(true);
  },
  watch: {
    search() {
      this.currentPage = 1;
      this.loadActions(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadActions(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadActions(true);
    },
    currentPage() {
      this.loadActions();
    },
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedActions = selection || [];
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
      this.loadActions(true);
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedActions() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedActions.map((action) => deleteAction(action.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadActions(true);
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
        await Promise.all(this.selectedActions.map((action) => {
          const payload = {
            name: action.name,
            media_id: action.media_id,
            template: action.template,
            enabled: enabledOverride === 'nochange' ? action.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? action.status : statusOverride,
            description: action.description,
          };
          return updateAction(action.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadActions(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.actionsTableRef && this.$refs.actionsTableRef.clearSelection) {
        this.$refs.actionsTableRef.clearSelection();
      }
      this.selectedActions = [];
    },
    async loadActions(reset = false) {
      if (reset) {
        this.actions = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const [actionResp, mediaResp] = await Promise.all([
          fetchActionData({
            q: this.search || undefined,
            status: this.statusFilter === 'all' ? undefined : this.statusFilter,
            limit: this.pageSize,
            offset: (this.currentPage - 1) * this.pageSize,
            sort: this.sortBy || undefined,
            order: this.sortOrder || undefined,
            with_total: 1,
          }),
          this.mediaOptions.length === 0 ? fetchMediaData({ limit: 100, offset: 0 }) : Promise.resolve(null),
        ]);
        const data = Array.isArray(actionResp)
          ? actionResp
          : (actionResp.data?.items || actionResp.items || actionResp.data || actionResp.actions || []);
        const total = actionResp?.data?.total ?? actionResp?.total ?? data.length;
        const mapped = data.map((a) => ({
          id: a.ID || a.id || 0,
          name: a.Name || a.name || '',
          media_id: a.MediaID || a.media_id || 0,
          template: a.Template || a.template || '',
          enabled: a.Enabled ?? a.enabled ?? 1,
          status: a.Status ?? a.status ?? 0,
          status_reason: a.Reason || a.reason || a.Error || a.error || a.ErrorMessage || a.error_message || a.LastError || a.last_error || '',
          description: a.Description || a.description || '',
        }));
        this.actions = mapped;
        this.totalActions = Number.isFinite(total) ? total : mapped.length;
        if (mediaResp) {
          const mediaData = Array.isArray(mediaResp) ? mediaResp : (mediaResp.data || mediaResp.media || []);
          this.mediaOptions = mediaData.map((m) => ({
            id: m.ID || m.id || 0,
            name: m.Name || m.name || '',
          }));
        }
      } catch (err) {
        this.error = err.message || this.$t('actions.loadFailed');
      } finally {
        this.loading = false;
      }
    },
    mediaName(id) {
      const match = this.mediaOptions.find((m) => m.id === id);
      return match ? match.name : id;
    },
    openCreate() {
      this.createDialogVisible = true;
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newAction = { name: '', media_id: 0, template: '', enabled: 1, status: 1, description: '' };
    },
    async onCreate() {
      if (!this.newAction.name) {
        ElMessage.warning(this.$t('actions.validationName'));
        return;
      }
      try {
        await addAction(this.newAction);
        await this.loadActions(true);
        this.createDialogVisible = false;
        this.newAction = { name: '', media_id: 0, template: '', enabled: 1, status: 1, description: '' };
        ElMessage.success(this.$t('actions.created'));
      } catch (err) {
        ElMessage.error(this.$t('actions.createFailed') + ': ' + (err.message || ''));
      }
    },
    openProperties(action) {
      this.selectedAction = { ...action };
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        await updateAction(this.selectedAction.id, {
          name: this.selectedAction.name,
          media_id: this.selectedAction.media_id,
          template: this.selectedAction.template,
          enabled: this.selectedAction.enabled,
          status: this.selectedAction.status,
          description: this.selectedAction.description,
        });
        await this.loadActions(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('actions.updated'));
      } catch (err) {
        ElMessage.error(this.$t('actions.updateFailed') + ': ' + (err.message || ''));
      }
    },
    onDelete(action) {
      ElMessageBox.confirm(
        `${this.$t('actions.delete')} ${action.name}?`,
        this.$t('actions.delete'),
        {
          confirmButtonText: this.$t('actions.delete'),
          cancelButtonText: this.$t('actions.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteAction(action.id);
          await this.loadActions(true);
          ElMessage.success(this.$t('actions.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('actions.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('actions.deleteCanceled'));
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
  },
};
</script>

<style scoped>
.actions-content {
  margin-top: 8px;
}

.actions-pagination {
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
