<template>
    <div class="chat-panel">
        <div class="chat-header">
            <div class="chat-header-row">
                <div class="chat-title">
                    <h3>{{ t('chat.title') }}</h3>
                    <p>{{ t('chat.subtitle') }}</p>
                </div>
                <div class="chat-tools">
                    <div class="chat-tool-group">
                        <span class="tool-label">{{ t('chat.toneLabel') }}</span>
                        <el-radio-group v-model="toneMode" size="small">
                            <el-radio-button label="professional">{{ t('chat.toneProfessional') }}</el-radio-button>
                            <el-radio-button label="roast">{{ t('chat.toneRoast') }}</el-radio-button>
                        </el-radio-group>
                    </div>
                    <div class="chat-tool-group">
                        <span class="tool-label">{{ t('chat.toolMode') }}</span>
                        <el-switch v-model="toolModeEnabled" size="small" />
                        <span v-if="toolModeEnabled" class="tool-badge">{{ t('chat.toolBadge') }}</span>
                    </div>
                </div>
            </div>
        </div>
        <div class="chat-provider">
            <el-select 
                v-model="selectedProviderId" 
                :placeholder="t('chat.selectProvider')" 
                class="full-width"
                @change="onProviderChange"
            >
                <el-option 
                    v-for="provider in providers" 
                    :key="provider.id" 
                    :label="provider.name" 
                    :value="provider.id" 
                />
            </el-select>
        </div>
        <el-divider content-position="center">{{ t('chat.section') }}</el-divider>
        <div class="chat-messages" ref="messagesContainer" @scroll="onScroll">
            <div v-if="loadingHistory" class="loading-history">
                <el-icon class="is-loading"><Loading /></el-icon>
                <span>{{ t('chat.loadingHistory') }}</span>
            </div>
            <div v-if="historyLoaded && messages.length > 0" class="history-divider">
                <el-divider content-position="center">{{ t('chat.historyLoaded') }}</el-divider>
            </div>
            <div v-if="!loadingHistory && messages.length === 0" class="empty-state">
                {{ t('chat.empty') }}
            </div>
            <div v-for="message in messages" :key="message.id" 
                 :class="['message', message.role === 'user' ? 'user-message' : 'assistant-message']">
                <div class="message-content">{{ message.content }}</div>
            </div>
            <div v-if="loading" class="message assistant-message">
                <div class="message-content">{{ t('chat.thinking') }}</div>
            </div>
        </div>
        <div class="chat-footer">
            <el-input 
                v-model="talkInput" 
                class="chat-input" 
                :autosize="{ minRows: 1, maxRows: 4 }" 
                type="textarea"
                :placeholder="t('chat.placeholder')" 
                @keyup.enter.ctrl="sendMessage"
                :disabled="loading"
            />
            <el-button 
                class="chat-send"
                type="primary" 
                @click="sendMessage" 
                :loading="loading"
                :disabled="!talkInput.trim()"
            >{{ t('chat.send') }}</el-button>
        </div>
    </div>
</template>

<script lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { sendChatMessage, fetchChatHistory } from '@/api/chats';
import { fetchProviderData } from '@/api/providers';
import { getToken } from '@/utils/auth';

interface Chat {
    id: number | string;
    provider_id: number;
    role: 'user' | 'assistant';
    model: string;
    content: string;
}

interface Provider {
    id: number;
    name: string;
    default_model: string;
    url: string;
    api_key: string;
    description: string;
    status: number;
}

export default {
    name: 'SideBarChat',
    components: {
        Loading,
    },
    data() {
        return {
            talkInput: '',
            messages: [] as Chat[],
            loading: false,
            loadingHistory: false,
            historyLoaded: false,
            providers: [] as Provider[],
            selectedProviderId: null as number | null,
            toolModeEnabled: true,
            toneMode: 'professional',
        };
    },
    setup() {
        const { t } = useI18n();
        return { t };
    },
    created() {
        if (getToken()) {
            this.loadProviders();
        }
    },
    methods: {
        async loadProviders() {
            try {
                const response = await fetchProviderData();
                const data = Array.isArray(response) ? response : (response.data || response.providers || []);
                this.providers = data.map((p) => ({
                    id: p.ID || p.id,
                    name: p.Name || p.name || '',
                    default_model: p.DefaultModel || p.default_model || p.Model || p.model || '',
                    url: p.URL || p.url || '',
                    api_key: p.APIKey || p.api_key || '',
                    description: p.Description || p.description || '',
                    status: p.Status ?? p.status ?? 0,
                }));
                // Auto-select first provider if available
                if (this.providers.length > 0 && !this.selectedProviderId) {
                    this.selectedProviderId = this.providers[0].id;
                    this.onProviderChange(this.selectedProviderId);
                }
            } catch (err) {
                console.error('Error loading providers:', err);
            }
        },
        onProviderChange(providerId: number) {
            // Provider changed, model will be taken from provider's default
        },
        async loadChatHistory() {
            if (this.historyLoaded || this.loadingHistory) return;
            
            this.loadingHistory = true;
            try {
                const response = await fetchChatHistory();
                const data = Array.isArray(response) ? response : (response.data || response.messages || []);
                // Map PascalCase fields from backend and reverse to show oldest first
                const historyMessages = data.map((msg, index) => ({
                    id: msg.ID || msg.id || index,
                    provider_id: msg.ProviderID || msg.provider_id || 0,
                    role: (msg.Role || msg.role || 'user').toLowerCase(),
                    model: msg.Model || msg.model || '',
                    content: msg.Content || msg.content || '',
                })).reverse();
                // Prepend history to current messages
                this.messages = [...historyMessages, ...this.messages];
                this.historyLoaded = true;
            } catch (err) {
                console.error('Error loading chat history:', err);
            } finally {
                this.loadingHistory = false;
            }
        },
        onScroll() {
            const container = this.$refs.messagesContainer;
            if (container && container.scrollTop === 0 && !this.historyLoaded && !this.loadingHistory) {
                this.loadChatHistory();
            }
        },
        async sendMessage() {
            if (!this.talkInput.trim() || this.loading) return;
            
            if (!this.selectedProviderId) {
                ElMessage({
                    type: 'warning',
                    message: this.t('chat.selectProviderWarn'),
                });
                return;
            }
            
            // Get the selected provider's default model
            const selectedProvider = this.providers.find(p => p.id === this.selectedProviderId);
            const model = selectedProvider?.default_model || '';
            
            const userMessage = this.talkInput.trim();
            this.talkInput = '';
            
            // Add user message to chat
            const userMsgId = Date.now();
            this.messages.push({
                id: userMsgId,
                provider_id: this.selectedProviderId,
                role: 'user',
                model: model,
                content: userMessage,
            });
            this.$nextTick(() => this.scrollToBottom());
            
            this.loading = true;
            try {
                const locale = localStorage.getItem('nagare_locale') || 'en';
                const response = await sendChatMessage({ 
                    content: userMessage,
                    provider_id: this.selectedProviderId,
                    model: model,
                    role: 'user',
                    use_tools: this.toolModeEnabled,
                    mode: this.toneMode,
                    locale: locale,
                });
                
                // Backend returns response - handle different formats
                if (response) {
                    // Get data array from response
                    let dataArray = [];
                    if (Array.isArray(response)) {
                        dataArray = response;
                    } else if (Array.isArray(response.data)) {
                        dataArray = response.data;
                    } else if (response.data && typeof response.data === 'object') {
                        // Single message response - add assistant message
                        const assistantMsg = {
                            id: Date.now() + 1,
                            provider_id: this.selectedProviderId,
                            role: 'assistant',
                            model: model,
                            content: response.data.Content || response.data.content || response.data.message || '',
                        };
                        this.messages.push(assistantMsg);
                        return;
                    } else if (response.message || response.content) {
                        // Direct message response
                        const assistantMsg = {
                            id: Date.now() + 1,
                            provider_id: this.selectedProviderId,
                            role: 'assistant',
                            model: model,
                            content: response.message || response.content || '',
                        };
                        this.messages.push(assistantMsg);
                        return;
                    }
                    
                    if (dataArray.length > 0) {
                        // Map PascalCase fields and reverse to show oldest first
                        this.messages = dataArray.map((msg, index) => ({
                            id: msg.ID || msg.id || index,
                            provider_id: msg.ProviderID || msg.provider_id || 0,
                            role: (msg.Role || msg.role || 'user').toLowerCase(),
                            model: msg.Model || msg.model || '',
                            content: msg.Content || msg.content || '',
                        })).reverse();
                    }
                }
                
                this.$nextTick(() => this.scrollToBottom());
            } catch (err) {
                ElMessage({
                    type: 'error',
                    message: 'Failed to send message: ' + (err.message || 'Unknown error'),
                });
                console.error('Error sending message:', err);
            } finally {
                this.loading = false;
            }
        },
        scrollToBottom() {
            const container = this.$refs.messagesContainer;
            if (container) {
                container.scrollTop = container.scrollHeight;
            }
        }
    },
};
</script>

<style scoped>
.chat-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 12px;
    gap: 8px;
    background: var(--surface-2);
}

.chat-header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
}

.chat-tools {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
}

.chat-tool-group {
    display: flex;
    align-items: center;
    gap: 8px;
}

.tool-label {
    font-size: 12px;
    color: var(--text-muted);
}

.tool-badge {
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 999px;
    background: var(--text-strong);
    color: #ffffff;
}

.chat-header h3 {
    margin-bottom: 4px;
    font-weight: 600;
    color: var(--text-strong);
}

.chat-header p {
    color: var(--text-muted);
    font-size: 12px;
}

.chat-provider {
    padding-top: 4px;
}

.full-width {
    width: 100%;
}

.chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    border-radius: 10px;
    background: var(--surface-1);
    border: 1px solid var(--border-1);
}

.message {
    margin-bottom: 12px;
    padding: 10px 14px;
    border-radius: 12px;
    max-width: 85%;
    word-wrap: break-word;
}

.user-message {
    background: linear-gradient(180deg, var(--brand-500) 0%, var(--brand-600) 100%);
    color: #ffffff;
    margin-left: auto;
    text-align: right;
}

.assistant-message {
    background-color: var(--surface-2);
    color: var(--text-strong);
    border: 1px solid var(--border-1);
    margin-right: auto;
}

.message-content {
    white-space: pre-wrap;
}

.chat-footer {
    display: flex;
    align-items: flex-end;
    gap: 8px;
    padding-top: 6px;
}

.chat-input {
    flex: 1;
}

.chat-send {
    flex-shrink: 0;
}

.loading-history {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px;
    color: var(--text-muted);
}

.history-divider {
    margin-bottom: 8px;
}

.empty-state {
    text-align: center;
    color: var(--text-muted);
    font-size: 12px;
    padding: 12px 8px;
}
</style>