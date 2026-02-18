<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <div class="host-info">
        <el-button icon="ArrowLeft" circle @click="$router.back()" />
        <span class="host-name">Terminal: {{ hostName || 'Loading...' }}</span>
        <span class="host-ip">({{ hostIp }})</span>
      </div>
      <div class="terminal-actions">
        <el-button type="danger" @click="handleDisconnect">Disconnect</el-button>
        <el-button type="primary" @click="handleReconnect" :disabled="connected">Reconnect</el-button>
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
import axios from 'axios'

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
    const response = await axios.get(`/api/v1/hosts/${hostId}`)
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
      selection: '#264f78',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#e5e5e5'
    }
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalElement.value)
  fitAddon.fit()

  term.onData((data) => {
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify({ type: 'data', data }))
    }
  })

  window.addEventListener('resize', handleResize)
}

const handleResize = () => {
  if (fitAddon) {
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
  // Note: Adjust URL according to your backend proxy/server setup
  const url = `${protocol}//${host}/api/v1/hosts/${hostId}/ssh?token=${token}&cols=${term.cols}&rows=${term.rows}`

  socket = new WebSocket(url)
  socket.binaryType = 'arraybuffer'

  socket.onopen = () => {
    connected.value = true
    term.write('
*** Connected to host ***
')
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
    term.write(`
*** Connection closed: ${event.reason || 'No reason'} ***
`)
  }

  socket.onerror = (error) => {
    console.error('WebSocket Error:', error)
    term.write('
*** WebSocket error ***
')
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
  await nextTick()
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
  height: 100%;
  background-color: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background-color: #333;
  color: #fff;
  border-bottom: 1px solid #444;
}

.host-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.host-name {
  font-weight: bold;
  font-size: 1.1em;
}

.host-ip {
  color: #aaa;
  font-size: 0.9em;
}

.terminal-actions {
  display: flex;
  gap: 10px;
}

.terminal-body {
  flex: 1;
  padding: 5px;
  overflow: hidden;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
