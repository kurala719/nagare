<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.hostStatus') }}</span>
        <el-button type="primary" text @click="$router.push('/host')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="hosts" style="width: 100%" max-height="250" v-loading="loading" :empty-text="$t('dashboard.noHosts')">
      <el-table-column prop="id" :label="$t('dashboard.id')" width="60" sortable />
      <el-table-column prop="name" :label="$t('dashboard.name')" show-overflow-tooltip sortable />
      <el-table-column :label="$t('dashboard.ip')" width="140" sortable>
        <template #default="{ row }">
          {{ row.ip_addr || row.ip || row.IPAddr || '-' }}
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
  name: 'RecentHosts',
  props: {
    hosts: {
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

      }
      return map[status] || map[0]
    }

    const getHealthStatus = (score) => {
      if (score >= 90) return 'success'
      if (score >= 70) return 'warning'
      return 'exception'
    }

    return { getStatusInfo, getHealthStatus }
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
