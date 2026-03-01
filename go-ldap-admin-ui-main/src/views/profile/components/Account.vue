<template>
  <div>
    <div class="card-row">
      <user-card v-if="userCardData" :user="userCardData" class="account-card" />

      <el-card class="account-card span-two password-card">
        <div slot="header" class="clearfix header-flex">
          <span>{{ $t('account.changeAccountPassword') }}</span>
          <span class="header-tip">{{ $t('account.supportsVerification') }}</span>
        </div>

        <el-form
          ref="dialogForm"
          size="small"
          :model="dialogFormData"
          :rules="dialogFormRules"
          label-width="100px"
          class="password-form"
        >
          <div class="form-grid">
            <div class="verify-column">
              <el-form-item :label="$t('account.verificationMethod')">
                <el-radio-group v-model="verifyMethod" size="small" @change="handleVerifyChange">
                  <el-radio-button label="password">{{ $t('account.originalPassword') }}</el-radio-button>
                  <el-radio-button label="email" :disabled="!posixForm.mail">{{ $t('account.emailCode') }}</el-radio-button>
                  <el-radio-button label="oauth" :disabled="!connectors.length">OAuth</el-radio-button>
                </el-radio-group>
              </el-form-item>

              <el-form-item v-if="verifyMethod === 'password'" :label="$t('account.originalPassword')" prop="oldPassword">
                <el-input
                  v-model.trim="dialogFormData.oldPassword"
                  autocomplete="on"
                  :type="passwordTypeOld"
                  :placeholder="$t('account.pleaseEnterOriginalPassword')"
                />
                <span class="show-pwd" @click="showPwdOld">
                  <svg-icon :icon-class="passwordTypeOld === 'password' ? 'eye' : 'eye-open'" />
                </span>
              </el-form-item>

              <template v-else-if="verifyMethod === 'email'">
                <el-form-item :label="$t('changePassword.email')">
                  <el-input v-model="posixForm.mail" disabled :placeholder="$t('account.pleaseBindEmailFirst')" />
                </el-form-item>
                <el-form-item :label="$t('account.verificationCode')" prop="otpCode" class="code-item">
                  <el-input
                    v-model.trim="dialogFormData.otpCode"
                    autocomplete="off"
                    :placeholder="$t('account.pleaseEnterCode')"
                    maxlength="6"
                  />
                  <el-button class="inline-btn" :disabled="otpCountdown>0" size="mini" @click="sendPasswordCode">
                    {{ otpCountdown>0 ? $t('login.resend', {seconds: otpCountdown}) : $t('login.getCode') }}
                  </el-button>
                </el-form-item>
              </template>

              <template v-else>
                <div class="oauth-verify">
                  <div v-loading="oauthLoading" class="oauth-list">
                    <div v-for="item in connectors" :key="item.id" class="oauth-chip" @click="startOAuthVerification(item)">
                      <div class="oauth-avatar">{{ item.provider.slice(0,1).toUpperCase() }}</div>
                      <div>
                        <div class="oauth-name">{{ item.name }}</div>
                        <div class="oauth-meta">{{ item.provider }} · {{ $t('account.clickToVerify') }}</div>
                      </div>
                    </div>
                    <el-empty v-if="!connectors.length && !oauthLoading" :description="$t('account.noOauthConnectors')" />
                  </div>
                </div>
              </template>
            </div>

            <div class="password-column">
              <div class="password-grid">
                <el-form-item :label="$t('account.newPassword')" prop="newPassword">
                  <el-input
                    v-model.trim="dialogFormData.newPassword"
                    autocomplete="on"
                    :type="passwordTypeNew"
                    :placeholder="$t('account.pleaseEnterNewPassword')"
                  />
                  <span class="show-pwd" @click="showPwdNew">
                    <svg-icon :icon-class="passwordTypeNew === 'password' ? 'eye' : 'eye-open'" />
                  </span>
                </el-form-item>

                <el-form-item :label="$t('account.confirmPassword')" prop="confirmPassword">
                  <el-input
                    v-model.trim="dialogFormData.confirmPassword"
                    autocomplete="on"
                    :type="passwordTypeConfirm"
                    :placeholder="$t('account.pleaseConfirmNewPassword')"
                  />
                  <span class="show-pwd" @click="showPwdConfirm">
                    <svg-icon :icon-class="passwordTypeConfirm === 'password' ? 'eye' : 'eye-open'" />
                  </span>
                </el-form-item>
              </div>
            </div>
          </div>

          <div class="btn-row">
            <el-button :loading="submitLoading" type="primary" @click="submitForm">{{ $t('common.confirm') }}</el-button>
            <el-button @click="cancelForm">{{ $t('common.cancel') }}</el-button>
          </div>
        </el-form>
      </el-card>

      <el-card v-if="isPosixUser" class="account-card">
        <div slot="header" class="clearfix">
          <span>{{ $t('account.terminalSettings') }}</span>
        </div>

        <el-form label-width="110px" size="small">
          <el-form-item label="uid">
            <el-input v-model="posixForm.uidNumber" disabled />
          </el-form-item>
          <el-form-item label="gid">
            <el-input v-model="posixForm.gidNumber" disabled />
          </el-form-item>
          <el-form-item label="homeDir">
            <el-input v-model="posixForm.homeDirectory" disabled />
          </el-form-item>
          <el-form-item label="Shell">
            <el-input
              v-model.trim="posixForm.loginShell"
              :placeholder="$t('account.enterLoginShellExample')"
            />
          </el-form-item>
          <el-form-item>
            <el-button :loading="posixSaveLoading" type="primary" @click="savePosixShell">{{ $t('common.save') }}</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card v-if="isPosixUser" class="account-card span-two">
        <div slot="header" class="clearfix">
          <span>{{ $t('account.generateAndDownloadSshCert') }}</span>
        </div>

        <el-alert
          type="info"
          :closable="false"
          style="margin-bottom: 12px;"
          :title="$t('account.generateSshCertDescription')"
        />

        <div style="margin-bottom: 12px; display: flex; gap: 8px;">
          <el-button type="primary" size="small" :loading="issueLoading" @click="handleIssue">{{ $t('account.generate') }}</el-button>
          <el-button size="small" :disabled="!sshData.zipBase64" @click="downloadZip">{{ $t('account.download') }}</el-button>
          <el-button size="small" @click="openReadme">{{ $t('account.importGuide') }}</el-button>
        </div>

        <el-form label-width="110px" size="small">
          <el-form-item :label="$t('account.privateKey')">
            <el-input v-model="sshData.privateKey" type="textarea" :rows="4" readonly :placeholder="$t('account.shownAfterGeneration')" />
          </el-form-item>
          <el-form-item :label="$t('account.userCertificate')">
            <el-input v-model="sshData.certificate" type="textarea" :rows="3" readonly :placeholder="$t('account.shownAfterGeneration')" />
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script>
import { changePwd, issueSSHPubKey, updateUserById, getInfo, sendPasswordChangeCode } from '@/api/system/user'
import { publicConnectors, startPasswordOAuth } from '@/api/system/oauth'
import store from '@/store'
import JSEncrypt from 'jsencrypt'
import { Message } from 'element-ui'
import { saveAs } from 'file-saver'
import UserCard from './UserCard'

export default {
  components: { UserCard },
  props: {
    user: {
      type: Object,
      default: () => ({})
    }
  },
  data() {
    const confirmPass = (rule, value, callback) => {
      if (value) {
        if (this.dialogFormData.newPassword !== value) {
          callback(new Error(this.$t('account.passwordsDoNotMatch')))
        } else {
          callback()
        }
      } else {
        callback(new Error(this.$t('account.pleaseEnterNewPasswordAgain')))
      }
    }
    const verifySwitcher = (rule, value, callback) => {
      if (rule.field === 'oldPassword') {
        if (this.verifyMethod !== 'password') return callback()
        if (!this.dialogFormData.oldPassword || this.dialogFormData.oldPassword.length < 6) {
          return callback(new Error(this.$t('account.pleaseEnterOriginalPassword')))
        }
      }
      if (rule.field === 'otpCode') {
        if (this.verifyMethod !== 'email') return callback()
        if (!this.dialogFormData.otpCode || this.dialogFormData.otpCode.length !== 6) {
          return callback(new Error(this.$t('account.pleaseEnterEmailCode')))
        }
      }
      if (rule.field === 'oauthCode') {
        if (this.verifyMethod !== 'oauth') return callback()
        if (!this.dialogFormData.oauthCode) {
          return callback(new Error(this.$t('account.pleaseCompleteOauth')))
        }
      }
      callback()
    }
    return {
      submitLoading: false,
      verifyMethod: 'password',
      dialogFormData: {
        oldPassword: '',
        otpCode: '',
        oauthCode: '',
        newPassword: '',
        confirmPassword: ''
      },
      dialogFormRules: {
        oldPassword: [{ validator: verifySwitcher, trigger: 'blur' }],
        otpCode: [{ validator: verifySwitcher, trigger: 'blur' }],
        oauthCode: [{ validator: verifySwitcher, trigger: 'change' }],
        newPassword: [
          { required: true, message: this.$t('account.pleaseEnterNewPassword'), trigger: 'blur' },
          { min: 6, max: 30, message: this.$t('common.lengthBetween', { min: 6, max: 30 }), trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, validator: confirmPass, trigger: 'blur' }
        ]
      },
      publicKey: process.env.VUE_APP_PUBLIC_KEY,
      passwordTypeOld: 'password',
      passwordTypeNew: 'password',
      passwordTypeConfirm: 'password',
      otpCountdown: 0,
      otpTimer: null,
      connectors: [],
      oauthLoading: false,
      pendingStates: {},
      oauthWindow: null,
      issueLoading: false,
      isPosixUser: false,
      posixSaveLoading: false,
      posixForm: {
        ID: '',
        username: '',
        nickname: '',
        givenName: '',
        mail: '',
        jobNumber: '',
        postalAddress: '',
        departments: '',
        position: '',
        mobile: '',
        avatar: '',
        introduction: '',
        departmentId: [],
        source: '',
        roleIds: [],
        loginShell: '/bin/bash',
        uidNumber: '',
        gidNumber: '',
        homeDirectory: ''
      },
      sshData: {
        privateKey: '',
        publicKey: '',
        certificate: '',
        caPublicKey: '',
        zipBase64: '',
        zipName: ''
      }
    }
  },
  computed: {
    userCardData() {
      const name = this.user?.name || this.posixForm.nickname || this.posixForm.username || ''
      const role = this.user?.role || ''
      const avatar = this.user?.avatar || this.posixForm.avatar || ''
      return name || role || avatar ? { name, role, avatar } : null
    }
  },
  created() {
    this.loadProfile()
    this.loadConnectors()
  },
  mounted() {
    window.addEventListener('message', this.handleOAuthMessage)
  },
  beforeDestroy() {
    window.removeEventListener('message', this.handleOAuthMessage)
    this.stopCountdown()
    if (this.oauthWindow && !this.oauthWindow.closed) {
      this.oauthWindow.close()
    }
  },
  methods: {
    submitForm() {
      this.$refs['dialogForm'].validate(async valid => {
        if (valid) {
          const verification = this.getVerificationValue()
          if (!verification) {
            Message({ showClose: true, message: this.$t('account.pleaseCompleteVerification'), type: 'error' })
            return
          }
          const dialogFormDataCopy = { ...this.dialogFormData }

          const encryptor = new JSEncrypt()
          const publicKey = (this.publicKey || '').replace(/\\n/g, '\n')
          encryptor.setPublicKey(publicKey)
          const oldPasswd = encryptor.encrypt(verification)
          const newPasswd = encryptor.encrypt(this.dialogFormData.newPassword)
          const confirmPasswd = encryptor.encrypt(this.dialogFormData.confirmPassword)
          if (!oldPasswd || !newPasswd || !confirmPasswd) {
            this.$message.error(this.$t('account.invalidPublicKey'))
            this.submitLoading = false
            return
          }
          dialogFormDataCopy.oldPassword = oldPasswd
          dialogFormDataCopy.newPassword = newPasswd
          dialogFormDataCopy.confirmPassword = confirmPasswd

          this.submitLoading = true
          const { code, msg } = await changePwd(dialogFormDataCopy)

          this.submitLoading = false
          if (code !== 0) {
            return Message({
              showClose: true,
              message: msg,
              type: 'error'
            })
          }
          this.resetForm()
          Message({
            showClose: true,
            message: this.$t('account.passwordChangedRelogin'),
            type: 'success'
          })
          // 重新登录
          setTimeout(() => {
            store.dispatch('user/logout').then(() => {
              location.reload() // 为了重新实例化vue-router对象 避免bug
            })
          }, 1500)
        } else {
          this.$message({
            showClose: true,
            message: this.$t('common.formValidationFailed'),
            type: 'warn'
          })
          return false
        }
      })
    },
    cancelForm() {
      this.resetForm()
    },
    resetForm() {
      this.$refs['dialogForm'].resetFields()
      this.dialogFormData = {
        oldPassword: '',
        otpCode: '',
        oauthCode: '',
        newPassword: '',
        confirmPassword: ''
      }
      this.verifyMethod = 'password'
      this.stopCountdown()
    },
    getVerificationValue() {
      if (this.verifyMethod === 'password') return this.dialogFormData.oldPassword
      if (this.verifyMethod === 'email') return this.dialogFormData.otpCode
      if (this.verifyMethod === 'oauth') return this.dialogFormData.oauthCode
      return ''
    },
    handleVerifyChange() {
      this.dialogFormData.oldPassword = ''
      this.dialogFormData.otpCode = ''
      this.dialogFormData.oauthCode = ''
      if (this.verifyMethod !== 'email') {
        this.stopCountdown()
      }
    },
    showPwdOld() {
      this.passwordTypeOld = this.passwordTypeOld === 'password' ? '' : 'password'
    },
    showPwdNew() {
      this.passwordTypeNew = this.passwordTypeNew === 'password' ? '' : 'password'
    },
    showPwdConfirm() {
      this.passwordTypeConfirm = this.passwordTypeConfirm === 'password' ? '' : 'password'
    },
    async sendPasswordCode() {
      if (this.verifyMethod !== 'email') return
      if (!this.posixForm.mail) {
        Message({ showClose: true, message: this.$t('account.pleaseBindEmailFirst'), type: 'error' })
        return
      }
      if (this.otpCountdown > 0) return
      try {
        await sendPasswordChangeCode()
        Message({ showClose: true, message: this.$t('account.codeSent'), type: 'success' })
        this.startCountdown()
      } catch (e) {
        Message({ showClose: true, message: e?.message || this.$t('login.sendFailed'), type: 'error' })
      }
    },
    startCountdown() {
      this.stopCountdown()
      this.otpCountdown = 60
      this.otpTimer = setInterval(() => {
        if (this.otpCountdown <= 1) {
          this.stopCountdown()
        } else {
          this.otpCountdown -= 1
        }
      }, 1000)
    },
    stopCountdown() {
      if (this.otpTimer) {
        clearInterval(this.otpTimer)
        this.otpTimer = null
      }
      this.otpCountdown = 0
    },
    async startOAuthVerification(connector) {
      if (!connector || !connector.id) return
      this.oauthLoading = true
      try {
        const { data } = await startPasswordOAuth({ connectorId: connector.id, intent: 'password' })
        if (!data || !data.authUrl) {
          Message({ showClose: true, message: this.$t('account.failedGetAuthUrl'), type: 'error' })
          this.oauthLoading = false
          return
        }
        this.pendingStates[data.state] = { connectorId: connector.id }
        const features = 'width=520,height=640,menubar=no,toolbar=no'
        this.oauthWindow = window.open(data.authUrl, '_blank', features)
        if (!this.oauthWindow) {
          Message({ showClose: true, message: this.$t('account.allowPopups'), type: 'error' })
        }
      } catch (e) {
        Message({ showClose: true, message: e?.message || this.$t('account.failedStartOauthVerification'), type: 'error' })
      } finally {
        this.oauthLoading = false
      }
    },
    handleOAuthMessage(event) {
      const payload = event?.data
      if (!payload || payload.intent !== 'password' || !payload.state) return
      const pending = this.pendingStates[payload.state]
      if (!pending) return
      delete this.pendingStates[payload.state]
      if (payload.code) {
        this.dialogFormData.oauthCode = payload.code
        this.verifyMethod = 'oauth'
        Message({ showClose: true, message: this.$t('account.oauthVerificationComplete'), type: 'success' })
      }
    },
    async handleIssue() {
      this.issueLoading = true
      try {
        const { code, msg, data } = await issueSSHPubKey()
        this.issueLoading = false
        if (code !== 0) {
          return Message({ showClose: true, message: msg || this.$t('account.generationFailed'), type: 'error' })
        }
        this.sshData = { ...data }
        Message({ showClose: true, message: this.$t('account.generationSuccess'), type: 'success' })
        this.downloadZip()
      } catch (err) {
        this.issueLoading = false
        Message({ showClose: true, message: err?.message || this.$t('account.generationFailed'), type: 'error' })
      }
    },
    downloadZip() {
      if (!this.sshData.zipBase64) return
      try {
        const byteChars = atob(this.sshData.zipBase64)
        const byteNumbers = new Array(byteChars.length)
        for (let i = 0; i < byteChars.length; i++) {
          byteNumbers[i] = byteChars.charCodeAt(i)
        }
        const byteArray = new Uint8Array(byteNumbers)
        const blob = new Blob([byteArray], { type: 'application/zip' })
        const name = this.sshData.zipName || `ssh-${Date.now()}.zip`
        saveAs(blob, name)
      } catch (e) {
        Message({ showClose: true, message: this.$t('account.downloadFailed'), type: 'error' })
      }
    },
    async openReadme() {
      try {
        const res = await fetch('/static/ssh-import-readme.md')
        const text = await res.text()
        const encoded = window.btoa(unescape(encodeURIComponent(text)))
        const url = `https://markdownreader.mutantcat.org/?base64=${encoded}`
        window.open(url, '_blank')
      } catch (e) {
        Message({ showClose: true, message: this.$t('account.openGuideFailed'), type: 'error' })
      }
    },
    async loadConnectors() {
      try {
        const { data } = await publicConnectors()
        this.connectors = data || []
      } catch (e) {
        this.connectors = []
      }
    },
    async loadProfile() {
      try {
        const { code, data } = await getInfo()
        if (code === 0 && data) {
          this.isPosixUser = Boolean(data.uidNumber && data.gidNumber)
          // 预填POSIX表单信息
          this.posixForm.ID = data.ID || data.id
          this.posixForm.username = data.username
          this.posixForm.nickname = data.nickname
          this.posixForm.givenName = data.givenName
          this.posixForm.mail = data.mail
          this.posixForm.jobNumber = data.jobNumber
          this.posixForm.postalAddress = data.postalAddress
          this.posixForm.departments = data.departments
          this.posixForm.position = data.position
          this.posixForm.mobile = data.mobile
          this.posixForm.avatar = data.avatar
          this.posixForm.introduction = data.introduction
          this.posixForm.departmentId = this.normalizeDepartmentIds(data.departmentId)
          this.posixForm.source = data.source
          this.posixForm.roleIds = (data.roles || []).map(r => r.ID || r.id).filter(Boolean)
          this.posixForm.loginShell = data.loginShell || this.posixForm.loginShell || '/bin/bash'
          this.posixForm.uidNumber = data.uidNumber || ''
          this.posixForm.gidNumber = data.gidNumber || ''
          this.posixForm.homeDirectory = data.homeDirectory || ''
        }
      } catch (e) {
        this.isPosixUser = false
      }
    },
    normalizeDepartmentIds(deptId) {
      if (!deptId) return []
      if (Array.isArray(deptId)) return deptId.map(d => +d).filter(Boolean)
      if (typeof deptId === 'string') {
        return deptId.split(',').map(d => +d).filter(Boolean)
      }
      return []
    },
    async savePosixShell() {
      if (!this.posixForm.loginShell) {
        Message({ showClose: true, message: this.$t('account.pleaseEnterLoginShell'), type: 'error' })
        return
      }
      if (!this.posixForm.roleIds || !this.posixForm.roleIds.length) {
        Message({ showClose: true, message: this.$t('account.missingRoleInfo'), type: 'error' })
        return
      }
      if (!this.posixForm.departmentId || !this.posixForm.departmentId.length) {
        Message({ showClose: true, message: this.$t('account.missingDepartmentInfo'), type: 'error' })
        return
      }
      // 组装更新请求，沿用完整更新接口避免字段缺失
      const payload = {
        ID: this.posixForm.ID,
        username: this.posixForm.username,
        nickname: this.posixForm.nickname || this.posixForm.username,
        givenName: this.posixForm.givenName,
        mail: this.posixForm.mail,
        jobNumber: this.posixForm.jobNumber,
        postalAddress: this.posixForm.postalAddress,
        departments: this.posixForm.departments,
        position: this.posixForm.position,
        mobile: this.posixForm.mobile,
        avatar: this.posixForm.avatar,
        introduction: this.posixForm.introduction,
        departmentId: this.posixForm.departmentId,
        source: this.posixForm.source,
        roleIds: this.posixForm.roleIds,
        loginShell: this.posixForm.loginShell
      }
      this.posixSaveLoading = true
      try {
        const { code, msg } = await updateUserById(payload)
        if (code !== 0) {
          return Message({ showClose: true, message: msg || this.$t('account.saveFailed'), type: 'error' })
        }
        Message({ showClose: true, message: this.$t('account.saveSuccessful'), type: 'success' })
      } catch (e) {
        Message({ showClose: true, message: e?.message || this.$t('account.saveFailed'), type: 'error' })
      } finally {
        this.posixSaveLoading = false
      }
    }
  }
}
</script>

<style scoped>
.card-row {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: 16px;
  row-gap: 20px;
  align-items: stretch;
}

.account-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.password-card {
  min-height: 320px;
}

.header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.header-tip {
  font-size: 12px;
  color: #909399;
}

.password-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.form-grid {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 16px;
  align-items: flex-start;
  flex: 1;
}

.verify-column,
.password-column {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.password-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.btn-row {
  margin-top: auto;
  padding-top: 8px;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  flex-wrap: nowrap;
  width: 100%;
}

.code-item {
  display: block;
}

.inline-btn {
  margin-left: 8px;
}

.code-item .el-form-item__label {
  text-align: right;
}

.code-item ::v-deep .el-form-item__content {
  display: flex;
  align-items: center;
}

.code-item .el-input {
  flex: 1;
}

.code-item .inline-btn {
  flex-shrink: 0;
  white-space: nowrap;
}

.hint-text {
  font-size: 12px;
  color: #909399;
  margin-top: -8px;
}

.oauth-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
}

.oauth-chip {
  display: flex;
  gap: 10px;
  align-items: center;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.oauth-chip:hover {
  border-color: #409eff;
  box-shadow: 0 6px 18px rgba(64, 158, 255, 0.12);
}

.oauth-avatar {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #7bd5f5, #7879ff);
  color: #0c1234;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.oauth-name {
  font-weight: 700;
  color: #303133;
}

.oauth-meta {
  font-size: 12px;
  color: #909399;
}

.full-span {
  grid-column: 1 / -1;
}

.span-two {
  grid-column: span 2;
}

@media (max-width: 900px) {
  .span-two {
    grid-column: span 1;
  }

  .password-card {
    min-height: auto;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .password-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 1200px) {
  .card-row {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .card-row {
    grid-template-columns: 1fr;
  }
}

.account-card ::v-deep .el-card__body {
  flex: 1;
  display: flex;
  flex-direction: column;
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
