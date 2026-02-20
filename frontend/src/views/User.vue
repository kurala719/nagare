<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('users.title') }}</h1>
      <p class="page-subtitle">{{ totalUsers }} {{ $t('users.roleUser') }}</p>
    </div>

    <div class="standard-toolbar">
      <div class="filter-group">
        <el-input v-model="search" :placeholder="$t('users.search')" clearable style="width: 240px">
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
      </div>

      <div class="action-group">
        <el-button-group style="margin-right: 8px">
          <el-button @click="selectAll">{{ $t('common.selectAll') || 'Select All' }}</el-button>
          <el-button @click="clearSelection">{{ $t('common.deselectAll') || 'Deselect All' }}</el-button>
        </el-button-group>
        <el-button v-if="isSuperAdmin" type="primary" :icon="Plus" @click="openCreate">
          {{ $t('users.create') }}
        </el-button>
        <el-dropdown trigger="click" v-if="isSuperAdmin && selectedCount > 0" style="margin-left: 8px">
          <el-button>
            {{ $t('common.selectedCount', { count: selectedCount }) }}<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :icon="Edit" @click="openBulkUpdateDialog">{{ $t('common.bulkUpdate') }}</el-dropdown-item>
              <el-dropdown-item :icon="Delete" @click="openBulkDeleteDialog" style="color: var(--el-color-danger)">{{ $t('common.bulkDelete') }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <el-icon class="is-loading" size="50" color="#409EFF"><Loading /></el-icon>
      <p>{{ $t('users.loading') }}</p>
    </div>

    <el-alert
      v-if="error && !loading"
      :title="error"
      type="error"
      show-icon
      :closable="false"
      class="users-error"
    >
      <template #default>
        <el-button size="small" @click="loadUsers">{{ $t('users.retry') }}</el-button>
      </template>
    </el-alert>

    <el-empty
      v-if="!loading && !error && filteredUsers.length === 0"
      :description="$t('users.noUsers')"
      class="users-empty"
    />

  <div v-if="!loading && !error" class="users-content">
    <el-table
      v-if="filteredUsers.length > 0"
      :data="filteredUsers"
      border
      ref="usersTableRef"
      row-key="id"
      @selection-change="onSelectionChange"
      @sort-change="onSortChange"
    >
      <el-table-column v-if="isSuperAdmin" type="selection" width="50" align="center" />
      <el-table-column prop="id" :label="$t('users.id')" width="90" align="center" sortable="custom" />
      <el-table-column prop="username" :label="$t('users.username')" min-width="140" sortable="custom" />
      <el-table-column prop="nickname" :label="$t('users.nickname')" min-width="140" show-overflow-tooltip sortable="custom" />
      <el-table-column prop="email" :label="$t('users.email')" min-width="180" show-overflow-tooltip sortable="custom" />
      <el-table-column prop="phone" :label="$t('users.phone')" width="140" align="center" show-overflow-tooltip sortable="custom" />
      <el-table-column prop="privileges" :label="$t('users.role')" width="140" align="center" sortable="custom">
        <template #default="{ row }">
          <el-tag :type="roleTagType(row.privileges)" size="small" effect="dark">{{ roleLabel(row.privileges) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" :label="$t('users.status')" width="120" align="center" sortable="custom">
        <template #default="{ row }">
          <el-tooltip :content="statusReason(row.status)" placement="top">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
              {{ row.status === 1 ? $t('users.active') : $t('users.inactive') }}
            </el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column v-if="isSuperAdmin" :label="$t('users.actions')" width="200" fixed="right" align="center">
        <template #default="{ row }">
          <el-button-group>
            <el-button size="small" :icon="Edit" @click="openEdit(row)">{{ $t('users.edit') }}</el-button>
            <el-button size="small" type="danger" :icon="Delete" @click="confirmDelete(row)">{{ $t('users.delete') }}</el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>
  </div>
    <div v-if="!loading && !error && totalUsers > 0" class="users-pagination">
      <el-pagination
        background
        layout="sizes, prev, pager, next"
        :page-sizes="[10, 20, 50, 100]"
        v-model:page-size="pageSize"
        v-model:current-page="currentPage"
        :total="totalUsers"
      />
    </div>

    <el-dialog v-model="dialogVisible" :title="isEditing ? $t('users.editTitle') : $t('users.createTitle')" width="600px">
      <el-form :model="form" label-width="120px">
        <el-form-item :label="$t('users.username')">
          <el-input v-model="form.username" :disabled="isEditing" />
        </el-form-item>
        <el-form-item v-if="!isEditing" :label="$t('users.password')">
          <el-input v-model="form.password" type="password" show-password :placeholder="$t('users.passwordHint')" />
        </el-form-item>
        <el-form-item :label="$t('users.nickname')">
          <el-input v-model="form.nickname" />
        </el-form-item>
        <el-form-item :label="$t('users.email')">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item :label="$t('users.phone')">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item :label="$t('users.role')">
          <el-select v-model="form.privileges" style="width: 100%">
            <el-option v-for="option in privilegeOptions" :key="option.value" :label="option.label" :value="option.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('users.status')">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item :label="$t('users.avatar')">
          <el-input v-model="form.avatar" />
        </el-form-item>
        <el-form-item :label="$t('users.address')">
          <el-input v-model="form.address" />
        </el-form-item>
        <el-form-item :label="$t('users.introduction')">
          <el-input v-model="form.introduction" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ $t('users.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="saveUser">
          {{ isEditing ? $t('users.update') : $t('users.create') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- Bulk Update Dialog -->
    <el-dialog v-model="bulkDialogVisible" :title="$t('common.bulkUpdateTitle')" width="420px">
      <el-form :model="bulkForm" label-width="140px">
        <el-form-item :label="$t('users.status')">
          <el-select v-model="bulkForm.status" style="width: 100%;">
            <el-option :label="$t('common.bulkUpdateNoChange')" value="nochange" />
            <el-option :label="$t('users.active')" :value="1" />
            <el-option :label="$t('users.inactive')" :value="0" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bulkDialogVisible = false">{{ $t('users.cancel') }}</el-button>
        <el-button type="primary" @click="applyBulkUpdate" :loading="bulkUpdating">{{ $t('common.apply') }}</el-button>
      </template>
    </el-dialog>

    <!-- Bulk Delete Confirmation Dialog -->
    <el-dialog v-model="bulkDeleteDialogVisible" :title="$t('common.bulkDeleteConfirmTitle')" width="420px">
      <p>{{ $t('common.bulkDeleteConfirmText', { count: selectedCount }) }}</p>
      <template #footer>
        <el-button @click="bulkDeleteDialogVisible = false">{{ $t('users.cancel') }}</el-button>
        <el-button type="danger" @click="deleteSelectedUsers" :loading="bulkDeleting">{{ $t('users.delete') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ElMessage, ElMessageBox } from 'element-plus'
import { markRaw } from 'vue'
import { searchUsers, addUser, updateUser, deleteUser } from '@/api/users'
import { getUserInformationByUserID, updateUserInformationByUserID } from '@/api/userInformation'
import { getUserPrivileges } from '@/utils/auth'
import { Loading, Plus, Search, Edit, Delete, ArrowDown } from '@element-plus/icons-vue'

export default {
  name: 'User',
  components: {
    Loading,
    Plus,
    Search,
    Edit,
    Delete,
    ArrowDown
  },
  data() {
    return {
      users: [],
      loading: false,
      saving: false,
      error: null,
      search: '',
      pageSize: 20,
      currentPage: 1,
      totalUsers: 0,
      sortBy: '',
      sortOrder: '',
      bulkDialogVisible: false,
      bulkDeleteDialogVisible: false,
      bulkUpdating: false,
      bulkDeleting: false,
      selectedUserRows: [],
      dialogVisible: false,
      isEditing: false,
      form: {
        id: null,
        username: '',
        password: '',
        privileges: 1,
        status: 1,
        email: '',
        phone: '',
        avatar: '',
        address: '',
        introduction: '',
        nickname: '',
      },
      bulkForm: {
        status: 'nochange',
      },
      // Icons for template usage
      Plus: markRaw(Plus),
      Search: markRaw(Search),
      Edit: markRaw(Edit),
      Delete: markRaw(Delete),
      ArrowDown: markRaw(ArrowDown),
      Loading: markRaw(Loading)
    }
  },
  computed: {
    isSuperAdmin() {
      return getUserPrivileges() >= 3
    },
    filteredUsers() {
      const searchTerm = this.search.trim().toLowerCase()
      if (!searchTerm) {
        return this.users
      }
      // Perform client-side filtering on the current page of users.
      return this.users.filter((user) => {
        const username = user.username || ''
        const nickname = user.nickname || ''
        const email = user.email || ''
        return (
          username.toLowerCase().includes(searchTerm) || nickname.toLowerCase().includes(searchTerm) || email.toLowerCase().includes(searchTerm)
        )
      })
    },
    privilegeOptions() {
      const current = getUserPrivileges()
      const options = [
        { value: 1, label: this.$t('users.roleUser') },
        { value: 2, label: this.$t('users.roleAdmin') },
        { value: 3, label: this.$t('users.roleSuper') },
      ]
      if (current >= 3) {
        return options
      }
      return options.filter((opt) => opt.value < current)
    },
    selectedCount() {
      return this.selectedUserRows.length
    },
  },
  created() {
    this.loadUsers(true)
  },
  watch: {
    // The search is now handled on the client-side by the `filteredUsers` computed property.
    // We reset to page 1 when search is initiated to avoid confusion.
    search() {
      this.currentPage = 1
    },
    pageSize() {
      this.currentPage = 1
      this.loadUsers(true)
    },
    currentPage() {
      this.loadUsers()
    },
  },
  methods: {
    statusReason(status) {
      return status === 1 ? this.$t('common.reasonActive') : this.$t('common.reasonInactive')
    },
    onSelectionChange(selection) {
      this.selectedUserRows = selection || []
    },
    selectAll() {
      if (this.$refs.usersTableRef) {
        this.users.forEach((row) => {
          this.$refs.usersTableRef.toggleRowSelection(row, true)
        })
      }
    },
    onSortChange({ prop, order }) {
      if (!prop || !order) {
        this.sortBy = ''
        this.sortOrder = ''
      } else {
        this.sortBy = prop
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc'
      }
      this.currentPage = 1
      this.loadUsers(true)
    },
    openBulkDeleteDialog() {
      if (!this.isSuperAdmin) return
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'))
        return
      }
      this.bulkDeleteDialogVisible = true
    },
    async deleteSelectedUsers() {
      if (!this.isSuperAdmin || this.selectedCount === 0) return

      this.bulkDeleting = true
      try {
        await Promise.all(this.selectedUserRows.map((user) => deleteUser(user.id)))
        ElMessage.success(this.$t('common.bulkDeleteSuccess', { count: this.selectedCount }))
        this.bulkDeleteDialogVisible = false
        this.clearSelection()
        await this.loadUsers(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('common.bulkDeleteFailed'))
      } finally {
        this.bulkDeleting = false
      }
    },
    openBulkUpdateDialog() {
      if (!this.isSuperAdmin) return
      if (this.selectedCount === 0) {
        ElMessage.warning(this.$t('common.selectAtLeastOne'))
        return
      }
      this.bulkForm = { status: 'nochange' }
      this.bulkDialogVisible = true
    },
    async applyBulkUpdate() {
      if (!this.isSuperAdmin || this.selectedCount === 0) return
      if (this.bulkForm.status === 'nochange') {
        ElMessage.warning(this.$t('common.bulkUpdateNoChanges'))
        return
      }

      this.bulkUpdating = true
      try {
        const statusOverride = this.bulkForm.status
        await Promise.all(this.selectedUserRows.map((user) => {
          return updateUser(user.id, { status: statusOverride })
        }))
        ElMessage.success(this.$t('common.bulkUpdateSuccess', { count: this.selectedCount }))
        this.bulkDialogVisible = false
        this.clearSelection()
        await this.loadUsers(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('common.bulkUpdateFailed'))
      } finally {
        this.bulkUpdating = false
      }
    },
    clearSelection() {
      if (this.$refs.usersTableRef && this.$refs.usersTableRef.clearSelection) {
        this.$refs.usersTableRef.clearSelection()
      }
      this.selectedUserRows = []
    },
    async loadUsers(reset = false) {
      if (reset) {
        this.users = []
      }
      this.loading = true
      this.error = null
      try {
        const response = await searchUsers({
          // The search is now primarily client-side for responsiveness.
          // The server-side query can be kept to fetch a broadly relevant set of users.
          q: this.search || undefined,
          limit: this.pageSize,
          offset: (this.currentPage - 1) * this.pageSize,
          sort: this.sortBy || undefined,
          order: this.sortOrder || undefined,
          with_total: 1,
        })
        let payload = []
        let total = 0
        
        // Handle various response formats
        // response might be the axios response object, so response.data is the body
        const resBody = response?.data || response
        
        if (resBody?.success && resBody.data) {
          const resData = resBody.data
          if (Array.isArray(resData)) {
            payload = resData
            total = resData.length
          } else if (resData.items && Array.isArray(resData.items)) {
            payload = resData.items
            total = resData.total ?? resData.items.length
          }
        } else if (resBody && (resBody.items || Array.isArray(resBody))) {
          if (Array.isArray(resBody)) {
            payload = resBody
            total = resBody.length
          } else if (resBody.items && Array.isArray(resBody.items)) {
            payload = resBody.items
            total = resBody.total ?? resBody.items.length
          }
        }
        this.users = payload.map((u) => ({
          id: u.ID || u.id,
          username: u.Username || u.username,
          nickname: u.Nickname || u.nickname || '',
          email: u.Email || u.email || '',
          phone: u.Phone || u.phone || '',
          privileges: u.Privileges ?? u.privileges ?? 1,
          status: u.Status ?? u.status ?? 0,
        }))
        this.totalUsers = Number.isFinite(total) ? total : this.users.length
      } catch (err) {
        this.error = err?.message || this.$t('users.loadFailed')
      } finally {
        this.loading = false
      }
    },
    openCreate() {
      this.isEditing = false
      this.form = {
        id: null,
        username: '',
        password: '',
        privileges: this.privilegeOptions[0]?.value || 1,
        status: 1,
        email: '',
        phone: '',
        avatar: '',
        address: '',
        introduction: '',
        nickname: '',
      }
      this.dialogVisible = true
    },
    openEdit(row) {
      this.isEditing = true
      this.form = {
        id: row.id,
        username: row.username || '',
        password: '',
        privileges: row.privileges ?? 1,
        status: row.status ?? 1,
        email: row.email || '',
        phone: row.phone || '',
        avatar: row.avatar || '',
        address: row.address || '',
        introduction: row.introduction || '',
        nickname: row.nickname || '',
      }
      this.loadUserInfo(row.id)
      this.dialogVisible = true
    },
    async loadUserInfo(userId) {
      if (!userId) return
      try {
        const { data } = await getUserInformationByUserID(userId)
        const payload = data?.data || data
        Object.assign(this.form, {
          nickname: payload?.nickname || '',
          email: payload?.email || '',
          phone: payload?.phone || '',
          avatar: payload?.avatar || '',
          address: payload?.address || '',
          introduction: payload?.introduction || '',
        })
      } catch (err) {
        if (err?.response?.status !== 404) {
          ElMessage.error(err?.response?.data?.error || err.message || this.$t('users.loadInfoFailed'))
        }
      }
    },
    async saveUser() {
      this.saving = true
      try {
        if (!this.form.username) {
          ElMessage.warning(this.$t('users.usernameRequired'))
          return
        }
        if (!this.isEditing && !this.form.password) {
          ElMessage.warning(this.$t('users.passwordRequired'))
          return
        }
        const authPayload = {
          username: this.form.username,
          privileges: this.form.privileges,
          status: this.form.status,
        }
        if (!this.isEditing && this.form.password) {
          authPayload.password = this.form.password
        }
        const infoPayload = {
          nickname: this.form.nickname,
          email: this.form.email,
          phone: this.form.phone,
          avatar: this.form.avatar,
          address: this.form.address,
          introduction: this.form.introduction,
        }
        if (this.isEditing && this.form.id) {
          await updateUser(this.form.id, authPayload)
          await updateUserInformationByUserID(this.form.id, infoPayload)
          ElMessage.success(this.$t('users.updated'))
        } else {
          await addUser(authPayload)
          ElMessage.success(this.$t('users.created'))
        }
        this.dialogVisible = false
        await this.loadUsers(true)
      } catch (err) {
        ElMessage.error(err?.response?.data?.error || err.message || this.$t('users.operationFailed'))
      } finally {
        this.saving = false
      }
    },
    async confirmDelete(row) {
      try {
        await ElMessageBox.confirm(
          this.$t('users.confirmDeleteText'),
          this.$t('users.confirmDelete'),
          { type: 'warning' }
        )
        await deleteUser(row.id)
        ElMessage.success(this.$t('users.deleted'))
        await this.loadUsers(true)
      } catch (err) {
        if (err !== 'cancel' && err !== 'close') {
          ElMessage.error(err?.response?.data?.error || err.message || this.$t('users.deleteFailed'))
        }
      }
    },
    roleLabel(privileges) {
      switch (privileges) {
        case 3:
          return this.$t('users.roleSuper')
        case 2:
          return this.$t('users.roleAdmin')
        case 1:
          return this.$t('users.roleUser')
        default:
          return this.$t('users.roleUnauthorized')
      }
    },
    roleTagType(privileges) {
      if (privileges >= 3) return 'danger'
      if (privileges === 2) return 'warning'
      if (privileges === 1) return 'success'
      return 'info'
    },
  },
}
</script>

<style scoped>
.users-content {
  margin-top: 8px;
}

.users-pagination {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}

.loading-state {
  text-align: center;
  padding: 60px;
}

:deep(.el-table__row) {
  cursor: pointer;
  transition: all 0.2s ease;
}

:deep(.el-table__row:hover) {
  background-color: var(--brand-50) !important;
}
</style>
