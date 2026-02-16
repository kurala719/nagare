<template>
  <el-row :gutter="20" class="summary-row">
    <el-col :xs="24" :sm="12" :md="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon alerts-icon">
          <el-icon size="40"><Bell /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">{{ summary.alerts.total }}</div>
          <div class="summary-label">{{ $t('dashboard.alerts') }}</div>
          <div class="summary-detail">
            <el-tag type="danger" size="small">{{ summary.alerts.critical }} {{ $t('dashboard.critical') }}</el-tag>
          </div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon hosts-icon">
          <el-icon size="40"><Monitor /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">{{ summary.hosts.total }}</div>
          <div class="summary-label">{{ $t('dashboard.hosts') }}</div>
          <div class="summary-detail">
            <el-tag type="success" size="small">{{ summary.hosts.online }} {{ $t('dashboard.online') }}</el-tag>
          </div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon monitors-icon">
          <el-icon size="40"><DataLine /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">{{ summary.monitors.total }}</div>
          <div class="summary-label">{{ $t('dashboard.monitors') }}</div>
          <div class="summary-detail">
            <el-tag type="success" size="small">{{ summary.monitors.active }} {{ $t('dashboard.active') }}</el-tag>
          </div>
        </div>
      </el-card>
    </el-col>
    <el-col :xs="24" :sm="12" :md="6">
      <el-card class="summary-card" shadow="hover">
        <div class="summary-icon providers-icon">
          <el-icon size="40"><Connection /></el-icon>
        </div>
        <div class="summary-info">
          <div class="summary-value">{{ summary.providers.total }}</div>
          <div class="summary-label">{{ $t('dashboard.providers') }}</div>
          <div class="summary-detail">
            <el-tag type="success" size="small">{{ summary.providers.active }} {{ $t('dashboard.active') }}</el-tag>
          </div>
        </div>
      </el-card>
    </el-col>
  </el-row>
</template>

<script>
import { defineComponent } from 'vue'
import { Bell, Monitor, DataLine, Connection } from '@element-plus/icons-vue'

export default defineComponent({
  name: 'SummaryCards',
  components: { Bell, Monitor, DataLine, Connection },
  props: {
    summary: {
      type: Object,
      required: true
    }
  }
})
</script>

<style scoped>
.summary-row {
  margin-bottom: 20px;
}

.summary-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  padding: 20px;
}

.summary-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
  flex-shrink: 0;
}

.alerts-icon { background: linear-gradient(135deg, #f56c6c, #e74c3c); }
.hosts-icon { background: linear-gradient(135deg, #67c23a, #2ecc71); }
.monitors-icon { background: linear-gradient(135deg, #409eff, #3498db); }
.providers-icon { background: linear-gradient(135deg, #e6a23c, #f39c12); }

.summary-info {
  flex: 1;
  min-width: 0;
}

.summary-value {
  font-size: 28px;
  font-weight: bold;
  color: var(--el-text-color-primary);
  line-height: 1.2;
}

.summary-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}
</style>
