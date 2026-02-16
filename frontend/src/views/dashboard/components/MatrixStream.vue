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

export default defineComponent({
  name: 'MatrixStream',
  setup() {
    const { t } = useI18n()
    const logs = ref([])
    const streaming = ref(true)
    const streamRef = ref(null)
    let timer = null
    let logSeed = 0

    const toggleStream = () => {
      streaming.value = !streaming.value
    }

    const buildLogLine = () => {
      const phrases = [
        'Analyzing AI context window', 'Syncing monitor node', 'Optimizing token usage',
        'Indexing anomaly signatures', 'Merging health telemetry', 'Calibrating alert thresholds',
        'Reconciling host heartbeat', 'Streaming topology edges', 'Normalizing metric deltas',
        'Rebuilding signal graph',
      ]
      const phrase = phrases[Math.floor(Math.random() * phrases.length)]
      return `${phrase} :: [${Math.random().toString(16).substr(2, 8)}]`
    }

    const appendLog = () => {
      if (!streaming.value) return
      logs.value.push({
        id: logSeed++,
        time: new Date().toLocaleTimeString(),
        text: buildLogLine(),
      })
      if (logs.value.length > 200) {
        logs.value.splice(0, logs.value.length - 200)
      }
      nextTick(() => {
        if (streamRef.value) {
          streamRef.value.scrollTop = streamRef.value.scrollHeight
        }
      })
    }

    onMounted(() => {
      timer = setInterval(appendLog, 200)
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
  height: 240px;
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
