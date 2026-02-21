<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('triggers.title') }}</h1>
      <p class="page-subtitle">{{ totalTriggers }} {{ $t('triggers.title') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('triggers.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('triggers.filterStatus')" style="width: 120px">
          <el-option :label="$t('triggers.filterAll')" value="all" />
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
        <el-button type="primary" :icon="Plus" @click="openCreate">
          {{ $t('triggers.create') }}
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
      <p>{{ $t('triggers.loading') }}</p>
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
      <el-button size="small" @click="loadTriggers(true)">{{ $t('triggers.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && triggers.length === 0"
    :description="$t('triggers.noTriggers')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && triggers.length > 0 && filteredTriggers.length === 0"
    :description="$t('triggers.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="triggers-content">
    <el-table
      v-if="filteredTriggers.length > 0"
      :data="filteredTriggers"
      border
      ref="triggersTableRef"
      row-key="id"
      @selection-change="onSelectionChange"
      @sort-change="onSortChange"
    >
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column prop="name" :label="$t('triggers.name')" min-width="160" sortable="custom" />
      <!-- Entity column removed as it's always 'item' -->
      <el-table-column prop="severity_min" :label="$t('triggers.severityMin')" width="140" align="center" sortable="custom" />
      <el-table-column :label="$t('common.enabled')" width="110" align="center" prop="enabled" sortable="custom">
        <template #default="{ row }">
          <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
            {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('triggers.status')" width="160" align="center" prop="status" sortable="custom">
        <template #default="{ row }">
          <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column :label="$t('triggers.actions')" width="200" fixed="right" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" :icon="Setting" @click="openProperties(row)">{{ $t('triggers.properties') }}</el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="onDelete(row)">{{ $t('triggers.delete') }}</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <div v-if="!loading && !error && totalTriggers > 0" class="triggers-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalTriggers"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('triggers.createTitle')" width="520px" align-center>
    <el-form :model="newTrigger" label-width="130px">
      <el-form-item :label="$t('triggers.name')">
        <el-input v-model="newTrigger.name" :placeholder="$t('triggers.name')" />
      </el-form-item>
      
      <!-- Entity is strictly 'item', hidden from UI -->
      
      <el-form-item :label="$t('triggers.severityMin')">
        <el-input-number v-model="newTrigger.severity_min" :min="0" :max="10" style="width: 100%;" />
      </el-form-item>

      <el-form-item :label="$t('triggers.itemStatus')">
        <el-select v-model="newTrigger.item_status" clearable :placeholder="$t('triggers.filterAll')" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('triggers.itemHostId')">
        <el-input-number v-model="newTrigger.alert_host_id" :min="0" :controls="false" style="width: 100%;" />
      </el-form-item>
      <el-form-item :label="$t('triggers.itemId')">
        <el-input-number v-model="newTrigger.alert_item_id" :min="0" :controls="false" style="width: 100%;" />
      </el-form-item>
      <el-form-item :label="$t('triggers.itemValueOperator')">
        <el-select v-model="newTrigger.item_value_operator" clearable :placeholder="$t('triggers.itemValueOperatorHint')" style="width: 100%;">
          <el-option v-for="op in itemOperators" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('triggers.itemValueThreshold')">
        <div class="threshold-row">
          <el-input-number v-model="newTrigger.item_value_threshold" :min="0" :controls="false" style="width: 100%;" />
          <el-input-number
            v-if="isBetweenOperator(newTrigger.item_value_operator)"
            v-model="newTrigger.item_value_threshold_max"
            :min="0"
            :controls="false"
            style="width: 100%;"
          />
        </div>
      </el-form-item>
      
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newTrigger.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('triggers.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('triggers.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('triggers.properties')} - ${selectedTrigger?.name || ''}`" width="560px">
    <el-form :model="selectedTrigger" label-width="130px">
      <el-form-item :label="$t('triggers.name')">
        <el-input v-model="selectedTrigger.name" />
      </el-form-item>
      
      <el-form-item :label="$t('triggers.severityMin')">
        <el-input-number v-model="selectedTrigger.severity_min" :min="0" :max="10" style="width: 100%;" />
      </el-form-item>

      <el-form-item :label="$t('triggers.itemStatus')">
        <el-select v-model="selectedTrigger.item_status" clearable :placeholder="$t('triggers.filterAll')" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('triggers.itemHostId')">
        <el-input-number v-model="selectedTrigger.alert_host_id" :min="0" :controls="false" style="width: 100%;" />
      </el-form-item>
      <el-form-item :label="$t('triggers.itemId')">
        <el-input-number v-model="selectedTrigger.alert_item_id" :min="0" :controls="false" style="width: 100%;" />
      </el-form-item>
      <el-form-item :label="$t('triggers.itemValueOperator')">
        <el-select v-model="selectedTrigger.item_value_operator" clearable :placeholder="$t('triggers.itemValueOperatorHint')" style="width: 100%;">
          <el-option v-for="op in itemOperators" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('triggers.itemValueThreshold')">
        <div class="threshold-row">
          <el-input-number v-model="selectedTrigger.item_value_threshold" :min="0" :controls="false" style="width: 100%;" />
          <el-input-number
            v-if="isBetweenOperator(selectedTrigger.item_value_operator)"
            v-model="selectedTrigger.item_value_threshold_max"
            :min="0"
            :controls="false"
            style="width: 100%;"
          />
        </div>
      </el-form-item>
      
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedTrigger.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('triggers.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('triggers.save') }}</el-button>
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
      <el-button @click="bulkDialogVisible = false">{{ $t('triggers.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('triggers.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedTriggers" :loading="bulkDeleting">{{ $t('triggers.delete') }}</el-button>
    </template>
  </el-dialog>
  </div>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Search, Plus, Edit, Delete, ArrowDown, Setting } from '@element-plus/icons-vue';
import { fetchTriggerData, addTrigger, updateTrigger, deleteTrigger } from '@/api/triggers';

export default {
  name: 'Trigger',
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
      triggers: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 20,
      currentPage: 1,
      totalTriggers: 0,
      sortBy: '',
      sortOrder: '',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedTriggers: [],
      newTrigger: {
        name: '',
        entity: 'item',
        severity_min: 1,
        alert_host_id: null,
        alert_item_id: null,
        item_status: null,
        item_value_threshold: null,
        item_value_threshold_max: null,
        item_value_operator: '',
        enabled: 1,
      },
      selectedTrigger: {
        id: 0,
        name: '',
        entity: 'item',
        severity_min: 1,
        alert_host_id: null,
        alert_item_id: null,
        item_status: null,
        item_value_threshold: null,
        item_value_threshold_max: null,
        item_value_operator: '',
        enabled: 1,
        status: 1,
      },
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
    filteredTriggers() {
      return this.triggers;
    },
    selectedCount() {
      return this.selectedTriggers.length;
    },
    itemOperators() {
      return [
        { label: '>', value: '>' },
        { label: '>=', value: '>=' },
        { label: '<', value: '<' },
        { label: '<=', value: '<=' },
        { label: '==', value: '==' },
        { label: '!=', value: '!=' },
        { label: this.$t('triggers.operatorBetween'), value: 'between' },
        { label: this.$t('triggers.operatorOutside'), value: 'outside' },
      ];
    },
  },
  created() {
    this.loadTriggers(true);
  },
  watch: {
    search() {
      this.currentPage = 1;
      this.loadTriggers(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadTriggers(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadTriggers(true);
    },
    currentPage() {
      this.loadTriggers();
    },
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedTriggers = selection || [];
    },
    selectAll() {
      if (this.$refs.triggersTableRef) {
        this.triggers.forEach((row) => {
          this.$refs.triggersTableRef.toggleRowSelection(row, true);
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
      this.loadTriggers(true);
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedTriggers() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedTriggers.map((trigger) => deleteTrigger(trigger.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadTriggers(true);
      } catch (err) {
        ElMessage.error(this.$t('common.bulkDeleteFailed') + ': ' + (err.message || ''));
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
        await Promise.all(this.selectedTriggers.map((trigger) => {
          const payload = this.buildTriggerPayload(trigger, {
            enabled: enabledOverride === 'nochange' ? trigger.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? trigger.status : statusOverride,
          });
          return updateTrigger(trigger.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadTriggers(true);
      } catch (err) {
        ElMessage.error(this.$t('common.bulkUpdateFailed') + ': ' + (err.message || ''));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.triggersTableRef && this.$refs.triggersTableRef.clearSelection) {
        this.$refs.triggersTableRef.clearSelection();
      }
      this.selectedTriggers = [];
    },
    async loadTriggers(reset = false) {
      if (reset) {
        this.triggers = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const triggerResp = await fetchTriggerData({
            q: this.search || undefined,
            status: this.statusFilter === 'all' ? undefined : this.statusFilter,
            limit: this.pageSize,
            offset: (this.currentPage - 1) * this.pageSize,
            sort: this.sortBy || undefined,
            order: this.sortOrder || undefined,
            with_total: 1,
          });
        const triggerData = Array.isArray(triggerResp)
          ? triggerResp
          : (triggerResp.data?.items || triggerResp.items || triggerResp.data || triggerResp.triggers || []);
        const total = triggerResp?.data?.total ?? triggerResp?.total ?? triggerData.length;
        const mapped = triggerData.map((t) => ({
          id: t.ID || t.id || 0,
          name: t.Name || t.name || '',
          entity: t.Entity || t.entity || 'item',
          severity_min: t.SeverityMin ?? t.severity_min ?? 0,
          alert_host_id: t.AlertHostID ?? t.alert_host_id ?? null,
          alert_item_id: t.AlertItemID ?? t.alert_item_id ?? null,
          item_status: t.ItemStatus ?? t.item_status ?? null,
          item_value_threshold: t.ItemValueThreshold ?? t.item_value_threshold ?? null,
          item_value_threshold_max: t.ItemValueThresholdMax ?? t.item_value_threshold_max ?? null,
          item_value_operator: t.ItemValueOperator || t.item_value_operator || '',
          enabled: t.Enabled ?? t.enabled ?? 1,
          status: t.Status ?? t.status ?? 0,
          status_reason: t.Reason || t.reason || t.Error || t.error || t.ErrorMessage || t.error_message || t.LastError || t.last_error || '',
        }));
        this.triggers = mapped;
        this.totalTriggers = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('triggers.loadFailed');
      } finally {
        this.loading = false;
      }
    },
    openCreate() {
      this.createDialogVisible = true;
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newTrigger = {
        name: '',
        entity: 'item',
        severity_min: 1,
        alert_host_id: null,
        alert_item_id: null,
        item_status: null,
        item_value_threshold: null,
        item_value_threshold_max: null,
        item_value_operator: '',
        enabled: 1,
        status: 1,
      };
    },
    async onCreate() {
      if (!this.newTrigger.name) {
        ElMessage.warning(this.$t('triggers.validationName'));
        return;
      }
      try {
        await addTrigger(this.buildTriggerPayload(this.newTrigger));
        await this.loadTriggers(true);
        this.createDialogVisible = false;
        this.newTrigger = {
          name: '',
          entity: 'item',
          severity_min: 1,
          alert_host_id: null,
          alert_item_id: null,
          item_status: null,
          item_value_threshold: null,
          item_value_threshold_max: null,
          item_value_operator: '',
          enabled: 1,
          status: 1,
        };
        ElMessage.success(this.$t('triggers.created'));
      } catch (err) {
        ElMessage.error(this.$t('triggers.createFailed') + ': ' + (err.message || ''));
      }
    },
    openProperties(trigger) {
      this.selectedTrigger = { ...trigger };
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        await updateTrigger(this.selectedTrigger.id, this.buildTriggerPayload(this.selectedTrigger));
        await this.loadTriggers(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('triggers.updated'));
      } catch (err) {
        ElMessage.error(this.$t('triggers.updateFailed') + ': ' + (err.message || ''));
      }
    },
    onDelete(trigger) {
      ElMessageBox.confirm(
        `${this.$t('triggers.delete')} ${trigger.name}?`,
        this.$t('triggers.delete'),
        {
          confirmButtonText: this.$t('triggers.delete'),
          cancelButtonText: this.$t('triggers.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteTrigger(trigger.id);
          await this.loadTriggers(true);
          ElMessage.success(this.$t('triggers.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('triggers.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('triggers.deleteCanceled'));
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
    buildTriggerPayload(trigger, overrides = {}) {
      const payload = {
        name: trigger.name,
        entity: 'item',
        severity_min: trigger.severity_min,
        enabled: trigger.enabled,
        status: trigger.status,
        alert_host_id: trigger.alert_host_id,
        alert_item_id: trigger.alert_item_id,
        item_status: trigger.item_status,
        item_value_threshold: trigger.item_value_threshold,
        item_value_threshold_max: trigger.item_value_threshold_max,
        item_value_operator: trigger.item_value_operator,
        ...overrides,
      };
      const optionalKeys = [
        'alert_host_id',
        'alert_item_id',
      ];
      Object.keys(payload).forEach((key) => {
        if (payload[key] === null || payload[key] === undefined || payload[key] === '') {
          delete payload[key];
        }
      });
      optionalKeys.forEach((key) => {
        if (payload[key] !== undefined && payload[key] <= 0) {
          delete payload[key];
        }
      });
      return payload;
    },
    isBetweenOperator(operator) {
      return operator === 'between' || operator === 'outside';
    },
  },
};
</script>

<style scoped>
.triggers-content {
  margin-top: 8px;
}

.triggers-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.loading-state {
  text-align: center;
  padding: 60px;
}

.threshold-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

:deep(.el-table__row) {
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  background-color: var(--brand-50) !important;
}
</style>
