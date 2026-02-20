<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('ansible.jobsTitle') }}</h1>
      <p class="page-subtitle">{{ $t('ansible.jobsSubtitle') }}</p>
    </div>

    <div class="standard-toolbar" v-if="!currentJobId">
      <div class="filter-group">
        <el-button :icon="ArrowLeft" @click="$router.back()">{{ $t('common.back') }}</el-button>
      </div>
      <div class="action-group">
        <el-button type="primary" :icon="Refresh" @click="loadJobs">
          {{ $t('common.refresh') }}
        </el-button>
      </div>
    </div>

    <!-- Job List View -->
    <el-card v-if="!currentJobId" class="jobs-card">
      <el-table :data="jobs" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="playbook_name" :label="$t('ansible.pbName')" min-width="150" />
        <el-table-column prop="host_filter" :label="$t('ansible.hostFilter')" width="150" />
        <el-table-column prop="status" :label="$t('ansible.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" :label="$t('ansible.colTime')" width="180" />
        <el-table-column :label="$t('common.actions')" width="120" align="center">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewJobDetail(row.id)">{{ $t('ansible.viewLogs') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Job Detail / Console View -->
    <div v-else class="console-view">
      <div class="console-header">
        <div class="header-left">
          <el-button :icon="ArrowLeft" circle @click="currentJobId = null" />
          <span class="job-info">Job #{{ currentJobId }} - {{ activeJob?.playbook_name }}</span>
          <el-tag :type="getStatusType(activeJob?.status)" size="small" style="margin-left: 10px">
            {{ activeJob?.status }}
          </el-tag>
        </div>
        <div class="header-right">
          <el-button v-if="activeJob?.status === 'running'" type="danger" size="small" disabled>Stop</el-button>
        </div>
      </div>
      
      <div ref="terminalElement" class="terminal-body"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Refresh } from '@element-plus/icons-vue'
import { fetchAnsibleJobs, getAnsibleJob } from '@/api/ansible'
import { getToken } from '@/utils/auth'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const jobs = ref([])
const currentJobId = ref(route.params.id)
const activeJob = ref(null)

const terminalElement = ref(null)
let term = null
let fitAddon = null
let ws = null

const loadJobs = async () => {
  loading.value = true
  try {
    const res = await fetchAnsibleJobs()
    if (res && res.success) {
      jobs.value = res.data || []
    }
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const viewJobDetail = (id) => {
  router.push(`/ansible/jobs/${id}`)
}

const getStatusType = (status) => {
  switch (status) {
    case 'success': return 'success'
    case 'failed': return 'danger'
    case 'running': return 'warning'
    default: return 'info'
  }
}

const initTerminal = () => {
  term = new Terminal({
    cursorBlink: true,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    fontSize: 13,
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4'
    },
    convertEol: true,
    rows: 30
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalElement.value)
  
  nextTick(() => {
    fitAddon.fit()
  })
}

const connectWebSocket = () => {
  const token = getToken()
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  // Reuse the site-message WS or add specific one. 
  // For simplicity here, we'll use site-message WS as it already broadcasts everything
  const url = `${protocol}//${host}/api/v1/site-messages/ws?token=${token}`

  ws = new WebSocket(url)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.event === 'ansible_log' && data.job_id === Number(currentJobId.value)) {
        term.write(data.data)
      }
      if (data.event === 'ansible_status' && data.job_id === Number(currentJobId.value)) {
        activeJob.value.status = data.status
      }
    } catch (e) {
      console.error(e)
    }
  }
}

const loadActiveJob = async () => {
  try {
    const res = await getAnsibleJob(currentJobId.value)
    if (res && res.success) {
      activeJob.value = res.data
      if (term) {
        term.clear()
        term.write(activeJob.value.output || '')
      }
    }
  } catch (e) {
    console.error(e)
  }
}

watch(() => route.params.id, (newId) => {
  currentJobId.value = newId
  if (newId) {
    nextTick(() => {
      if (!term) initTerminal()
      loadActiveJob()
    })
  }
})

onMounted(() => {
  if (currentJobId.value) {
    initTerminal()
    loadActiveJob()
    connectWebSocket()
  } else {
    loadJobs()
  }
})

onUnmounted(() => {
  if (ws) ws.close()
  if (term) term.dispose()
})
</script>

<style scoped>
.jobs-card {
  margin-top: 20px;
}

.console-view {
  background-color: #1e1e1e;
  border-radius: 8px;
  overflow: hidden;
  margin-top: 20px;
  display: flex;
  flex-direction: column;
}

.console-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 15px;
  background-color: #333;
  color: #fff;
  border-bottom: 1px solid #444;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.job-info {
  font-weight: 600;
}

.terminal-body {
  padding: 10px;
  background-color: #1e1e1e;
}

:deep(.xterm-viewport) {
  overflow-y: auto !important;
}
</style>
