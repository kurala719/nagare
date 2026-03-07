<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.recentAlerts') }}</span>
        <el-button type="primary" text @click="$router.push('/alert')">{{ $t('dashboard.viewAll') }}</el-button>
      </div>
    </template>
    <el-table :data="alerts" style="width: 100%" max-height="250" v-loading="loading" :empty-text="$t('dashboard.noAlerts')">
      <el-table-column prop="id" :label="$t('dashboard.id')" width="60" sortable />
      <el-table-column prop="message" :label="$t('dashboard.message')" show-overflow-tooltip sortable />
      <el-table-column prop="severity" :label="$t('dashboard.severity')" width="100" sortable>
        <template #default="{ row }">
          <el-tag :type="getSeverityType(row.severity)" size="small" style="display: flex; align-items: center; gap: 4px;">
            <el-icon><component :is="getSeverityIcon(row.severity)" /></el-icon>
            {{ getSeverityLabel(row.severity) }}
          </el-tag>
        </template>
      </el-table-column>

    </el-table>
  </el-card>
</template>

<script>
import { defineComponent } from 'vue'
import { WarningFilled, Warning, InfoFilled, QuestionFilled } from '@element-plus/icons-vue'

export default defineComponent({
  components: {
    WarningFilled,
    Warning,
    InfoFilled,
    QuestionFilled
  },
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
    const getSeverityLabel = (severity) => {
      if (typeof severity === 'number') {
        const map = { 0: 'info', 1: 'low', 2: 'medium', 3: 'high', 4: 'critical' }
        return map[severity] || String(severity)
      }
      return String(severity || '')
    }

    const getSeverityType = (severity) => {
      let s = ''
      if (typeof severity === 'number') {
        const map = { 0: 'info', 1: 'info', 2: 'warning', 3: 'danger', 4: 'danger' }
        return map[severity] || 'info'
      } else {
        s = String(severity || '').toLowerCase()
      }
      if (s === 'critical' || s === 'high') return 'danger'
      if (s === 'medium' || s === 'warning') return 'warning'
      if (s === 'low' || s === 'info') return 'info'
      return 'info'
    }

    const getSeverityIcon = (severity) => {
      let s = ''
      if (typeof severity === 'number') {
        const map = { 0: 'InfoFilled', 1: 'InfoFilled', 2: 'Warning', 3: 'WarningFilled', 4: 'WarningFilled' }
        return map[severity] || 'QuestionFilled'
      } else {
        s = String(severity || '').toLowerCase()
      }
      if (s === 'critical' || s === 'high') return 'WarningFilled'
      if (s === 'medium' || s === 'warning') return 'Warning'
      if (s === 'low' || s === 'info') return 'InfoFilled'
      return 'InfoFilled'
    }

    return { getSeverityLabel, getSeverityType, getSeverityIcon }
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
