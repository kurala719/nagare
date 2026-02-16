<template>
  <div class="configuration-page">
    <div class="config-header">
      <h2>{{ $t('configuration.title') }}</h2>
      <div class="config-actions">
        <el-button @click="loadConfig" :loading="loading" :disabled="editing">
          <el-icon><Refresh /></el-icon>
          {{ $t('configuration.reload') }}
        </el-button>
        <el-button v-if="!editing" type="primary" @click="startEdit">
          <el-icon><Edit /></el-icon>
          {{ $t('configuration.edit') }}
        </el-button>
        <template v-if="editing">
          <el-button @click="cancelEdit">
            {{ $t('configuration.cancel') }}
          </el-button>
          <el-button type="primary" @click="saveChanges" :loading="saving">
            <el-icon><Check /></el-icon>
            {{ $t('configuration.save') }}
          </el-button>
        </template>
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
      <!-- System Settings -->
      <el-card class="config-section">
        <template #header>
          <div class="card-header">
            <el-icon><Setting /></el-icon>
            <span>{{ $t('configuration.system') }}</span>
          </div>
        </template>
        <el-form :model="editableConfig.system" label-width="150px" :disabled="!editing">
          <el-form-item :label="$t('configuration.systemName')">
            <el-input v-model="editableConfig.system.system_name" />
          </el-form-item>
          <el-form-item :label="$t('configuration.ipAddress')">
            <el-input v-model="editableConfig.system.ip_address" />
          </el-form-item>
          <el-form-item :label="$t('configuration.port')">
            <el-input-number v-model="editableConfig.system.port" :min="1" :max="65535" />
          </el-form-item>
          <el-form-item :label="$t('configuration.availability')">
            <el-switch v-model="editableConfig.system.availability" />
          </el-form-item>
        </el-form>
      </el-card>

      <!-- Database Settings -->
      <el-card class="config-section">
        <template #header>
          <div class="card-header">
            <el-icon><DataLine /></el-icon>
            <span>{{ $t('configuration.database') }}</span>
          </div>
        </template>
        <el-form :model="editableConfig.database" label-width="150px" :disabled="!editing">
          <el-form-item :label="$t('configuration.dbVersion')">
            <el-input v-model="editableConfig.database.version" />
          </el-form-item>
          <el-form-item :label="$t('configuration.dbHost')">
            <el-input v-model="editableConfig.database.host" />
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
      </el-card>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh, Edit, Check, Setting, DataLine, Loading } from '@element-plus/icons-vue';
import { getMainConfig, updateConfig, saveConfig } from '@/api/config';

export default {
  name: 'Configuration',
  components: {
    Refresh,
    Edit,
    Check,
    Setting,
    DataLine,
    Loading,
  },
  setup() {
    const { t } = useI18n();
    const config = ref(null);
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
    });
    const loading = ref(false);
    const saving = ref(false);
    const editing = ref(false);
    const error = ref(null);

    const loadConfig = async () => {
      loading.value = true;
      error.value = null;
      try {
        const response = await getMainConfig();
        const data = response.data || response;
        config.value = data;
        
        // Update editable config
        if (data.system) {
          Object.assign(editableConfig.system, data.system);
        }
        if (data.database) {
          Object.assign(editableConfig.database, data.database);
        }
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
      // Reset to original values
      if (config.value) {
        if (config.value.system) {
          Object.assign(editableConfig.system, config.value.system);
        }
        if (config.value.database) {
          Object.assign(editableConfig.database, config.value.database);
        }
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

        // Prepare the update payload
        const payload = {
          system: { ...editableConfig.system },
          database: {
            host: editableConfig.database.host,
            port: editableConfig.database.port,
            username: editableConfig.database.username,
            password: editableConfig.database.password,
          },
        };

        // Update configuration
        await updateConfig(payload);
        
        // Save to disk
        await saveConfig();

        // Reload to get updated values
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

    onMounted(() => {
      loadConfig();
    });

    return {
      config,
      editableConfig,
      loading,
      saving,
      editing,
      error,
      loadConfig,
      startEdit,
      cancelEdit,
      saveChanges,
    };
  },
};
</script>

<style scoped>
.configuration-page {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.config-header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.config-actions {
  display: flex;
  gap: 10px;
}

.config-loading {
  text-align: center;
  padding: 60px 20px;
}

.config-loading p {
  margin-top: 16px;
  color: #909399;
  font-size: 14px;
}

.config-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.config-section {
  margin-bottom: 20px;
}

.config-section:last-child {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.el-form {
  margin-top: 20px;
}

.el-form-item {
  margin-bottom: 18px;
}
</style>
