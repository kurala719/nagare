<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.groups') }}</span>
        <el-button link type="primary" @click="$router.push('/group')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="groups" style="width: 100%" v-loading="loading" size="small" :empty-text="$t('dashboard.noGroups', '暂无分组')">
      <el-table-column prop="name" :label="$t('dashboard.name')" show-overflow-tooltip sortable />
      <el-table-column prop="status" :label="$t('dashboard.status')" width="100" sortable>
        <template #default="{ row }">
          <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="Health" width="80">
        <template #default="{ row }">
          <el-progress 
            type="circle" 
            :percentage="row.health_score ?? row.HealthScore ?? 100" 
            :width="24" 
            :stroke-width="3" 
            :show-text="false"
            :status="getHealthStatus(row.health_score ?? row.HealthScore ?? 100)" 
          />
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script>
import { defineComponent } from 'vue'
import { useI18n } from 'vue-i18n'

export default defineComponent({
  name: 'RecentGroups',
  props: {
    groups: {
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
      switch (status) {
        case 1:
          return { label: t('common.statusActive'), type: 'success', reason: t('common.reasonActive') }
        case 0:
          return { label: t('common.statusInactive'), type: 'info', reason: t('common.reasonInactive') }
        case 2:
          return { label: t('common.statusError'), type: 'danger', reason: t('common.reasonError') }
        case 3:
          return { label: t('common.statusSyncing'), type: 'warning', reason: t('common.reasonSyncing') }
        default:
          return { label: t('common.unknown', 'Unknown'), type: 'info', reason: '' }
      }
    }

    const getHealthStatus = (score) => {
      if (score >= 90) return 'success'
      if (score >= 70) return 'warning'
      return 'exception'
    }

    return {
      getStatusInfo,
      getHealthStatus
    }
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
