<template>
  <el-card class="detail-card voice-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <span>{{ $t('dashboard.voiceTitle') }}</span>
        <div class="card-actions">
          <el-button size="small" type="primary" :disabled="!voiceSupported" @click="toggleVoiceListening">
            {{ voiceListening ? $t('dashboard.voiceStop') : $t('dashboard.voiceStart') }}
          </el-button>
        </div>
      </div>
    </template>

    <div class="voice-body">
      <el-alert
        v-if="!voiceSupported"
        type="warning"
        :title="$t('dashboard.voiceNotSupported')"
        :closable="false"
        show-icon
      />
      <div v-else class="voice-status">
        <div class="voice-status-pill" :class="voiceListening ? 'active' : 'idle'">
          {{ voiceListening ? $t('dashboard.voiceListening') : $t('dashboard.voiceIdle') }}
        </div>
        <p class="voice-hint">{{ $t('dashboard.voiceHint') }}</p>
        <div v-if="voiceTranscript" class="voice-transcript">
          <span class="voice-label">{{ $t('dashboard.voiceTranscript') }}:</span>
          <span>{{ voiceTranscript }}</span>
        </div>
        <div v-if="voiceLastAction" class="voice-action">
          <span class="voice-label">{{ $t('dashboard.voiceAction') }}:</span>
          <span>{{ voiceLastAction }}</span>
        </div>
        <div v-if="voiceError" class="voice-error">{{ voiceError }}</div>
      </div>
    </div>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

export default defineComponent({
  name: 'VoiceControl',
  setup() {
    const { t } = useI18n()
    const router = useRouter()
    const voiceSupported = ref(false)
    const voiceListening = ref(false)
    const voiceTranscript = ref('')
    const voiceLastAction = ref('')
    const voiceError = ref('')
    let voiceRecognizer = null

    const initVoiceRecognition = () => {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition
      if (!SpeechRecognition) {
        voiceSupported.value = false
        return
      }
      voiceSupported.value = true
      const recognizer = new SpeechRecognition()
      recognizer.lang = 'zh-CN'
      recognizer.interimResults = false
      recognizer.maxAlternatives = 1
      recognizer.continuous = false
      recognizer.onresult = (event) => {
        const transcript = event?.results?.[0]?.[0]?.transcript || ''
        voiceTranscript.value = transcript
        executeVoiceAction(transcript)
      }
      recognizer.onend = () => {
        voiceListening.value = false
      }
      recognizer.onerror = (event) => {
        voiceError.value = event?.error || t('dashboard.voiceFailed')
        voiceListening.value = false
      }
      voiceRecognizer = recognizer
    }

    const toggleVoiceListening = () => {
      if (!voiceSupported.value || !voiceRecognizer) {
        voiceError.value = t('dashboard.voiceNotSupported')
        return
      }
      if (voiceListening.value) {
        voiceRecognizer.stop()
        voiceListening.value = false
        return
      }
      voiceError.value = ''
      voiceTranscript.value = ''
      voiceLastAction.value = ''
      try {
        voiceListening.value = true
        voiceRecognizer.start()
      } catch (err) {
        voiceListening.value = false
        voiceError.value = t('dashboard.voiceFailed')
      }
    }

    const executeVoiceAction = async (text) => {
      const normalized = String(text || '').toLowerCase()
      if (!normalized) return
      if (normalized.includes('health') || normalized.includes('健康') || normalized.includes('状态')) {
        voiceLastAction.value = t('dashboard.voiceActionHealth')
        await fetchQuickHealth()
        return
      }
      if (normalized.includes('alert') || normalized.includes('告警')) {
        voiceLastAction.value = t('dashboard.voiceActionAlerts')
        router.push('/alert')
        return
      }
      if (normalized.includes('switch') || normalized.includes('交换机')) {
        voiceLastAction.value = t('dashboard.voiceActionSwitch')
        router.push({ path: '/host', query: { q: 'switch' } })
        return
      }
      if (normalized.includes('topology') || normalized.includes('拓扑')) {
        voiceLastAction.value = t('dashboard.voiceActionTopology')
        // Emit event or handle in parent? 
        // For simplicity, just show message
        ElMessage.info(t('dashboard.voiceActionTopology'))
        return
      }
      voiceLastAction.value = t('dashboard.voiceNoMatch')
    }

    const fetchQuickHealth = async () => {
      try {
        const data = await request({
          url: '/system/health',
          method: 'GET'
        })
        const payload = data?.data || data || {}
        const score = payload.score ?? '--'
        ElMessage.success(`${t('dashboard.voiceHealthResult')}: ${score}`)
      } catch (err) {
        ElMessage.error(t('dashboard.voiceHealthFailed'))
      }
    }

    onMounted(() => {
      initVoiceRecognition()
    })

    onBeforeUnmount(() => {
      if (voiceRecognizer && voiceListening.value) {
        try {
          voiceRecognizer.stop()
        } catch (err) {
          console.warn('Failed to stop voice recognition:', err)
        }
      }
    })

    return {
      voiceSupported,
      voiceListening,
      voiceTranscript,
      voiceLastAction,
      voiceError,
      toggleVoiceListening
    }
  }
})
</script>

<style scoped>
.voice-card :deep(.el-card__body) {
  padding: 18px;
}
.voice-body {
  min-height: 180px;
}
.voice-status {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.voice-status-pill {
  width: fit-content;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}
.voice-status-pill.idle {
  background: #f3f4f6;
  color: #4b5563;
}
.voice-status-pill.active {
  background: linear-gradient(135deg, #10b981, #22c55e);
  color: #ffffff;
  box-shadow: 0 6px 14px rgba(34, 197, 94, 0.35);
}
.voice-hint {
  color: #6b7280;
  font-size: 13px;
}
.voice-transcript,
.voice-action {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 13px;
  color: #1f2937;
}
.voice-label {
  font-weight: 600;
  color: #111827;
}
.voice-error {
  color: #b91c1c;
  font-size: 12px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>

