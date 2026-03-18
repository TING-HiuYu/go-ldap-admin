<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('group.name')">
          <el-input v-model.trim="params.groupName" style="width: 100px;" clearable :placeholder="$t('group.name')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('common.description')">
          <el-input v-model.trim="params.remark" style="width: 100px;" clearable :placeholder="$t('common.description')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('user.syncStatus')">
          <el-select v-model.trim="params.syncState" style="width: 110px;" clearable :placeholder="$t('user.syncStatus')" @change="search" @clear="search">
            <el-option :label="$t('user.synced')" value="1" />
            <el-option :label="$t('user.notSynced')" value="2" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-search" type="primary" @click="search">{{ $t('common.search') }}</el-button>
        </el-form-item>
        <!-- <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="resetData">{{ $t('common.reset') }}</el-button>
        </el-form-item> -->
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
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncOpenLdapDepts">{{ $t('group.syncLdapDepartments') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.dingTalkEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncDingTalkDepts">{{ $t('group.syncDingTalkDepartments') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.feiShuEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncFeiShuDepts">{{ $t('group.syncFeishuDepartments') }}</el-button>
        </el-form-item>
        <el-form-item v-if="syncConfig.weComEnableSync">
          <el-button :loading="loading" icon="el-icon-download" type="warning" @click="syncWeComDepts">{{ $t('group.syncWeComDepartments') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :default-expand-all="true" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" row-key="ID" :data="infoTableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="groupName" :label="$t('group.name')" />
        <el-table-column show-overflow-tooltip sortable prop="groupType" :label="$t('group.type')" />
        <el-table-column show-overflow-tooltip sortable prop="groupDn" label="DN" />
        <el-table-column show-overflow-tooltip sortable prop="remark" :label="$t('common.description')" />
        <el-table-column show-overflow-tooltip sortable prop="CreatedAt" :label="$t('common.createdAt')" />
        <el-table-column show-overflow-tooltip sortable prop="UpdatedAt" :label="$t('common.updatedAt')" />
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="220">
          <template #default="scope">
            <el-tooltip v-if="scope.row.groupType != 'ou' && scope.row.groupName != 'root'" :content="$t('group.addMember')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-setting" circle type="info" @click="addUp(scope.row)" />
            </el-tooltip>
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="update(scope.row)" />
            </el-tooltip>
            <el-tooltip class="delete-popover" :content="$t('common.delete')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmDelete')" @onConfirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
              </el-popconfirm>
            </el-tooltip>
            <el-tooltip v-if="scope.row.syncState === 2" class="delete-popover" :content="$t('common.sync')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmSync')" @onConfirm="singleSync(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-upload2" circle type="success" />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>
      <!-- Add -->
      <el-dialog :title="dialogFormTitle" :visible.sync="updateLoading">
        <el-form ref="dialogForm" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="120px">
          <el-form-item :label="$t('group.name')" prop="groupName">
            <el-input v-model.trim="dialogFormData.groupName" :placeholder="$t('group.namePinyin')" />
          </el-form-item>
          <el-form-item :label="$t('group.groupType')" prop="groupType">
            <el-select v-model.trim="dialogFormData.groupType" :placeholder="$t('group.groupTypeHint')" style="width:100%">
              <el-option :label="$t('group.cnGroup')" value="cn" />
              <el-option :label="$t('group.ouOrganization')" value="ou" />
            </el-select>
          </el-form-item>
          <el-form-item v-if="dialogFormData.groupType === 'cn'" :label="$t('group.cnObjectType')" prop="groupClass">
            <div style="display:flex;align-items:center;">
              <el-select v-model.trim="dialogFormData.groupClass" :placeholder="$t('group.selectLdapGroupType')" style="flex:1;">
                <el-option label="groupOfUniqueNames" value="groupOfUniqueNames" />
                <el-option label="posixGroup" value="posixGroup" />
              </el-select>
              <el-tooltip :content="$t('group.posixGroupHint')" placement="top">
                <i class="el-icon-question" style="margin-left:8px;color:#909399;" />
              </el-tooltip>
            </div>
          </el-form-item>
          <el-form-item :label="$t('group.parentGroup')" prop="parentId">
            <treeselect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              :placeholder="$t('group.pleaseSelectParentGroup')"
              @input="treeselectInput"
            />
          </el-form-item>
          <el-form-item :label="$t('common.description')" prop="remark">
            <el-input v-model.trim="dialogFormData.remark" type="textarea" :placeholder="$t('common.description')" :autosize="{minRows: 3, maxRows: 6}" show-word-limit maxlength="100" />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button size="mini" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>
      <!-- Edit -->
      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible">
        <el-form ref="dialogForm" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="120px">
          <el-form-item :label="$t('group.name')" prop="groupName">
            <el-input v-model.trim="dialogFormData.groupName" :disabled="true" :placeholder="$t('group.name')" />
          </el-form-item>
          <el-form-item :label="$t('group.groupType')">
            <el-input v-model.trim="dialogFormData.groupType" disabled />
          </el-form-item>
          <el-form-item :label="$t('group.ldapType')">
            <el-input v-model.trim="dialogFormData.groupClass" disabled />
          </el-form-item>
          <el-form-item v-if="dialogFormData.groupClass === 'posixGroup'" label="gidNumber">
            <el-input v-model.trim="dialogFormData.gidNumber" disabled />
          </el-form-item>
          <el-form-item v-if="dialogFormData.groupClass === 'posixGroup'" :label="$t('group.homeDirectory')">
            <el-input :value="dialogFormData.homePrefix || ('/home/' + dialogFormData.groupName)" disabled />
          </el-form-item>
          <el-form-item :label="$t('common.description')" prop="remark">
            <el-input v-model.trim="dialogFormData.remark" type="textarea" :placeholder="$t('common.description')" :autosize="{minRows: 3, maxRows: 6}" show-word-limit maxlength="100" />
          </el-form-item>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button size="mini" :loading="submitLoading" type="primary" @click="submitForm()">{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import Treeselect from '@riophae/vue-treeselect'
import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import { getGroupTree, groupAdd, groupUpdate, groupDel, syncDingTalkDeptsApi, syncWeComDeptsApi, syncFeiShuDeptsApi, syncOpenLdapDeptsApi, syncSqlGroups } from '@/api/personnel/group'
import { getConfig } from '@/api/system/base'
import { Message } from 'element-ui'

export default {
  name: 'Group',
  components: {
    Treeselect
  },
  filters: {
    methodTagFilter(val) {
      if (val === 'GET') {
        return ''
      } else if (val === 'POST') {
        return 'success'
      } else {
        return 'info'
      }
    }
  },
  data() {
    return {
      // Query parameters
      params: {
        groupName: undefined,
        remark: undefined,
        syncState: undefined,
        pageNum: 1,
        pageSize: 1000// Normal usage should not exceed this; backend limits max 1000 per request
      },
      // Table data
      tableData: [],
      infoTableData: [],
      total: 0,
      loading: false,
      // Parent directory data
      treeselectData: [],
      treeselectValue: 0,
      updateLoading: false, // Add new
      // Dialog
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        ID: '',
        groupName: '',
        parentId: 0,
        syncState: 1,
        groupType: '',
        groupClass: 'groupOfUniqueNames',
        gidNumber: '',
        homePrefix: '',
        remark: ''
      },
      dialogFormRules: {

        groupName: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', {min: 1, max: 50}), trigger: 'blur' }
        ],
        groupType: [
          { required: true, message: this.$t('group.groupTypeValidation'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('group.ouCnOrOther'), trigger: 'blur' }
        ],
        groupClass: [
          { validator: (rule, value, callBack) => {
            if (this.dialogFormData.groupType === 'cn' && !value) {
              callBack(this.$t('group.pleaseSelectGroupObjectType'))
            } else {
              callBack()
            }
          }, trigger: 'change' }
        ],
        parentId: [
          { required: true, message: this.$t('group.pleaseSelectParent'), trigger: 'blur' },
          { validator: (rule, value, callBack) => {
            if (value >= 0) {
              callBack()
            } else {
              callBack(this.$t('group.pleaseSelectValidDepartment'))
            }
          } }
        ],
        remark: [
          { required: false, message: this.$t('common.remark'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('common.lengthBetween', {min: 0, max: 100}), trigger: 'blur' }
        ]
      },

      // Delete button popover
      popoverVisible: false,
      // Table multi-selection
      multipleSelection: [],
      dialogTransfer: '', // Transfer dialog header
      dialogTransferVisible: false,

      transParams: {
        groupId: '',
        nickname: ''
      },
      renderFunc(h, option) {
        return <span>{option.key} - {option.label}</span>
      },
      userArrInfo: [], // Initial personnel list data
      data: [], // Transformed personnel list data
      value3: [], // Right side default personnel list data
      userId: [], // Selected user code array for backend
      ui: {
        submitLoading: false
      },
      statusTrans: '',

      // Sync configuration
      syncConfig: {
        ldapEnableSync: false,
        dingTalkEnableSync: false,
        feiShuEnableSync: false,
        weComEnableSync: false
      }
    }
  },
  created() {
    this.getTableData()
    this.getSyncConfig()
  },
  methods: {
    // Get sync configuration
    async getSyncConfig() {
      try {
        const { data } = await getConfig()
        this.syncConfig = data
      } catch (error) {
        console.error(this.$t('group.failedGetSyncConfig'), error)
      }
    },
    // Search
    search() {
      // Initialize table data
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData))
      this.infoTableData = this.deal(this.infoTableData, node => node.groupName.includes(this.params.groupName) || node.remark.includes(this.params.remark) || node.syncState.toString().includes(this.params.syncState))
    },
    resetData() {
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData))
    },
    // Page data filter
    deal(nodes, predicate) {
      // If no nodes left, end recursion
      if (!(nodes && nodes.length)) {
        return []
      }
      const newChildren = []
      for (const node of nodes) {
        if (predicate(node)) {
          // If node matches condition, add directly to new node set
          newChildren.push(node)
          node.children = this.deal(node.children, predicate)
        } else {
          // If current node doesn't match, recursively filter children,
          // promote matching children up into the new node set
          newChildren.push(...this.deal(node.children, predicate))
        }
      }
      return newChildren
    },
    // Get table data
    async getTableData() {
      this.loading = true
      try {
        const { data } = await getGroupTree(this.params)
        this.tableData = data
        this.infoTableData = JSON.parse(JSON.stringify(data))
        this.treeselectData = [{ ID: 0, groupName: this.$t('group.topLevel'), children: data }]
      } finally {
        this.loading = false
      }
    },

    // Add new
    create() {
      this.dialogFormTitle = this.$t('group.addGroup')
      this.updateLoading = true // Show add dialog
      this.dialogType = 'create'
      this.dialogFormData.groupClass = 'groupOfUniqueNames'
      this.dialogFormData.gidNumber = ''
      this.dialogFormData.homePrefix = ''
      this.dialogFormData.groupType = ''
      this.dialogFormData.parentId = 0
    },
    // Edit
    update(row) {
      this.dialogFormData.ID = row.ID
      this.dialogFormData.groupName = row.groupName
      this.dialogFormData.remark = row.remark
      this.dialogFormData.groupType = row.groupType
      this.dialogFormData.groupClass = row.groupClass || 'groupOfUniqueNames'
      this.dialogFormData.gidNumber = row.gidNumber || ''
      this.dialogFormData.homePrefix = row.homePrefix || ('/home/' + row.groupName)
      this.dialogFormTitle = this.$t('group.editGroup')
      this.dialogType = 'update'
      this.dialogFormVisible = true
    },
    // Transfer dialog
    addUp(row) {
      this.dialogTransfer = this.$t('group.userManagement')
      this.dialogTransferVisible = true
      this.transParams.groupId = row.ID
      this.transParams.nickname = row.remark
      this.$router.push({ path: '/userList', query: row })
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
      this.$refs['dialogForm'].validate(async valid => {
        if (valid) {
          this.submitLoading = true
          try {
            if (this.dialogFormData.groupType !== 'cn') {
              this.dialogFormData.groupClass = 'groupOfUniqueNames'
            }
            if (this.dialogType === 'create') {
              await groupAdd(this.dialogFormData).then(res => {
                this.judgeResult(res)
              })
            } else {
              await groupUpdate(this.dialogFormData).then(res => {
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
      this.updateLoading = false
      this.$refs['dialogForm'].resetFields()
      this.dialogFormData = {

        ID: '',
        groupName: '',
        parentId: 0,
        syncState: 1,
        groupType: '',
        groupClass: 'groupOfUniqueNames',
        gidNumber: '',
        homePrefix: '',
        remark: ''
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
        const groupIds = []
        this.multipleSelection.forEach(x => {
          groupIds.push(x.ID)
        })
        try {
          await groupDel({ groupIds: groupIds }).then(res => {
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
      this.$confirm(this.$t('group.batchSyncToLdap'), this.$t('common.notice'), {
        confirmButtonText: this.$t('common.confirm'),
        cancelButtonText: this.$t('common.cancel'),
        type: 'warning'
      }).then(async res => {
        this.loading = true
        const groupIds = []
        this.multipleSelection.forEach(x => {
          groupIds.push(x.ID)
        })
        try {
          await syncSqlGroups({ groupIds: groupIds }).then(res => {
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

    // Table multi-selection
    handleSelectionChange(val) {
      this.multipleSelection = val
    },

    // Single delete
    async singleDelete(Id) {
      this.loading = true
      try {
        await groupDel({ groupIds: [Id] }).then(res => {
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
        await syncSqlGroups({ groupIds: [Id] }).then(res => {
          this.judgeResult(res)
        })
      } finally {
        this.loading = false
      }
      this.getTableData()
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
    // treeselect
    normalizer(node) {
      return {
        id: node.ID,
        label: (node.groupType ? node.groupType + '=' : '') + node.groupName,
        // Disable posixGroup as parent node; allow ou/cn as parent
        isDisabled: node.groupClass === 'posixGroup',
        children: node.children
      }
    },
    treeselectInput(value) {
      this.treeselectValue = value
    },
    syncDingTalkDepts() {
      this.loading = true
      syncDingTalkDeptsApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncWeComDepts() {
      this.loading = true
      syncWeComDeptsApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncFeiShuDepts() {
      this.loading = true
      syncFeiShuDeptsApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
    },
    syncOpenLdapDepts() {
      this.loading = true
      syncOpenLdapDeptsApi().then(res => {
        this.judgeResult(res)
        this.loading = false
        this.getTableData()
      })
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
   .transfer-footer {
    margin-left: 20px;
    padding: 6px 5px;
  }
</style>
