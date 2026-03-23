<template>
  <div class="mcp-servers-container">
    <el-card class="box-card">
      <template #header>
        <div class="card-header">
          <span class="header-title">{{ t('mcp.title') || 'MCP Servers' }}</span>
          <div class="header-actions">
            <el-button type="info" plain @click="fetchData">
              <el-icon><Refresh /></el-icon> {{ t('common.refresh') }}
            </el-button>
            <el-button type="primary" @click="handleAdd">
              <el-icon><Plus /></el-icon> {{ t('common.add') }}
            </el-button>
            <el-button type="success" :loading="saving" @click="handleSave">
              <el-icon><Check /></el-icon> {{ t('common.save') }}
            </el-button>
          </div>
        </div>
      </template>

      <el-alert
        :title="t('mcp.alertInfo')"
        type="info"
        show-icon
        style="margin-bottom: 20px;"
      />

      <el-table
        v-loading="loading"
        :data="servers"
        style="width: 100%"
        border
        stripe
      >
        <el-table-column prop="name" :label="t('mcp.serviceName')" width="180">
          <template #default="scope">
            <strong>{{ scope.row.name }}</strong>
          </template>
        </el-table-column>
        
        <el-table-column prop="command" :label="t('mcp.command')">
          <template #default="scope">
            <span>{{ scope.row.command }} {{ scope.row.args.join(' ') }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" :label="t('common.status')" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'">
              {{ scope.row.enabled ? t('common.enabled') : t('common.disabled') }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column :label="t('common.actions')" width="250" fixed="right">
          <template #default="scope">
            <el-button size="small" type="primary" plain @click="handleEdit(scope.$index, scope.row)">
              {{ t('common.edit') }}
            </el-button>
            <el-button size="small" type="warning" plain :loading="testing === scope.$index" @click="handleTest(scope.$index, scope.row)">
              {{ t('common.test') }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(scope.$index)">
              {{ t('common.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog :title="dialogTitle" v-model="dialogVisible" width="600px" destroy-on-close>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" class="demo-ruleForm">
        <el-form-item :label="t('mcp.name')" prop="name">
          <el-input v-model="form.name" :placeholder="t('mcp.namePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('mcp.command')" prop="command">
          <el-input v-model="form.command" :placeholder="t('mcp.commandPlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('mcp.args')">
          <el-select
            v-model="form.args"
            multiple
            filterable
            allow-create
            default-first-option
            :placeholder="t('mcp.argsPlaceholder')"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item :label="t('mcp.envVars')">
          <div>
            <div v-for="(item, index) in form.envList" :key="index" style="display: flex; gap: 8px; margin-bottom: 8px;">
              <el-input v-model="item.key" :placeholder="t('mcp.envKey')" style="width: 40%" />
              <el-input v-model="item.value" :placeholder="t('mcp.envValue')" style="width: 50%" />
              <el-button type="danger" plain circle @click="form.envList.splice(index, 1)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button type="info" plain @click="form.envList.push({ key: '', value: '' })" size="small">
              <el-icon><Plus /></el-icon> {{ t('mcp.addEnv') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="t('common.enabled')">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" @click="submitForm">{{ t('common.confirm') }}</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { defineComponent, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Plus, Refresh, Check, Delete } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMCPServers, saveMCPServers, testMCPServer } from '@/api/mcp'

export default defineComponent({
  name: 'MCPServer',
  components: {
    Plus,
    Refresh,
    Check,
    Delete
  },
  setup() {
    const { t } = useI18n()
    const loading = ref(false)
    const saving = ref(false)
    const testing = ref(-1)
    const servers = ref([])

    const dialogVisible = ref(false)
    const dialogTitle = ref('')
    const formRef = ref(null)
    const editingIndex = ref(-1)

    const form = ref({
      name: '',
      command: '',
      args: [],
      enabled: true,
      env: {},
      envList: []
    })

    const rules = {
      name: [{ required: true, message: 'Please input server name', trigger: 'blur' }],
      command: [{ required: true, message: 'Please input execution command', trigger: 'blur' }]
    }

    const fetchData = async () => {
      loading.value = true
      try {
        const response = await getMCPServers()
        // API returns a generic response format
        let payload = response.data?.data || response.data || []
        servers.value = Array.isArray(payload) ? payload : []
      } catch (error) {
        ElMessage.error(error.message || 'Failed to fetch MCP servers')
      } finally {
        loading.value = false
      }
    }

    const handleAdd = () => {
      dialogTitle.value = t('mcp.addTitle')
      form.value = {
        name: '',
        command: '',
        args: [],
        enabled: true,
        env: {},
        envList: []
      }
      editingIndex.value = -1
      dialogVisible.value = true
    }

    const handleEdit = (index, row) => {
      dialogTitle.value = t('mcp.editTitle')
      form.value = JSON.parse(JSON.stringify(row))
      if (!Array.isArray(form.value.args)) form.value.args = []
      
      form.value.envList = []
      if (form.value.env) {
        for (const [key, value] of Object.entries(form.value.env)) {
          form.value.envList.push({ key, value })
        }
      }
      editingIndex.value = index
      dialogVisible.value = true
    }

    const handleDelete = (index) => {
      ElMessageBox.confirm(
        t('mcp.deleteConfirmText'),
        t('mcp.deleteTitle'),
        { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
      ).then(() => {
        servers.value.splice(index, 1)
        ElMessage.success(t('mcp.removedLocally'))
      }).catch(() => {})
    }

    const handleTest = async (index, row) => {
      testing.value = index
      try {
        const payload = JSON.parse(JSON.stringify(row))
        const envMap = {}
        if (payload.envList) {
          for (const item of payload.envList) {
            if (item.key && item.key.trim() !== '') {
              envMap[item.key.trim()] = item.value
            }
          }
          payload.env = envMap
          delete payload.envList
        }
        const res = await testMCPServer(payload)
        const data = res.data?.data || res.data
        if (data.ok || data.result?.connected) {
          ElMessage.success(t('mcp.testPassed', { count: data.result?.tool_count || 0 }))
        } else {
          ElMessage.error(t('mcp.testFailed', { message: data.message || data.result?.error || 'Unknown error' }))
        }
      } catch (error) {
        ElMessage.error(error.message || 'Test connection error')
      } finally {
        testing.value = -1
      }
    }

    const submitForm = () => {
      formRef.value.validate((valid) => {
        if (valid) {
          const envMap = {}
          if (form.value.envList) {
            for (const item of form.value.envList) {
              if (item.key && item.key.trim() !== '') {
                envMap[item.key.trim()] = item.value
              }
            }
          }
          form.value.env = envMap
          delete form.value.envList

          if (editingIndex.value >= 0) {
            servers.value.splice(editingIndex.value, 1, JSON.parse(JSON.stringify(form.value)))
          } else {
            servers.value.push(JSON.parse(JSON.stringify(form.value)))
          }
          dialogVisible.value = false
          ElMessage.success(t('mcp.savedToList'))
        }
      })
    }

    const handleSave = async () => {
      saving.value = true
      try {
        await saveMCPServers(servers.value)
        ElMessage.success(t('mcp.saveSuccess'))
      } catch (error) {
        ElMessage.error(error.message || 'Failed to save configuration')
      } finally {
        saving.value = false
      }
    }

    onMounted(() => {
      fetchData()
    })

    return {
      t,
      loading,
      saving,
      testing,
      servers,
      dialogVisible,
      dialogTitle,
      formRef,
      form,
      rules,
      fetchData,
      handleAdd,
      handleEdit,
      handleDelete,
      handleTest,
      submitForm,
      handleSave
    }
  }
})
</script>

<style scoped>
.mcp-servers-container {
  padding: 24px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-title {
  font-size: 18px;
  font-weight: bold;
}
.header-actions {
  display: flex;
  gap: 12px;
}
</style>
