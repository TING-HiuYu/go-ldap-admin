<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" class="demo-form-inline">
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="create">{{ $t('common.add') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :disabled="multipleSelection.length === 0" :loading="loading" icon="el-icon-delete" type="danger" @click="batchDelete">{{ $t('common.batchDelete') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :tree-props="{children: 'children', hasChildren: 'hasChildren'}" row-key="ID" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip prop="title" :label="$t('menu.menuTitle')" width="150" />
        <el-table-column show-overflow-tooltip prop="name" :label="$t('common.name')" />
        <el-table-column show-overflow-tooltip prop="icon" :label="$t('menu.icon')" />
        <el-table-column show-overflow-tooltip prop="path" :label="$t('menu.routePath')" />
        <el-table-column show-overflow-tooltip prop="component" :label="$t('menu.componentPath')" />
        <el-table-column show-overflow-tooltip prop="redirect" :label="$t('menu.redirect')" />
        <el-table-column show-overflow-tooltip prop="sort" :label="$t('menu.sort')" align="center" width="80" />
        <el-table-column show-overflow-tooltip prop="status" :label="$t('common.disabled')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.status === 1 ? 'success':'danger'">{{ scope.row.status === 1 ? $t('common.no'):$t('common.yes') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="hidden" :label="$t('menu.hidden')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.hidden === 1 ? 'danger':'success'">{{ scope.row.hidden === 1 ? $t('common.yes'):$t('common.no') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="noCache" :label="$t('menu.cache')" align="center" width="80">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.noCache === 1 ? 'danger':'success'">{{ scope.row.noCache === 1 ? $t('common.no'):$t('common.yes') }}</el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column show-overflow-tooltip prop="activeMenu" label="Highlighted menu" /> -->
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="120">
          <template slot-scope="scope">
            <el-tooltip fixed :content="$t('common.edit')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="update(scope.row)" />
            </el-tooltip>
            <el-tooltip class="delete-popover" fixed :content="$t('common.delete')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmDelete')" @onConfirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible" width="580px">
        <el-form ref="dialogForm" :inline="true" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="80px">
          <el-form-item :label="$t('menu.menuTitle')" prop="title">
            <el-input v-model.trim="dialogFormData.title" :placeholder="$t('menu.menuTitleLabel')" style="width: 440px" />
          </el-form-item>
          <el-form-item :label="$t('common.name')" prop="name">
            <el-input v-model.trim="dialogFormData.name" :placeholder="$t('menu.nameLabel')" style="width: 220px" />
          </el-form-item>
          <el-form-item :label="$t('menu.sort')" prop="sort">
            <el-input-number v-model.number="dialogFormData.sort" controls-position="right" :min="1" :max="999" />
          </el-form-item>
          <el-form-item :label="$t('menu.icon')" prop="icon">
            <el-popover
              placement="bottom-start"
              width="450"
              trigger="click"
              @show="$refs['iconSelect'].reset()"
            >
              <IconSelect ref="iconSelect" @selected="selected" />
              <el-input slot="reference" v-model="dialogFormData.icon" style="width: 440px;" :placeholder="$t('menu.clickToSelectIcon')" readonly>
                <svg-icon v-if="dialogFormData.icon" slot="prefix" :icon-class="dialogFormData.icon" class="el-input__icon" style="height: 32px;width: 16px;" />
                <i v-else slot="prefix" class="el-icon-search el-input__icon" />
              </el-input>
            </el-popover>
          </el-form-item>
          <el-form-item :label="$t('menu.routePath')" prop="path">
            <el-input v-model.trim="dialogFormData.path" :placeholder="$t('menu.routePathLabel')" style="width: 440px" />
          </el-form-item>
          <el-form-item :label="$t('menu.componentPath')" prop="component">
            <el-input v-model.trim="dialogFormData.component" :placeholder="$t('menu.componentPathLabel')" style="width: 440px" />
          </el-form-item>
          <el-form-item :label="$t('menu.redirect')" prop="redirect">
            <el-input v-model.trim="dialogFormData.redirect" :placeholder="$t('menu.redirectLabel')" style="width: 440px" />
          </el-form-item>
          <el-form-item :label="$t('common.disabled')" prop="status">
            <el-radio-group v-model="dialogFormData.status">
              <el-radio-button label="Yes">{{ $t('common.yes') }}</el-radio-button>
              <el-radio-button label="No">{{ $t('common.no') }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item :label="$t('menu.hidden')" prop="hidden">
            <el-radio-group v-model="dialogFormData.hidden">
              <el-radio-button label="Yes">{{ $t('common.yes') }}</el-radio-button>
              <el-radio-button label="No">{{ $t('common.no') }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item :label="$t('menu.cache')" prop="noCache">
            <el-radio-group v-model="dialogFormData.noCache">
              <el-radio-button label="Yes">{{ $t('common.yes') }}</el-radio-button>
              <el-radio-button label="No">{{ $t('common.no') }}</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <!-- <el-form-item label="Highlighted menu" prop="activeMenu">
            <el-input v-model.trim="dialogFormData.activeMenu" placeholder="Highlighted menu(activeMenu)" style="width: 440px" />
          </el-form-item> -->
          <el-form-item :label="$t('menu.parentDirectory')" prop="parentId">
            <!-- <el-cascader
              v-model="dialogFormData.parentId"
              :show-all-levels="false"
              :options="treeselectData"
              :props="{ checkStrictly: true, label:'title', value:'ID', emitPath:false}"
              clearable
              filterable
            /> -->
            <treeselect
              v-model="dialogFormData.parentId"
              :options="treeselectData"
              :normalizer="normalizer"
              style="width:440px"
              @input="treeselectInput"
            />
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
import IconSelect from '@/components/IconSelect'
import Treeselect from '@riophae/vue-treeselect'
import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import { getMenuTree, createMenu, updateMenuById, batchDeleteMenuByIds } from '@/api/system/menu'
import { Message } from 'element-ui'

export default {
  name: 'Menu',
  components: {
    IconSelect,
    Treeselect
  },
  data() {
    return {
      // Table data
      tableData: [],
      loading: false,

      // Parent directory data
      treeselectData: [],
      treeselectValue: 0,

      // Dialog
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        ID: '',
        title: '',
        name: '',
        icon: '',
        path: '',
        component: 'Layout',
        redirect: '',
        sort: 999,
        status: 'No',
        hidden: 'No',
        noCache: 'Yes',
        alwaysShow: 2,
        breadcrumb: 1,
        // activeMenu: '',
        parentId: 0
      },
      dialogFormRules: {
        title: [
          { required: true, message: this.$t('menu.pleaseEnterTitle'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', {min: 1, max: 50}), trigger: 'blur' }
        ],
        name: [
          { required: true, message: this.$t('menu.pleaseEnterName'), trigger: 'blur' },
          { min: 1, max: 100, message: this.$t('common.lengthBetween', {min: 1, max: 100}), trigger: 'blur' }
        ],
        path: [
          { required: true, message: this.$t('menu.pleaseEnterPath'), trigger: 'blur' },
          { min: 1, max: 100, message: this.$t('common.lengthBetween', {min: 1, max: 100}), trigger: 'blur' }
        ],
        component: [
          { required: false, message: this.$t('menu.pleaseEnterComponentPath'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('common.lengthBetween', {min: 0, max: 100}), trigger: 'blur' }
        ],
        redirect: [
          { required: false, message: this.$t('menu.pleaseEnterRedirect'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('common.lengthBetween', {min: 0, max: 100}), trigger: 'blur' }
        ],
        // activeMenu: [
        //   { required: false, message: 'Please enter highlighted menu', trigger: 'blur' },
        //   { min: 0, max: 100, message: this.$t('common.lengthBetween', {min: 0, max: 100}), trigger: 'blur' }
        // ],
        parentId: [
          { required: true, message: this.$t('menu.pleaseSelectParentDirectory'), trigger: 'change' }
        ]

      },

      // Delete button popover
      popoverVisible: false,
      // Table multi-select
      multipleSelection: []
    }
  },
  created() {
    this.getTableData()
  },
  methods: {
    // Get table data
    async getTableData() {
      this.loading = true
      try {
        const { data } = await getMenuTree()

        this.tableData = data
        this.treeselectData = [{ ID: 0, title: this.$t('menu.topLevel'), children: data }]
      } finally {
        this.loading = false
      }
    },

    // Create
    create() {
      this.dialogFormTitle = this.$t('menu.addMenu')
      this.dialogType = 'create'
      this.dialogFormVisible = true
    },

    // Update
    update(row) {
      this.dialogFormData.ID = row.ID
      this.dialogFormData.title = row.title
      this.dialogFormData.name = row.name
      this.dialogFormData.icon = row.icon
      this.dialogFormData.path = row.path
      this.dialogFormData.component = row.component
      this.dialogFormData.redirect = row.redirect
      this.dialogFormData.sort = row.sort
      this.dialogFormData.status = row.status === 1 ? 'No' : 'Yes'
      this.dialogFormData.hidden = row.hidden === 1 ? 'Yes' : 'No'
      this.dialogFormData.noCache = row.noCache === 1 ? 'No' : 'Yes'
      // this.dialogFormData.activeMenu = row.activeMenu
      this.dialogFormData.parentId = row.parentId

      this.dialogFormTitle = this.$t('menu.editMenu')
      this.dialogType = 'update'
      this.dialogFormVisible = true
    },

    // Judge result
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
          if (this.dialogFormData.ID === this.dialogFormData.parentId) {
            return Message({
              showClose: true,
              message: this.$t('menu.cannotSelectSelfAsParent'),
              type: 'error'
            })
          }
          if (this.dialogFormData.component === '') {
            this.dialogFormData.component = 'Layout'
          }
          this.dialogFormData.status = this.dialogFormData.status === 'Yes' ? 2 : 1
          this.dialogFormData.hidden = this.dialogFormData.hidden === 'Yes' ? 1 : 2
          this.dialogFormData.noCache = this.dialogFormData.noCache === 'Yes' ? 2 : 1
          const dialogFormDataCopy = { ...this.dialogFormData, parentId: this.treeselectValue }
          try {
            if (this.dialogType === 'create') {
              await createMenu(dialogFormDataCopy).then(res => {
                this.judgeResult(res)
              })
            } else {
              await updateMenuById(dialogFormDataCopy).then(res => {
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
            type: 'error'
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
        title: '',
        name: '',
        icon: '',
        path: '',
        component: 'Layout',
        redirect: '',
        sort: 999,
        status: 'No',
        hidden: 'No',
        noCache: 'Yes',
        alwaysShow: 2,
        breadcrumb: 1,
        // Highlighted menu
        // activeMenu: '',
        parentId: 0
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
        const menuIds = []
        this.multipleSelection.forEach(x => {
          menuIds.push(x.ID)
        })
        try {
          await batchDeleteMenuByIds({ menuIds: menuIds }).then(res => {
            this.judgeResult(res)
          })
        } finally {
          this.loading = false
        }
        this.getTableData()
      }).catch(() => {
        Message({
          type: 'info',
          message: this.$t('common.deleteCancelled')
        })
      })
    },

    // Table multi-select
    handleSelectionChange(val) {
      this.multipleSelection = val
    },

    // Single delete
    async singleDelete(Id) {
      this.loading = true
      try {
        await batchDeleteMenuByIds({ menuIds: [Id] }).then(res => {
          this.judgeResult(res)
        })
      } finally {
        this.loading = false
      }
      this.getTableData()
    },

    // Select icon
    selected(name) {
      this.dialogFormData.icon = name
    },

    // treeselect
    normalizer(node) {
      return {
        id: node.ID,
        label: node.title,
        children: node.children
      }
    },
    treeselectInput(value) {
      this.treeselectValue = value
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
</style>
