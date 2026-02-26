<template>
  <div class="nagare-container">
    <div class="page-header">
      <div class="header-main">
        <h1 class="page-title">{{ $t('packets.title') }}</h1>
        <p class="page-subtitle">{{ $t('packets.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" :icon="Plus" @click="openUploadDialog">
          {{ $t('packets.upload') }}
        </el-button>
        <el-button @click="loadData" :loading="loading" :icon="Refresh" circle />
      </div>
    </div>

    <div class="packets-list animate-slide-up">
      <el-empty v-if="items.length === 0" :description="$t('packets.noItems')" />
      
      <el-table v-else :data="items" stripe style="width: 100%" v-loading="loading">
        <el-table-column prop="ID" label="ID" width="80" />
        <el-table-column prop="name" :label="$t('packets.name')" min-width="150" />
        
        <el-table-column prop="status" :label="$t('packets.status')" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="risk_level" :label="$t('packets.risk')" width="120">
          <template #default="{ row }">
            <el-tag :type="getRiskType(row.risk_level)" effect="dark" v-if="row.status === 2">
              {{ $t('packets.' + row.risk_level) || row.risk_level }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>

        <el-table-column prop="CreatedAt" label="Time" width="180">
          <template #default="{ row }">
            {{ new Date(row.CreatedAt).toLocaleString() }}
          </template>
        </el-table-column>

        <el-table-column :label="$t('packets.actions')" width="200" align="right">
          <template #default="{ row }">
            <el-button link type="primary" :icon="Search" @click="viewAnalysis(row)" v-if="row.status === 2">
              {{ $t('packets.result') }}
            </el-button>
            <el-button link type="warning" :icon="VideoPlay" @click="handleAnalyze(row)" v-if="row.status !== 1">
              {{ $t('packets.reAnalyze') }}
            </el-button>
            <el-button link type="danger" :icon="Delete" @click="handleDelete(row)">
              {{ $t('packets.delete') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Upload Dialog -->
    <el-dialog v-model="uploadDialogVisible" :title="$t('packets.uploadTitle')" width="600px">
      <el-form :model="uploadForm" label-width="120px" ref="uploadFormRef" :rules="uploadRules">
        <el-form-item :label="$t('packets.name')" prop="name">
          <el-input v-model="uploadForm.name" placeholder="e.g. Suspicious TCP Flow" />
        </el-form-item>
        
        <el-form-item :label="$t('packets.provider')" prop="provider_id">
          <el-select v-model="uploadForm.provider_id" @change="onProviderChange" style="width: 100%">
            <el-option v-for="p in aiProviders" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('packets.model')" prop="model">
          <el-select v-model="uploadForm.model" style="width: 100%" filterable allow-create>
            <el-option v-for="m in availableModels" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>

        <el-form-item :label="$t('packets.file')">
          <el-upload
            class="packet-uploader"
            drag
            action="#"
            :auto-upload="false"
            :on-change="handleFileChange"
            :limit="1"
            :file-list="fileList"
          >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
              Drop file here or <em>click to upload</em>
            </div>
          </el-upload>
        </el-form-item>

        <el-form-item :label="$t('packets.rawContent')">
          <el-input v-model="uploadForm.raw_content" type="textarea" :rows="4" placeholder="Paste hex or flow data here..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitUpload" :loading="uploading">{{ $t('packets.analyze') }}</el-button>
      </template>
    </el-dialog>

    <!-- Result Dialog -->
    <el-dialog v-model="resultDialogVisible" :title="$t('packets.summary')" width="800px" custom-class="analysis-dialog">
      <div v-if="selectedItem" class="analysis-container">
        <div class="analysis-header-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item :label="$t('packets.name')">{{ selectedItem.name }}</el-descriptions-item>
            <el-descriptions-item :label="$t('packets.risk')">
              <el-tag :type="getRiskType(selectedItem.risk_level)" effect="dark">
                {{ $t('packets.' + selectedItem.risk_level) || selectedItem.risk_level }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="Provider">{{ selectedItem.provider_name }} ({{ selectedItem.model }})</el-descriptions-item>
            <el-descriptions-item label="Source File">{{ selectedItem.file_path || 'Manual Content' }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <div v-if="selectedItem.raw_content" class="raw-data-section">
          <h4>Raw Data / Flow</h4>
          <div class="raw-box">
            <code>{{ selectedItem.raw_content }}</code>
          </div>
        </div>

        <el-divider content-position="left">AI Intelligence Analysis</el-divider>
        <div class="ai-analysis-content markdown-body">
          <p style="white-space: pre-wrap;">{{ selectedItem.analysis }}</p>
        </div>
      </div>
      <template #footer>
        <el-button @click="resultDialogVisible = false">{{ $t('packets.close') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { Plus, Refresh, Search, VideoPlay, Delete, UploadFilled, Loading } from '@element-plus/icons-vue'
import { fetchPacketAnalyses, uploadPacket, deletePacketAnalysis, startPacketAnalysis } from '@/api/packetAnalysis'
import { fetchProviderData } from '@/api/providers'
import { getMainConfig } from '@/api/config'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const uploading = ref(false)
const items = ref([])
const aiProviders = ref([])
const availableModels = ref([])
const uploadDialogVisible = ref(false)
const resultDialogVisible = ref(false)
const selectedItem = ref(null)
const fileList = ref([])

const uploadForm = reactive({
  name: '',
  provider_id: null,
  model: '',
  raw_content: '',
  file: null
})

const uploadRules = {
  name: [{ required: true, message: 'Please enter a name', trigger: 'blur' }],
  provider_id: [{ required: true, message: 'Please select an AI provider', trigger: 'change' }]
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await fetchPacketAnalyses()
    if (res && res.success) {
      items.value = res.data || []
    }
  } catch (err) {
    ElMessage.error('Failed to load data')
  } finally {
    loading.value = false
  }
}

const loadProviders = async () => {
  try {
    const configRes = await getMainConfig()
    const config = configRes.data?.data || configRes.data || configRes
    const defProviderId = config.ai?.provider_id || 1
    const defModel = config.ai?.model || ''

    const res = await fetchProviderData({ enabled: 1 })
    const list = res.data?.items || res.items || res.data || []
    aiProviders.value = list.map(p => ({
      id: p.ID || p.id,
      name: p.name || p.Name,
      models: p.models || p.Models || []
    }))

    uploadForm.provider_id = defProviderId
    uploadForm.model = defModel
    onProviderChange(defProviderId)
  } catch (err) {
    console.error('Failed to load AI providers', err)
  }
}

const onProviderChange = (val) => {
  const p = aiProviders.value.find(prov => prov.id === val)
  if (p) {
    availableModels.value = p.models || []
    if (availableModels.value.length > 0 && !availableModels.value.includes(uploadForm.model)) {
      uploadForm.model = availableModels.value[0]
    }
  }
}

const handleFileChange = (file) => {
  uploadForm.file = file.raw
}

const openUploadDialog = () => {
  uploadForm.name = ''
  uploadForm.raw_content = ''
  uploadForm.file = null
  fileList.value = []
  uploadDialogVisible.value = true
}

const submitUpload = async () => {
  if (!uploadForm.name) return ElMessage.warning('Name is required')
  
  const formData = new FormData()
  formData.append('name', uploadForm.name)
  formData.append('provider_id', uploadForm.provider_id)
  formData.append('model', uploadForm.model)
  formData.append('raw_content', uploadForm.raw_content)
  if (uploadForm.file) {
    formData.append('file', uploadForm.file)
  }

  uploading.value = true
  try {
    const res = await uploadPacket(formData)
    if (res && res.success) {
      ElMessage.success(t('packets.success'))
      uploadDialogVisible.value = false
      loadData()
    }
  } catch (err) {
    ElMessage.error('Upload failed')
  } finally {
    uploading.value = false
  }
}

const handleAnalyze = async (row) => {
  try {
    const res = await startPacketAnalysis(row.ID)
    if (res && res.success) {
      ElMessage.success('Analysis started')
      loadData()
    }
  } catch (err) {
    ElMessage.error('Failed to start analysis')
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(t('common.deleteConfirm') || 'Delete this record?', 'Warning', {
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deletePacketAnalysis(row.ID)
      if (res && res.success) {
        ElMessage.success(t('packets.deleted'))
        loadData()
      }
    } catch (err) {
      ElMessage.error('Delete failed')
    }
  })
}

const viewAnalysis = (row) => {
  selectedItem.value = row
  const prov = aiProviders.value.find(p => p.id === row.provider_id)
  selectedItem.value.provider_name = prov ? prov.name : 'Unknown'
  resultDialogVisible.value = true
}

const getStatusType = (status) => {
  const map = { 0: 'info', 1: 'warning', 2: 'success', 3: 'danger' }
  return map[status] || 'info'
}

const getStatusLabel = (status) => {
  const map = { 
    0: t('packets.pending'), 
    1: t('packets.analyzing'), 
    2: t('packets.completed'), 
    3: t('packets.failed') 
  }
  return map[status] || 'Unknown'
}

const getRiskType = (risk) => {
  const map = { 'clean': 'success', 'notable': 'warning', 'malicious': 'danger' }
  return map[risk] || 'info'
}

onMounted(() => {
  loadData()
  loadProviders()
})
</script>

<style scoped>
.packets-list {
  padding: 20px;
  background: var(--surface-1);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-1);
}

.analysis-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.raw-box {
  background: #1e293b;
  color: #e2e8f0;
  padding: 16px;
  border-radius: 8px;
  max-height: 200px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 13px;
  word-break: break-all;
  white-space: pre-wrap;
}

.ai-analysis-content {
  background: var(--surface-2);
  padding: 24px;
  border-radius: 8px;
  line-height: 1.8;
  font-size: 15px;
}

.markdown-body h1, .markdown-body h2, .markdown-body h3 {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
}

.packet-uploader {
  width: 100%;
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
