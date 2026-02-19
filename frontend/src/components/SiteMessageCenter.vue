<template>
  <div class="message-center">
    <el-popover
      placement="bottom-end"
      :width="400"
      trigger="click"
      @show="onPopoverShow"
    >
      <template #reference>
        <div class="notification-badge">
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="item">
            <el-icon :size="20"><Bell /></el-icon>
          </el-badge>
        </div>
      </template>

      <div class="message-list-container">
        <div class="message-header">
          <h3>{{ $t('message.title') }}</h3>
          <el-button v-if="unreadCount > 0" link type="primary" @click="handleMarkAllRead">
            {{ $t('message.markAllRead') }}
          </el-button>
        </div>

        <el-scrollbar max-height="400px">
          <div v-if="messages.length === 0" class="empty-messages">
            <el-empty :description="$t('message.empty')" :image-size="60" />
          </div>
          <div v-else class="message-items">
            <div 
              v-for="msg in messages" 
              :key="msg.id" 
              class="message-item"
              :class="{ 'is-unread': msg.is_read === 0 }"
              @click="handleMessageClick(msg)"
            >
              <div class="message-icon">
                <el-icon :class="getSeverityClass(msg.severity)">
                  <component :is="getTypeIcon(msg.type)" />
                </el-icon>
              </div>
              <div class="message-content">
                <div class="message-item-title">{{ msg.title }}</div>
                <div class="message-text">{{ msg.content }}</div>
                <div class="message-time">{{ msg.created_at }}</div>
              </div>
            </div>
          </div>
        </el-scrollbar>

        <div class="message-footer">
          <el-button link @click="goToHistory">{{ $t('message.viewAll') }}</el-button>
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { Bell, InfoFilled, Warning, CircleCloseFilled, SuccessFilled, Management, Operation } from '@element-plus/icons-vue'
import { fetchSiteMessages, getUnreadCount, markAsRead, markAllAsRead } from '@/api/siteMessage'
import { ElNotification, ElMessage } from 'element-plus'
import { getToken } from '@/utils/auth'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const unreadCount = ref(0)
const messages = ref([])
const loading = ref(false)
let ws = null

const isAuthenticated = ref(!!getToken())

watch(() => isAuthenticated.value, (val) => {
  if (val) {
    connectWebSocket()
  } else {
    if (ws) ws.close()
  }
})

const loadUnreadCount = async () => {
  try {
    const res = await getUnreadCount()
    if (res.data.success) {
      unreadCount.value = res.data.data.count
    }
  } catch (e) {
    console.error(e)
  }
}

const loadRecentMessages = async () => {
  loading.value = true
  try {
    const res = await fetchSiteMessages({ limit: 5 })
    if (res.data.success) {
      messages.value = res.data.data.items || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const onPopoverShow = () => {
  loadRecentMessages()
}

const handleMarkAllRead = async () => {
  try {
    const res = await markAllAsRead()
    if (res.data.success) {
      unreadCount.value = 0
      messages.value.forEach(m => m.is_read = 1)
      ElMessage.success(t('message.markAllReadSuccess'))
    }
  } catch (e) {
    ElMessage.error(t('message.operationFailed'))
  }
}

const handleMessageClick = async (msg) => {
  if (msg.is_read === 0) {
    try {
      await markAsRead(msg.id)
      msg.is_read = 1
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    } catch (e) {
      console.error(e)
    }
  }
}

const goToHistory = () => {
  router.push('/site-messages')
}

const getSeverityClass = (severity) => {
  switch (severity) {
    case 3: return 'color-danger'
    case 2: return 'color-warning'
    case 1: return 'color-success'
    default: return 'color-info'
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

const connectWebSocket = () => {
  const token = getToken()
  if (!token) return

  // Try direct connection to backend if proxy fails, or consistent address
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  // Use 127.0.0.1:8080 for direct backend access if needed, but proxy is preferred
  const backendHost = window.location.hostname === 'localhost' ? '127.0.0.1:8080' : window.location.host
  const url = `${protocol}//${backendHost}/api/v1/site-messages/ws?token=${token}`

  console.log('Connecting to notification WebSocket:', url)
  ws = new WebSocket(url)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.event === 'site_message') {
        const msg = data.data
        unreadCount.value++
        
        // Add to list if already showing
        if (Array.isArray(messages.value)) {
          messages.value.unshift(msg)
          if (messages.value.length > 5) {
            messages.value.pop()
          }
        }

        // Show toast
        ElNotification({
          title: msg.title,
          message: msg.content,
          type: getSeverityType(msg.severity),
          position: 'bottom-right'
        })
      }
    } catch (e) {
      console.error('WS parse error', e)
    }
  }

  ws.onclose = () => {
    // Reconnect after 5 seconds
    setTimeout(connectWebSocket, 5000)
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

const handleAuthChanged = () => {
  isAuthenticated.value = !!getToken()
}

onMounted(() => {
  loadUnreadCount()
  connectWebSocket()
  window.addEventListener('auth-changed', handleAuthChanged)
})

onUnmounted(() => {
  if (ws) ws.close()
  window.removeEventListener('auth-changed', handleAuthChanged)
})
</script>

<style scoped>
.notification-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  transition: background-color 0.3s;
  color: var(--text-strong);
}

.notification-badge:hover {
  background-color: var(--surface-2);
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-1);
  margin-bottom: 10px;
}

.message-header h3 {
  margin: 0;
  font-size: 16px;
}

.message-items {
  display: flex;
  flex-direction: column;
}

.message-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.message-item:hover {
  background-color: var(--surface-2);
}

.message-item.is-unread {
  background-color: var(--brand-50);
}

.message-icon {
  display: flex;
  align-items: flex-start;
  padding-top: 2px;
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-item-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
}

.message-text {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.message-time {
  font-size: 12px;
  color: var(--text-muted);
}

.message-footer {
  display: flex;
  justify-content: center;
  padding-top: 12px;
  border-top: 1px solid var(--border-1);
  margin-top: 10px;
}

.empty-messages {
  padding: 40px 0;
}

.color-danger { color: var(--el-color-danger); }
.color-warning { color: var(--el-color-warning); }
.color-success { color: var(--el-color-success); }
.color-info { color: var(--el-color-info); }
</style>
