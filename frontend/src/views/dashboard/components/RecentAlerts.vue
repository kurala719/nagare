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
import { defineComponent, getCurrentInstance } from 'vue'
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
    const instance = getCurrentInstance()
    const t = (key) => instance?.proxy?.$t?.(key) || key

    const normalizeSeverity = (severity) => {
      if (typeof severity === 'number') return severity
      const parsed = Number.parseInt(String(severity ?? '').trim(), 10)
      if (Number.isFinite(parsed)) return parsed
      const label = String(severity || '').toLowerCase()
      const map = {
        disaster: 5,
        critical: 5,
        high: 4,
        average: 3,
        medium: 3,
        warning: 2,
        info: 1,
        low: 1,
        none: 0,
        unknown: 0
      }
      return map[label] ?? 0
    }

    const getSeverityLabel = (severity) => {
      const level = normalizeSeverity(severity)
      const map = {
        0: t('alerts.severityNotClassified'),
        1: t('alerts.severityInfo'),
        2: t('alerts.severityWarning'),
        3: t('alerts.severityAverage'),
        4: t('alerts.severityHigh'),
        5: t('alerts.severityDisaster')
      }
      return map[level] || t('alerts.severityNotClassified')
    }

    const getSeverityType = (severity) => {
      const level = normalizeSeverity(severity)
      const map = { 0: 'info', 1: 'info', 2: 'warning', 3: 'warning', 4: 'danger', 5: 'danger' }
      return map[level] || 'info'
    }

    const getSeverityIcon = (severity) => {
      const level = normalizeSeverity(severity)
      const map = {
        0: 'InfoFilled',
        1: 'InfoFilled',
        2: 'Warning',
        3: 'Warning',
        4: 'WarningFilled',
        5: 'WarningFilled'
      }
      return map[level] || 'QuestionFilled'
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
