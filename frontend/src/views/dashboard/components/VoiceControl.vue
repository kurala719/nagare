<template>
  <el-card class="detail-card voice-card" shadow="hover">
    <template #header>
      <div class="card-header">
        <div class="header-title">
          <el-icon class="header-icon"><Microphone /></el-icon>
          <span>{{ $t('dashboard.voiceTitle') }}</span>
        </div>
        <div class="card-actions">
          <el-tooltip :content="voiceSupported ? (voiceListening ? $t('dashboard.voiceStop') : $t('dashboard.voiceStart')) : $t('dashboard.voiceNotSupported')" placement="top">
            <el-button 
              circle 
              :type="voiceListening ? 'danger' : 'primary'" 
              :disabled="!voiceSupported" 
              @click="toggleVoiceListening"
              class="mic-btn"
              :class="{ 'is-listening': voiceListening }"
            >
              <el-icon :size="20">
                <Microphone v-if="!voiceListening" />
                <Close v-else />
              </el-icon>
            </el-button>
          </el-tooltip>
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
        class="mb-4"
      />
      
      <div class="voice-interaction-area">
        <div class="status-indicator">
          <div class="status-dot" :class="voiceListening ? 'active' : 'idle'"></div>
          <span class="status-text">{{ voiceListening ? $t('dashboard.voiceListening') : $t('dashboard.voiceIdle') }}</span>
        </div>

        <div class="transcript-box" :class="{ 'has-content': voiceTranscript }">
          <p v-if="voiceTranscript" class="transcript-text">"{{ voiceTranscript }}"</p>
          <p v-else class="voice-hint">{{ $t('dashboard.voiceHint') || 'Try saying "Health status" or "Show alerts"' }}</p>
        </div>

        <div v-if="voiceLastAction" class="action-result animate-fade-in">
          <el-icon class="action-icon"><Check /></el-icon>
          <span>{{ voiceLastAction }}</span>
        </div>

        <div v-if="voiceError" class="error-msg animate-fade-in">
          <el-icon><Warning /></el-icon>
          <span>{{ voiceError }}</span>
        </div>
      </div>
      
      <div class="suggested-commands">
        <el-tag size="small" effect="plain" round class="cmd-tag">"Health status"</el-tag>
        <el-tag size="small" effect="plain" round class="cmd-tag">"Show alerts"</el-tag>
        <el-tag size="small" effect="plain" round class="cmd-tag">"Find switches"</el-tag>
      </div>
    </div>
  </el-card>
</template>

<script>
import { defineComponent, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Microphone, Close, Check, Warning } from '@element-plus/icons-vue'
import request from '@/utils/request'

export default defineComponent({
  name: 'VoiceControl',
  components: {
    Microphone,
    Close,
    Check,
    Warning
  },
  setup() {
    const { t } = useI18n()
    const router = useRouter()
    const voiceSupported = ref(false)
    const voiceListening = ref(false)
    const voiceTranscript = ref('')
    const voiceLastAction = ref('')
    const voiceError = ref('')
    let voiceRecognizer = null
    let listeningTimeout = null

    const initVoiceRecognition = () => {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition
      if (!SpeechRecognition) {
        voiceSupported.value = false
        return
      }
      voiceSupported.value = true
      const recognizer = new SpeechRecognition()
      // Use user's locale or default to English/Chinese based on i18n
      recognizer.lang = t('common.language') === '语言' ? 'zh-CN' : 'en-US'
      recognizer.interimResults = true // Enable interim results for snappier UI
      recognizer.maxAlternatives = 1
      recognizer.continuous = false

      recognizer.onresult = (event) => {
        let interimTranscript = ''
        let finalTranscript = ''

        for (let i = event.resultIndex; i < event.results.length; ++i) {
          if (event.results[i].isFinal) {
            finalTranscript += event.results[i][0].transcript
          } else {
            interimTranscript += event.results[i][0].transcript
          }
        }
        
        voiceTranscript.value = finalTranscript || interimTranscript
        
        if (finalTranscript) {
          executeVoiceAction(finalTranscript)
          // Auto-stop after a final result
          setTimeout(() => {
            if (voiceListening.value) {
              toggleVoiceListening()
            }
          }, 1000)
        }
      }

      recognizer.onend = () => {
        voiceListening.value = false
        clearTimeout(listeningTimeout)
      }

      recognizer.onerror = (event) => {
        if (event.error !== 'no-speech') {
          voiceError.value = event?.error || t('dashboard.voiceFailed')
        }
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
        clearTimeout(listeningTimeout)
        return
      }
      
      voiceError.value = ''
      voiceTranscript.value = ''
      voiceLastAction.value = ''
      
      try {
        voiceRecognizer.lang = t('common.language') === '语言' ? 'zh-CN' : 'en-US'
        voiceRecognizer.start()
        voiceListening.value = true
        
        // Auto timeout after 10 seconds if no speech
        listeningTimeout = setTimeout(() => {
          if (voiceListening.value) {
            voiceRecognizer.stop()
            voiceListening.value = false
            if (!voiceTranscript.value) {
              voiceError.value = 'Listening timeout'
            }
          }
        }, 10000)
      } catch (err) {
        voiceListening.value = false
        voiceError.value = t('dashboard.voiceFailed')
      }
    }

    const executeVoiceAction = async (text) => {
      const normalized = String(text || '').toLowerCase()
      if (!normalized) return
      
      // Health / Status
      if (normalized.match(/health|status|健康|状态/i)) {
        voiceLastAction.value = t('dashboard.voiceActionHealth') || 'Checking system health...'
        await fetchQuickHealth()
        return
      }
      // Alerts / Incidents
      if (normalized.match(/alert|incident|告警|报警/i)) {
        voiceLastAction.value = t('dashboard.voiceActionAlerts') || 'Opening alerts page...'
        setTimeout(() => router.push('/alert'), 800)
        return
      }
      // Switch / Networking
      if (normalized.match(/switch|network|交换机|网络/i)) {
        voiceLastAction.value = t('dashboard.voiceActionSwitch') || 'Finding network switches...'
        setTimeout(() => router.push({ path: '/host', query: { q: 'switch' } }), 800)
        return
      }
      // Topology / Map
      if (normalized.match(/topology|map|拓扑|地图/i)) {
        voiceLastAction.value = t('dashboard.voiceActionTopology') || 'Opening network topology...'
        ElMessage.info(voiceLastAction.value)
        return
      }
      // Assets / Hosts
      if (normalized.match(/asset|host|device|资产|主机|设备/i)) {
        voiceLastAction.value = 'Opening assets page...'
        setTimeout(() => router.push('/host'), 800)
        return
      }
      
      voiceLastAction.value = t('dashboard.voiceNoMatch') || 'Command not recognized.'
    }

    const fetchQuickHealth = async () => {
      try {
        const data = await request({
          url: '/system/health',
          method: 'GET'
        })
        const payload = data?.data || data || {}
        const score = payload.score ?? '--'
        ElMessage.success(`${t('dashboard.voiceHealthResult') || 'System Health Score'}: ${score}`)
      } catch (err) {
        ElMessage.error(t('dashboard.voiceHealthFailed') || 'Failed to fetch health score')
      }
    }

    onMounted(() => {
      initVoiceRecognition()
    })

    onBeforeUnmount(() => {
      clearTimeout(listeningTimeout)
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
.voice-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.voice-card :deep(.el-card__body) {
  padding: 20px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.header-icon {
  color: var(--el-color-primary);
  font-size: 18px;
}

.mic-btn {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.mic-btn.is-listening {
  animation: pulse-ring 2s cubic-bezier(0.215, 0.61, 0.355, 1) infinite;
  transform: scale(1.1);
}

@keyframes pulse-ring {
  0% { box-shadow: 0 0 0 0 rgba(245, 108, 108, 0.7); }
  70% { box-shadow: 0 0 0 10px rgba(245, 108, 108, 0); }
  100% { box-shadow: 0 0 0 0 rgba(245, 108, 108, 0); }
}

.voice-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.voice-interaction-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 16px;
  margin-top: 10px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  transition: all 0.3s;
}

.status-dot.idle {
  background-color: var(--el-color-info);
}

.status-dot.active {
  background-color: var(--el-color-danger);
  box-shadow: 0 0 8px var(--el-color-danger);
}

.transcript-box {
  width: 100%;
  min-height: 60px;
  padding: 16px;
  border-radius: var(--radius-lg);
  background: var(--surface-2);
  border: 1px dashed var(--border-1);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.transcript-box.has-content {
  background: var(--brand-50);
  border-color: var(--brand-200);
}

.transcript-text {
  margin: 0;
  font-size: 16px;
  font-style: italic;
  color: var(--el-color-primary);
  font-weight: 500;
}

.voice-hint {
  margin: 0;
  color: var(--text-muted);
  font-size: 13px;
}

.action-result {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.error-msg {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--el-color-danger);
  font-size: 13px;
}

.suggested-commands {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-1);
}

.cmd-tag {
  color: var(--text-muted);
  background: transparent;
  border-color: var(--border-2);
}

.animate-fade-in {
  animation: fadeIn 0.4s ease-out forwards;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(5px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>

