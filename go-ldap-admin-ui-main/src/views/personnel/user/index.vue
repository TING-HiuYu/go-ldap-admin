<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('user.username')">
          <el-input v-model.trim="params.username" style="width: 100px;" clearable :placeholder="$t('user.username')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('user.nickname')">
          <el-input v-model.trim="params.nickname" style="width: 100px;" clearable :placeholder="$t('user.nickname')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('common.status')">
          <el-select v-model.trim="params.status" style="width: 100px;" clearable :placeholder="$t('common.status')" @change="search" @clear="search">
            <el-option :label="$t('common.normal')" value="1" />
            <el-option :label="$t('common.disabled')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('user.syncStatus')">
          <el-select v-model.trim="params.syncState" style="width: 100px;" clearable :placeholder="$t('user.syncStatus')" @change="search" @clear="search">
            <el-option :label="$t('user.synced')" value="1" />
            <el-option :label="$t('user.notSynced')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-search" type="primary" @click="search">{{ $t('common.search') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="create">{{ $t('common.add') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" icon="el-icon-delete" type="danger" @click="batchDelete">{{ $t('common.batchDelete') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" icon="el-icon-upload2" type="success" @click="batchSync">{{ $t('common.batchSync') }}</el-button>
        </el-form-item>
        <br>
        <el-form-item v-if="syncConfig.ldapEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncOpenLdapUsers">{{ $t('user.syncLdapUsers') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.dingTalkEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncDingTalkUsers">{{ $t('user.syncDingTalkUsers') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.feiShuEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncFeiShuUsers">{{ $t('user.syncFeishuUsers') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.weComEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncWeComUsers">{{ $t('user.syncWeComUsers') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="username" :label="$t('user.username')" />
        <el-table-column show-overflow-tooltip sortable prop="nickname" :label="$t('user.chineseName')" />
        <el-table-column show-overflow-tooltip sortable prop="givenName" :label="$t('user.alias')" />
        <!-- Displayed as switch; consider using boolean parameter in the future -->
        <el-table-column :label="$t('common.status')" align="center">
          <template slot-scope="scope">
            <el-switch v-model="scope.row.status" :active-value="1" :inactive-value="2" @change="userStateChanged(scope.row)" />
          </template>
        </el-table-column>
        <!-- <el-table-column show-overflow-tooltip sortable prop="status" :label="$t('common.status')" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.status === 1 ? 'success':'danger'" disable-transitions>{{ scope.row.status === 1 ? $t('common.normal') : $t('common.disabled') }}</el-tag>
          </template>
        </el-table-column> -->
        <el-table-column show-overflow-tooltip sortable prop="mail" :label="$t('user.email')" />

        <el-table-column show-overflow-tooltip sortable prop="departments" :label="$t('user.department')" />
        <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('common.creator')" />
        <el-table-column show-overflow-tooltip sortable prop="introduction" :label="$t('common.remark')" />
        <el-table-column show-overflow-tooltip sortable prop="userDn" label="DN" />
        <el-table-column show-overflow-tooltip sortable prop="CreatedAt" :label="$t('common.createdAt')" />
        <el-table-column show-overflow-tooltip sortable prop="UpdatedAt" :label="$t('common.updatedAt')" />
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="190">
          <template slot-scope="scope">
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="update(scope.row)" />
            </el-tooltip>
            <el-tooltip class="delete-popover" :content="$t('user.resetPassword')" effect="dark" placement="top">
              <el-popconfirm :title="$t('user.confirmResetPassword')" @onConfirm="resetUserPassword(scope.row.username)">
                <el-button slot="reference" size="mini" icon="el-icon-key" circle type="warning" />
              </el-popconfirm>
            </el-tooltip>
            <el-tooltip class="delete-popover" :content="$t('common.delete')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmDelete')" @onConfirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
              </el-popconfirm>
            </el-tooltip>
            <el-tooltip v-if="scope.row.syncState == 2" class="delete-popover" :content="$t('common.sync')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmSync')" @onConfirm="singleSync(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-upload2" circle type="success" />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        :current-page="params.pageNum"
        :page-size="params.pageSize"
        :total="total"
        :page-sizes="[1, 5, 10, 30]"
        layout="total, prev, pager, next, sizes"
        background
        style="margin-top: 10px;float:right;margin-bottom: 10px;"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />

      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible" width="50%">
        <el-form ref="dialogForm" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="80px">
          <el-row>
            <el-col :span="12">
              <el-form-item :label="$t('user.username')" prop="username">
                <el-input ref="password" v-model.trim="dialogFormData.username" :disabled="disabled" :placeholder="$t('user.usernamePinyin')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.chineseNameFull')" prop="nickname">
                <el-input v-model.trim="dialogFormData.nickname" :placeholder="$t('user.chineseNameFull')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.alias')" prop="givenName">
                <el-input v-model.trim="dialogFormData.givenName" :placeholder="$t('user.alias')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.email')" prop="mail">
                <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('user.email')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <!-- Hide password field when editing user -->
              <el-form-item v-if="dialogType === 'create'" :label="dialogType === 'create' ? $t('user.newPassword') : $t('user.resetPassword')" prop="password">
                <el-input v-model.trim="dialogFormData.password" autocomplete="off" :type="passwordType" :placeholder="dialogType === 'create' ? $t('user.newPassword') : $t('user.resetPassword')" />
                <span class="show-pwd" @click="showPwd">
                  <svg-icon :icon-class="passwordType === 'password' ? 'eye' : 'eye-open'" />
                </span>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('user.role')" prop="roleIds">
                <el-select v-model.trim="dialogFormData.roleIds" multiple :placeholder="$t('user.pleaseSelectRole')" style="width:100%">
                  <el-option
                    v-for="item in roles"
                    :key="item.ID"
                    :label="item.name"
                    :value="item.ID"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('common.status')" prop="status">
                <el-select v-model.trim="dialogFormData.status" :placeholder="$t('user.pleaseSelectStatus')" style="width:100%">
                  <el-option :label="$t('common.normal')" :value="1" />
                  <el-option :label="$t('common.disabled')" :value="2" />
                </el-select>
              </el-form-item>
            </el-col>

            <el-col v-if="dialogType === 'update'" :span="12">
              <el-form-item label="uid">
                <el-input v-model.trim="dialogFormData.uidNumber" disabled />
              </el-form-item>
            </el-col>
            <el-col v-if="dialogType === 'update'" :span="12">
              <el-form-item label="gid">
                <el-input v-model.trim="dialogFormData.gidNumber" disabled />
              </el-form-item>
            </el-col>
            <el-col v-if="dialogType === 'update'" :span="24">
              <el-form-item label="homeDir">
                <el-input v-model.trim="dialogFormData.homeDirectory" disabled />
              </el-form-item>
            </el-col>

            <el-col :span="24">
              <el-form-item label="Shell">
                <el-input
                  v-model.trim="dialogFormData.loginShell"
                  :placeholder="defaultLoginShell ? $t('user.defaultShellPlaceholder', { shell: defaultLoginShell }) : 'Shell'"
                />
              </el-form-item>
            </el-col>

            <el-col :span="24">
              <el-form-item :label="$t('user.belongDepartment')" prop="departmentId">
                <treeselect
                  v-model="dialogFormData.departmentId"
                  :options="departmentsOptions"
                  :placeholder="$t('user.pleaseSelectDepartment')"
                  :normalizer="normalizer"
                  value-consists-of="ALL"
                  :multiple="true"
                  :flat="true"
                  :no-children-text="$t('common.noMoreOptions')"
                  :no-results-text="$t('common.noMatchingOptions')"
                  @input="treeselectInput"
                />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('user.address')" prop="postalAddress">
                <el-input v-model.trim="dialogFormData.postalAddress" type="textarea" :placeholder="$t('user.address')" :autosize="{minRows: 3, maxRows: 6}" show-word-limit maxlength="100" />
              </el-form-item>
            </el-col>
            <el-col :span="24">
              <el-form-item :label="$t('common.remark')" prop="introduction">
                <el-input v-model.trim="dialogFormData.introduction" type="textarea" :placeholder="$t('common.remark')" :autosize="{minRows: 3, maxRows: 6}" show-word-limit maxlength="100" />
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button size="mini" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>

      <!-- Reset password result dialog -->
      <el-dialog
        :title="$t('user.passwordResetSuccessful')"
        :visible.sync="resetPasswordDialogVisible"
        width="400px"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        @close="closeResetPasswordDialog"
      >
        <div style="text-align: center;">
          <el-alert
            :title="$t('user.pleaseSaveNewPassword')"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 20px;"
          />
          <p style="margin-bottom: 10px; font-weight: bold;">{{ $t('user.userLabel') }}{{ resetUsername }}</p>
          <p style="margin-bottom: 20px; color: #606266;">{{ $t('user.newPassword') }}:</p>
          <el-input
            v-model="newPassword"
            readonly
            style="margin-bottom: 20px;"
          >
            <el-button
              slot="append"
              icon="el-icon-document-copy"
              @click="copyPassword"
            >
              {{ $t('common.copy') }}
            </el-button>
          </el-input>
          <el-alert
            :title="$t('user.savePasswordWarning')"
            type="info"
            :closable="false"
            show-icon
          />
        </div>
        <div slot="footer" class="dialog-footer">
          <el-button type="primary" @click="closeResetPasswordDialog">{{ $t('user.iHaveSaved') }}</el-button>
        </div>
      </el-dialog>

    </el-card>
  </div>
</template>

<script>
import JSEncrypt from 'jsencrypt'
import Treeselect from '@riophae/vue-treeselect'
import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import { getUsers, createUser, updateUserById, batchDeleteUserByIds, changeUserStatus, syncDingTalkUsersApi, syncWeComUsersApi, syncFeiShuUsersApi, syncOpenLdapUsersApi, syncSqlUsers } from '@/api/personnel/user'
import { resetPassword } from '@/api/system/user'
import { getRoles } from '@/api/system/role'
import { getGroupTree } from '@/api/personnel/group'
import { getConfig } from '@/api/system/base'
import { Message } from 'element-ui'

export default {
  name: 'User',
  components: {
    Treeselect
  },
  props: {
    disabled: { // username is not editable by default. To make it editable, remove this control (in create and edit) and use with backend ldap-user-name-modify config
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      // Query parameters
      params: {
        username: '',
        nickname: '',
        status: '',
        syncState: '',
        pageNum: 1,
        pageSize: 10
      },
      // Table data
      tableData: [],
      total: 0,
      loading: false,
      isUpdate: false,
      // Department data
      treeselectValue: 0,
      // Roles
      roles: [],
      // Department info
      departmentsOptions: [],

      passwordType: 'password',

      publicKey: process.env.VUE_APP_PUBLIC_KEY,

      // Dialog
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        username: '',
        password: '',
        nickname: '',
        status: 1,
        avatar: '',
        introduction: '',
        roleIds: '',
        ID: '',
        mail: '',
        givenName: '',
        postalAddress: '',
        departments: '',
        departmentId: undefined,
        uidNumber: '',
        gidNumber: '',
        homeDirectory: '',
        loginShell: ''
      },
      dialogFormRules: {
        username: [
          { required: true, message: this.$t('user.pleaseEnterUsername'), trigger: 'blur' },
          { min: 2, max: 20, message: this.$t('common.lengthBetween', { min: 2, max: 20 }), trigger: 'blur' }
        ],
        password: [
          { required: false, message: this.$t('user.pleaseEnterPassword'), trigger: 'blur' },
          { min: 6, max: 30, message: this.$t('common.lengthBetween', { min: 6, max: 30 }), trigger: 'blur' }
        ],
        mail: [
          { required: true, message: this.$t('user.pleaseEnterEmail'), trigger: 'blur' }
        ],
        nickname: [
          { required: true, message: this.$t('user.pleaseEnterNickname'), trigger: 'blur' },
          { min: 2, max: 20, message: this.$t('common.lengthBetween', { min: 2, max: 20 }), trigger: 'blur' }
        ],
        status: [
          { required: true, message: this.$t('user.pleaseSelectStatus'), trigger: 'change' }
        ],
        departmentId: [
          { required: true, message: this.$t('user.pleaseSelectDepartment'), trigger: 'change' },
          { validator: (rule, value, callBack) => {
            if (value < 1) {
              callBack(this.$t('group.pleaseSelectValidDepartment'))
            } else {
              callBack()
            }
          }
          }
        ],
        introduction: [
          { required: false, message: this.$t('common.remark'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('common.lengthBetween', { min: 0, max: 100 }), trigger: 'blur' }
        ]
      },

      // Delete button popover
      popoverVisible: false,
      // Table multi-select
      multipleSelection: [],
      changeUserStatusFormData: {
        id: '',
        status: ''
      },

      // Reset password result dialog
      resetPasswordDialogVisible: false,
      newPassword: '',
      resetUsername: '',

      // Sync configuration
      syncConfig: {
        ldapEnableSync: false,
        dingTalkEnableSync: false,
        feiShuEnableSync: false,
        weComEnableSync: false
      },
      defaultLoginShell: '/bin/bash'
    }
  },
  created() {
    this.getTableData()
    this.getRoles()
    this.getSyncConfig()
  },
  methods: {
    // Get sync configuration
    async getSyncConfig() {
      try {
        const { data } = await getConfig()
        this.syncConfig = {
          ldapEnableSync: data.ldapEnableSync,
          dingTalkEnableSync: data.dingTalkEnableSync,
          feiShuEnableSync: data.feiShuEnableSync,
          weComEnableSync: data.weComEnableSync
        }
        this.defaultLoginShell = data.defaultLoginShell || this.defaultLoginShell
        if (!this.dialogFormData.loginShell) {
          this.dialogFormData.loginShell = this.defaultLoginShell
        }
      } catch (error) {
        console.error(this.$t('group.failedGetSyncConfig') + ':', error)
      }
    },
    // Search
    search() {
      this.params.pageNum = 1
      this.getTableData()
    },

    // Get table data
    async getTableData() {
      this.loading = true
      try {
        const { data } = await getUsers(this.params)
        data.users.forEach(item => {
          const dataStrArr = item.departmentId.split(',')
          const dataIntArr = []
          dataStrArr.forEach(item => {
            dataIntArr.push(+item)
          })
          item.departmentId = dataIntArr
        })
        this.tableData = data.users
        this.total = data.total
      } finally {
        this.loading = false
      }
    },
    // Get all groups for department selection in dialog
    async getAllGroups() {
      this.loading = true
      try {
        const checkParams = {
          pageNum: 1,
          pageSize: 1000 // Most organizations should not have this many entries
        }
        const { data } = await getGroupTree(checkParams)
        this.departmentsOptions = [{ ID: 0, groupName: this.$t('user.pleaseSelectDepartmentInfo'), groupType: 'T', children: data }]
      } finally {
        this.loading = false
      }
    },
    // Get role data
    async getRoles() {
      const res = await getRoles(null)

      this.roles = res.data.roles
    },

    // Create
    create() {
      this.dialogFormTitle = this.$t('user.addUser')
      this.dialogType = 'create'
      this.disabled = false
      this.getAllGroups()
      this.dialogFormVisible = true
      this.dialogFormData.loginShell = this.defaultLoginShell
    },

    // Edit
    update(row) {
      this.disabled = true
      this.getAllGroups()
      this.dialogFormData.ID = row.ID
      this.dialogFormData.username = row.username
      this.dialogFormData.password = ''
      this.dialogFormData.nickname = row.nickname
      this.dialogFormData.status = row.status

      this.dialogFormData.introduction = row.introduction
      // Get role IDs from role array
      this.dialogFormData.roleIds = row.roles.map(item => item.ID)

      this.dialogFormTitle = this.$t('user.editUser')
      this.dialogType = 'update'
      this.passwordType = 'password'
      this.dialogFormVisible = true

      this.dialogFormData.mail = row.mail
      this.dialogFormData.givenName = row.givenName

      this.dialogFormData.postalAddress = row.postalAddress
      this.dialogFormData.departments = row.departments
      this.dialogFormData.departmentId = row.departmentId

      this.dialogFormData.uidNumber = row.uidNumber || ''
      this.dialogFormData.gidNumber = row.gidNumber || ''
      this.dialogFormData.homeDirectory = row.homeDirectory || ''
      this.dialogFormData.loginShell = row.loginShell || this.defaultLoginShell
    },

    // Convert department ID to department name
    setDepartmentNameByDepartmentId() {
      const ids = this.dialogFormData.departmentId
      if (!ids || !ids.length) return
      const departments = []
      // Depth-first traversal
      const dfs = (node, cb) => {
        if (!node) return
        cb(node)
        if (node.children && node.children.length) {
          node.children.forEach(item => {
            dfs(item, cb)
          })
        }
      }
      dfs(this.departmentsOptions[0], node => {
        if (ids.includes(node.ID)) {
          departments.push(node.groupName)
        }
      })
      this.dialogFormData.departments = departments.join(',')
    },

    // Check result
    judgeResult(res) {
      if (res.code === 0) {
        Message({
          showClose: true,
          message: this.$t('common.operationSuccessful'),
          type: 'success'
        })
      }
    },

    // Submit form
    submitForm() {
      if (this.dialogFormData.nickname === '') {
        Message({
          showClose: true,
          message: this.$t('user.pleaseEnterNickname'),
          type: 'error'
        })
        return false
      }
      if (this.dialogFormData.username === '') {
        Message({
          showClose: true,
          message: this.$t('user.pleaseEnterUsername'),
          type: 'error'
        })
        return false
      }
      if (this.dialogFormData.mail === '') {
        Message({
          showClose: true,
          message: this.$t('user.pleaseEnterEmail'),
          type: 'error'
        })
        return false
      }
      if (this.dialogFormData.status === '') {
        Message({
          showClose: true,
          message: this.$t('user.pleaseEnterStatus'),
          type: 'error'
        })
        return false
      }
      if (this.dialogFormData.roleIds === '') {
        Message({
          showClose: true,
          message: this.$t('user.pleaseSelectRoles'),
          type: 'error'
        })
        return false
      }
      this.$refs['dialogForm'].validate(async valid => {
        if (valid) {
          this.submitLoading = true
          // Auto-fill department field
          this.setDepartmentNameByDepartmentId()
          this.dialogFormDataCopy = { ...this.dialogFormData }
          if (this.dialogFormData.password !== '') {
          // RSA encrypt password
            const encryptor = new JSEncrypt()
            // Set public key
            const publicKey = (this.publicKey || '').replace(/\\n/g, '\n')
            encryptor.setPublicKey(publicKey)
            // Encrypt password
            const encPassword = encryptor.encrypt(this.dialogFormData.password)
            if (!encPassword) {
              Message({
                showClose: true,
                message: this.$t('user.invalidPublicKey'),
                type: 'error'
              })
              this.submitLoading = false
              return
            }
            this.dialogFormDataCopy.password = encPassword
          }
          try {
            if (this.dialogType === 'create') {
              await createUser(this.dialogFormDataCopy).then(res => {
                this.judgeResult(res)
              })
            } else {
              await updateUserById(this.dialogFormDataCopy).then(res => {
                this.judgeResult(res)
              })
            }
          } finally {
            this.submitLoading = false
          }
          this.resetForm()
          this.getTableData()
        } else {
          Message({
            showClose: true,
            message: this.$t('common.formValidationFailed'),
            type: 'warn'
          })
          return false
        }
      })
    },

    // Cancel form
    cancelForm() {
      this.resetForm()
    },

    resetForm() {
      this.dialogFormVisible = false
      this.$refs['dialogForm'].resetFields()
      this.dialogFormData = {
        username: '',
        password: '',
        nickname: '',
        status: 1,
        avatar: '',
        introduction: '',
        roleIds: '',
        mail: '',
        givenName: '',
        postalAddress: '',
        departments: '',
        departmentId: undefined,
        uidNumber: '',
        gidNumber: '',
        homeDirectory: '',
        loginShell: this.defaultLoginShell
      }
    },

    // Batch delete
    batchDelete() {
      this.$confirm(this.$t('common.permanentDeleteWarning'), this.$t('common.notice'), {
        confirmButtonText: this.$t('common.confirm'),
        cancelButtonText: this.$t('common.cancel'),
        type: 'warning'
      }).then(async res => {
        this.loading = true
        const userIds = []
        this.multipleSelection.forEach(x => {
          userIds.push(x.ID)
        })
        try {
          await batchDeleteUserByIds({ userIds: userIds }).then(res => {
            this.judgeResult(res)
          })
        } finally {
          this.loading = false
        }
        this.getTableData()
      }).catch(() => {
        Message({
          showClose: true,
          type: 'info',
          message: this.$t('common.deleteCancelled')
        })
      })
    },
    // Batch sync
    batchSync() {
      this.$confirm(this.$t('user.batchSyncToLdap'), this.$t('common.notice'), {
        confirmButtonText: this.$t('common.confirm'),
        cancelButtonText: this.$t('common.cancel'),
        type: 'warning'
      }).then(async res => {
        this.loading = true
        const userIds = []
        this.multipleSelection.forEach(x => {
          userIds.push(x.ID)
        })
        try {
          await syncSqlUsers({ userIds: userIds }).then(res => {
            this.judgeResult(res)
          })
        } finally {
          this.loading = false
        }
        this.getTableData()
      }).catch(() => {
        Message({
          showClose: true,
          type: 'info',
          message: this.$t('common.syncCancelled')
        })
      })
    },

    // Watch switch status change
    async userStateChanged(userInfo) {
      this.changeUserStatusFormData.id = userInfo.ID
      this.changeUserStatusFormData.status = userInfo.status
      const { code } = await changeUserStatus(this.changeUserStatusFormData)
      if (code !== 0) {
        return Message.error(this.$t('user.updateStatusFailed'))
      }
      Message.success(this.$t('user.updateStatusSuccessful'))
    },

    // Table multi-select
    handleSelectionChange(val) {
      this.multipleSelection = val
    },

    // Single delete
    async singleDelete(Id) {
      this.loading = true
      try {
        await batchDeleteUserByIds({ userIds: [Id] }).then(res => {
          this.judgeResult(res)
        })
      } finally {
        this.loading = false
      }
      this.getTableData()
    },
    // Single sync
    async singleSync(Id) {
      this.loading = true
      try {
        await syncSqlUsers({ userIds: [Id] }).then(res => {
          this.judgeResult(res)
        })
      } finally {
        this.loading = false
      }
      this.getTableData()
    },

    showPwd() {
      if (this.passwordType === 'password') {
        this.passwordType = ''
      } else {
        this.passwordType = 'password'
      }
    },

    // Pagination
    handleSizeChange(val) {
      this.params.pageSize = val
      this.getTableData()
    },
    handleCurrentChange(val) {
      this.params.pageNum = val
      this.getTableData()
    },
    // Treeselect
    normalizer(node) {
      return {
        id: node.ID,
        label: node.groupType + '=' + node.groupName,
        isDisabled: node.groupType === 'ou' || node.groupName === 'root',
        children: node.children
      }
    },
    treeselectInput(value) {
      this.treeselectValue = value
    },
    syncDingTalkUsers() {
      this.loading = true
      syncDingTalkUsersApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncWeComUsers() {
      this.loading = true
      syncWeComUsersApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncFeiShuUsers() {
      this.loading = true
      syncFeiShuUsersApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncOpenLdapUsers() {
      this.loading = true
      syncOpenLdapUsersApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },

    // Reset user password
    async resetUserPassword(username) {
      this.loading = true
      try {
        const res = await resetPassword({ username: username })
        if (res.code === 0) {
          this.newPassword = res.data.newPassword
          this.resetUsername = username
          this.resetPasswordDialogVisible = true
          Message({
            showClose: true,
            message: this.$t('user.passwordResetSuccessful'),
            type: 'success'
          })
        } else {
          Message({
            showClose: true,
            message: res.msg || this.$t('user.passwordResetFailed'),
            type: 'error'
          })
        }
      } finally {
        this.loading = false
      }
      this.getTableData()
    },

    // Copy password to clipboard
    copyPassword() {
      const textArea = document.createElement('textarea')
      textArea.value = this.newPassword
      document.body.appendChild(textArea)
      textArea.select()
      try {
        document.execCommand('copy')
        Message({
          showClose: true,
          message: this.$t('user.passwordCopied'),
          type: 'success'
        })
      } catch (err) {
        Message({
          showClose: true,
          message: this.$t('user.copyFailed'),
          type: 'error'
        })
      }
      document.body.removeChild(textArea)
    },

    // Close reset password dialog
    closeResetPasswordDialog() {
      this.resetPasswordDialogVisible = false
      this.newPassword = ''
      this.resetUsername = ''
    }
  }
}
</script>

<style scoped>
  .container-card{
    margin: 10px;
    margin-bottom: 100px;
  }

  .delete-popover{
    margin-left: 10px;
  }

  .show-pwd {
    position: absolute;
    right: 10px;
    top: 3px;
    font-size: 16px;
    color: #889aa4;
    cursor: pointer;
    user-select: none;
  }
</style>
