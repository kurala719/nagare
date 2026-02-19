<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('message.historyTitle') }}</h1>
      <p class="page-subtitle">{{ $t('message.historySubtitle') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-radio-group v-model="filterStatus" @change="loadMessages">
          <el-radio-button label="all">{{ $t('message.filterAll') }}</el-radio-button>
          <el-radio-button label="unread">{{ $t('message.filterUnread') }}</el-radio-button>
        </el-radio-group>
      </div>
      <div class="action-group">
        <el-button @click="handleMarkAllRead" :disabled="messages.every(m => m.is_read === 1)">
          {{ $t('message.markAllRead') }}
        </el-button>
        <el-button type="primary" :icon="Refresh" @click="loadMessages">
          {{ $t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <el-card v-loading="loading" class="history-card">
      <el-table :data="messages" style="width: 100%" @row-click="handleMessageClick">
        <el-table-column width="40">
          <template #default="{ row }">
            <div v-if="row.is_read === 0" class="unread-dot"></div>
          </template>
        </el-table-column>
        <el-table-column width="50">
          <template #default="{ row }">
            <el-icon :class="getSeverityClass(row.severity)">
              <component :is="getTypeIcon(row.type)" />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column prop="title" :label="$t('message.colTitle')" min-width="200">
          <template #default="{ row }">
            <span :class="{ 'font-bold': row.is_read === 0 }">{{ row.title }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="content" :label="$t('message.colContent')" min-width="400" show-overflow-tooltip />
        <el-table-column prop="type" :label="$t('message.colType')" width="120">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" :label="$t('message.colTime')" width="180" />
        <el-table-column width="100" align="center">
          <template #default="{ row }">
            <el-button link type="danger" :icon="Delete" @click.stop="handleDelete(row)" />
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          background
          layout="total, sizes, prev, pager, next"
          :total="total"
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          @current-change="loadMessages"
          @size-change="loadMessages"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Refresh, Delete, Warning, Operation, Management, InfoFilled } from '@element-plus/icons-vue'
import { fetchSiteMessages, markAsRead, markAllAsRead, deleteSiteMessage } from '@/api/siteMessage'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const messages = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const filterStatus = ref('all')

const loadMessages = async () => {
  loading.value = true
  try {
    const params = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      unread_only: filterStatus.value === 'unread' ? 1 : 0
    }
    const res = await fetchSiteMessages(params)
    if (res.data.success) {
      messages.value = res.data.data.items || []
      total.value = res.data.data.total || 0
    }
  } catch (e) {
    ElMessage.error(t('message.loadFailed'))
  } finally {
    loading.value = false
  }
}

const handleMessageClick = async (row) => {
  if (row.is_read === 0) {
    try {
      await markAsRead(row.id)
      row.is_read = 1
    } catch (e) {
      console.error(e)
    }
  }
  
  ElMessageBox.alert(row.content, row.title, {
    confirmButtonText: t('common.ok'),
    type: getSeverityType(row.severity)
  })
}

const handleMarkAllRead = async () => {
  try {
    await markAllAsRead()
    messages.value.forEach(m => m.is_read = 1)
    ElMessage.success(t('message.markAllReadSuccess'))
  } catch (e) {
    ElMessage.error(t('message.operationFailed'))
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(t('message.deleteConfirm'), t('common.warning'), {
    type: 'warning'
  }).then(async () => {
    try {
      await deleteSiteMessage(row.id)
      ElMessage.success(t('message.deleteSuccess'))
      loadMessages()
    } catch (e) {
      ElMessage.error(t('message.deleteFailed'))
    }
  }).catch(() => {})
}

const getSeverityClass = (severity) => {
  switch (severity) {
    case 3: return 'color-danger'
    case 2: return 'color-warning'
    case 1: return 'color-success'
    default: return 'color-info'
  }
}

const getSeverityType = (severity) => {
  switch (severity) {
    case 3: return 'error'
    case 2: return 'warning'
    case 1: return 'success'
    default: return 'info'
  }
}

const getTypeIcon = (type) => {
  switch (type) {
    case 'alert': return Warning
    case 'sync': return Operation
    case 'report': return Management
    default: return InfoFilled
  }
}

onMounted(() => {
  loadMessages()
})
</script>

<style scoped>
.history-card {
  margin-top: 20px;
}

.unread-dot {
  width: 8px;
  height: 8px;
  background-color: var(--el-color-primary);
  border-radius: 50%;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.font-bold {
  font-weight: bold;
}

.color-danger { color: var(--el-color-danger); }
.color-warning { color: var(--el-color-warning); }
.color-success { color: var(--el-color-success); }
.color-info { color: var(--el-color-info); }

.history-card :deep(.el-table__row) {
  cursor: pointer;
}
</style>
