<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.aiProviders') }}</span>
        <el-button type="primary" text @click="$router.push('/provider')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="providers" style="width: 100%" max-height="200" v-loading="loading">
      <el-table-column prop="id" :label="$t('dashboard.id')" width="60" sortable />
      <el-table-column prop="name" :label="$t('dashboard.name')" sortable />
      <el-table-column :label="$t('dashboard.model')" sortable>
        <template #default="{ row }">
          {{ row.default_model || row.model || row.DefaultModel || row.Model || '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('dashboard.status')" width="100" sortable>
        <template #default="{ row }">
          <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && providers.length === 0" :description="$t('dashboard.noProviders')" />
  </el-card>
</template>

<script>
import { defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'

export default defineComponent({
  name: 'RecentProviders',
  props: {
    providers: {
      type: Array,
      default: () => []
    },
    loading: {
      type: Boolean,
      default: false
    }
  },
  setup() {
    const { t } = useI18n()
    const getStatusInfo = (status) => {
      const map = {
        0: { label: t('common.statusInactive'), reason: t('common.reasonInactive'), type: 'info' },
        1: { label: t('common.statusActive'), reason: t('common.reasonActive'), type: 'success' },
        2: { label: t('common.statusError'), reason: t('common.reasonError'), type: 'danger' },
        3: { label: t('common.statusSyncing'), reason: t('common.reasonSyncing'), type: 'warning' },
      }
      return map[status] || map[0]
    }
    return { getStatusInfo }
  }
})
</script>

<style scoped>
.detail-card {
  height: 100%;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
