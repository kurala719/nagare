<template>
  <div class="nagare-container">
    <div class="page-header config-header">
      <div class="title-section">
        <h1 class="page-title">{{ $t('configuration.title') }}</h1>
        <p class="page-subtitle">{{ $t('configuration.subtitle') }}</p>
      </div>
      <div class="action-group">
        <el-button-group>
          <el-button @click="loadConfig" :loading="loading" :disabled="editing" :icon="Refresh">
            {{ $t('configuration.reload') }}
          </el-button>
          <el-button v-if="!editing" type="primary" @click="startEdit" :icon="Edit">
            {{ $t('configuration.edit') }}
          </el-button>
        </el-button-group>

        <template v-if="editing">
          <el-button @click="cancelEdit">
            {{ $t('configuration.cancel') }}
          </el-button>
          <el-button type="primary" @click="saveChanges" :loading="saving" :icon="Check">
            {{ $t('configuration.save') }}
          </el-button>
        </template>

        <el-button type="danger" plain @click="handleReset" :disabled="editing" :icon="Warning">
          {{ $t('configuration.reset') }}
        </el-button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && !config" class="config-loading">
      <el-icon class="is-loading" :size="50" color="#409EFF">
        <Loading />
      </el-icon>
      <p>{{ $t('configuration.loading') }}</p>
    </div>

    <!-- Error State -->
    <el-alert 
      v-if="error && !loading" 
      :title="error" 
      type="error" 
      show-icon
      style="margin: 20px 0;"
      :closable="false"
    >
      <template #default>
        <el-button size="small" @click="loadConfig">{{ $t('monitors.retry') }}</el-button>
      </template>
    </el-alert>

    <!-- Configuration Content -->
    <div v-if="config && !loading" class="config-content">
      <el-tabs v-model="activeTab" class="config-tabs" type="border-card">
        <!-- System Settings -->
        <el-tab-pane name="system">
          <template #label>
            <span class="tab-label">
              <el-icon><Setting /></el-icon>
              <span>{{ $t('configuration.system') }}</span>
            </span>
          </template>
          <div class="tab-pane-content">
            <el-form :model="editableConfig.system" label-width="180px" label-position="left">
              <el-form-item :label="$t('configuration.systemName')">
                <el-input v-model="editableConfig.system.system_name" placeholder="Nagare System" />
              </el-form-item>
              <el-form-item :label="$t('configuration.ipAddress')">
                <el-input v-model="editableConfig.system.ip_address" placeholder="127.0.0.1" />
              </el-form-item>
              <el-form-item :label="$t('configuration.port')">
                <el-input-number v-model="editableConfig.system.port" :min="1" :max="65535" />
              </el-form-item>
              <el-form-item :label="$t('configuration.availability')">
                <el-switch v-model="editableConfig.system.availability" />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- Database Settings -->
        <el-tab-pane name="database">
          <template #label>
            <span class="tab-label">
              <el-icon><DataLine /></el-icon>
              <span>{{ $t('configuration.database') }}</span>
            </span>
          </template>
          <div class="tab-pane-content">
            <el-form :model="editableConfig.database" label-width="180px" label-position="left">
              <el-form-item :label="$t('configuration.dbVersion')">
                <el-input v-model="editableConfig.database.version" placeholder="MYSQL 8.0" />
              </el-form-item>
              <el-form-item :label="$t('configuration.dbHost')">
                <el-input v-model="editableConfig.database.host" placeholder="127.0.0.1" />
              </el-form-item>
              <el-form-item :label="$t('configuration.dbPort')">
                <el-input-number v-model="editableConfig.database.port" :min="1" :max="65535" />
              </el-form-item>
              <el-form-item :label="$t('configuration.dbUsername')">
                <el-input v-model="editableConfig.database.username" />
              </el-form-item>
              <el-form-item :label="$t('configuration.dbPassword')">
                <el-input v-model="editableConfig.database.password" type="password" show-password />
              </el-form-item>
              <el-form-item :label="$t('configuration.dbName')">
                <el-input v-model="editableConfig.database.database_name" />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- Service Settings -->
        <el-tab-pane name="services">
          <template #label>
            <span class="tab-label">
              <el-icon><Compass /></el-icon>
              <span>{{ $t('configuration.services') }}</span>
            </span>
          </template>
          <div class="tab-pane-content">
            <div class="section-divider">{{ $t('configuration.syncService') }}</div>
            <el-form :model="editableConfig.sync" label-width="180px" label-position="left">
              <el-form-item :label="$t('system.syncEnabled')">
                <el-switch v-model="editableConfig.sync.enabled" />
              </el-form-item>
              <el-form-item :label="$t('system.syncInterval')">
                <el-input-number v-model="editableConfig.sync.interval_seconds" :min="1" />
              </el-form-item>
              <el-form-item :label="$t('system.syncConcurrency')">
                <el-input-number v-model="editableConfig.sync.concurrency" :min="1" />
              </el-form-item>
            </el-form>

            <div class="section-divider">{{ $t('configuration.statusCheckService') }}</div>
            <el-form :model="editableConfig.status_check" label-width="180px" label-position="left">
              <el-form-item :label="$t('system.statusCheckEnabled')">
                <el-switch v-model="editableConfig.status_check.enabled" />
              </el-form-item>
              <el-form-item :label="$t('system.statusCheckProviderEnabled')">
                <el-switch v-model="editableConfig.status_check.provider_enabled" />
              </el-form-item>
              <el-form-item :label="$t('system.statusCheckInterval')">
                <el-input-number v-model="editableConfig.status_check.interval_seconds" :min="1" />
              </el-form-item>
              <el-form-item :label="$t('system.statusCheckConcurrency')">
                <el-input-number v-model="editableConfig.status_check.concurrency" :min="1" />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- AI & Integrations -->
        <el-tab-pane name="integrations">
          <template #label>
            <span class="tab-label">
              <el-icon><Connection /></el-icon>
              <span>{{ $t('configuration.integrations') }}</span>
            </span>
          </template>
          <div class="tab-pane-content">
            <div class="section-divider">{{ $t('configuration.aiAnalysis') }}</div>
            <el-form :model="editableConfig.ai" label-width="180px" label-position="left">
              <el-form-item :label="$t('system.aiEnabled')">
                <el-switch v-model="editableConfig.ai.analysis_enabled" />
              </el-form-item>
              <el-form-item :label="$t('system.aiProviderId')">
                <el-input-number v-model="editableConfig.ai.provider_id" :min="0" />
              </el-form-item>
              <el-form-item :label="$t('system.aiModel')">
                <el-input v-model="editableConfig.ai.model" placeholder="gemini-1.5-pro" />
              </el-form-item>
              <el-form-item :label="$t('system.aiTimeout')">
                <el-input-number v-model="editableConfig.ai.analysis_timeout_seconds" :min="1" />
              </el-form-item>
              <el-form-item :label="$t('system.aiMinSeverity')">
                <el-input-number v-model="editableConfig.ai.analysis_min_severity" :min="0" :max="4" />
              </el-form-item>
            </el-form>

            <div class="section-divider">{{ $t('configuration.mcpSettings') }}</div>
            <el-form :model="editableConfig.mcp" label-width="180px" label-position="left">
              <el-form-item :label="$t('system.mcpEnabled')">
                <el-switch v-model="editableConfig.mcp.enabled" />
              </el-form-item>
              <el-form-item :label="$t('system.mcpApiKey')">
                <el-input v-model="editableConfig.mcp.api_key" type="password" show-password />
              </el-form-item>
              <el-form-item :label="$t('system.mcpMaxConcurrency')">
                <el-input-number v-model="editableConfig.mcp.max_concurrency" :min="1" />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- External Infrastructure -->
        <el-tab-pane name="external">
          <template #label>
            <span class="tab-label">
              <el-icon><Share /></el-icon>
              <span>{{ $t('configuration.external') }}</span>
            </span>
          </template>
          <div class="tab-pane-content" style="max-width: 900px;">
            <div class="section-header">
              <div class="section-divider">{{ $t('configuration.externalInfrastructure') }}</div>
              <el-button type="primary" size="small" @click="addExternalItem" :disabled="!editing">
                <el-icon><Plus /></el-icon>
                {{ $t('configuration.addItem') }}
              </el-button>
            </div>
            
            <el-table :data="editableConfig.external" border stripe style="width: 100%; margin-top: 16px;">
              <el-table-column :label="$t('configuration.itemType')" width="150">
                <template #default="{ row }">
                  <el-select v-model="row.type" size="small" :disabled="!editing">
                    <el-option label="Monitor" value="monitor" />
                    <el-option label="Alarm" value="alarm" />
                    <el-option label="Provider" value="provider" />
                    <el-option label="Media" value="media" />
                  </el-select>
                </template>
              </el-table-column>
              <el-table-column :label="$t('configuration.itemKey')" width="150">
                <template #default="{ row }">
                  <el-input v-model="row.key" size="small" :disabled="!editing" placeholder="zabbix" />
                </template>
              </el-table-column>
              <el-table-column :label="$t('configuration.itemName')" min-width="150">
                <template #default="{ row }">
                  <el-input v-model="row.name" size="small" :disabled="!editing" placeholder="Zabbix Server" />
                </template>
              </el-table-column>
              <el-table-column :label="$t('configuration.itemId')" width="100">
                <template #default="{ row }">
                  <el-input-number v-model="row.id" :controls="false" size="small" :disabled="!editing" style="width: 100%;" />
                </template>
              </el-table-column>
              <el-table-column label="Actions" width="80" align="center">
                <template #default="{ $index }">
                  <el-button type="danger" :icon="Delete" circle size="small" @click="removeExternalItem($index)" :disabled="!editing" />
                </template>
              </el-table-column>
            </el-table>
            <p class="help-text">These definitions define the available types for monitors, alarms, AI providers, and notification media.</p>
          </div>
        </el-tab-pane>

        <!-- Advanced -->
        <el-tab-pane name="advanced">
          <template #label>
            <span class="tab-label">
              <el-icon><Operation /></el-icon>
              <span>{{ $t('configuration.advanced') }}</span>
            </span>
          </template>
          <div class="tab-pane-content">
            <div class="section-divider">{{ $t('system.mediaRateLimitSettings') }}</div>
            <el-form :model="editableConfig.media_rate_limit" label-width="180px" label-position="left">
              <el-form-item label="Global Interval (s)">
                <el-input-number v-model="editableConfig.media_rate_limit.global_interval_seconds" :min="0" />
              </el-form-item>
              <el-form-item label="Media Type Interval (s)">
                <el-input-number v-model="editableConfig.media_rate_limit.media_type_interval_seconds" :min="0" />
              </el-form-item>
              <el-form-item label="Media Interval (s)">
                <el-input-number v-model="editableConfig.media_rate_limit.media_interval_seconds" :min="0" />
              </el-form-item>
            </el-form>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  Refresh, Edit, Check, Setting, DataLine, Loading, 
  Monitor, Bell, Warning, Compass, Connection, Operation,
  Share, Plus, Delete
} from '@element-plus/icons-vue';
import { getMainConfig, updateConfig, saveConfig, resetConfig } from '@/api/config';

export default {
  name: 'Configuration',
  components: {
    Refresh,
    Edit,
    Check,
    Setting,
    DataLine,
    Loading,
    Monitor,
    Bell,
    Warning,
    Compass,
    Connection,
    Operation,
    Share,
    Plus,
    Delete,
  },
  setup() {
    const { t } = useI18n();
    const config = ref(null);
    const activeTab = ref('system');
    const editableConfig = reactive({
      system: {
        system_name: '',
        ip_address: '',
        port: 8080,
        availability: true,
      },
      database: {
        version: '',
        host: '',
        port: 3306,
        username: '',
        password: '',
        database_name: '',
      },
      sync: {
        enabled: true,
        interval_seconds: 60,
        concurrency: 2,
      },
      status_check: {
        enabled: true,
        provider_enabled: false,
        interval_seconds: 60,
        concurrency: 4,
      },
      mcp: {
        enabled: true,
        api_key: '',
        max_concurrency: 4,
      },
      ai: {
        analysis_enabled: true,
        provider_id: 1,
        model: '',
        analysis_timeout_seconds: 60,
        analysis_min_severity: 2,
      },
      media_rate_limit: {
        global_interval_seconds: 30,
        media_type_interval_seconds: 30,
        media_interval_seconds: 30,
      },
      external: [],
    });
    const loading = ref(false);
    const saving = ref(false);
    const editing = ref(false);
    const error = ref(null);

    const mapData = (source, target, fieldMap) => {
      if (!source) return;
      Object.entries(fieldMap).forEach(([targetKey, sourceKeys]) => {
        for (const sKey of sourceKeys) {
          if (source[sKey] !== undefined) {
            target[targetKey] = source[sKey];
            break;
          }
        }
      });
    };

    const performMapping = (data) => {
      if (!data) return;

      const systemSource = data.system || data.System || {};
      mapData(systemSource, editableConfig.system, {
        system_name: ['system_name', 'SystemName'],
        ip_address: ['ip_address', 'IPAddress'],
        port: ['port', 'Port'],
        availability: ['availability', 'Availability']
      });

      const dbSource = data.database || data.Database || {};
      mapData(dbSource, editableConfig.database, {
        version: ['version', 'Version'],
        host: ['host', 'Host'],
        port: ['port', 'Port'],
        username: ['username', 'Username'],
        password: ['password', 'Password'],
        database_name: ['database_name', 'DatabaseName']
      });

      const syncSource = data.sync || data.Sync || {};
      mapData(syncSource, editableConfig.sync, {
        enabled: ['enabled', 'Enabled'],
        interval_seconds: ['interval_seconds', 'IntervalSeconds'],
        concurrency: ['concurrency', 'Concurrency']
      });

      const statusSource = data.status_check || data.StatusCheck || {};
      mapData(statusSource, editableConfig.status_check, {
        enabled: ['enabled', 'Enabled'],
        provider_enabled: ['provider_enabled', 'ProviderEnabled'],
        interval_seconds: ['interval_seconds', 'IntervalSeconds'],
        concurrency: ['concurrency', 'Concurrency']
      });

      const mcpSource = data.mcp || data.MCP || {};
      mapData(mcpSource, editableConfig.mcp, {
        enabled: ['enabled', 'Enabled'],
        api_key: ['api_key', 'APIKey'],
        max_concurrency: ['max_concurrency', 'MaxConcurrency']
      });

      const aiSource = data.ai || data.AI || {};
      mapData(aiSource, editableConfig.ai, {
        analysis_enabled: ['analysis_enabled', 'AnalysisEnabled'],
        provider_id: ['provider_id', 'ProviderID'],
        model: ['model', 'Model'],
        analysis_timeout_seconds: ['analysis_timeout_seconds', 'AnalysisTimeoutSeconds'],
        analysis_min_severity: ['analysis_min_severity', 'AnalysisMinSeverity']
      });

      const mediaSource = data.media_rate_limit || data.MediaRateLimit || {};
      mapData(mediaSource, editableConfig.media_rate_limit, {
        global_interval_seconds: ['global_interval_seconds', 'GlobalIntervalSeconds'],
        media_type_interval_seconds: ['media_type_interval_seconds', 'MediaTypeIntervalSeconds'],
        media_interval_seconds: ['media_interval_seconds', 'MediaIntervalSeconds']
      });

      if (data.external || data.External) {
        editableConfig.external = JSON.parse(JSON.stringify(data.external || data.External));
      } else {
        editableConfig.external = [];
      }
    };

    const loadConfig = async () => {
      loading.value = true;
      error.value = null;
      try {
        const response = await getMainConfig();
        const data = response.data?.data || response.data || response;
        console.log('Loaded config data:', data);
        config.value = data;
        performMapping(data);
      } catch (err) {
        error.value = err.message || 'Failed to load configuration';
        console.error('Error loading configuration:', err);
      } finally {
        loading.value = false;
      }
    };

    const startEdit = () => {
      editing.value = true;
    };

    const cancelEdit = () => {
      editing.value = false;
      if (config.value) {
        performMapping(config.value);
      }
    };

    const saveChanges = async () => {
      try {
        await ElMessageBox.confirm(
          t('configuration.confirmSaveText'),
          t('configuration.confirmSave'),
          {
            confirmButtonText: t('configuration.yes'),
            cancelButtonText: t('configuration.no'),
            type: 'warning',
          }
        );

        saving.value = true;
        const payload = JSON.parse(JSON.stringify(editableConfig));
        await updateConfig(payload);
        await saveConfig();
        await loadConfig();
        
        editing.value = false;
        ElMessage({
          type: 'success',
          message: t('configuration.saveSuccess'),
        });
      } catch (err) {
        if (err !== 'cancel') {
          ElMessage({
            type: 'error',
            message: t('configuration.saveFailed') + ': ' + (err.message || 'Unknown error'),
          });
          console.error('Error saving configuration:', err);
        }
      } finally {
        saving.value = false;
      }
    };

    const handleReset = async () => {
      try {
        await ElMessageBox.confirm(
          t('configuration.resetConfirmText'),
          t('configuration.resetConfirmTitle'),
          {
            confirmButtonText: t('configuration.resetConfirmYes'),
            cancelButtonText: t('configuration.resetConfirmNo'),
            type: 'danger',
          }
        );

        loading.value = true;
        await resetConfig();
        await loadConfig();
        
        ElMessage({
          type: 'success',
          message: t('configuration.resetSuccess'),
        });
      } catch (err) {
        if (err !== 'cancel') {
          ElMessage.error('Reset failed: ' + (err.message || 'Unknown error'));
        }
      } finally {
        loading.value = false;
      }
    };

    const addExternalItem = () => {
      editableConfig.external.push({
        type: 'monitor',
        key: '',
        name: '',
        id: 0
      });
    };

    const removeExternalItem = (index) => {
      editableConfig.external.splice(index, 1);
    };

    onMounted(() => {
      loadConfig();
    });

    return {
      config,
      activeTab,
      editableConfig,
      loading,
      saving,
      editing,
      error,
      loadConfig,
      startEdit,
      cancelEdit,
      saveChanges,
      handleReset,
      addExternalItem,
      removeExternalItem,
      Delete,
      Refresh,
      Edit,
      Warning,
      Check,
    };
  },
};
</script>

<style scoped>
.config-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  margin-bottom: 32px;
}

.config-loading {
  text-align: center;
  padding: 100px 20px;
  color: var(--text-muted);
}

.config-tabs {
  border: 1px solid var(--border-1);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-md);
  background: var(--surface-1);
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
}

.tab-pane-content {
  padding: 32px;
  max-width: 800px;
}

.section-divider {
  margin: 40px 0 24px;
  font-weight: 700;
  font-size: 18px;
  color: var(--text-strong);
  padding-bottom: 12px;
  border-bottom: 2px solid var(--border-1);
  font-family: var(--font-display);
}

.section-divider:first-child {
  margin-top: 0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.help-text {
  margin-top: 20px;
  font-size: 14px;
  color: var(--text-muted);
  font-style: italic;
  padding: 12px;
  background: var(--surface-2);
  border-radius: var(--radius-sm);
}

:deep(.el-tabs--border-card) {
  background: transparent;
  border: none;
}

:deep(.el-tabs--border-card > .el-tabs__header) {
  background: var(--surface-3);
  border-bottom: 1px solid var(--border-1);
}

:deep(.el-tabs--border-card > .el-tabs__header .el-tabs__item.is-active) {
  background: var(--surface-1);
  border-right-color: var(--border-1);
  border-left-color: var(--border-1);
}

:deep(.el-form-item__label) {
  font-weight: 600;
  color: var(--text-strong);
}
</style>
