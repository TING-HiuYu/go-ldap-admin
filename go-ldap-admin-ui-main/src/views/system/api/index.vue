<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form size="mini" :inline="true" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('api.accessPath')">
          <el-input v-model.trim="params.path" clearable :placeholder="$t('api.accessPath')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('api.category')">
          <el-input v-model.trim="params.category" clearable :placeholder="$t('api.category')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('api.requestMethod')">
          <el-select v-model.trim="params.method" clearable :placeholder="$t('api.requestMethod')" @change="search" @clear="search">
            <el-option :label="$t('api.getResource')" value="GET" />
            <el-option :label="$t('api.postResource')" value="POST" />
            <el-option :label="$t('api.putResource')" value="PUT" />
            <el-option :label="$t('api.patchResource')" value="PATCH" />
            <el-option :label="$t('api.deleteResource')" value="DELETE" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('common.creator')">
          <el-input v-model.trim="params.creator" clearable :placeholder="$t('common.creator')" @keyup.enter.native="search" @clear="search" />
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
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column show-overflow-tooltip sortable prop="path" :label="$t('api.accessPath')" />
        <el-table-column show-overflow-tooltip sortable prop="category" :label="$t('api.category')" />
        <el-table-column show-overflow-tooltip sortable prop="method" :label="$t('api.requestMethod')" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.method | methodTagFilter" disable-transitions>{{ scope.row.method }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip sortable prop="creator" :label="$t('common.creator')" />
        <el-table-column show-overflow-tooltip sortable prop="remark" :label="$t('common.remark')" />
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="120">
          <template slot-scope="scope">
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="update(scope.row)" />
            </el-tooltip>
            <el-tooltip class="delete-popover" :content="$t('common.delete')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmDelete')" @onConfirm="singleDelete(scope.row.ID)">
                <el-button slot="reference" size="mini" icon="el-icon-delete" circle type="danger" />
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

      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible">
        <el-form ref="dialogForm" size="small" :model="dialogFormData" :rules="dialogFormRules" label-width="120px">
          <el-form-item :label="$t('api.accessPath')" prop="path">
            <el-input v-model.trim="dialogFormData.path" :placeholder="$t('api.accessPath')" />
          </el-form-item>
          <el-form-item :label="$t('api.category')" prop="category">
            <el-input v-model.trim="dialogFormData.category" :placeholder="$t('api.category')" />
          </el-form-item>
          <el-form-item :label="$t('api.requestMethod')" prop="method">
            <el-select v-model.trim="dialogFormData.method" :placeholder="$t('api.pleaseSelectRequestMethod')">
              <el-option :label="$t('api.getResource')" value="GET" />
              <el-option :label="$t('api.postResource')" value="POST" />
              <el-option :label="$t('api.putResource')" value="PUT" />
              <el-option :label="$t('api.patchResource')" value="PATCH" />
              <el-option :label="$t('api.deleteResource')" value="DELETE" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('common.remark')" prop="remark">
            <el-input v-model.trim="dialogFormData.remark" type="textarea" :placeholder="$t('common.remark')" show-word-limit maxlength="100" />
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
import { getApis, createApi, updateApiById, batchDeleteApiByIds } from '@/api/system/api'
import { Message } from 'element-ui'

export default {
  name: 'Api',
  filters: {
    methodTagFilter(val) {
      if (val === 'GET') {
        return ''
      } else if (val === 'POST') {
        return 'success'
      } else if (val === 'PUT') {
        return 'info'
      } else if (val === 'PATCH') {
        return 'warning'
      } else if (val === 'DELETE') {
        return 'danger'
      } else {
        return 'info'
      }
    }
  },
  data() {
    return {
      // Query parameters
      params: {
        path: '',
        method: '',
        category: '',
        creator: '',
        pageNum: 1,
        pageSize: 10
      },
      // Table data
      tableData: [],
      total: 0,
      loading: false,

      // Dialog
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        ID: '',
        path: '',
        category: '',
        method: '',
        remark: ''
      },
      dialogFormRules: {
        path: [
          { required: true, message: this.$t('api.pleaseEnterAccessPath'), trigger: 'blur' },
          { min: 1, max: 100, message: this.$t('common.lengthBetween', {min: 1, max: 100}), trigger: 'blur' }
        ],
        category: [
          { required: true, message: this.$t('api.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', {min: 1, max: 50}), trigger: 'blur' }
        ],
        method: [
          { required: true, message: this.$t('api.pleaseSelectRequestMethod'), trigger: 'change' }
        ],
        remark: [
          { required: false, message: this.$t('common.remark'), trigger: 'blur' },
          { min: 0, max: 100, message: this.$t('common.lengthBetween', {min: 0, max: 100}), trigger: 'blur' }
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
    // Search
    search() {
      this.params.pageNum = 1
      this.getTableData()
    },

    // Get table data
    async getTableData() {
      this.loading = true
      try {
        const { data } = await getApis(this.params)
        this.tableData = data.apis
        this.total = data.total
      } finally {
        this.loading = false
      }
    },

    // Create
    create() {
      this.dialogFormTitle = this.$t('api.addApi')
      this.dialogType = 'create'
      this.dialogFormVisible = true
    },

    // Update
    update(row) {
      this.dialogFormData.ID = row.ID
      this.dialogFormData.path = row.path
      this.dialogFormData.category = row.category
      this.dialogFormData.method = row.method
      this.dialogFormData.remark = row.remark

      this.dialogFormTitle = this.$t('api.editApi')
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
          try {
            if (this.dialogType === 'create') {
              await createApi(this.dialogFormData).then(res => {
                this.judgeResult(res)
              })
            } else {
              await updateApiById(this.dialogFormData).then(res => {
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
        ID: '',
        path: '',
        category: '',
        method: '',
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
        const apiIds = []
        this.multipleSelection.forEach(x => {
          apiIds.push(x.ID)
        })
        try {
          await batchDeleteApiByIds({ apiIds: apiIds }).then(res => {
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

    // Table multi-select
    handleSelectionChange(val) {
      this.multipleSelection = val
    },

    // Single delete
    async singleDelete(Id) {
      this.loading = true
      try {
        await batchDeleteApiByIds({ apiIds: [Id] }).then(res => {
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
