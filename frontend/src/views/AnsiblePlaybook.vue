<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('ansible.playbooksTitle') }}</h1>
      <p class="page-subtitle">{{ $t('ansible.playbooksSubtitle') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="searchQuery" :placeholder="$t('ansible.searchPlaceholder')" clearable style="width: 300px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>

      <div class="action-group">
        <el-button @click="openJobsHistory" :icon="Histogram">
          {{ $t('ansible.viewJobs') }}
        </el-button>
        <el-button type="primary" :icon="Plus" @click="openCreateDialog">
          {{ $t('ansible.createPlaybook') }}
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" size="40"><Loading /></el-icon>
      <p>{{ $t('common.loading') }}</p>
    </div>

    <el-empty v-else-if="playbooks.length === 0" :description="$t('ansible.noPlaybooks')" />

    <div v-else class="playbook-grid">
      <el-card v-for="pb in playbooks" :key="pb.id" class="playbook-card">
        <template #header>
          <div class="card-header">
            <span class="pb-name">{{ pb.name }}</span>
            <el-tag size="small" v-if="pb.tags">{{ pb.tags }}</el-tag>
          </div>
        </template>
        <div class="pb-desc">{{ pb.description || 'No description' }}</div>
        <div class="pb-actions">
          <el-button size="small" type="success" :icon="CaretRight" @click="handleRun(pb)">{{ $t('ansible.run') }}</el-button>
          <el-button size="small" :icon="Edit" @click="openEditDialog(pb)">{{ $t('common.edit') }}</el-button>
          <el-button size="small" type="danger" :icon="Delete" @click="handleDelete(pb)">{{ $t('common.delete') }}</el-button>
        </div>
      </el-card>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? $t('ansible.editPlaybook') : $t('ansible.createPlaybook')" width="80%">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item :label="$t('ansible.pbName')" prop="name">
              <el-input v-model="form.name" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="$t('ansible.pbTags')" prop="tags">
              <el-input v-model="form.tags" placeholder="e.g. web, db, maintenance" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="$t('ansible.pbDesc')" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
        
        <div class="editor-toolbar">
          <span>{{ $t('ansible.pbContent') }} (YAML)</span>
          <el-button size="small" type="primary" link :icon="MagicStick" @click="openAiDialog">
            {{ $t('ansible.aiGenerate') }}
          </el-button>
        </div>
        
        <div class="code-editor-container">
          <el-input 
            v-model="form.content" 
            type="textarea" 
            :rows="15" 
            class="yaml-editor"
            placeholder="---
- hosts: all
  tasks:
    - name: Hello World
      debug:
        msg: 'Hello from Nagare'"
          />
        </div>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- Run Dialog -->
    <el-dialog v-model="runDialogVisible" :title="$t('ansible.runPlaybook')" width="400px">
      <el-form :model="runForm" label-position="top">
        <el-form-item :label="$t('ansible.hostFilter')">
          <el-input v-model="runForm.hostFilter" placeholder="all, web_servers, 192.168.1.10" />
          <p class="help-text">{{ $t('ansible.hostFilterHelp') }}</p>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="runDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="confirmRun" :loading="running">{{ $t('ansible.startJob') }}</el-button>
      </template>
    </el-dialog>

    <!-- AI Generator Dialog -->
    <el-dialog v-model="aiDialogVisible" :title="$t('ansible.aiGenerateTitle')" width="500px">
      <el-form :model="aiForm" label-position="top">
        <el-form-item :label="$t('ansible.aiRequirement')">
          <el-input 
            v-model="aiForm.requirement" 
            type="textarea" 
            :rows="4" 
            placeholder="e.g. Restart nginx service on all web servers" 
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="aiDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="generateAiPlaybook" :loading="generating">
          {{ $t('ansible.generate') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Search, Plus, Edit, Delete, Loading, CaretRight, Histogram, MagicStick } from '@element-plus/icons-vue'
import { fetchPlaybooks, createPlaybook, updatePlaybook, deletePlaybook, runPlaybook, recommendPlaybook } from '@/api/ansible'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const loading = ref(false)
const playbooks = ref([])
const searchQuery = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)

const runDialogVisible = ref(false)
const running = ref(false)
const activePlaybook = ref(null)
const runForm = ref({
  hostFilter: 'all'
})

const aiDialogVisible = ref(false)
const generating = ref(false)
const aiForm = ref({
  requirement: ''
})

const form = ref({
  id: null,
  name: '',
  description: '',
  content: '',
  tags: ''
})

const rules = {
  name: [{ required: true, message: 'Name is required', trigger: 'blur' }],
  content: [{ required: true, message: 'Content is required', trigger: 'blur' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await fetchPlaybooks({ q: searchQuery.value })
    if (res && res.success) {
      const data = res.data
      if (data.items && Array.isArray(data.items)) {
        playbooks.value = data.items
      } else if (Array.isArray(data)) {
        playbooks.value = data
      } else {
        playbooks.value = []
      }
    }
  } catch (err) {
    ElMessage.error(t('ansible.loadFailed'))
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  form.value = { id: null, name: '', description: '', content: '', tags: '' }
  dialogVisible.value = true
}

const openEditDialog = (pb) => {
  isEdit.value = true
  form.value = { ...pb }
  dialogVisible.value = true
}

const handleDelete = (pb) => {
  ElMessageBox.confirm(t('ansible.deleteConfirm'), t('common.warning'), {
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deletePlaybook(pb.id)
      if (res && res.success) {
        ElMessage.success(t('ansible.deleteSuccess'))
        loadData()
      }
    } catch (err) {
      ElMessage.error(t('ansible.deleteFailed'))
    }
  }).catch(() => {})
}

const submitForm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const payload = {
          name: form.value.name,
          description: String(form.value.description || ''),
          content: form.value.content,
          tags: String(form.value.tags || '')
        }
        
        
        let res
        if (isEdit.value) {
          res = await updatePlaybook(form.value.id, payload)
        } else {
          res = await createPlaybook(payload)
        }
        
        if (res && res.success) {
          ElMessage.success(isEdit.value ? t('ansible.updateSuccess') : t('ansible.createSuccess'))
          dialogVisible.value = false
          loadData()
        }
      } catch (err) {
        console.error('Submit playbook error details:', err.response?.data)
        const errorMsg = err.response?.data?.error || err.response?.data?.message || t('common.operationFailed')
        ElMessage.error(errorMsg)
      } finally {
        submitting.value = false
      }
    }
  })
}

const handleRun = (pb) => {
  activePlaybook.value = pb
  runForm.value.hostFilter = 'all'
  runDialogVisible.value = true
}

const confirmRun = async () => {
  running.value = true
  try {
    const res = await runPlaybook(activePlaybook.value.id, { host_filter: runForm.value.hostFilter })
    if (res && res.success) {
      const jobId = res.data.job_id
      runDialogVisible.value = false
      router.push(`/ansible/jobs/${jobId}`)
    }
  } catch (err) {
    ElMessage.error(t('ansible.runFailed'))
  } finally {
    running.value = false
  }
}

const openJobsHistory = () => {
  router.push('/ansible/jobs')
}

const openAiDialog = () => {
  aiDialogVisible.value = true
}

const generateAiPlaybook = async () => {
  if (!aiForm.value.requirement) return
  generating.value = true
  try {
    const res = await recommendPlaybook({ context: aiForm.value.requirement })
    if (res && res.success) {
      form.value.content = res.data.content
      aiDialogVisible.value = false
      ElMessage.success(t('ansible.aiSuccess'))
    }
  } catch (err) {
    ElMessage.error(t('ansible.aiFailed'))
  } finally {
    generating.value = false
  }
}

watch(searchQuery, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.playbook-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  padding: 20px 0;
}

.playbook-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pb-name {
  font-weight: bold;
  font-size: 1.1em;
}

.pb-desc {
  color: var(--text-secondary);
  font-size: 0.9em;
  margin-bottom: 20px;
  flex: 1;
}

.pb-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--text-strong);
}

.code-editor-container {
  border: 1px solid var(--border-color);
  border-radius: 4px;
  overflow: hidden;
}

.yaml-editor :deep(.el-textarea__inner) {
  font-family: 'Fira Code', 'Courier New', monospace;
  background-color: #1e1e1e;
  color: #d4d4d4;
  font-size: 14px;
  line-height: 1.5;
}

.help-text {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 4px;
}
</style>
