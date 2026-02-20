<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.monitors') }}</span>
        <el-button link type="primary" @click="$router.push('/monitor')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="monitors" style="width: 100%" v-loading="loading" size="small">
      <el-table-column prop="name" :label="$t('dashboard.name')" show-overflow-tooltip sortable />
      <el-table-column prop="type" :label="$t('monitors.type')" width="100" sortable>
        <template #default="{ row }">
          <el-tag size="small" effect="plain">{{ row.type }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('dashboard.status')" width="80" sortable>
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)" size="small">
            {{ getStatusLabel(row.status) }}
          </el-tag>
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
  name: 'RecentMonitors',
  props: {
    monitors: {
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
    
    const getStatusType = (status) => {
      return status === 1 ? 'success' : 'info'
    }
    
    const getStatusLabel = (status) => {
      return status === 1 ? t('dashboard.activeLabel') : t('dashboard.inactiveLabel')
    }

    const getHealthStatus = (score) => {
      if (score >= 90) return 'success'
      if (score >= 70) return 'warning'
      return 'exception'
    }
    
    return {
      getStatusType,
      getStatusLabel,
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