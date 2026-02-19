<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('kb.title') }}</h1>
      <p class="page-subtitle">{{ $t('kb.subtitle') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="searchQuery" :placeholder="$t('kb.searchPlaceholder')" clearable style="width: 300px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>

      <div class="action-group">
        <el-dropdown v-if="selectedIds.length > 0" class="batch-actions">
          <el-button type="warning">
            {{ $t('common.selectedCount', { count: selectedIds.length }) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="Delete" @click="handleBulkDelete" style="color: var(--el-color-danger)">
                {{ $t('common.bulkDelete') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <el-button v-if="items.length > 0" @click="toggleSelectAll">
          {{ isAllSelected ? $t('common.deselectAll') : $t('common.selectAll') }}
        </el-button>

        <el-button type="primary" :icon="Plus" @click="openCreateDialog">
          {{ $t('kb.create') }}
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" size="40"><Loading /></el-icon>
      <p>{{ $t('common.loading') }}</p>
    </div>

    <el-empty v-else-if="items.length === 0" :description="$t('kb.noItems')" />

    <div v-else class="kb-grid">
      <el-card v-for="item in items" :key="item.ID" 
        class="kb-card" 
        :class="{ 'is-selected': isSelected(item.ID) }"
        @click="toggleSelection(item.ID)"
      >
        <template #header>
          <div class="card-header">
            <div class="header-left">
              <el-checkbox 
                :model-value="isSelected(item.ID)" 
                @change="() => toggleSelection(item.ID)" 
                @click.stop
              />
              <span class="topic">{{ item.Topic }}</span>
            </div>
            <el-tag size="small">{{ item.Category }}</el-tag>
          </div>
        </template>
        <div class="content">{{ item.Content }}</div>
        <div class="keywords">
          <el-tag v-for="kw in splitKeywords(item.Keywords)" :key="kw" size="small" type="info" class="kw-tag">
            {{ kw }}
          </el-tag>
        </div>
        <div class="actions">
          <el-button size="small" :icon="Edit" @click="openEditDialog(item)">{{ $t('common.edit') }}</el-button>
          <el-button size="small" type="danger" :icon="Delete" @click="handleDelete(item)">{{ $t('common.delete') }}</el-button>
        </div>
      </el-card>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? $t('kb.editTitle') : $t('kb.createTitle')" width="600px">
      <el-form :model="form" label-width="100px" :rules="rules" ref="formRef">
        <el-form-item :label="$t('kb.topic')" prop="topic">
          <el-input v-model="form.topic" />
        </el-form-item>
        <el-form-item :label="$t('kb.category')" prop="category">
          <el-select v-model="form.category" style="width: 100%">
            <el-option label="Network" value="Network" />
            <el-option label="Database" value="Database" />
            <el-option label="Application" value="Application" />
            <el-option label="System" value="System" />
            <el-option label="Other" value="Other" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('kb.keywords')" prop="keywords">
          <el-input v-model="form.keywords" placeholder="Comma separated: OSPF, MTU, Error 1002" />
        </el-form-item>
        <el-form-item :label="$t('kb.content')" prop="content">
          <el-input v-model="form.content" type="textarea" :rows="6" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { Search, Plus, Edit, Delete, Loading, ArrowDown } from '@element-plus/icons-vue'
import { fetchKnowledgeBase, addKnowledgeBase, updateKnowledgeBase, deleteKnowledgeBase, bulkDeleteKnowledgeBase } from '@/api/knowledgeBase'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const items = ref([])
const searchQuery = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const selectedIds = ref([])

const isAllSelected = computed(() => {
  return items.value.length > 0 && selectedIds.value.length === items.value.length
})

const isSelected = (id) => selectedIds.value.includes(id)

const toggleSelection = (id) => {
  const index = selectedIds.value.indexOf(id)
  if (index > -1) {
    selectedIds.value.splice(index, 1)
  } else {
    selectedIds.value.push(id)
  }
}

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = items.value.map(item => item.ID)
  }
}

const form = ref({
  id: null,
  topic: '',
  content: '',
  keywords: '',
  category: 'Other'
})

const rules = {
  topic: [{ required: true, message: t('kb.topicRequired'), trigger: 'blur' }],
  content: [{ required: true, message: t('kb.contentRequired'), trigger: 'blur' }],
  category: [{ required: true, message: t('kb.categoryRequired'), trigger: 'change' }]
}

const loadData = async () => {
  loading.value = true
  selectedIds.value = []
  try {
    const response = await fetchKnowledgeBase({ q: searchQuery.value })
    if (response.data && response.data.success) {
      items.value = response.data.data || []
    } else {
      ElMessage.error(t('kb.loadFailed'))
    }
  } catch (err) {
    ElMessage.error(t('kb.loadFailed'))
  } finally {
    loading.value = false
  }
}

const splitKeywords = (kw) => {
  if (!kw) return []
  return kw.split(',').map(s => s.trim()).filter(s => s !== '')
}

const openCreateDialog = () => {
  isEdit.value = false
  form.value = { id: null, topic: '', content: '', keywords: '', category: 'Other' }
  dialogVisible.value = true
}

const openEditDialog = (item) => {
  isEdit.value = true
  form.value = {
    id: item.ID,
    topic: item.Topic,
    content: item.Content,
    keywords: item.Keywords,
    category: item.Category
  }
  dialogVisible.value = true
}

const handleDelete = (item) => {
  ElMessageBox.confirm(t('kb.deleteConfirm'), t('common.warning'), {
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteKnowledgeBase(item.ID)
      if (res.data && res.data.success) {
        ElMessage.success(t('kb.deleteSuccess'))
        loadData()
      } else {
        ElMessage.error(t('kb.deleteFailed'))
      }
    } catch (err) {
      ElMessage.error(t('kb.deleteFailed'))
    }
  }).catch(() => {})
}

const handleBulkDelete = () => {
  if (selectedIds.value.length === 0) return
  
  ElMessageBox.confirm(
    t('common.bulkDeleteConfirmText', { count: selectedIds.value.length }),
    t('common.bulkDeleteConfirmTitle'),
    { type: 'warning' }
  ).then(async () => {
    try {
      await bulkDeleteKnowledgeBase(selectedIds.value)
      ElMessage.success(t('common.bulkDeleteSuccess', { count: selectedIds.value.length }))
      loadData()
    } catch (err) {
      ElMessage.error(t('common.bulkDeleteFailed'))
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
          topic: form.value.topic,
          content: form.value.content,
          keywords: form.value.keywords,
          category: form.value.category
        }
        let res
        if (isEdit.value) {
          res = await updateKnowledgeBase(form.value.id, payload)
        } else {
          res = await addKnowledgeBase(payload)
        }
        
        if (res.data && res.data.success) {
          ElMessage.success(isEdit.value ? t('kb.updateSuccess') : t('kb.createSuccess'))
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(t('kb.saveFailed'))
        }
      } catch (err) {
        ElMessage.error(t('kb.saveFailed'))
      } finally {
        submitting.value = false
      }
    }
  })
}

watch(searchQuery, () => {
  loadData()
})

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px;
  color: #909399;
}

.kb-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
  padding: 20px;
}

.kb-card {
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.kb-card {
  display: flex;
  flex-direction: column;
  transition: all 0.3s;
  cursor: pointer;
  border: 1px solid transparent;
}

.kb-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.kb-card.is-selected {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.topic {
  font-weight: bold;
  font-size: 1.1em;
}

.content {
  margin-bottom: 15px;
  color: #606266;
  white-space: pre-wrap;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.keywords {
  margin-bottom: 15px;
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: auto;
}
</style>
