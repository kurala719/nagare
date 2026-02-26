<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div class="host-info">
        <el-button icon="ArrowLeft" circle @click="$router.back()" />
        <span class="host-name">{{ $t('terminal.title') }}: {{ hostName || $t('terminal.selectHost') }}</span>
        <span class="host-ip" v-if="hostIp">({{ hostIp }})</span>
        <el-button v-if="!route.params.id" size="small" style="margin-left: 10px" @click="showHostSelector = true">
          {{ $t('terminal.switchHost') }}
        </el-button>
      </div>
      <div class="terminal-actions">
        <el-button type="danger" size="small" @click="handleDisconnect" :disabled="!connected">{{ $t('terminal.disconnect') }}</el-button>
        <el-button type="primary" size="small" @click="handleReconnect" :disabled="connected || !currentHostId">{{ $t('terminal.reconnect') }}</el-button>
      </div>
    </div>
    <div ref="terminalElement" class="terminal-body"></div>

    <!-- Host Selection Dialog -->
    <el-dialog v-model="showHostSelector" :title="$t('terminal.connectTitle')" width="500px" :close-on-click-modal="false" :show-close="!!currentHostId">
      <el-tabs v-model="connectionMode">
        <el-tab-pane :label="$t('terminal.savedHost')" name="saved">
          <div style="padding: 20px 0">
            <el-select v-model="selectedHostId" :placeholder="$t('terminal.selectHostPlaceholder')" style="width: 100%" filterable>
              <el-option v-for="h in availableHosts" :key="h.id" :label="`${h.name} (${h.ip_addr})`" :value="h.id" />
            </el-select>
          </div>
        </el-tab-pane>
        <el-tab-pane :label="$t('terminal.directConnect')" name="direct">
          <div style="padding: 20px 0">
            <el-form :model="directConfig" label-width="100px">
              <el-form-item :label="$t('hosts.ip')">
                <el-input v-model="directConfig.ip" placeholder="192.168.1.1" />
              </el-form-item>
              <el-form-item :label="$t('system.port')">
                <el-input-number v-model="directConfig.port" :min="1" :max="65535" style="width: 100%" />
              </el-form-item>
              <el-form-item :label="$t('auth.username')">
                <el-input v-model="directConfig.user" placeholder="root" />
              </el-form-item>
              <el-form-item :label="$t('auth.password')">
                <el-input v-model="directConfig.password" type="password" show-password />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="showHostSelector = false" v-if="!!currentHostId">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmHostSelection" :disabled="!canConnect">{{ $t('terminal.connect') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getToken } from '@/utils/auth'
import request from '@/utils/request'
import { fetchHostData } from '@/api/hosts'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const currentHostId = ref(route.params.id)
const hostName = ref('')
const hostIp = ref('')
const connected = ref(false)
const showHostSelector = ref(false)
const selectedHostId = ref(null)
const availableHosts = ref([])
const connectionMode = ref('saved')
const directConfig = ref({
  ip: '',
  port: 22,
  user: '',
  password: ''
})

const canConnect = computed(() => {
  if (connectionMode.value === 'saved') {
    return !!selectedHostId.value
  } else {
    return !!directConfig.value.ip && !!directConfig.value.user
  }
})

const terminalElement = ref(null)
let term = null
let fitAddon = null
let socket = null

const hostSSHUser = ref('')

const fetchHostInfo = async (id) => {
  try {
    const response = await request.get(`/hosts/${id}`)
    if (response && response.success) {
      hostName.value = response.data.name
      hostIp.value = response.data.ip_addr
      hostSSHUser.value = response.data.ssh_user
      
      if (!hostSSHUser.value) {
        ElMessage.warning(t('terminal.noCredentials'))
      }
    }
  } catch (err) {
    console.error('Failed to fetch host info', err)
    ElMessage.error(t('terminal.fetchInfoFailed'))
  }
}

const loadAvailableHosts = async () => {
  try {
    const res = await fetchHostData()
    if (res.success) {
      availableHosts.value = (res.data || []).filter(h => h.ssh_user) // Only show hosts with SSH user if from maintenance menu
      if (availableHosts.value.length === 0 && res.data?.length > 0) {
        availableHosts.value = res.data // Fallback to all hosts if none have SSH user configured
      }
    }
  } catch (err) {
    console.error('Failed to load hosts', err)
  }
}

const confirmHostSelection = async () => {
  if (!canConnect.value) return
  
  if (connectionMode.value === 'saved') {
    currentHostId.value = selectedHostId.value
    await fetchHostInfo(currentHostId.value)
  } else {
    currentHostId.value = 'direct'
    hostName.value = directConfig.value.ip
    hostIp.value = directConfig.value.ip
    hostSSHUser.value = directConfig.value.user
  }
  
  showHostSelector.value = false
  
  if (term) {
    term.clear()
  }
  
  handleDisconnect()
  connectWebSocket()
}

const initTerminal = () => {
  term = new Terminal({
    cursorBlink: true,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    fontSize: 14,
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#aeafad',
      selection: '#264f78'
    }
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalElement.value)
  
  // Initial fit
  nextTick(() => {
    fitAddon.fit()
  })

  term.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: 'data', data }))
    }
  })

  window.addEventListener('resize', handleResize)
}

const handleResize = () => {
  if (fitAddon && term) {
    fitAddon.fit()
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({
        type: 'resize',
        cols: term.cols,
        rows: term.rows
      }))
    }
  }
}

const connectWebSocket = () => {
  if (!currentHostId.value) return
  const token = getToken()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  
  let url = ''
  if (currentHostId.value === 'direct') {
    const { ip, port, user, password } = directConfig.value
    url = `${protocol}//${host}/api/v1/terminal/ssh?token=${token}&ip=${ip}&port=${port}&user=${user}&password=${encodeURIComponent(password)}&cols=${term.cols}&rows=${term.rows}`
  } else {
    url = `${protocol}//${host}/api/v1/hosts/${currentHostId.value}/ssh?token=${token}&cols=${term.cols}&rows=${term.rows}`
  }

  socket = new WebSocket(url)
  socket.binaryType = 'arraybuffer'

  socket.onopen = () => {
    connected.value = true
    term.write(`\r\n*** ${t('terminal.connected')} ***\r\n`)
    handleResize()
  }

  socket.onmessage = (event) => {
    if (typeof event.data === 'string') {
      term.write(event.data)
    } else {
      const decoder = new TextDecoder('utf-8')
      term.write(decoder.decode(event.data))
    }
  }

  socket.onclose = (event) => {
    connected.value = false
    const reason = event.reason ? `: ${event.reason}` : ''
    term.write(`\r\n*** ${t('terminal.closed')}${reason} ***\r\n`)
  }

  socket.onerror = (error) => {
    console.error('WebSocket Error:', error)
    term.write(`\r\n*** ${t('terminal.wsError')} ***\r\n`)
    ElMessage.error(t('terminal.connFailed'))
  }
}

const handleDisconnect = () => {
  if (socket) {
    socket.close()
  }
}

const handleReconnect = () => {
  if (term) {
    term.clear()
  }
  connectWebSocket()
}

onMounted(async () => {
  initTerminal()
  if (currentHostId.value) {
    await fetchHostInfo(currentHostId.value)
    connectWebSocket()
  } else {
    await loadAvailableHosts()
    showHostSelector.value = true
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (socket) {
    socket.close()
  }
  if (term) {
    term.dispose()
  }
})
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background-color: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  margin: 10px;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 15px;
  background-color: #333;
  color: #fff;
  border-bottom: 1px solid #444;
}

.host-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.host-name {
  font-weight: 600;
  font-size: 1.05em;
}

.host-ip {
  color: #aaa;
  font-size: 0.9em;
}

.terminal-actions {
  display: flex;
  gap: 8px;
}

.terminal-body {
  flex: 1;
  padding: 4px;
  overflow: hidden;
  background-color: #1e1e1e;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
