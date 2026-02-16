<template>
  <div class="media-types-toolbar">
    <div class="media-types-filters">
      <span class="filter-label">{{ $t('mediaTypes.search') }}</span>
      <el-input v-model="search" :placeholder="$t('mediaTypes.search')" clearable class="media-types-search" />
      <span class="filter-label">{{ $t('mediaTypes.filterStatus') }}</span>
      <el-select v-model="statusFilter" :placeholder="$t('mediaTypes.filterStatus')" class="media-types-filter">
        <el-option :label="$t('mediaTypes.filterAll')" value="all" />
        <el-option :label="$t('common.statusInactive')" :value="0" />
        <el-option :label="$t('common.statusActive')" :value="1" />
        <el-option :label="$t('common.statusError')" :value="2" />
        <el-option :label="$t('common.statusSyncing')" :value="3" />
      </el-select>
      <span class="filter-label">{{ $t('common.sort') }}</span>
      <el-select v-model="sortKey" class="media-types-filter">
        <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
        <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
        <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
        <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
        <el-option :label="$t('common.sortStatusAsc')" value="status_asc" />
        <el-option :label="$t('common.sortStatusDesc')" value="status_desc" />
      </el-select>
      <div class="media-types-bulk-actions">
        <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
        <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
          {{ $t('common.bulkUpdate') }}
        </el-button>
        <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
          {{ $t('common.bulkDelete') }}
        </el-button>
      </div>
    </div>
    <el-button type="primary" @click="openCreate">
      {{ $t('mediaTypes.create') }}
    </el-button>
  </div>

  <div v-if="loading" class="loading-state">
    <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
    <p>{{ $t('mediaTypes.loading') }}</p>
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
      <el-button size="small" @click="loadMediaTypes">{{ $t('mediaTypes.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && mediaTypes.length === 0"
    :description="$t('mediaTypes.noMediaTypes')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && mediaTypes.length > 0 && filteredMediaTypes.length === 0"
    :description="$t('mediaTypes.noResults')"
    style="margin: 40px;"
  />

  <div
    v-if="!loading && !error"
    class="media-types-scroll"
  >
    <el-table
      v-if="filteredMediaTypes.length > 0"
      :data="filteredMediaTypes"
      border
      style="margin: 20px; width: calc(100% - 40px);"
      ref="mediaTypesTableRef"
      row-key="id"
      @selection-change="onSelectionChange"
    >
    <el-table-column type="selection" width="50" />
    <el-table-column prop="name" :label="$t('mediaTypes.name')" min-width="160" />
    <el-table-column prop="key" :label="$t('mediaTypes.key')" min-width="160" />
    <el-table-column :label="$t('common.enabled')" min-width="110">
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column :label="$t('mediaTypes.status')" min-width="160">
      <template #default="{ row }">
        <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
          <el-tag :type="getStatusInfo(row.status).type">
            {{ getStatusInfo(row.status).label }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column prop="description" :label="$t('mediaTypes.description')" min-width="200" show-overflow-tooltip />
    <el-table-column :label="$t('mediaTypes.actions')" min-width="200" fixed="right">
      <template #default="{ row }">
        <el-button size="small" @click="openProperties(row)">{{ $t('mediaTypes.properties') }}</el-button>
        <el-button size="small" type="danger" @click="onDelete(row)">{{ $t('mediaTypes.delete') }}</el-button>
      </template>
    </el-table-column>
    </el-table>
  </div>
  <div v-if="!loading && !error && totalMediaTypes > 0" class="media-types-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalMediaTypes"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('mediaTypes.createTitle')" width="500px" align-center>
    <el-form :model="newMediaType" label-width="120px">
      <el-form-item :label="$t('mediaTypes.name')">
        <el-input v-model="newMediaType.name" :placeholder="$t('mediaTypes.name')" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.key')">
        <el-input v-model="newMediaType.key" :placeholder="$t('mediaTypes.keyHint')" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.template')">
        <el-input
          v-model="newMediaType.template"
          type="textarea"
          :placeholder="$t('mediaTypes.templateHint')"
        />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.fields')">
        <el-input
          v-model="newMediaType.fieldsText"
          type="textarea"
          :placeholder="$t('mediaTypes.fieldsHint')"
        />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newMediaType.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.status')">
        <el-select v-model="newMediaType.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.description')">
        <el-input v-model="newMediaType.description" type="textarea" :placeholder="$t('mediaTypes.description')" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('mediaTypes.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('mediaTypes.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('mediaTypes.properties')} - ${selectedMediaType?.name || ''}`" width="600px">
    <el-form :model="selectedMediaType" label-width="120px">
      <el-form-item :label="$t('mediaTypes.name')">
        <el-input v-model="selectedMediaType.name" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.key')">
        <el-input v-model="selectedMediaType.key" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.template')">
        <el-input
          v-model="selectedMediaType.template"
          type="textarea"
          :placeholder="$t('mediaTypes.templateHint')"
        />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.fields')">
        <el-input
          v-model="selectedMediaType.fieldsText"
          type="textarea"
          :placeholder="$t('mediaTypes.fieldsHint')"
        />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedMediaType.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.status')">
        <el-select v-model="selectedMediaType.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('mediaTypes.description')">
        <el-input type="textarea" v-model="selectedMediaType.description" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('mediaTypes.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('mediaTypes.save') }}</el-button>
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
      <el-form-item :label="$t('mediaTypes.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('mediaTypes.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('mediaTypes.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedMediaTypes" :loading="bulkDeleting">{{ $t('mediaTypes.delete') }}</el-button>
    </template>
  </el-dialog>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { fetchMediaTypeData, addMediaType, updateMediaType, deleteMediaType } from '@/api/mediaTypes';

export default {
  name: 'MediaType',
  components: { Loading },
  data() {
    return {
      mediaTypes: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 20,
      currentPage: 1,
      totalMediaTypes: 0,
      sortKey: 'updated_desc',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedMediaTypeRows: [],
      newMediaType: { name: '', key: '', enabled: 1, status: 1, description: '', template: '', fieldsText: '[]' },
      selectedMediaType: { id: 0, name: '', key: '', enabled: 1, status: 1, description: '', template: '', fieldsText: '[]' },
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
    };
  },
  computed: {
    filteredMediaTypes() {
      return this.mediaTypes;
    },
    selectedCount() {
      return this.selectedMediaTypeRows.length;
    },
  },
  created() {
    this.loadMediaTypes(true);
  },
  watch: {
    search() {
      this.currentPage = 1;
      this.loadMediaTypes(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadMediaTypes(true);
    },
    sortKey() {
      this.currentPage = 1;
      this.loadMediaTypes(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadMediaTypes(true);
    },
    currentPage() {
      this.loadMediaTypes();
    },
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedMediaTypeRows = selection || [];
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedMediaTypes() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedMediaTypeRows.map((mediaType) => deleteMediaType(mediaType.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadMediaTypes(true);
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
        await Promise.all(this.selectedMediaTypeRows.map((mediaType) => {
          const payload = {
            name: mediaType.name,
            key: mediaType.key,
            enabled: enabledOverride === 'nochange' ? mediaType.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? mediaType.status : statusOverride,
            description: mediaType.description,
          };
          return updateMediaType(mediaType.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadMediaTypes(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.mediaTypesTableRef && this.$refs.mediaTypesTableRef.clearSelection) {
        this.$refs.mediaTypesTableRef.clearSelection();
      }
      this.selectedMediaTypeRows = [];
    },
    async loadMediaTypes(reset = false) {
      if (reset) {
        this.mediaTypes = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
        const response = await fetchMediaTypeData({
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
          : (response.data?.items || response.items || response.data || response.mediaTypes || []);
        const total = response?.data?.total ?? response?.total ?? data.length;
        const mapped = data.map((t) => ({
          id: t.ID || t.id || 0,
          name: t.Name || t.name || '',
          key: t.Key || t.key || '',
          enabled: t.Enabled ?? t.enabled ?? 1,
          status: t.Status ?? t.status ?? 0,
          status_reason: t.Reason || t.reason || t.Error || t.error || t.ErrorMessage || t.error_message || t.LastError || t.last_error || '',
          description: t.Description || t.description || '',
          template: t.Template || t.template || '',
          fields: t.Fields || t.fields || [],
        }));
        this.mediaTypes = mapped;
        this.totalMediaTypes = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('mediaTypes.loadFailed');
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
    openCreate() {
      this.createDialogVisible = true;
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newMediaType = { name: '', key: '', enabled: 1, status: 1, description: '', template: '', fieldsText: '[]' };
    },
    async onCreate() {
      if (!this.newMediaType.name || !this.newMediaType.key) {
        ElMessage.warning(this.$t('mediaTypes.validationName'));
        return;
      }
      try {
        const fields = this.parseJsonInput(this.newMediaType.fieldsText, []);
        if (fields === null) return;
        await addMediaType({
          name: this.newMediaType.name,
          key: this.newMediaType.key,
          enabled: this.newMediaType.enabled,
          status: this.newMediaType.status,
          description: this.newMediaType.description,
          template: this.newMediaType.template,
          fields,
        });
        await this.loadMediaTypes(true);
        this.createDialogVisible = false;
        this.newMediaType = { name: '', key: '', enabled: 1, status: 1, description: '', template: '', fieldsText: '[]' };
        ElMessage.success(this.$t('mediaTypes.created'));
      } catch (err) {
        ElMessage.error(this.$t('mediaTypes.createFailed') + ': ' + (err.message || ''));
      }
    },
    openProperties(mediaType) {
      this.selectedMediaType = {
        ...mediaType,
        template: mediaType.template || '',
        fieldsText: JSON.stringify(mediaType.fields || [], null, 2),
      };
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        const fields = this.parseJsonInput(this.selectedMediaType.fieldsText, []);
        if (fields === null) return;
        await updateMediaType(this.selectedMediaType.id, {
          name: this.selectedMediaType.name,
          key: this.selectedMediaType.key,
          enabled: this.selectedMediaType.enabled,
          status: this.selectedMediaType.status,
          description: this.selectedMediaType.description,
          template: this.selectedMediaType.template,
          fields,
        });
        await this.loadMediaTypes(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('mediaTypes.updated'));
      } catch (err) {
        ElMessage.error(this.$t('mediaTypes.updateFailed') + ': ' + (err.message || ''));
      }
    },
    parseJsonInput(value, fallback) {
      if (value === undefined || value === null) return fallback;
      const text = String(value).trim();
      if (text === '') return fallback;
      try {
        return JSON.parse(text);
      } catch (err) {
        ElMessage.error(this.$t('common.invalidJson'));
        return null;
      }
    },
    onDelete(mediaType) {
      ElMessageBox.confirm(
        `${this.$t('mediaTypes.delete')} ${mediaType.name}?`,
        this.$t('mediaTypes.delete'),
        {
          confirmButtonText: this.$t('mediaTypes.delete'),
          cancelButtonText: this.$t('mediaTypes.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteMediaType(mediaType.id);
          await this.loadMediaTypes(true);
          ElMessage.success(this.$t('mediaTypes.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('mediaTypes.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('mediaTypes.deleteCanceled'));
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
.media-types-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 16px 20px 0;
}

.media-types-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.media-types-bulk-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.media-types-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 16px;
}

.selected-count {
  color: #606266;
  font-size: 13px;
}

.media-types-search {
  width: 240px;
}

.media-types-filter {
  min-width: 160px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}
</style>
