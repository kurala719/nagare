<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('media.title') }}</h1>
      <p class="page-subtitle">{{ totalMedia }} {{ $t('media.title') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('media.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>

        <el-select v-model="statusFilter" :placeholder="$t('media.filterStatus')" style="width: 140px">
          <el-option :label="$t('media.filterAll')" value="all" />
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
          {{ $t('media.create') }}
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
      <p>{{ $t('media.loading') }}</p>
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
      <el-button size="small" @click="loadMedia(true)">{{ $t('media.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && mediaList.length === 0"
    :description="$t('media.noMedia')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && mediaList.length > 0 && filteredMedia.length === 0"
    :description="$t('media.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="media-content">
    <el-table
      v-if="filteredMedia.length > 0"
      :data="filteredMedia"
      border
      ref="mediaTableRef"
      row-key="id"
      @selection-change="onSelectionChange"
      @sort-change="onSortChange"
    >
      <el-table-column type="selection" width="50" align="center" />
      <el-table-column prop="name" :label="$t('media.name')" min-width="160" sortable="custom" />
      <el-table-column :label="$t('media.type')" width="140" prop="media_type_id" sortable="custom">
        <template #default="{ row }">
          {{ mediaTypeName(row.media_type_id) }}
        </template>
      </el-table-column>
      <el-table-column prop="target" :label="$t('media.target')" min-width="200" show-overflow-tooltip sortable="custom" />
      <el-table-column :label="$t('common.enabled')" width="110" align="center" prop="enabled" sortable="custom">
        <template #default="{ row }">
          <el-tag :type="row.enabled === 1 ? 'success' : 'info'" size="small">
            {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('media.status')" width="160" align="center" prop="status" sortable="custom">
        <template #default="{ row }">
          <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column :label="$t('media.actions')" width="200" fixed="right" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" :icon="Setting" @click="openProperties(row)">{{ $t('media.properties') }}</el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="onDelete(row)">{{ $t('media.delete') }}</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </div>
  <div v-if="!loading && !error && totalMedia > 0" class="media-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalMedia"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('media.createTitle')" width="500px" align-center>
    <el-form :model="newMedia" label-width="120px">
      <el-form-item :label="$t('media.name')">
        <el-input v-model="newMedia.name" :placeholder="$t('media.name')" />
      </el-form-item>
      <el-form-item :label="$t('media.type')">
        <el-select v-model="newMedia.media_type_id" style="width: 100%;">
          <el-option v-for="mediaType in mediaTypeOptions" :key="mediaType.id" :label="mediaType.name" :value="mediaType.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('media.target')">
        <el-input
          v-model="newMedia.target"
          :placeholder="$t('media.targetHint')"
          :disabled="newMediaTypeHasTemplate"
        />
      </el-form-item>
      <el-form-item
        v-for="field in newMediaFields"
        :key="field.key"
        :label="fieldLabel(field)"
        :required="field.required"
      >
        <el-input v-model="newMedia.params[field.key]" :placeholder="fieldPlaceholder(field)" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newMedia.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('media.status')">
        <el-select v-model="newMedia.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('media.description')">
        <el-input v-model="newMedia.description" type="textarea" :placeholder="$t('media.description')" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('media.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('media.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('media.properties')} - ${selectedMedia?.name || ''}`" width="600px">
    <el-form :model="selectedMedia" label-width="120px">
      <el-form-item :label="$t('media.name')">
        <el-input v-model="selectedMedia.name" />
      </el-form-item>
      <el-form-item :label="$t('media.type')">
        <el-select v-model="selectedMedia.media_type_id" style="width: 100%;">
          <el-option v-for="mediaType in mediaTypeOptions" :key="mediaType.id" :label="mediaType.name" :value="mediaType.id" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('media.target')">
        <el-input
          v-model="selectedMedia.target"
          :placeholder="$t('media.targetHint')"
          :disabled="selectedMediaTypeHasTemplate"
        />
      </el-form-item>
      <el-form-item
        v-for="field in selectedMediaFields"
        :key="field.key"
        :label="fieldLabel(field)"
        :required="field.required"
      >
        <el-input v-model="selectedMedia.params[field.key]" :placeholder="fieldPlaceholder(field)" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedMedia.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('media.status')">
        <el-select v-model="selectedMedia.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
      <el-form-item :label="$t('media.description')">
        <el-input type="textarea" v-model="selectedMedia.description" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('media.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('media.save') }}</el-button>
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
      <el-form-item :label="$t('media.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('media.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('media.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedMedia" :loading="bulkDeleting">{{ $t('media.delete') }}</el-button>
    </template>
  </el-dialog>
  </div>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { markRaw } from 'vue';
import { Loading, Search, Plus, Edit, Delete, ArrowDown, Setting } from '@element-plus/icons-vue';
import { fetchMediaData, addMedia, updateMedia, deleteMedia } from '@/api/media';
import { fetchMediaTypeData } from '@/api/mediaTypes';

export default {
  name: 'Media',
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
      mediaList: [],
      mediaTypeOptions: [],
      loading: false,
      error: null,
      search: '',
      statusFilter: 'all',
      pageSize: 20,
      currentPage: 1,
      totalMedia: 0,
      sortBy: '',
      sortOrder: '',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedMediaRows: [],
      newMedia: { name: '', media_type_id: 0, target: '', params: {}, enabled: 1, status: 1, description: '' },
      selectedMedia: { id: 0, name: '', media_type_id: 0, target: '', params: {}, enabled: 1, status: 1, description: '' },
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
    filteredMedia() {
      return this.mediaList;
    },
    selectedCount() {
      return this.selectedMediaRows.length;
    },
    newMediaType() {
      return this.mediaTypeOptions.find((t) => t.id === this.newMedia.media_type_id) || null;
    },
    selectedMediaType() {
      return this.mediaTypeOptions.find((t) => t.id === this.selectedMedia.media_type_id) || null;
    },
    newMediaFields() {
      return this.normalizeFields(this.newMediaType?.fields || []);
    },
    selectedMediaFields() {
      return this.normalizeFields(this.selectedMediaType?.fields || []);
    },
    newMediaTypeHasTemplate() {
      return Boolean(this.newMediaType && String(this.newMediaType.template || '').trim());
    },
    selectedMediaTypeHasTemplate() {
      return Boolean(this.selectedMediaType && String(this.selectedMediaType.template || '').trim());
    },
  },
  watch: {
    'newMedia.media_type_id'(value) {
      this.applyMediaTypeDefaults(this.newMedia, value, false);
    },
    'selectedMedia.media_type_id'(value) {
      this.applyMediaTypeDefaults(this.selectedMedia, value, true);
    },
    search() {
      this.currentPage = 1;
      this.loadMedia(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadMedia(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadMedia(true);
    },
    currentPage() {
      this.loadMedia();
    },
  },
  created() {
    this.loadMediaTypes();
    this.loadMedia(true);
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedMediaRows = selection || [];
    },
    selectAll() {
      if (this.$refs.mediaTableRef) {
        this.mediaList.forEach((row) => {
          this.$refs.mediaTableRef.toggleRowSelection(row, true);
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
      this.loadMedia(true);
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedMedia() {
      if (this.selectedCount === 0) return;

      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedMediaRows.map((media) => deleteMedia(media.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadMedia(true);
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
        await Promise.all(this.selectedMediaRows.map((media) => {
          const payload = {
            name: media.name,
            media_type_id: media.media_type_id,
            target: media.target,
            enabled: enabledOverride === 'nochange' ? media.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? media.status : statusOverride,
            description: media.description,
          };
          return updateMedia(media.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadMedia(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.mediaTableRef && this.$refs.mediaTableRef.clearSelection) {
        this.$refs.mediaTableRef.clearSelection();
      }
      this.selectedMediaRows = [];
    },
    async loadMediaTypes() {
      try {
        const mediaTypeResp = await fetchMediaTypeData({ limit: 100, offset: 0 });
        const typeData = Array.isArray(mediaTypeResp) ? mediaTypeResp : (mediaTypeResp.data || mediaTypeResp.mediaTypes || []);
        this.mediaTypeOptions = typeData.map((t) => ({
          id: t.ID || t.id || 0,
          name: t.Name || t.name || '',
          type: t.Type || t.type || '',
          description: t.Description || t.description || '',
          template: t.Template || t.template || '',
          fields: t.Fields || t.fields || [],
        }));
      } catch (err) {
        console.error('Error loading media types:', err);
      }
    },
    async loadMedia(reset = false) {
      if (reset) {
        this.mediaList = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const mediaResp = await fetchMediaData({
          q: this.search || undefined,
          status: this.statusFilter === 'all' ? undefined : this.statusFilter,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
          with_total: 1,
        });
        const data = Array.isArray(mediaResp)
          ? mediaResp
          : (mediaResp.data?.items || mediaResp.items || mediaResp.data || mediaResp.media || []);
        const total = mediaResp?.data?.total ?? mediaResp?.total ?? data.length;
        const mapped = data.map((m) => ({
          id: m.ID || m.id || 0,
          name: m.Name || m.name || '',
          media_type_id: m.MediaTypeID || m.media_type_id || 0,
          target: m.Target || m.target || '',
          enabled: m.Enabled ?? m.enabled ?? 1,
          status: m.Status ?? m.status ?? 0,
          status_reason: m.Reason || m.reason || m.Error || m.error || m.ErrorMessage || m.error_message || m.LastError || m.last_error || '',
          description: m.Description || m.description || '',
          params: m.Params || m.params || {},
        }));
        this.mediaList = mapped;
        this.totalMedia = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('media.loadFailed');
        ElMessage.error(this.error);
      } finally {
        this.loading = false;
      }
    },
    openCreate() {
      this.createDialogVisible = true;
      this.newMedia = { name: '', media_type_id: 0, target: '', params: {}, enabled: 1, status: 1, description: '' };
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newMedia = { name: '', media_type_id: 0, target: '', params: {}, enabled: 1, status: 1, description: '' };
    },
    async onCreate() {
      if (!this.newMedia.name) {
        ElMessage.warning(this.$t('media.validationName'));
        return;
      }
      if (!this.newMedia.media_type_id) {
        ElMessage.warning(this.$t('media.validationType'));
        return;
      }
      try {
        await addMedia({
          name: this.newMedia.name,
          media_type_id: this.newMedia.media_type_id,
          target: this.newMedia.target,
          params: this.newMedia.params || {},
          enabled: this.newMedia.enabled,
          status: this.newMedia.status,
          description: this.newMedia.description,
        });
        await this.loadMedia(true);
        this.createDialogVisible = false;
        this.newMedia = { name: '', media_type_id: 0, target: '', params: {}, enabled: 1, status: 1, description: '' };
        ElMessage.success(this.$t('media.created'));
      } catch (err) {
        ElMessage.error(this.$t('media.createFailed') + ': ' + (err.message || ''));
      }
    },
    openProperties(media) {
      this.selectedMedia = {
        ...media,
        params: media.params || {},
      };
      this.applyMediaTypeDefaults(this.selectedMedia, this.selectedMedia.media_type_id, true);
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        await updateMedia(this.selectedMedia.id, {
          name: this.selectedMedia.name,
          media_type_id: this.selectedMedia.media_type_id,
          target: this.selectedMedia.target,
          params: this.selectedMedia.params || {},
          enabled: this.selectedMedia.enabled,
          status: this.selectedMedia.status,
          description: this.selectedMedia.description,
        });
        await this.loadMedia(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('media.updated'));
      } catch (err) {
        ElMessage.error(this.$t('media.updateFailed') + ': ' + (err.message || ''));
      }
    },
    normalizeFields(fields) {
      return (fields || []).filter((field) => field && String(field.key || '').trim() !== '');
    },
    fieldLabel(field) {
      return field.label || field.key;
    },
    fieldPlaceholder(field) {
      if (field.default) return `${this.$t('media.paramsHint')}: ${field.default}`;
      return this.$t('media.paramsHint');
    },
    applyMediaTypeDefaults(model, mediaTypeId, preserveExisting) {
      const mediaType = this.mediaTypeOptions.find((t) => t.id === mediaTypeId);
      const fields = this.normalizeFields(mediaType?.fields || []);
      const nextParams = {};
      fields.forEach((field) => {
        const key = field.key;
        const existing = model.params ? model.params[key] : undefined;
        if (preserveExisting && existing !== undefined && existing !== '') {
          nextParams[key] = existing;
          return;
        }
        if (field.default !== undefined && field.default !== null && field.default !== '') {
          nextParams[key] = field.default;
          return;
        }
        nextParams[key] = '';
      });
      model.params = nextParams;
    },
    mediaTypeName(id) {
      const found = this.mediaTypeOptions.find((t) => t.id === id);
      return found ? found.name : id;
    },
    onDelete(media) {
      ElMessageBox.confirm(
        `${this.$t('media.delete')} ${media.name}?`,
        this.$t('media.delete'),
        {
          confirmButtonText: this.$t('media.delete'),
          cancelButtonText: this.$t('media.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteMedia(media.id);
          await this.loadMedia(true);
          ElMessage.success(this.$t('media.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('media.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('media.deleteCanceled'));
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
.media-content {
  margin-top: 8px;
}

.media-pagination {
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
