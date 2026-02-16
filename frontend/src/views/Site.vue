<template>
  <div class="sites-toolbar">
    <div class="sites-filters">
      <span class="filter-label">{{ $t('common.columns') }}</span>
      <el-select v-model="selectedColumns" multiple collapse-tags :placeholder="$t('common.search')" class="sites-filter" style="min-width: 220px;">
        <el-option v-for="col in columnOptions" :key="col.key" :label="col.label" :value="col.key" />
      </el-select>
      <span class="filter-label">{{ $t('common.search') }}</span>
      <el-select v-model="searchField" :placeholder="$t('common.search')" class="sites-filter">
        <el-option :label="$t('sites.filterAll')" value="all" />
        <el-option v-for="col in searchableColumns" :key="col.key" :label="col.label" :value="col.key" />
      </el-select>
      <span class="filter-label">{{ $t('common.search') }}</span>
      <el-input v-model="search" :placeholder="$t('sites.search')" clearable class="sites-search" />
      <span class="filter-label">{{ $t('sites.filterStatus') }}</span>
      <el-select v-model="statusFilter" :placeholder="$t('sites.filterStatus')" class="sites-filter">
        <el-option :label="$t('sites.filterAll')" value="all" />
        <el-option :label="$t('common.statusInactive')" :value="0" />
        <el-option :label="$t('common.statusActive')" :value="1" />
        <el-option :label="$t('common.statusError')" :value="2" />
        <el-option :label="$t('common.statusSyncing')" :value="3" />
      </el-select>
      <span class="filter-label">{{ $t('common.sort') }}</span>
      <el-select v-model="sortKey" class="sites-filter">
        <el-option :label="$t('common.sortUpdatedDesc')" value="updated_desc" />
        <el-option :label="$t('common.sortCreatedDesc')" value="created_desc" />
        <el-option :label="$t('common.sortNameAsc')" value="name_asc" />
        <el-option :label="$t('common.sortNameDesc')" value="name_desc" />
        <el-option :label="$t('common.sortStatusAsc')" value="status_asc" />
        <el-option :label="$t('common.sortStatusDesc')" value="status_desc" />
      </el-select>
      <div class="sites-bulk-actions">
        <span class="selected-count">{{ $t('common.selectedCount', { count: selectedCount }) }}</span>
        <el-button type="primary" plain :disabled="selectedCount === 0" @click="openBulkUpdateDialog">
          {{ $t('common.bulkUpdate') }}
        </el-button>
        <el-button type="danger" plain :disabled="selectedCount === 0" @click="openBulkDeleteDialog">
          {{ $t('common.bulkDelete') }}
        </el-button>
      </div>
    </div>
    <el-button type="primary" @click="createDialogVisible = true">
      {{ $t('sites.create') }}
    </el-button>
  </div>

  <div v-if="loading" class="loading-state">
    <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
    <p>{{ $t('sites.loading') }}</p>
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
      <el-button size="small" @click="loadSites">{{ $t('sites.retry') }}</el-button>
    </template>
  </el-alert>

  <el-empty
    v-if="!loading && !error && sites.length === 0"
    :description="$t('sites.noSites')"
    style="margin: 40px;"
  />

  <el-empty
    v-if="!loading && !error && sites.length > 0 && filteredSites.length === 0"
    :description="$t('sites.noResults')"
    style="margin: 40px;"
  />

  <div v-if="!loading && !error" class="sites-scroll">
  <el-table
    v-if="filteredSites.length > 0"
    :data="filteredSites"
    border
    style="margin: 20px; width: calc(100% - 40px);"
    ref="sitesTableRef"
    row-key="id"
    @selection-change="onSelectionChange"
  >
    <el-table-column type="selection" width="50" />
    <el-table-column v-if="isColumnVisible('name')" prop="name" :label="$t('sites.name')" min-width="160" />
    <el-table-column v-if="isColumnVisible('enabled')" :label="$t('common.enabled')" min-width="110">
      <template #default="{ row }">
        <el-tag :type="row.enabled === 1 ? 'success' : 'info'">
          {{ row.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('status')" :label="$t('sites.status')" min-width="160">
      <template #default="{ row }">
        <el-tooltip :content="row.status_reason || getStatusInfo(row.status).reason" placement="top">
          <el-tag :type="getStatusInfo(row.status).type">
            {{ getStatusInfo(row.status).label }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column v-if="isColumnVisible('description')" prop="description" :label="$t('sites.description')" min-width="200" show-overflow-tooltip />
    <el-table-column :label="$t('sites.actions')" min-width="240" fixed="right">
      <template #default="{ row }">
        <el-button size="small" type="primary" @click="openDetails(row)">{{ $t('sites.details') }}</el-button>
        <el-button size="small" @click="openProperties(row)">{{ $t('sites.properties') }}</el-button>
        <el-button size="small" type="danger" @click="onDelete(row)">{{ $t('sites.delete') }}</el-button>
      </template>
    </el-table-column>
  </el-table>
  </div>
  <div v-if="!loading && !error && totalSites > 0" class="sites-pagination">
    <el-pagination
      background
      layout="sizes, prev, pager, next"
      :page-sizes="[10, 20, 50, 100]"
      v-model:page-size="pageSize"
      v-model:current-page="currentPage"
      :total="totalSites"
    />
  </div>

  <el-dialog v-model="createDialogVisible" :title="$t('sites.createTitle')" width="500px" align-center>
    <el-form :model="newSite" label-width="120px">
      <el-form-item :label="$t('sites.name')">
        <el-input v-model="newSite.name" :placeholder="$t('sites.name')" />
      </el-form-item>
      <el-form-item :label="$t('sites.description')">
        <el-input v-model="newSite.description" type="textarea" :placeholder="$t('sites.description')" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="newSite.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('sites.status')">
        <el-select v-model="newSite.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelCreate">{{ $t('sites.cancel') }}</el-button>
      <el-button type="primary" @click="onCreate">{{ $t('sites.save') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="propertiesDialogVisible" :title="`${$t('sites.properties')} - ${selectedSite?.name || ''}`" width="600px">
    <el-form :model="selectedSite" label-width="120px">
      <el-form-item :label="$t('sites.name')">
        <el-input v-model="selectedSite.name" />
      </el-form-item>
      <el-form-item :label="$t('sites.description')">
        <el-input v-model="selectedSite.description" type="textarea" />
      </el-form-item>
      <el-form-item :label="$t('common.enabled')">
        <el-switch v-model="selectedSite.enabled" :active-value="1" :inactive-value="0" />
      </el-form-item>
      <el-form-item :label="$t('sites.status')">
        <el-select v-model="selectedSite.status" style="width: 100%;">
          <el-option :label="$t('common.statusInactive')" :value="0" />
          <el-option :label="$t('common.statusActive')" :value="1" />
          <el-option :label="$t('common.statusError')" :value="2" />
          <el-option :label="$t('common.statusSyncing')" :value="3" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="cancelProperties">{{ $t('sites.cancel') }}</el-button>
      <el-button type="primary" @click="saveProperties">{{ $t('sites.save') }}</el-button>
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
      <el-form-item :label="$t('sites.status')">
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
      <el-button @click="bulkDialogVisible = false">{{ $t('sites.cancel') }}</el-button>
      <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
    </template>
  </el-dialog>

  <!-- Bulk Delete Confirmation Dialog -->
  <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
    <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
    <template #footer>
      <el-button @click="bulkDeleteDialogVisible = false">{{ $t('sites.cancel') }}</el-button>
      <el-button type="danger" @click="deleteSelectedSites" :loading="bulkDeleting">{{ $t('sites.delete') }}</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="detailDialogVisible" :title="$t('sites.detailTitle')" width="800px">
    <div v-if="detailLoading" style="text-align: center; padding: 30px;">
      <el-icon class="is-loading" size="40" color="#409EFF"><Loading /></el-icon>
    </div>
    <div v-else-if="siteDetail">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="$t('sites.name')">{{ siteDetail.site.name }}</el-descriptions-item>
        <el-descriptions-item :label="$t('sites.status')">
          <el-tag :type="getStatusInfo(siteDetail.site.status).type">
            {{ getStatusInfo(siteDetail.site.status).label }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item :label="$t('common.enabled')">
          {{ siteDetail.site.enabled === 1 ? $t('common.enabled') : $t('common.disabled') }}
        </el-descriptions-item>
        <el-descriptions-item :label="$t('sites.description')">{{ siteDetail.site.description }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">{{ $t('sites.summary') }}</el-divider>
      <el-row :gutter="12">
        <el-col :span="8"><el-card>{{ $t('sites.totalHosts') }}: {{ siteDetail.summary.total_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card>{{ $t('sites.activeHosts') }}: {{ siteDetail.summary.active_hosts }}</el-card></el-col>
        <el-col :span="8"><el-card>{{ $t('sites.errorHosts') }}: {{ siteDetail.summary.error_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card>{{ $t('sites.syncingHosts') }}: {{ siteDetail.summary.syncing_hosts }}</el-card></el-col>
        <el-col :span="8" style="margin-top: 12px;"><el-card>{{ $t('sites.totalItems') }}: {{ siteDetail.summary.total_items }}</el-card></el-col>
      </el-row>

      <el-divider content-position="left">{{ $t('sites.hosts') }}</el-divider>
      <el-table :data="siteDetail.hosts" border>
        <el-table-column prop="name" :label="$t('hosts.name')" min-width="160" />
        <el-table-column prop="ip_addr" :label="$t('hosts.ip')" min-width="140" />
        <el-table-column :label="$t('hosts.status')" min-width="160">
          <template #default="{ row }">
            <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
              <el-tag :type="getStatusInfo(row.status).type">{{ getStatusInfo(row.status).label }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <div v-else class="empty-detail">
      {{ $t('sites.selectSite') }}
    </div>
  </el-dialog>
</template>

<script lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { fetchSiteData, fetchSiteDetail, addSite, updateSite, deleteSite } from '@/api/sites';

export default {
  name: 'Site',
  components: { Loading },
  data() {
    return {
      sites: [],
      pageSize: 20,
      currentPage: 1,
      totalSites: 0,
      sortKey: 'updated_desc',
      loading: false,
      error: null,
      search: '',
      searchField: 'all',
      selectedColumns: ['name', 'enabled', 'status', 'description'],
      columnOptions: [
        { key: 'name', label: this.$t('sites.name') },
        { key: 'enabled', label: this.$t('common.enabled') },
        { key: 'status', label: this.$t('sites.status') },
        { key: 'description', label: this.$t('sites.description') },
      ],
      statusFilter: 'all',
      createDialogVisible: false,
      propertiesDialogVisible: false,
      detailDialogVisible: false,
      detailLoading: false,
      siteDetail: null,
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedSiteRows: [],
      newSite: { name: '', description: '', enabled: 1, status: 1 },
      selectedSite: { id: 0, name: '', description: '', enabled: 1, status: 1 },
      bulkForm: {
        enabled: 'nochange',
        status: 'nochange',
      },
    };
  },
  computed: {
    filteredSites() {
      return this.sites;
    },
    searchableColumns() {
      return this.columnOptions;
    },
    selectedCount() {
      return this.selectedSiteRows.length;
    },
  },
  watch: {
    search() {
      this.currentPage = 1;
      this.loadSites(true);
    },
    statusFilter() {
      this.currentPage = 1;
      this.loadSites(true);
    },
    sortKey() {
      this.currentPage = 1;
      this.loadSites(true);
    },
    pageSize() {
      this.currentPage = 1;
      this.loadSites(true);
    },
    currentPage() {
      this.loadSites();
    },
  },
  created() {
    this.loadSites(true);
  },
  methods: {
    onSelectionChange(selection) {
      this.selectedSiteRows = selection || [];
    },
    openBulkDeleteDialog() {
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'));
        return;
      }
      this.bulkDeleteDialogVisible = true;
    },
    async deleteSelectedSites() {
      if (this.selectedCount === 0) return;
      this.bulkDeleting = true;
      try {
        await Promise.all(this.selectedSiteRows.map((site) => deleteSite(site.id)));
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }));
        this.bulkDeleteDialogVisible = false;
        this.clearSelection();
        await this.loadSites(true);
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
        await Promise.all(this.selectedSiteRows.map((site) => {
          const payload = {
            name: site.name,
            description: site.description,
            enabled: enabledOverride === 'nochange' ? site.enabled : (enabledOverride === 'enable' ? 1 : 0),
            status: statusOverride === 'nochange' ? site.status : statusOverride,
          };
          return updateSite(site.id, payload);
        }));
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }));
        this.bulkDialogVisible = false;
        this.clearSelection();
        await this.loadSites(true);
      } catch (err) {
        ElMessage.error(err.message || this.$t('common.bulkUpdateFailed'));
      } finally {
        this.bulkUpdating = false;
      }
    },
    clearSelection() {
      if (this.$refs.sitesTableRef && this.$refs.sitesTableRef.clearSelection) {
        this.$refs.sitesTableRef.clearSelection();
      }
      this.selectedSiteRows = [];
    },
    async loadSites(reset = false) {
      if (reset) {
        this.sites = [];
      }
      this.loading = reset;
      this.error = null;
      try {
        const { sortBy, sortOrder } = this.parseSortKey(this.sortKey);
        const response = await fetchSiteData({
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
          : (response.data?.items || response.items || response.data || response.sites || []);
        const total = response?.data?.total ?? response?.total ?? data.length;
        const mapped = data.map((s) => ({
          id: s.ID || s.id || 0,
          name: s.Name || s.name || '',
          description: s.Description || s.description || '',
          enabled: s.Enabled ?? s.enabled ?? 1,
          status: s.Status ?? s.status ?? 0,
          status_reason: s.Reason || s.reason || s.Error || s.error || s.ErrorMessage || s.error_message || s.LastError || s.last_error || '',
        }));
        this.sites = mapped;
        this.totalSites = Number.isFinite(total) ? total : mapped.length;
      } catch (err) {
        this.error = err.message || this.$t('sites.loadFailed');
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
    async openDetails(site) {
      this.$router.push({ path: `/site/${site.id}/detail` });
    },
    openProperties(site) {
      this.selectedSite = { ...site };
      this.propertiesDialogVisible = true;
    },
    cancelProperties() {
      this.propertiesDialogVisible = false;
    },
    async saveProperties() {
      try {
        await updateSite(this.selectedSite.id, {
          name: this.selectedSite.name,
          description: this.selectedSite.description,
          enabled: this.selectedSite.enabled,
          status: this.selectedSite.status,
        });
        await this.loadSites(true);
        this.propertiesDialogVisible = false;
        ElMessage.success(this.$t('sites.updated'));
      } catch (err) {
        ElMessage.error(this.$t('sites.updateFailed') + ': ' + (err.message || ''));
      }
    },
    cancelCreate() {
      this.createDialogVisible = false;
      this.newSite = { name: '', description: '', enabled: 1, status: 1 };
    },
    async onCreate() {
      if (!this.newSite.name) {
        ElMessage.warning(this.$t('sites.name'));
        return;
      }
      try {
        await addSite(this.newSite);
        await this.loadSites(true);
        this.createDialogVisible = false;
        this.newSite = { name: '', description: '', enabled: 1, status: 1 };
        ElMessage.success(this.$t('sites.created'));
      } catch (err) {
        ElMessage.error(this.$t('sites.createFailed') + ': ' + (err.message || ''));
      }
    },
    onDelete(site) {
      ElMessageBox.confirm(
        `${this.$t('sites.delete')} ${site.name}?`,
        this.$t('sites.delete'),
        {
          confirmButtonText: this.$t('sites.delete'),
          cancelButtonText: this.$t('sites.cancel'),
          type: 'warning',
        }
      ).then(async () => {
        try {
          await deleteSite(site.id);
          await this.loadSites();
          ElMessage.success(this.$t('sites.deleted'));
        } catch (err) {
          ElMessage.error(this.$t('sites.deleteFailed') + ': ' + (err.message || ''));
        }
      }).catch(() => {
        ElMessage.info(this.$t('sites.deleteCanceled'));
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
    isColumnVisible(key) {
      return this.selectedColumns.includes(key);
    },
  },
};
</script>

<style scoped>
.sites-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin: 16px 20px 0;
}

.sites-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.sites-bulk-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.sites-pagination {
  display: flex;
  justify-content: flex-end;
  padding: 0 20px 16px;
}

.selected-count {
  color: #606266;
  font-size: 13px;
}

.sites-search {
  width: 240px;
}

.sites-filter {
  min-width: 160px;
}

.loading-state {
  text-align: center;
  padding: 40px;
}

.empty-detail {
  text-align: center;
  padding: 20px;
  color: #909399;
}
</style>
