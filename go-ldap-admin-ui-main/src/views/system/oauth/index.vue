<template>
  <div>
    <el-card class="container-card" shadow="always">
      <el-form ref="searchForm" :inline="true" size="mini" :model="params" class="demo-form-inline">
        <el-form-item :label="$t('oauth.keyword')">
          <el-input v-model.trim="params.keyword" clearable :placeholder="$t('oauth.nameDescription')" @keyup.enter.native="search" @clear="search" />
        </el-form-item>
        <el-form-item :label="$t('oauth.provider')">
          <el-select v-model="params.provider" clearable :placeholder="$t('oauth.selectProvider')" @change="search" @clear="search">
            <el-option v-for="p in providers" :key="p.value" :label="p.label" :value="p.value" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('common.status')">
          <el-select v-model="params.enabled" clearable :placeholder="$t('oauth.enableStatus')" @change="search" @clear="search">
            <el-option :label="$t('common.enabled')" :value="true" />
            <el-option :label="$t('common.stopped')" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-search" type="primary" @click="search">{{ $t('common.search') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-refresh" @click="resetSearch">{{ $t('common.reset') }}</el-button>
        </el-form-item>
        <el-form-item>
          <el-button :loading="loading" icon="el-icon-plus" type="warning" @click="openCreate">{{ $t('common.add') }}</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column show-overflow-tooltip prop="name" :label="$t('common.name')" />
        <el-table-column show-overflow-tooltip prop="provider" :label="$t('oauth.provider')" width="120" />
        <el-table-column :label="$t('common.enabled')" width="100" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? $t('common.yes') : $t('common.no') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip :label="$t('oauth.defaultRoles')" min-width="200">
          <template slot-scope="scope">
            <el-tag v-for="rid in scope.row.defaultRoleIds" :key="rid" size="mini" class="role-tag">{{ roleMap[rid] || rid }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="departmentId" :label="$t('oauth.defaultDepartment')" :formatter="formatDept" />
        <el-table-column :label="$t('oauth.defaultStatus')" width="110" align="center">
          <template slot-scope="scope">
            <el-tag size="small" :type="scope.row.defaultStatus === 1 ? 'success' : 'warning'">{{ scope.row.defaultStatus === 1 ? $t('common.normal') : $t('common.disabled') }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column show-overflow-tooltip prop="defaultLoginShell" :label="$t('oauth.defaultShell')" width="140" />
        <el-table-column show-overflow-tooltip prop="creator" :label="$t('common.creator')" width="120" />
        <el-table-column show-overflow-tooltip prop="updatedAt" :label="$t('common.updatedAt')" width="180">
          <template slot-scope="scope">{{ formatDate(scope.row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column fixed="right" :label="$t('common.actions')" align="center" width="140">
          <template slot-scope="scope">
            <el-tooltip :content="$t('common.edit')" effect="dark" placement="top">
              <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="openEdit(scope.row)" />
            </el-tooltip>
            <el-tooltip :content="$t('common.delete')" effect="dark" placement="top">
              <el-popconfirm :title="$t('common.confirmDelete')" @onConfirm="handleDelete(scope.row)">
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
        :page-sizes="[5, 10, 20, 50]"
        layout="total, prev, pager, next, sizes"
        background
        style="margin-top: 12px;float:right;"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />

      <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="720px">
        <el-form ref="dialogForm" :model="dialogFormData" :rules="formRules" size="small" label-width="110px">
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item :label="$t('common.name')" prop="name">
                <el-input v-model.trim="dialogFormData.name" :placeholder="$t('oauth.connectorName')" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('oauth.provider')" prop="provider">
                <el-select v-model="dialogFormData.provider" :placeholder="$t('oauth.selectProvider')">
                  <el-option v-for="p in providers" :key="p.value" :label="p.label" :value="p.value" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item :label="$t('common.enabled')" prop="enabled">
                <el-switch v-model="dialogFormData.enabled" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('oauth.defaultStatus')" prop="defaultStatus">
                <el-radio-group v-model="dialogFormData.defaultStatus" size="small">
                  <el-radio-button :label="1">{{ $t('common.normal') }}</el-radio-button>
                  <el-radio-button :label="2">{{ $t('common.disabled') }}</el-radio-button>
                </el-radio-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item :label="$t('oauth.defaultRoles')" prop="defaultRoleIds">
                <el-select v-model="dialogFormData.defaultRoleIds" multiple filterable :placeholder="$t('oauth.selectRoles')">
                  <el-option v-for="role in roleOptions" :key="role.ID" :label="role.name" :value="role.ID" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('oauth.defaultDepartment')" prop="departmentId">
                <treeselect
                  v-model="dialogFormData.departmentId"
                  :options="groupTree"
                  :normalizer="normalizer"
                  :clearable="false"
                  :placeholder="$t('oauth.selectDepartment')"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="12">
            <el-col v-if="showShell" :span="12">
              <el-form-item :label="$t('oauth.defaultShell')" prop="defaultLoginShell">
                <el-input v-model.trim="dialogFormData.defaultLoginShell" placeholder="/bin/bash" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="$t('common.description')" prop="description">
                <el-input v-model.trim="dialogFormData.description" :placeholder="$t('oauth.briefDescription')" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-divider>{{ currentSchema ? currentSchema.title : $t('oauth.oauthConfiguration') }}</el-divider>
          <el-row :gutter="12">
            <el-col v-for="field in currentFields" :key="field.key" :span="12">
              <el-form-item :label="field.label" :prop="'config.' + field.key">
                <el-input
                  v-model.trim="dialogFormData.config[field.key]"
                  :placeholder="field.placeholder"
                  :type="field.inputType === 'password' ? 'password' : 'text'"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-divider>Webhooks</el-divider>
          <div style="margin-bottom: 12px;">
            <el-button size="mini" type="success" icon="el-icon-plus" @click="openAddWebhook">{{ $t('oauth.addWebhook') }}</el-button>
          </div>
          <el-table :data="dialogFormData.webhooks" border size="mini" style="width: 100%; margin-bottom: 12px;" :empty-text="$t('oauth.noWebhooks')">
            <el-table-column prop="url" :label="$t('oauth.targetUrl')" show-overflow-tooltip />
            <el-table-column :label="$t('oauth.events')" width="160">
              <template slot-scope="scope">
                <el-tag v-for="ev in scope.row.events" :key="ev" size="mini" style="margin-right:2px;">{{ eventLabel(ev) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('oauth.extraFields')" width="180">
              <template slot-scope="scope">
                <el-tag v-for="f in scope.row.fields" :key="f" size="mini" type="info" style="margin-right:2px;">{{ fieldLabel(f) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('common.enabled')" width="60" align="center">
              <template slot-scope="scope">
                <el-tag size="mini" :type="scope.row.enabled ? 'success' : 'info'">{{ scope.row.enabled ? $t('common.yes') : $t('common.no') }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column :label="$t('common.actions')" width="100" align="center">
              <template slot-scope="scope">
                <el-button size="mini" icon="el-icon-edit" circle type="primary" @click="openEditWebhook(scope.$index)" />
                <el-button size="mini" icon="el-icon-delete" circle type="danger" @click="removeWebhook(scope.$index)" />
              </template>
            </el-table-column>
          </el-table>
        </el-form>
        <div slot="footer">
          <el-button size="mini" @click="dialogVisible=false">{{ $t('common.cancel') }}</el-button>
          <el-button size="mini" type="primary" :loading="submitLoading" @click="submitForm">{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>

      <!-- Webhook 编辑弹窗 -->
      <el-dialog :title="$t('oauth.configureWebhook')" :visible.sync="webhookDialogVisible" width="520px" append-to-body>
        <el-form ref="webhookForm" :model="webhookFormData" :rules="webhookRules" size="small" label-width="100px">
          <el-form-item :label="$t('oauth.targetUrl')" prop="url">
            <el-input v-model.trim="webhookFormData.url" placeholder="https://example.com/webhook" />
          </el-form-item>
          <el-form-item :label="$t('common.enabled')" prop="enabled">
            <el-switch v-model="webhookFormData.enabled" />
          </el-form-item>
          <el-form-item :label="$t('common.description')" prop="description">
            <el-input v-model.trim="webhookFormData.description" :placeholder="$t('oauth.optionalDescription')" />
          </el-form-item>
          <el-form-item :label="$t('oauth.subscribeEvents')" prop="events">
            <el-checkbox-group v-model="webhookFormData.events">
              <el-checkbox v-for="ev in webhookEventOptions" :key="ev.value" :label="ev.value">{{ ev.label }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item :label="$t('oauth.extraFields')">
            <el-checkbox-group v-model="webhookFormData.fields">
              <el-checkbox v-for="f in webhookFieldOptions" :key="f.key" :label="f.key">{{ f.label }}</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button size="mini" @click="webhookDialogVisible=false">{{ $t('common.cancel') }}</el-button>
          <el-button size="mini" type="primary" @click="confirmWebhook">{{ $t('common.confirm') }}</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import Treeselect from '@riophae/vue-treeselect'
import '@riophae/vue-treeselect/dist/vue-treeselect.css'
import { listConnectors, createConnector, updateConnector, deleteConnector } from '@/api/system/oauth'
import { getRoles } from '@/api/system/role'
import { getGroupTree } from '@/api/personnel/group'
import { Message } from 'element-ui'

export default {
  name: 'OAuthManage',
  components: { Treeselect },
  data() {
    return {
      params: {
        keyword: '',
        provider: '',
        enabled: null,
        pageNum: 1,
        pageSize: 10
      },
      providers: [],
      providerSchemas: {},
      tableData: [],
      total: 0,
      loading: false,
      dialogVisible: false,
      dialogTitle: '',
      dialogType: '',
      submitLoading: false,
      roleOptions: [],
      roleMap: {},
      groupTree: [],
      groupMap: {},
      dialogFormData: {
        id: '',
        name: '',
        provider: '',
        enabled: true,
        defaultRoleIds: [],
        defaultStatus: 1,
        departmentId: null,
        defaultLoginShell: '/bin/bash',
        description: '',
        config: {},
        webhooks: []
      },
      baseRules: {
        name: [{ required: true, message: this.$t('oauth.pleaseEnterName'), trigger: 'blur' }],
        provider: [{ required: true, message: this.$t('oauth.pleaseSelectProvider'), trigger: 'change' }],
        defaultRoleIds: [{ type: 'array', required: true, message: this.$t('oauth.pleaseSelectDefaultRoles'), trigger: 'change' }],
        departmentId: [{ required: true, message: this.$t('oauth.pleaseSelectDefaultDepartment'), trigger: 'change' }],
        defaultStatus: [{ required: true, message: this.$t('oauth.pleaseSelectDefaultStatus'), trigger: 'change' }]
      },
      dialogRules: {},
      groupNodeMap: {},
      // Webhook 相关
      webhookEventOptions: [],
      webhookFieldOptions: [],
      webhookDialogVisible: false,
      webhookEditIndex: -1, // -1 表示新增，>=0 表示编辑
      webhookFormData: {
        url: '',
        events: [],
        fields: [],
        enabled: true,
        description: ''
      },
      webhookRules: {
        url: [{ required: true, message: this.$t('oauth.pleaseEnterWebhookUrl'), trigger: 'blur' }],
        events: [{ type: 'array', required: true, message: this.$t('oauth.pleaseSelectAtLeastOneEvent'), trigger: 'change' }]
      }
    }
  },
  computed: {
    formRules() {
      return { ...this.baseRules, ...this.dialogRules }
    },
    currentSchema() {
      return this.providerSchemas[this.dialogFormData.provider] || null
    },
    currentFields() {
      return this.currentSchema?.fields || []
    },
    showShell() {
      const node = this.groupNodeMap[this.dialogFormData.departmentId]
      return node?.groupClass === 'posixGroup'
    }
  },
  watch: {
    'dialogFormData.provider'(val) {
      this.applyProviderDefaults(val, this.dialogFormData.config)
    },
    'dialogFormData.departmentId'(val) {
      const node = this.groupNodeMap[val]
      if (!node || node.groupClass !== 'posixGroup') {
        this.dialogFormData.defaultLoginShell = ''
      } else if (!this.dialogFormData.defaultLoginShell) {
        this.dialogFormData.defaultLoginShell = '/bin/bash'
      }
    }
  },
  created() {
    this.getTableData()
    this.loadRoles()
    this.loadGroups()
  },
  methods: {
    getDefaultProvider() {
      if (this.providers && this.providers.length > 0) {
        return this.providers[0].value
      }
      return 'github'
    },
    encodeProviderConfig(configObj) {
      try {
        const json = JSON.stringify(configObj || {})
        const encoded = btoa(unescape(encodeURIComponent(json)))
        return { encoded }
      } catch (e) {
        return { encoded: '' }
      }
    },
    decodeProviderConfig(raw) {
      if (!raw) return {}
      if (raw.encoded) {
        try {
          const json = decodeURIComponent(escape(atob(raw.encoded)))
          return JSON.parse(json)
        } catch (e) {
          return {}
        }
      }
      return raw || {}
    },
    normalizeDefaultValue(value) {
      if (typeof value !== 'string') return value
      if (value.startsWith('/') && typeof window !== 'undefined' && window.location && window.location.origin) {
        return `${window.location.origin}${value}`
      }
      return value
    },
    buildConfigWithDefaults(providerId, incoming = {}) {
      const schema = this.providerSchemas[providerId] || {}
      const fields = Array.isArray(schema.fields) ? schema.fields : []
      if (!fields.length) return { ...incoming }
      const result = {}
      fields.forEach(field => {
        const val = incoming[field.key]
        if (val !== undefined && val !== null && val !== '') {
          result[field.key] = val
          return
        }
        const def = this.normalizeDefaultValue(field.defaultValue)
        result[field.key] = def !== undefined ? def : ''
      })
      return result
    },
    getEmptyForm() {
      const provider = this.getDefaultProvider()
      return {
        id: '',
        name: '',
        provider,
        enabled: true,
        defaultRoleIds: [],
        defaultStatus: 1,
        departmentId: null,
        defaultLoginShell: '/bin/bash',
        description: '',
        config: this.buildConfigWithDefaults(provider),
        webhooks: []
      }
    },
    normalizer(node) {
      return {
        id: node.ID,
        label: node.name || node.title || node.label || node.groupName || node.remark,
        children: node.children
      }
    },
    async loadRoles() {
      try {
        const { data } = await getRoles({ pageNum: 1, pageSize: 500 })
        this.roleOptions = data?.roles || []
        this.roleMap = this.roleOptions.reduce((acc, cur) => { acc[cur.ID] = cur.name; return acc }, {})
      } catch (e) {
        this.roleOptions = []
      }
    },
    async loadGroups() {
      try {
        const { data } = await getGroupTree()
        this.groupTree = data || []
        const map = {}
        this.groupNodeMap = {}
        const walk = nodes => {
          (nodes || []).forEach(n => {
            map[n.ID] = n.name || n.title || n.label || n.groupName || n.remark
            this.groupNodeMap[n.ID] = n
            if (n.children && n.children.length) walk(n.children)
          })
        }
        walk(this.groupTree)
        this.groupMap = map
      } catch (e) {
        this.groupTree = []
        this.groupMap = {}
      }
    },
    refreshFieldRules(providerId) {
      const schema = this.providerSchemas[providerId] || {}
      const fields = Array.isArray(schema.fields) ? schema.fields : []
      const rules = {}
      fields.forEach(f => {
        if (f && f.required) {
          rules['config.' + f.key] = [{ required: true, message: this.$t('oauth.pleaseEnterField', { label: f.label }), trigger: 'blur' }]
        }
      })
      this.dialogRules = rules
    },
    applyProviderDefaults(providerId, incomingConfig) {
      this.dialogFormData.config = this.buildConfigWithDefaults(providerId, incomingConfig)
      this.refreshFieldRules(providerId)
    },
    formatDept(row) {
      if (!row.departmentId) return '-'
      return this.groupMap[row.departmentId] || row.departmentId
    },
    formatDate(val) {
      if (!val) return ''
      return new Date(val).toLocaleString()
    },
    search() {
      this.params.pageNum = 1
      this.getTableData()
    },
    resetSearch() {
      this.params = { keyword: '', provider: '', enabled: null, pageNum: 1, pageSize: this.params.pageSize }
      this.getTableData()
    },
    async getTableData() {
      this.loading = true
      try {
        const { data } = await listConnectors(this.params)
        this.tableData = data?.list || []
        this.total = data?.total || 0
        if (Array.isArray(data?.providers) && data.providers.length > 0) {
          this.providers = data.providers
        }
        if (Array.isArray(data?.schemas)) {
          this.providerSchemas = data.schemas.reduce((acc, cur) => {
            if (cur && cur.id) acc[cur.id] = cur
            return acc
          }, {})
        }
        if (Array.isArray(data?.webhookEvents)) {
          this.webhookEventOptions = data.webhookEvents
        }
        if (Array.isArray(data?.webhookFields)) {
          this.webhookFieldOptions = data.webhookFields
        }
        if (!this.dialogFormData.provider) {
          const provider = this.getDefaultProvider()
          this.dialogFormData.provider = provider
          this.dialogFormData.config = this.buildConfigWithDefaults(provider, this.dialogFormData.config)
          this.refreshFieldRules(provider)
        }
      } finally {
        this.loading = false
      }
    },
    handleSizeChange(val) {
      this.params.pageSize = val
      this.getTableData()
    },
    handleCurrentChange(val) {
      this.params.pageNum = val
      this.getTableData()
    },
    openCreate() {
      this.dialogTitle = this.$t('oauth.addConnector')
      this.dialogType = 'create'
      this.dialogFormData = this.getEmptyForm()
      this.applyProviderDefaults(this.dialogFormData.provider, this.dialogFormData.config)
      this.dialogVisible = true
    },
    openEdit(row) {
      this.dialogTitle = this.$t('oauth.editConnector')
      this.dialogType = 'update'
      const cfg = this.decodeProviderConfig(row.config || {})
      this.dialogFormData = {
        id: row.id,
        name: row.name,
        provider: row.provider,
        enabled: row.enabled,
        defaultRoleIds: row.defaultRoleIds || [],
        defaultStatus: row.defaultStatus,
        departmentId: row.departmentId,
        defaultLoginShell: row.defaultLoginShell,
        description: row.description,
        config: this.buildConfigWithDefaults(row.provider, cfg),
        webhooks: (row.webhooks || []).map(wh => ({
          id: wh.id,
          url: wh.url,
          events: wh.events || [],
          fields: wh.fields || [],
          enabled: wh.enabled,
          description: wh.description || ''
        }))
      }
      this.refreshFieldRules(row.provider)
      this.dialogVisible = true
    },
    async submitForm() {
      this.$refs.dialogForm.validate(async valid => {
        if (!valid) return
        this.submitLoading = true
        const payload = { ...this.dialogFormData }
        payload.config = this.encodeProviderConfig(this.dialogFormData.config)
        // 确保 webhooks 传递给后端
        payload.webhooks = (this.dialogFormData.webhooks || []).map(wh => ({
          id: wh.id || 0,
          url: wh.url,
          events: wh.events || [],
          fields: wh.fields || [],
          enabled: wh.enabled,
          description: wh.description || ''
        }))
        try {
          if (this.dialogType === 'create') {
            await createConnector(payload)
          } else {
            await updateConnector(payload)
          }
          Message({ showClose: true, message: this.$t('common.operationSuccessful'), type: 'success' })
          this.dialogVisible = false
          this.getTableData()
        } catch (e) {
          Message({ showClose: true, message: e?.message || this.$t('common.operationFailed'), type: 'error' })
        } finally {
          this.submitLoading = false
        }
      })
    },
    async handleDelete(row) {
      if (!row?.id) return
      this.loading = true
      try {
        await deleteConnector({ ids: [row.id] })
        Message({ showClose: true, message: this.$t('oauth.deleteSuccessful'), type: 'success' })
        this.getTableData()
      } catch (e) {
        Message({ showClose: true, message: e?.message || this.$t('common.operationFailed'), type: 'error' })
      } finally {
        this.loading = false
      }
    },
    // ── Webhook 相关方法 ──
    eventLabel(val) {
      const opt = this.webhookEventOptions.find(e => e.value === val)
      return opt ? opt.label : val
    },
    fieldLabel(key) {
      const opt = this.webhookFieldOptions.find(f => f.key === key)
      return opt ? opt.label : key
    },
    getEmptyWebhook() {
      return { url: '', events: [], fields: [], enabled: true, description: '' }
    },
    openAddWebhook() {
      this.webhookEditIndex = -1
      this.webhookFormData = this.getEmptyWebhook()
      this.webhookDialogVisible = true
    },
    openEditWebhook(index) {
      this.webhookEditIndex = index
      const wh = this.dialogFormData.webhooks[index]
      this.webhookFormData = {
        url: wh.url,
        events: [...(wh.events || [])],
        fields: [...(wh.fields || [])],
        enabled: wh.enabled,
        description: wh.description || ''
      }
      this.webhookDialogVisible = true
    },
    removeWebhook(index) {
      this.dialogFormData.webhooks.splice(index, 1)
    },
    confirmWebhook() {
      this.$refs.webhookForm.validate(valid => {
        if (!valid) return
        const item = { ...this.webhookFormData }
        if (this.webhookEditIndex >= 0) {
          // 编辑：保留原有 id
          const orig = this.dialogFormData.webhooks[this.webhookEditIndex]
          item.id = orig?.id || 0
          this.$set(this.dialogFormData.webhooks, this.webhookEditIndex, item)
        } else {
          this.dialogFormData.webhooks.push(item)
        }
        this.webhookDialogVisible = false
      })
    }
  }
}
</script>

<style scoped>
.role-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}
</style>
