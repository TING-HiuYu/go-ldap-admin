<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form
        size="mini"
        :inline="true"
        :model="params"
        class="demo-form-inline"
      >
        <el-form-item :label="$t('fieldRelation.fieldId')">
          <el-input
            v-model.trim="params.remark"
            clearable
            :placeholder="$t('common.description')"
            @keyup.enter.native="search"
            @clear="search"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-search"
            type="primary"
            @click="search"
          >{{ $t('common.search') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            icon="el-icon-plus"
            type="warning"
            @click="create"
          >{{ $t('common.add') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            :disabled="multipleSelection.length === 0"
            :loading="loading"
            icon="el-icon-delete"
            type="danger"
            @click="batchDelete"
          >{{ $t('common.batchDelete') }}</el-button>
        </el-form-item>
        <br>
      </el-form>

      <el-table
        v-loading="loading"
        :default-expand-all="true"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        row-key="ID"
        :data="infoTableData"
        border
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column
          show-overflow-tooltip
          width="52"
          sortable
          prop="ID"
          :label="$t('fieldRelation.index')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="CreatedAt"
          :label="$t('common.createdAt')"
        />
        <el-table-column
          show-overflow-tooltip
          sortable
          prop="Flag"
          :label="$t('fieldRelation.fieldId')"
        />
        <el-table-column show-overflow-tooltip sortable :label="$t('fieldRelation.fieldAttributes')">
          <template slot-scope="props">
            <el-form>
              <el-form-item>
                <span>{{ props.row.Attributes }}</span>
              </el-form-item>
            </el-form>
          </template>
        </el-table-column>
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="120">
          <template #default="scope">
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
              <el-button
                size="mini"
                icon="el-icon-edit"
                circle
                type="primary"
                @click="update(scope.row)"
              />
            </el-tooltip>
            <el-tooltip
              class="delete-popover"
              :content="$t('common.delete')"
              effect="dark"
              placement="top"
            >
              <el-popconfirm
                :title="$t('common.confirmDelete')"
                @onConfirm="singleDelete(scope.row.ID)"
              >
                <el-button
                  slot="reference"
                  size="mini"
                  icon="el-icon-delete"
                  circle
                  type="danger"
                />
              </el-popconfirm>
            </el-tooltip>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="updateLoading">
        <div class="components-container">
          <aside>{{ $t('fieldRelation.documentReference') }} <a href="http://ldapdoc.eryajf.net/pages/84953d/" target="_blank">{{ $t('fieldRelation.dynamicFieldRelationManagement') }}</a></aside>
        </div>
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="120px"
        >
          <el-form-item :label="$t('fieldRelation.type')">
            <el-checkbox-group v-model="checked">
              <el-checkbox-button
                v-for="city in cities"
                :key="city"
                :label="city"
                @change="checkbox(city)"
              >
                {{ city === 'user' ? $t('fieldRelation.userFieldDynamicRelation') : $t('fieldRelation.groupFieldDynamicRelation') }}
              </el-checkbox-button>
            </el-checkbox-group>
          </el-form-item>

          <template v-if="checked == 'user'">
            <el-form-item :label="$t('fieldRelation.typeFlag')">
              <el-select
                v-model="userVal"
                :placeholder="$t('common.pleaseSelect')"
                @change="changeUser(userVal)"
              >
                <el-option
                  v-for="item in userOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.username')" prop="username">
              <el-input
                v-model.trim="dialogFormData.username"
                :placeholder="$t('fieldRelation.usernamePinyin')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.chineseName')" prop="nickname">
              <el-input
                v-model.trim="dialogFormData.nickname"
                :placeholder="$t('fieldRelation.chineseName')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.alias')" prop="givenName">
              <el-input
                v-model.trim="dialogFormData.givenName"
                :placeholder="$t('fieldRelation.alias')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.email')" prop="mail">
              <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('fieldRelation.email')" />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.employeeId')" prop="jobNumber">
              <el-input
                v-model.trim="dialogFormData.jobNumber"
                :placeholder="$t('fieldRelation.employeeId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.phone')" prop="mobile">
              <el-input
                v-model.trim="dialogFormData.mobile"
                :placeholder="$t('fieldRelation.phone')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.avatar')" prop="avatar">
              <el-input
                v-model.trim="dialogFormData.avatar"
                :placeholder="$t('fieldRelation.avatar')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.address')" prop="postalAddress">
              <el-input
                v-model.trim="dialogFormData.postalAddress"
                :placeholder="$t('fieldRelation.address')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.position')" prop="position">
              <el-input
                v-model.trim="dialogFormData.position"
                :placeholder="$t('fieldRelation.position')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.sourceUserId')" prop="sourceUserId">
              <el-input
                v-model.trim="dialogFormData.sourceUserId"
                :placeholder="$t('fieldRelation.sourceUserId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.sourceUnionId')" prop="sourceUnionId">
              <el-input
                v-model.trim="dialogFormData.sourceUnionId"
                :placeholder="$t('fieldRelation.sourceUnionId')"
              />
            </el-form-item>
            <el-form-item :label="$t('common.remark')" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                :placeholder="$t('common.remark')"
              />
            </el-form-item>
            <!-- <el-form-item :label="$t('common.remark')" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                type="textarea"
                :placeholder="$t('common.remark')"
                :autosize="{ minRows: 3, maxRows: 6 }"
                show-word-limit
                maxlength="100"
              />
            </el-form-item> -->
          </template>
          <template v-else>
            <el-form-item :label="$t('fieldRelation.typeFlag')">
              <el-select
                v-model="groupVal"
                :placeholder="$t('common.pleaseSelect')"
                @change="changeGroup(groupVal)"
              >
                <el-option
                  v-for="item in options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.groupName')" prop="groupName">
              <el-input
                v-model.trim="dialogFormData.groupName"
                :placeholder="$t('fieldRelation.groupName')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.parentDeptId')" prop="sourceDeptParentId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptParentId"
                :placeholder="$t('fieldRelation.parentDeptId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.deptId')" prop="sourceDeptId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptId"
                :placeholder="$t('fieldRelation.deptId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.groupDescription')" prop="remark">
              <el-input
                v-model.trim="dialogFormData.remark"
                :placeholder="$t('fieldRelation.groupDescription')"
              />
            </el-form-item>
          </template>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm('A')"
          >{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>

      <!-- 编辑 -->
      <el-dialog :title="dialogFormTitle" :visible.sync="dialogFormVisible">
        <div class="components-container">
          <aside>{{ $t('fieldRelation.documentReference') }} <a href="http://ldapdoc.eryajf.net/pages/84953d/" target="_blank">{{ $t('fieldRelation.dynamicFieldRelationManagement') }}</a></aside>
        </div>
        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="120px"
        >
          <template v-if="checked == 'user'">
            <el-form-item :label="$t('fieldRelation.type')">
              <el-button type="primary">{{ $t('fieldRelation.userFieldDynamicRelation') }}</el-button>
            </el-form-item>

            <el-form-item :label="$t('fieldRelation.typeFlag')">
              <el-select
                v-model="userVal"
                :placeholder="$t('common.pleaseSelect')"
                @change="changeUser(userVal)"
              >
                <el-option
                  v-for="item in userOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.username')" prop="username">
              <el-input
                v-model.trim="dialogFormData.username"
                :placeholder="$t('fieldRelation.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.chineseName')" prop="nickname">
              <el-input
                v-model.trim="dialogFormData.nickname"
                :placeholder="$t('fieldRelation.chineseName')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.alias')" prop="givenName">
              <el-input
                v-model.trim="dialogFormData.givenName"
                :placeholder="$t('fieldRelation.alias')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.email')" prop="mail">
              <el-input v-model.trim="dialogFormData.mail" :placeholder="$t('fieldRelation.email')" />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.employeeId')" prop="jobNumber">
              <el-input
                v-model.trim="dialogFormData.jobNumber"
                :placeholder="$t('fieldRelation.employeeId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.phone')" prop="mobile">
              <el-input
                v-model.trim="dialogFormData.mobile"
                :placeholder="$t('fieldRelation.phone')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.avatar')" prop="avatar">
              <el-input
                v-model.trim="dialogFormData.avatar"
                :placeholder="$t('fieldRelation.avatar')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.address')" prop="postalAddress">
              <el-input
                v-model.trim="dialogFormData.postalAddress"
                :placeholder="$t('fieldRelation.address')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.position')" prop="position">
              <el-input
                v-model.trim="dialogFormData.position"
                :placeholder="$t('fieldRelation.position')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.sourceUserId')" prop="sourceUserId">
              <el-input
                v-model.trim="dialogFormData.sourceUserId"
                :placeholder="$t('fieldRelation.sourceUserId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.sourceUnionId')" prop="sourceUnionId">
              <el-input
                v-model.trim="dialogFormData.sourceUnionId"
                :placeholder="$t('fieldRelation.sourceUnionId')"
              />
            </el-form-item>
            <el-form-item :label="$t('common.remark')" prop="introduction">
              <el-input
                v-model.trim="dialogFormData.introduction"
                :placeholder="$t('common.remark')"
              />
            </el-form-item>
          </template>
          <template v-else>
            <el-form-item :label="$t('fieldRelation.type')">
              <el-button type="primary">{{ $t('fieldRelation.groupFieldDynamicRelation') }}</el-button>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.typeFlag')">
              <el-select
                v-model="groupVal"
                :placeholder="$t('common.pleaseSelect')"
                @change="changeGroup(groupVal)"
              >
                <el-option
                  v-for="item in options"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.groupName')" prop="groupName">
              <el-input
                v-model.trim="dialogFormData.groupName"
                :placeholder="$t('fieldRelation.groupName')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.parentDeptId')" prop="sourceDeptParentId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptParentId"
                :placeholder="$t('fieldRelation.parentDeptId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.deptId')" prop="sourceDeptId">
              <el-input
                v-model.trim="dialogFormData.sourceDeptId"
                :placeholder="$t('fieldRelation.deptId')"
              />
            </el-form-item>
            <el-form-item :label="$t('fieldRelation.groupDescription')" prop="remark">
              <el-input
                v-model.trim="dialogFormData.remark"
                :placeholder="$t('fieldRelation.groupDescription')"
              />
            </el-form-item>
          </template>
        </el-form>
        <div slot="footer" class="dialog-footer">
          <el-button size="mini" @click="cancelForm()">{{ $t('common.cancel') }}</el-button>
          <el-button
            size="mini"
            :loading="submitLoading"
            type="primary"
            @click="submitForm('B')"
          >{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
// import Treeselect from '@riophae/vue-treeselect'
// import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import {
  relationList,
  relationAdd,
  relationUp,
  relationDel
} from '@/api/personnel/fieldRelation'
import { Message } from 'element-ui'

const cityOptions = ['user', 'group']
export default {
  name: 'FieldRelation',
  components: {
    // Treeselect
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
      userVal: '',
      groupVal: '',
      updateId: '',
      checked: ['user'],
      cities: cityOptions,
      // 查询参数
      params: {
        flag: '',
        pageNum: 1,
        pageSize: 1000
      },
      // 表格数据
      tableData: [],
      infoTableData: [],
      total: 0,
      loading: false,
      // 上级目录数据
      // treeselectData: [],
      // treeselectValue: 0,
      updateLoading: false,
      // dialog对话框
      submitLoading: false,
      dialogFormTitle: '',
      dialogType: '',
      dialogFormVisible: false,
      dialogFormData: {
        username: '',
        nickname: '',
        givenName: '',
        mail: '',
        jobNumber: '',
        mobile: '',
        avatar: '',
        postalAddress: '',
        position: '',
        introduction: '',
        sourceUserId: '',
        sourceUnionId: '',
        groupName: '',
        remark: '',
        sourceDeptId: '',
        sourceDeptParentId: ''
      },
      //   dialogFromGroup: {

      //   },
      // 表格多选
      multipleSelection: []
      // typeFlag:
    }
  },
  computed: {
    options() {
      return [
        { label: this.$t('fieldRelation.feishu'), value: 'feishu_group' },
        { label: this.$t('fieldRelation.dingTalk'), value: 'dingtalk_group' },
        { label: this.$t('fieldRelation.weCom'), value: 'wecom_group' }
      ]
    },
    userOptions() {
      return [
        { label: this.$t('fieldRelation.feishu'), value: 'feishu_user' },
        { label: this.$t('fieldRelation.dingTalk'), value: 'dingtalk_user' },
        { label: this.$t('fieldRelation.weCom'), value: 'wecom_user' }
      ]
    },
    dialogFormRules() {
      return {
        sourceDeptParentId: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterParentDeptId'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        sourceDeptId: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterDeptId'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        username: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterTypeFlag'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        givenName: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        avatar: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        postalAddress: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        position: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        sourceUserId: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: ['blur', 'change'] }
        ],
        sourceUnionId: [
          { required: true, message: this.$t('group.pleaseEnterCategory'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: ['blur', 'change'] }
        ],
        groupName: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterGroupName'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('common.lengthBetween', { min: 1, max: 50 }), trigger: 'blur' }
        ],
        remark: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterDescription'), trigger: 'blur' },
          {
            min: 1,
            max: 50,
            message: this.$t('common.lengthBetween', { min: 1, max: 50 }),
            trigger: 'blur'
          }
        ],
        // mail: [
        //   { required: true, message: this.$t('fieldRelation.pleaseEnterEmail'), trigger: 'blur' },
        //   { type: 'email', message: this.$t('fieldRelation.pleaseEnterEmailAddress'), trigger: ['blur', 'change'] }
        // ],
        mail: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterEmail'), trigger: 'blur' },
          { min: 1, max: 50, message: this.$t('fieldRelation.pleaseEnterEmailAddress'), trigger: 'blur' }
        ],
        jobNumber: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterEmployeeId'), trigger: 'blur' },
          {
            min: 0,
            max: 20,
            message: this.$t('common.lengthBetween', { min: 0, max: 20 }),
            trigger: 'blur'
          }
        ],
        nickname: [
          { required: true, message: this.$t('fieldRelation.pleaseEnterNickname'), trigger: 'blur' },
          {
            min: 2,
            max: 20,
            message: this.$t('common.lengthBetween', { min: 2, max: 20 }),
            trigger: 'blur'
          }
        ],
        mobile: [{ required: true, message: this.$t('fieldRelation.pleaseEnterPhone'), trigger: 'blur' }],
        introduction: [
          { required: true, message: this.$t('common.remark'), trigger: 'blur' },
          {
            min: 0,
            max: 100,
            message: this.$t('common.lengthBetween', { min: 0, max: 100 }),
            trigger: 'blur'
          }
        ]
      }
    }
  },
  created() {
    this.getTableData()
  },
  methods: {
    checkbox(city) {
      this.checked = this.checked.includes(city) ? [city] : []
      this.value = this.city
    },
    changeUser(e) {
      this.userVal = e
    },
    changeGroup(e) {
      this.groupVal = e
    },
    // 查询
    search() {
      // 初始化表格数据
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData))
      this.infoTableData = this.deal(this.infoTableData, (node) =>
        node.Flag.includes(this.params.flag)
      )
    },
    resetData() {
      this.infoTableData = JSON.parse(JSON.stringify(this.tableData))
    },
    // 页面数据过滤
    deal(nodes, predicate) {
      if (!(nodes && nodes.length)) {
        return []
      }
      const newChildren = []
      for (const node of nodes) {
        if (predicate(node)) {
          newChildren.push(node)
          node.children = this.deal(node.children, predicate)
        } else {
          newChildren.push(...this.deal(node.children, predicate))
        }
      }
      return newChildren
    },
    // 获取表格数据
    async getTableData() {
      this.loading = true
      try {
        const { data } = await relationList(this.params)
        this.tableData = data

        this.infoTableData = JSON.parse(JSON.stringify(data))
      } finally {
        this.loading = false
      }
    },

    // 新增
    create() {
      this.checked = ['user']
      this.userVal = ''
      this.groupVal = ''
      this.dialogFormData = {}
      this.dialogFromGroup = {}
      this.dialogFormTitle = this.$t('fieldRelation.add')
      this.updateLoading = true
      this.dialogType = 'create'
    },
    // 修改
    update(row) {
      const typeDialog = row.Flag.split('_')[1]

      const {
        avatar,
        givenName,
        introduction,
        jobNumber,
        mail,
        mobile,
        nickname,
        position,
        postalAddress,
        sourceUnionId,
        sourceUserId,
        username,
        groupName,
        remark,
        sourceDeptId,
        sourceDeptParentId
      } = row.Attributes

      if (typeDialog === 'user') {
        this.updateId = row.ID
        this.checked = ['user']

        this.userVal = row.Flag
        this.dialogFormData.username = username
        this.dialogFormData.nickname = nickname
        this.dialogFormData.givenName = givenName
        this.dialogFormData.mail = mail
        this.dialogFormData.jobNumber = jobNumber
        this.dialogFormData.mobile = mobile
        this.dialogFormData.avatar = avatar
        this.dialogFormData.postalAddress = postalAddress
        this.dialogFormData.position = position
        this.dialogFormData.introduction = introduction
        this.dialogFormData.sourceUserId = sourceUserId
        this.dialogFormData.sourceUnionId = sourceUnionId
      } else {
        this.updateId = row.ID
        this.checked = ['group']
        this.groupVal = row.Flag
        this.dialogFormData.groupName = groupName
        this.dialogFormData.remark = remark
        this.dialogFormData.sourceDeptId = sourceDeptId
        this.dialogFormData.sourceDeptParentId = sourceDeptParentId
      }

      this.dialogFormTitle = this.$t('fieldRelation.edit')
      this.dialogType = 'update'
      this.dialogFormVisible = true
    },

    // 提交表单
    submitForm(e) {
      let flag, attributes
      if (this.checked && this.checked[0] === 'user') {
        if (this.userVal === '') {
          Message({
            message: this.$t('fieldRelation.pleaseSelectTypeFlag'),
            type: 'warning'
          })
          return false
        }
        flag = this.userVal
        attributes = this.dialogFormData
      } else {
        if (this.groupVal === '') {
          Message({
            message: this.$t('fieldRelation.pleaseSelectTypeFlag'),
            type: 'warning'
          })
          return false
        }
        flag = this.groupVal
        attributes = this.dialogFormData
      }
      this.$refs['dialogForm'].validate(async(valid) => {
        if (valid) {
          this.submitLoading = true
          try {
            if (this.dialogType === 'create') {
              await relationAdd({
                flag: flag,
                attributes: attributes
              })
            } else {
              await relationUp({
                id: this.updateId,
                flag: flag,
                attributes: attributes
              })
            }
          } finally {
            this.submitLoading = false
          }
          this.resetForm()
          this.getTableData()
          Message({
            showClose: true,
            message: this.$t('common.operationSuccessful'),
            type: 'success'
          })
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

    // 提交表单
    cancelForm() {
      this.resetForm()
    },

    resetForm() {
      this.dialogFormVisible = false
      this.updateLoading = false
      this.$refs['dialogForm'].resetFields()
      this.dialogFormData = {
        groupName: '',
        remark: ''
      }
    },

    // 批量删除
    batchDelete() {
      this.$confirm(this.$t('common.permanentDeleteWarning'), this.$t('common.notice'), {
        confirmButtonText: this.$t('common.confirm'),
        cancelButtonText: this.$t('common.cancel'),
        type: 'warning'
      })
        .then(async(res) => {
          this.loading = true
          const groupIds = []
          this.multipleSelection.forEach((x) => {
            groupIds.push(x.ID)
          })
          try {
            await relationDel({ fieldRelationIds: groupIds })
          } finally {
            this.loading = false
          }
          this.getTableData()
          Message({
            showClose: true,
            message: this.$t('fieldRelation.deleteSuccessful'),
            type: 'success'
          })
        })
        .catch(() => {
          Message({
            showClose: true,
            type: 'info',
            message: this.$t('common.deleteCancelled')
          })
        })
    },
    // 单个删除
    async singleDelete(Id) {
      this.loading = true
      try {
        await relationDel({ fieldRelationIds: [Id] })
      } finally {
        this.loading = false
      }
      this.getTableData()
    },

    // 表格多选
    handleSelectionChange(val) {
      this.multipleSelection = val
    },

    // 分页
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
.container-card {
  margin: 10px;
  margin-bottom: 100px;
}

.delete-popover {
  margin-left: 10px;
}
.transfer-footer {
  margin-left: 20px;
  padding: 6px 5px;
}
.demo-table-expand {
  font-size: 0;
}
.demo-table-expand label {
  width: 90px;
  color: #99a9bf;
  text-align: left !important;
}
.demo-table-expand .el-form-item {
  margin-right: 0;
  margin-bottom: 0;
  width: 50%;
}
.link-title {
  margin-left: 30px;
  margin-bottom: 10px;
}

/* .el-form-item /deep/ label{
    label{

    }

} */
</style>
