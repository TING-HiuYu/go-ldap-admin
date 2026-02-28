<template>
  <div class="login-shell" :style="{ backgroundImage: `url(${imgSrc})` }">
    <div class="login-overlay" />
    <div class="login-body">
      <div class="brand-card">
        <div class="brand-title">Go-Ldap-Admin</div>
        <p class="brand-sub">{{ $t('login.slogan') }}</p>
        <div class="brand-metrics">
          <span class="pill">{{ $t('login.securityEmailOtp') }}</span>
          <span class="pill">GitHub OAuth</span>
          <span class="pill">{{ $t('login.casbinAuth') }}</span>
        </div>
      </div>

      <div class="auth-card">
        <div class="auth-header">
          <div class="mode-switch">
            <span :class="['mode-item', { active: loginMode === 'password' }]" @click="switchMode('password')">{{ $t('login.accountPassword') }}</span>
            <span :class="['mode-item', { active: loginMode === 'otp' }]" @click="switchMode('otp')">{{ $t('login.emailCode') }}</span>
            <span :class="['mode-item', { active: loginMode === 'oauth' }]" @click="switchMode('oauth')">{{ $t('login.oauthLogin') }}</span>
          </div>
          <p class="auth-hint">{{ modeHint }}</p>
        </div>

        <el-form
          v-if="loginMode === 'password'"
          ref="loginForm"
          :model="loginForm"
          :rules="loginRules"
          class="auth-form"
          autocomplete="on"
          label-position="left"
        >
          <el-form-item prop="username">
            <span class="svg-container"><svg-icon icon-class="user" /></span>
            <el-input
              ref="username"
              v-model="loginForm.username"
              :placeholder="$t('login.username')"
              name="username"
              type="text"
              tabindex="1"
              autocomplete="on"
            />
          </el-form-item>

          <el-tooltip v-model="capsTooltip" :content="$t('login.capsLockOn')" placement="right" manual>
            <el-form-item prop="password">
              <span class="svg-container"><svg-icon icon-class="password" /></span>
              <el-input
                :key="passwordType"
                ref="password"
                v-model="loginForm.password"
                :type="passwordType"
                :placeholder="$t('login.password')"
                name="password"
                tabindex="2"
                autocomplete="on"
                @keyup.native="checkCapslock"
                @blur="capsTooltip = false"
                @keyup.enter.native="handlePasswordLogin"
              />
              <span class="show-pwd" @click="showPwd">
                <svg-icon :icon-class="passwordType === 'password' ? 'eye' : 'eye-open'" />
              </span>
            </el-form-item>
          </el-tooltip>

          <div class="form-actions">
            <el-link type="primary" class="fixed-link" @click="changePass"><span class="link-label">{{ $t('login.forgotPassword') }}</span></el-link>
            <el-button :loading="loading" type="primary" class="full-btn" @click.native.prevent="handlePasswordLogin">{{ $t('login.login') }}</el-button>
          </div>
        </el-form>

        <el-form
          v-else-if="loginMode === 'otp'"
          ref="otpForm"
          :model="otpForm"
          :rules="otpRules"
          class="auth-form"
          label-position="left"
        >
          <el-form-item prop="mail">
            <span class="svg-container"><svg-icon icon-class="email" /></span>
            <el-input v-model.trim="otpForm.mail" :placeholder="$t('login.enterEmail')" autocomplete="on" />
          </el-form-item>
          <el-form-item prop="code">
            <span class="svg-container"><svg-icon icon-class="password" /></span>
            <el-input v-model.trim="otpForm.code" :placeholder="$t('login.enterSixDigitCode')" maxlength="6" />
            <el-button class="code-btn" :disabled="otpCountdown>0" @click="sendOtpCode">
              {{ otpCountdown>0 ? $t('login.resend', {seconds: otpCountdown}) : $t('login.getCode') }}
            </el-button>
          </el-form-item>
          <div class="form-actions">
            <el-link type="primary" class="fixed-link" @click="switchMode('password')"><span class="link-label">{{ $t('login.usePasswordInstead') }}</span></el-link>
            <el-button :loading="loading" type="primary" :class="['full-btn', { danger: isResetFlow }]" @click.native.prevent="handleOtpLogin">{{ otpButtonText }}</el-button>
          </div>
        </el-form>

        <div v-else class="oauth-panel">
          <div v-loading="oauthLoading" class="oauth-grid">
            <div v-for="item in connectors" :key="item.id" class="oauth-card" @click="startOAuthLogin(item)">
              <div class="avatar" :class="item.provider">
                {{ item.provider.slice(0,1).toUpperCase() }}
              </div>
              <div>
                <div class="name">{{ item.name }}</div>
                <div class="meta">{{ item.provider }} · {{ item.enabled ? $t('login.enabled') : $t('login.disabled') }}</div>
              </div>
            </div>
            <el-empty v-if="!connectors.length && !oauthLoading" :description="$t('login.noOauthConnectors')" />
          </div>
          <div class="form-actions">
            <el-link type="primary" @click="switchMode('password')">{{ $t('login.backToPassword') }}</el-link>
          </div>
        </div>

        <div v-if="loginMode!=='oauth' && connectors.length" class="oauth-inline">
          <span class="inline-label">{{ $t('login.quickOauth') }}</span>
          <el-button
            v-for="item in connectors"
            :key="item.id"
            size="mini"
            plain
            class="inline-oauth-btn"
            @click="switchToOAuth(item)"
          >
            {{ item.name }}
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import JSEncrypt from 'jsencrypt'
import { sendLoginOtp, otpLogin } from '@/api/system/base'
import { publicConnectors, startOAuth } from '@/api/system/oauth'
import { setToken } from '@/utils/auth'

export default {
  name: 'Login',
  data() {
    const validatePassword = (rule, value, callback) => {
      if (!value || value.length < 6) {
        callback(new Error(this.$t('login.passwordMinLength')))
      } else {
        callback()
      }
    }
    return {
      imgSrc: require('@/assets/backgd-image/login.jpeg'),
      loginMode: 'password',
      isResetFlow: false,
      loginForm: {
        username: '',
        password: ''
      },
      otpForm: {
        mail: '',
        code: ''
      },
      loginRules: {
        username: [{ required: true, trigger: 'submit', message: this.$t('login.pleaseEnterUsername') }],
        password: [{ required: true, trigger: 'submit', validator: validatePassword }]
      },
      otpRules: {
        mail: [{ required: true, trigger: 'submit', message: this.$t('login.enterEmail') }],
        code: [{ required: true, trigger: 'submit', message: this.$t('login.pleaseEnterCode'), min: 6, max: 6 }]
      },
      passwordType: 'password',
      publicKey: process.env.VUE_APP_PUBLIC_KEY,
      capsTooltip: false,
      loading: false,
      redirect: undefined,
      otherQuery: {},
      otpCountdown: 0,
      otpTimer: null,
      connectors: [],
      oauthLoading: false,
      pendingStates: {},
      oauthWindow: null
    }
  },
  computed: {
    modeHint() {
      if (this.loginMode === 'password') return ''
      if (this.loginMode === 'otp') return ''
      return this.$t('login.selectConnector')
    },
    otpButtonText() {
      return this.isResetFlow ? this.$t('login.resetPasswordAndLogin') : this.$t('login.login')
    }
  },
  watch: {
    $route: {
      handler(route) {
        const query = route.query
        if (query) {
          this.redirect = query.redirect
          this.otherQuery = this.getOtherQuery(query)
        }
      },
      immediate: true
    }
  },
  mounted() {
    this.prefillFocus()
    this.fetchConnectors()
    window.addEventListener('message', this.handleOAuthMessage)
  },
  beforeDestroy() {
    window.removeEventListener('message', this.handleOAuthMessage)
    if (this.otpTimer) {
      clearInterval(this.otpTimer)
      this.otpTimer = null
    }
    if (this.oauthWindow && !this.oauthWindow.closed) {
      this.oauthWindow.close()
    }
  },
  methods: {
    prefillFocus() {
      this.$nextTick(() => {
        if (this.loginForm.username === '' && this.$refs.username) {
          this.$refs.username.focus()
        } else if (this.loginForm.password === '' && this.$refs.password) {
          this.$refs.password.focus()
        }
      })
    },
    switchMode(mode) {
      this.loginMode = mode
      if (mode !== 'otp') {
        this.isResetFlow = false
      }
      if (mode === 'password') {
        this.$nextTick(() => this.prefillFocus())
      }
    },
    switchToOAuth(connector) {
      this.loginMode = 'oauth'
      this.$nextTick(() => this.startOAuthLogin(connector))
    },
    checkCapslock(e) {
      const { key } = e
      this.capsTooltip = key && key.length === 1 && (key >= 'A' && key <= 'Z')
    },
    showPwd() {
      this.passwordType = this.passwordType === 'password' ? '' : 'password'
      this.$nextTick(() => {
        if (this.$refs.password) this.$refs.password.focus()
      })
    },
    changePass() {
      this.isResetFlow = true
      this.switchMode('otp')
    },
    handlePasswordLogin() {
      this.$refs.loginForm.validate(valid => {
        if (!valid) return
        this.loading = true
        const encryptor = new JSEncrypt()
        const publicKey = (this.publicKey || '').replace(/\\n/g, '\n')
        encryptor.setPublicKey(publicKey)
        const encPassword = encryptor.encrypt(this.loginForm.password)
        if (!encPassword) {
          this.loading = false
          this.$message.error(this.$t('login.invalidPublicKey'))
          return
        }
        const payload = { username: this.loginForm.username, password: encPassword }
        this.$store.dispatch('user/login', payload)
          .then(() => {
            this.afterLogin()
          })
          .catch(() => {
            this.loading = false
          })
      })
    },
    sendOtpCode() {
      if (this.otpCountdown > 0) return
      if (!this.otpForm.mail) {
        this.$message.warning(this.$t('login.pleaseEnterEmailFirst'))
        return
      }
      this.$refs.otpForm.validateField('mail', async valid => {
        if (valid) return
        try {
          await sendLoginOtp({ mail: this.otpForm.mail })
          this.$message.success(this.$t('login.codeSent'))
          this.startCountdown()
        } catch (e) {
          this.$message.error(e?.message || this.$t('login.sendFailed'))
        }
      })
    },
    startCountdown() {
      this.otpCountdown = 60
      if (this.otpTimer) clearInterval(this.otpTimer)
      this.otpTimer = setInterval(() => {
        if (this.otpCountdown <= 1) {
          clearInterval(this.otpTimer)
          this.otpTimer = null
          this.otpCountdown = 0
        } else {
          this.otpCountdown -= 1
        }
      }, 1000)
    },
    handleOtpLogin() {
      this.$refs.otpForm.validate(async valid => {
        if (!valid) return
        this.loading = true
        try {
          const payload = { mail: this.otpForm.mail, code: this.otpForm.code }
          if (this.isResetFlow) payload.reset = true
          const { data } = await otpLogin(payload)
          const token = data?.token
          if (!token) {
            this.$message.error(this.$t('login.loginFailedNoToken'))
            this.loading = false
            return
          }
          this.applyToken(token)
          if (this.isResetFlow && data?.newPassword) {
            // Save new password to sessionStorage, show dialog after entering main UI
            sessionStorage.setItem('generatedPassword', data.newPassword)
          }
          this.afterLogin()
        } catch (e) {
          this.$message.error(e?.message || this.$t('login.loginFailed'))
          this.loading = false
        }
      })
    },
    fetchConnectors() {
      this.oauthLoading = true
      publicConnectors()
        .then(({ data }) => {
          this.connectors = data || []
        })
        .finally(() => {
          this.oauthLoading = false
        })
    },
    async startOAuthLogin(connector) {
      if (!connector || !connector.id) return
      this.oauthLoading = true
      try {
        const { data } = await startOAuth({ connectorId: connector.id, intent: 'login' })
        if (!data || !data.authUrl) {
          this.$message.error(this.$t('login.failedGetAuthUrl'))
          this.oauthLoading = false
          return
        }
        this.pendingStates[data.state] = { connectorId: connector.id }
        const features = 'width=520,height=640,menubar=no,toolbar=no'
        this.oauthWindow = window.open(data.authUrl, '_blank', features)
        if (!this.oauthWindow) {
          this.$message.error(this.$t('login.allowPopups'))
        }
      } catch (e) {
        this.$message.error(e?.message || this.$t('login.failedStartAuth'))
      } finally {
        this.oauthLoading = false
      }
    },
    handleOAuthMessage(event) {
      const payload = event?.data
      if (!payload || !payload.state || !payload.intent) return
      const pending = this.pendingStates[payload.state]
      if (!pending) return
      delete this.pendingStates[payload.state]
      if (payload.intent === 'login' && payload.token) {
        this.applyToken(payload.token)
        // If user was newly created, backend returns the generated random password
        if (payload.generatedPassword) {
          sessionStorage.setItem('generatedPassword', payload.generatedPassword)
        }
        this.$message.success(this.$t('login.oauthLoginSuccess'))
        this.afterLogin()
      } else {
        this.$message.warning(this.$t('login.noCredentials'))
      }
    },
    applyToken(token) {
      this.$store.commit('user/SET_TOKEN', token)
      setToken(token)
    },
    afterLogin() {
      this.$router.push({ path: this.redirect || '/', query: this.otherQuery })
      this.loading = false
      this.isResetFlow = false
    },
    getOtherQuery(query) {
      return Object.keys(query).reduce((acc, cur) => {
        if (cur !== 'redirect') acc[cur] = query[cur]
        return acc
      }, {})
    }
  }
}
</script>

<style lang="scss">
.login-shell .el-input input {
  background: transparent;
  border: 0;
  border-radius: 0;
  color: #ecf1ff;
  caret-color: #8aa2ff;
  padding-left: 8px;
}
.login-shell .el-form-item {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 10px;
  margin-bottom: 18px;
}
.login-shell .el-form-item__content {
  display: flex;
  align-items: center;
}
.login-shell .el-link {
  font-size: 13px;
}

.fixed-link {
  width: 150px;
  display: inline-block;
  text-align: center;
}

.fixed-link .link-label {
  display: inline;
}

.fixed-link:hover .link-label {
}

.fixed-link:hover,
.fixed-link:focus {
  text-decoration: none;
}

.fixed-link .el-link--inner {
  display: inline-flex;
  width: 100%;
  justify-content: center;
}

.fixed-link.is-underline:after,
.fixed-link .el-link--inner:after {
  display: none !important;
  border: none !important;
}
</style>

<style scoped lang="scss">
.login-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-size: cover;
  background-position: center;
  position: relative;
  overflow: hidden;
  padding: 40px 24px;
  color: #ecf1ff;
}

.login-overlay {
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 20% 20%, rgba(99, 123, 255, 0.25), transparent 35%),
    radial-gradient(circle at 80% 10%, rgba(108, 255, 204, 0.18), transparent 30%),
    linear-gradient(135deg, rgba(12, 19, 52, 0.85), rgba(18, 21, 42, 0.9));
  backdrop-filter: blur(2px);
}

.login-body {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 24px;
  width: 100%;
  max-width: 1200px;
}

.brand-card {
  background: linear-gradient(135deg, rgba(98, 122, 255, 0.8), rgba(39, 203, 255, 0.75));
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 20px 50px rgba(0, 0, 0, 0.28);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 12px;
}

.brand-title {
  font-size: 28px;
  font-weight: 700;
  letter-spacing: 1px;
}

.brand-sub {
  font-size: 15px;
  color: #e7ecff;
  margin: 0;
}

.brand-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.pill {
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 999px;
  padding: 8px 12px;
  font-size: 12px;
  letter-spacing: 0.3px;
}

.auth-card {
  background: rgba(16, 18, 35, 0.9);
  border-radius: 16px;
  padding: 28px;
  box-shadow: 0 18px 45px rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.04);
  display: flex;
  flex-direction: column;
  min-height: 520px;
}

.auth-header {
  margin-bottom: 16px;
}

.mode-switch {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.mode-item {
  padding: 8px 14px;
  border-radius: 12px;
  cursor: pointer;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid transparent;
  transition: all 0.2s ease;
  font-weight: 600;
}

.mode-item.active {
  background: linear-gradient(120deg, #7bd5f5, #7879ff);
  color: #0c1234;
  border-color: rgba(255, 255, 255, 0.16);
  box-shadow: 0 10px 30px rgba(120, 121, 255, 0.3);
}

.auth-hint {
  margin-top: 8px;
  color: #c6cee8;
  font-size: 13px;
}

.auth-form {
  margin-top: 8px;
}

.svg-container {
  padding: 6px 8px 6px 14px;
  color: #9fb4ff;
  width: 34px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.show-pwd {
  position: absolute;
  right: 10px;
  top: 9px;
  font-size: 16px;
  color: #b8c3e5;
  cursor: pointer;
  user-select: none;
}

.form-actions {
  margin-top: 8px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.full-btn {
  flex: 1;
  height: 44px;
  border-radius: 12px;
  background: linear-gradient(120deg, #7bd5f5, #7879ff);
  border: none;
  font-weight: 700;
}

.full-btn.danger {
  background: linear-gradient(120deg, #ff6b6b, #ff8a65);
}

.code-btn {
  margin-left: 8px;
  background: rgba(255, 255, 255, 0.08);
  color: #dbe6ff;
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.tip-alert {
  margin: 4px 0 12px;
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.08);
}

.oauth-panel {
  margin-top: 12px;
}

.oauth-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.oauth-card {
  display: flex;
  gap: 12px;
  align-items: center;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  padding: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.oauth-card:hover {
  transform: translateY(-2px);
  border-color: rgba(120, 121, 255, 0.5);
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  background: linear-gradient(135deg, #7bd5f5, #7879ff);
  color: #0c1234;
}

.avatar.github {
  background: linear-gradient(135deg, #24292e, #4c5866);
  color: #f5f7fa;
}

.name {
  font-weight: 700;
  color: #f2f6ff;
}

.meta {
  font-size: 12px;
  color: #aeb9d7;
}

.oauth-inline {
  margin-top: 10px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  color: #c5cde7;
}

.inline-label {
  font-size: 13px;
}

.inline-oauth-btn {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #e9edff;
  border-radius: 999px;
  padding: 6px 12px;
}

@media (max-width: 1024px) {
  .login-body {
    grid-template-columns: 1fr;
  }
  .brand-card {
    order: 2;
  }
}

@media (max-width: 640px) {
  .auth-card {
    padding: 20px;
    min-height: auto;
  }
  .mode-switch {
    gap: 8px;
  }
  .mode-item {
    padding: 6px 10px;
  }
  .brand-card {
    padding: 20px;
  }
}
</style>
