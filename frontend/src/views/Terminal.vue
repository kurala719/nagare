<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div class="host-info">
        <el-button icon="ArrowLeft" circle @click="$router.back()" />
        <span class="host-name">Terminal: {{ hostName || 'Loading...' }}</span>
        <span class="host-ip" v-if="hostIp">({{ hostIp }})</span>
      </div>
      <div class="terminal-actions">
        <el-button type="danger" size="small" @click="handleDisconnect">Disconnect</el-button>
        <el-button type="primary" size="small" @click="handleReconnect" :disabled="connected">Reconnect</el-button>
      </div>
    </div>
    <div ref="terminalElement" class="terminal-body"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { getToken } from '@/utils/auth'
import request from '@/utils/request'

const route = useRoute()
const router = useRouter()
const hostId = route.params.id
const hostName = ref('')
const hostIp = ref('')
const connected = ref(false)

const terminalElement = ref(null)
let term = null
let fitAddon = null
let socket = null

const fetchHostInfo = async () => {
  try {
    const response = await request.get(`/hosts/${hostId}`)
    if (response.data.success) {
      hostName.value = response.data.data.name
      hostIp.value = response.data.data.ip_addr
    }
  } catch (err) {
    console.error('Failed to fetch host info', err)
  }
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
  const token = getToken()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  
  // Use term dimensions in query
  const url = `${protocol}//${host}/api/v1/hosts/${hostId}/ssh?token=${token}&cols=${term.cols}&rows=${term.rows}`

  socket = new WebSocket(url)
  socket.binaryType = 'arraybuffer'

  socket.onopen = () => {
    connected.value = true
    term.write('\r\n*** Connected to host ***\r\n')
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
    term.write(`\r\n*** Connection closed${reason} ***\r\n`)
  }

  socket.onerror = (error) => {
    console.error('WebSocket Error:', error)
    term.write('\r\n*** WebSocket error ***\r\n')
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
  await fetchHostInfo()
  initTerminal()
  connectWebSocket()
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
