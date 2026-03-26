<template>
  <el-card class="detail-card matrix-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.matrixTitle') }}</span>
        <div class="card-actions">
          <el-button size="small" @click="toggleStream">
            {{ streaming ? $t('dashboard.matrixPause') : $t('dashboard.matrixResume') }}
          </el-button>
        </div>
      </div>
    </template>
    <div class="matrix-stream" ref="streamRef">
      <div v-for="(line, index) in logs" :key="`${line.id}-${index}`" class="matrix-line">
        <span class="matrix-time">{{ line.time }}</span>
        <span class="matrix-text">{{ line.text }}</span>
      </div>
    </div>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchSystemLogs } from '@/api/system'
import { getToken } from '@/utils/auth'

export default defineComponent({
  name: 'MatrixStream',
  setup() {
    const { t } = useI18n()
    const logs = ref([])
    const streaming = ref(true)
    const streamRef = ref(null)
    const error = ref(null)
    let timer = null
    let lastLogId = 0

    const toggleStream = () => {
      streaming.value = !streaming.value
      if (streaming.value) {
        appendLog()
      }
    }

    const appendLog = async () => {
      if (!streaming.value || !getToken()) return
      
      try {
        const res = await fetchSystemLogs({ limit: 50, offset: 0 })
        const data = Array.isArray(res?.data || res) ? (res?.data || res) : []
        
        if (data.length > 0) {
          const getSeverityLabel = (level) => {
            switch (level) {
              case 0: return 'NOT_CLAS'
              case 1: return 'INFO'
              case 2: return 'WARN'
              case 3: return 'AVG'
              case 4: return 'HIGH'
              case 5: return 'CRIT'
              default: return 'INFO'
            }
          }

          // Reverse to show oldest first in this batch, if API returns newest first
          const newLogs = [...data].reverse().map(log => ({
            id: log.ID || log.id,
            time: new Date(log.CreatedAt || log.created_at).toLocaleTimeString(),
            text: `[${getSeverityLabel(log.Severity || log.severity)}] ${log.Message || log.message}`
          }))

          // For the Matrix stream effect, we only want to add genuinely new logs.
          // Assuming logs have incrementing IDs, filter for IDs greater than last seen.
          const freshLogs = newLogs.filter(l => l.id > lastLogId)
          
          if (freshLogs.length > 0) {
            lastLogId = freshLogs[freshLogs.length - 1].id
            
            logs.value.push(...freshLogs)
            if (logs.value.length > 200) {
              logs.value.splice(0, logs.value.length - 200)
            }
            nextTick(() => {
              if (streamRef.value) {
                streamRef.value.scrollTop = streamRef.value.scrollHeight
              }
            })
          }
        }
      } catch (err) {
        console.error('Failed to fetch system logs:', err)
      }
    }

    onMounted(async () => {
      // Fetch initial batch
      await appendLog()
      // Poll every 3 seconds for new logs
      timer = setInterval(appendLog, 3000)
    })

    onBeforeUnmount(() => {
      if (timer) clearInterval(timer)
    })

    return {
      logs,
      streaming,
      streamRef,
      toggleStream
    }
  }
})
</script>

<style scoped>
.matrix-card {
  background: radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 60%),
    radial-gradient(circle at bottom right, rgba(14, 116, 144, 0.18), transparent 55%),
    #050b08;
}
.matrix-card :deep(.el-card__body) {
  background: transparent;
  padding: 18px;
}
.matrix-stream {
  max-height: 240px;
  overflow-y: auto;
  padding: 12px;
  border-radius: 12px;
  background: rgba(2, 9, 6, 0.9);
  border: 1px solid rgba(16, 185, 129, 0.2);
  font-family: "JetBrains Mono", "Fira Code", monospace;
  font-size: 12px;
  color: #7bf7b0;
  text-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}
.matrix-line {
  display: flex;
  gap: 10px;
  padding: 2px 0;
  opacity: 0.9;
}
.matrix-time {
  color: rgba(123, 247, 176, 0.7);
  min-width: 78px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
