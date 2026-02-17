<template>
  <el-row :gutter="20" class="health-stats-row">
    <el-col :xs="24" :sm="12" :md="8" :lg="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon score-icon">
          <el-icon size="32"><DataLine /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">{{ loading ? '--' : (score.score ?? 0) }}</div>
          <div class="summary-label">{{ $t('dashboard.healthScore') }}</div>
          <el-tag :type="status.tagType" size="small">{{ status.label }}</el-tag>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6" v-if="alerts">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon alerts-icon">
          <el-icon size="32"><Bell /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : alerts.total }}
          </div>
          <div class="summary-label">{{ $t('dashboard.alerts') }}</div>
          <el-tag v-if="!loading && alerts.critical > 0" type="danger" size="small">
            {{ $t('dashboard.critical') }}: {{ alerts.critical }}
          </el-tag>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6" v-if="providers">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon providers-icon">
          <el-icon size="32"><Connection /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : `${providers.active} / ${providers.total}` }}
          </div>
          <div class="summary-label">{{ $t('dashboard.providers') }}</div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon monitors-icon">
          <el-icon size="32"><Monitor /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : (monitors ? `${monitors.active} / ${monitors.total}` : `${score.monitor_active ?? 0} / ${score.monitor_total ?? 0}`) }}
          </div>
          <div class="summary-label">{{ $t('dashboard.activeMonitors') }}</div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon groups-icon">
          <el-icon size="32"><Folder /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : `${score.group_active ?? 0} / ${score.group_total ?? 0}` }}
          </div>
          <div class="summary-label">{{ $t('dashboard.activeGroups') }}</div>
          <el-tag
            v-if="!loading"
            :type="(score.group_impacted ?? 0) > 0 ? 'danger' : 'success'"
            size="small"
          >
            {{ $t('dashboard.impactedGroups') }}: {{ score.group_impacted ?? 0 }}
          </el-tag>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon hosts-icon">
          <el-icon size="32"><Platform /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : `${score.host_active ?? 0} / ${score.host_total ?? 0}` }}
          </div>
          <div class="summary-label">{{ $t('dashboard.activeHosts') }}</div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="8" :lg="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon item-icon">
          <el-icon size="32"><Collection /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">
            {{ loading ? '--' : `${score.item_active ?? 0} / ${score.item_total ?? 0}` }}
          </div>
          <div class="summary-label">{{ $t('dashboard.activeItems') }}</div>
        </div>
      </el-card>
    </el-col>
  </el-row>
</template>

<script>
import { defineComponent, computed } from 'vue'
import { DataLine, Monitor, Collection, Bell, Connection, Folder, Platform } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

export default defineComponent({
  name: 'HealthStats',
  components: { DataLine, Monitor, Collection, Bell, Connection, Folder, Platform },
  props: {
    score: {
      type: Object,
      required: true
    },
    loading: {
      type: Boolean,
      default: false
    },
    alerts: {
      type: Object,
      default: null
    },
    providers: {
      type: Object,
      default: null
    },
    monitors: {
      type: Object,
      default: null
    }
  },
  setup(props) {
    const { t } = useI18n()

    const status = computed(() => {
      const s = Number(props.score?.score ?? 0)
      if (s >= 85) return { label: t('dashboard.healthGood'), tagType: 'success' }
      if (s >= 60) return { label: t('dashboard.healthWarn'), tagType: 'warning' }
      return { label: t('dashboard.healthBad'), tagType: 'danger' }
    })

    return { status }
  }
})
</script>

<style scoped>
.health-stats-row {
  margin-bottom: 20px;
}

.summary-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  padding: 20px;
}

.summary-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
  flex-shrink: 0;
}

.score-icon { background: linear-gradient(135deg, #409eff, #36d1dc); }
.monitors-icon { background: linear-gradient(135deg, #409eff, #3498db); }
.alerts-icon { background: linear-gradient(135deg, #f56c6c, #fd79a8); }
.providers-icon { background: linear-gradient(135deg, #e6a23c, #f1c40f); }
.hosts-icon { background: linear-gradient(135deg, #67c23a, #2ecc71); }
.groups-icon { background: linear-gradient(135deg, #8e44ad, #9b59b6); }
.item-icon { background: linear-gradient(135deg, #f56c6c, #e74c3c); }

.summary-info {
  flex: 1;
  min-width: 0;
}

.summary-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  line-height: 1.2;
}

.summary-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}
</style>
