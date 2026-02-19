<template>
  <div class="audit-log-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('auditLog.title') || 'Security Audit Log' }}</span>
          <div class="header-actions">
            <el-input
              v-model="searchQuery"
              :placeholder="$t('common.search') || 'Search...'"
              clearable
              style="width: 250px"
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button type="primary" @click="fetchAuditLogs">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
        </div>
      </template>

      <div class="timeline-container">
        <el-empty v-if="auditLogs.length === 0" description="No audit logs found" />
        <el-timeline v-else>
          <el-timeline-item
            v-for="log in auditLogs"
            :key="log.id"
            :timestamp="log.created_at"
            :type="getLogType(log.method)"
            placement="top"
          >
            <el-card>
              <div class="log-item">
                <div class="log-header">
                  <el-tag :type="getLogType(log.method)" size="small" effect="dark">
                    {{ log.method }}
                  </el-tag>
                  <span class="log-action">{{ log.action }}</span>
                  <span class="log-user">
                    <el-icon><User /></el-icon>
                    {{ log.username || 'System' }}
                  </span>
                </div>
                <div class="log-details">
                  <p><strong>Path:</strong> {{ log.path }}</p>
                  <p><strong>IP:</strong> {{ log.ip }}</p>
                  <p><strong>Status:</strong> 
                    <el-tag :type="log.status >= 400 ? 'danger' : 'success'" size="small">
                      {{ log.status }}
                    </el-tag>
                  </p>
                  <p><strong>Latency:</strong> {{ (log.latency / 1000).toFixed(2) }} ms</p>
                </div>
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <div class="pagination-container">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next"
            :total="total"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/request'
import { Search, Refresh, User } from '@element-plus/icons-vue'

const auditLogs = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')
const loading = ref(false)

const fetchAuditLogs = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/audit-logs', {
      params: {
        limit: pageSize.value,
        offset: (currentPage.value - 1) * pageSize.value,
        query: searchQuery.value
      }
    })
    if (res.success) {
      auditLogs.value = res.data.items
      total.value = res.data.total
    }
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
  } finally {
    loading.value = false
  }
}

const getLogType = (method) => {
  switch (method) {
    case 'POST': return 'success'
    case 'PUT':
    case 'PATCH': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}

const handleSearch = () => {
  currentPage.value = 1
  fetchAuditLogs()
}

const handleSizeChange = (val) => {
  pageSize.value = val
  fetchAuditLogs()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  fetchAuditLogs()
}

onMounted(() => {
  fetchAuditLogs()
})
</script>

<style scoped>
.audit-log-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.timeline-container {
  padding: 20px 0;
}

.log-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.log-action {
  font-weight: bold;
  font-size: 16px;
}

.log-user {
  margin-left: auto;
  color: #606266;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.log-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 4px;
  font-size: 13px;
  color: #666;
}

.log-details p {
  margin: 0;
}

.pagination-container {
  margin-top: 30px;
  display: flex;
  justify-content: center;
}
</style>
