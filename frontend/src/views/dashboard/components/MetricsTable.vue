<template>
  <el-card class="detail-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.networkMetrics') }}</span>
        <div class="card-actions">
          <el-input
            v-model="query"
            clearable
            size="small"
            class="metrics-search"
            :placeholder="$t('dashboard.metricsSearch')"
            @keyup.enter="loadMetrics"
          />
          <el-input
            v-model.number="limit"
            type="number"
            size="small"
            class="metrics-limit"
            :min="10"
            :max="500"
            @change="loadMetrics"
          />
          <el-button size="small" @click="loadMetrics" :loading="loading">
            {{ $t('dashboard.refresh') }}
          </el-button>
        </div>
      </div>
    </template>

    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading" size="40" color="#409EFF">
        <Loading />
      </el-icon>
      <p>{{ $t('dashboard.loadingMetrics') }}</p>
    </div>

    <el-alert
      v-else-if="error"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="metrics-alert"
    />

    <template v-else>
      <el-table :data="metrics" style="width: 100%" max-height="320">
        <el-table-column prop="host_name" :label="$t('dashboard.host')" min-width="140" show-overflow-tooltip />
        <el-table-column prop="item_name" :label="$t('dashboard.metric')" min-width="160" show-overflow-tooltip />
        <el-table-column :label="$t('dashboard.value')" min-width="140">
          <template #default="{ row }">
            <span>{{ row.value }} {{ row.units }}</span>
          </template>
        </el-table-column>
        <el-table-column :label="$t('dashboard.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusInfo(row.status).type" size="small">
              {{ getStatusInfo(row.status).label }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="$t('dashboard.updated')" min-width="160">
          <template #default="{ row }">
            {{ formatTimestamp(row.updated_at) }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="metrics.length === 0" :description="$t('dashboard.noMetrics')" />
    </template>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Loading } from '@element-plus/icons-vue'
import { authFetch } from '@/utils/authFetch'

export default defineComponent({
  name: 'MetricsTable',
  components: { Loading },
  setup() {
    const { t } = useI18n()
    const loading = ref(false)
    const error = ref(null)
    const metrics = ref([])
    const query = ref('')
    const limit = ref(200)

    const getStatusInfo = (status) => {
      const map = {
        0: { label: t('common.statusInactive'), type: 'info' },
        1: { label: t('common.statusActive'), type: 'success' },
        2: { label: t('common.statusError'), type: 'danger' },
        3: { label: t('common.statusSyncing'), type: 'warning' },
      }
      return map[status] || map[0]
    }

    const formatTimestamp = (value) => {
      if (!value) return '--'
      const date = new Date(value)
      if (Number.isNaN(date.getTime())) return value
      return date.toLocaleString()
    }

    const loadMetrics = async () => {
      loading.value = true
      error.value = null
      try {
        const params = new URLSearchParams()
        if (query.value.trim()) {
          params.set('q', query.value.trim())
        }
        if (limit.value) {
          params.set('limit', String(limit.value))
        }
        const url = `/api/v1/system/metrics${params.toString() ? `?${params}` : ''}`
        const response = await authFetch(url, {
          method: 'GET',
          headers: { 'Accept': 'application/json' },
        })
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        const data = await response.json().catch(() => ({}))
        if (data?.success === false) {
          throw new Error(data.error || t('dashboard.metricsLoadFailed'))
        }
        const payload = data?.data || data || []
        metrics.value = Array.isArray(payload) ? payload : []
      } catch (err) {
        error.value = err.message || t('dashboard.metricsLoadFailed')
        console.error('Error loading metrics:', err)
      } finally {
        loading.value = false
      }
    }

    onMounted(() => {
      loadMetrics()
    })

    return {
      loading,
      error,
      metrics,
      query,
      limit,
      loadMetrics,
      getStatusInfo,
      formatTimestamp
    }
  }
})
</script>

<style scoped>
.loading-container {
  text-align: center;
  padding: 60px;
  color: #909399;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.metrics-search {
  width: 180px;
}
.metrics-limit {
  width: 90px;
}
.metrics-alert {
  margin-bottom: 12px;
}
</style>
