<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.recentAlerts') }}</span>
        <el-button type="primary" text @click="$router.push('/alert')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="alerts" style="width: 100%" max-height="250" v-loading="loading">
      <el-table-column prop="id" :label="$t('dashboard.id')" width="60" />
      <el-table-column prop="message" :label="$t('dashboard.message')" show-overflow-tooltip />
      <el-table-column prop="severity" :label="$t('dashboard.severity')" width="100">
        <template #default="{ row }">
          <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-if="!loading && alerts.length === 0" :description="$t('dashboard.noAlerts')" />
  </el-card>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'RecentAlerts',
  props: {
    alerts: {
      type: Array,
      default: () => []
    },
    loading: {
      type: Boolean,
      default: false
    }
  },
  setup() {
    const getSeverityType = (severity) => {
      const s = String(severity || '').toLowerCase()
      if (s === 'critical' || s === 'high') return 'danger'
      if (s === 'medium' || s === 'warning') return 'warning'
      if (s === 'low' || s === 'info') return 'info'
      return 'info'
    }
    return { getSeverityType }
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
