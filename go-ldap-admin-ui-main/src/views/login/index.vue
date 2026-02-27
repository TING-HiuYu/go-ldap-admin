<template>
  <div class="login-shell" :style="{ backgroundImage: `url(${imgSrc})` }">
    <div class="login-overlay" />
    <div class="login-body">
      <div class="brand-card">
        <div class="brand-title">Go-Ldap-Admin</div>
        <p class="brand-sub">一个账号，连接 LDAP、OAuth 与团队。</p>
        <div class="brand-metrics">
          <span class="pill">安全 · 邮箱 OTP</span>
          <span class="pill">GitHub OAuth</span>
          <span class="pill">Casbin 授权</span>
        </div>
      </div>

      <div class="auth-card">
        <div class="auth-header">
          <div class="mode-switch">
            <span :class="['mode-item', { active: loginMode === 'password' }]" @click="switchMode('password')">账号密码</span>
            <span :class="['mode-item', { active: loginMode === 'otp' }]" @click="switchMode('otp')">邮箱验证码</span>
            <span :class="['mode-item', { active: loginMode === 'oauth' }]" @click="switchMode('oauth')">OAuth 登录</span>
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
              placeholder="用户名"
              name="username"
              type="text"
              tabindex="1"
              autocomplete="on"
            />
          </el-form-item>

          <el-tooltip v-model="capsTooltip" content="Caps Lock 已开启" placement="right" manual>
            <el-form-item prop="password">
              <span class="svg-container"><svg-icon icon-class="password" /></span>
              <el-input
                :key="passwordType"
                ref="password"
                v-model="loginForm.password"
                :type="passwordType"
                placeholder="密码"
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
            <el-link type="primary" class="fixed-link" @click="changePass"><span class="link-label">忘记密码?</span></el-link>
            <el-button :loading="loading" type="primary" class="full-btn" @click.native.prevent="handlePasswordLogin">登录</el-button>
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
            <el-input v-model.trim="otpForm.mail" placeholder="请输入邮箱" autocomplete="on" />
          </el-form-item>
          <el-form-item prop="code">
            <span class="svg-container"><svg-icon icon-class="password" /></span>
            <el-input v-model.trim="otpForm.code" placeholder="请输入6位验证码" maxlength="6" />
            <el-button class="code-btn" :disabled="otpCountdown>0" @click="sendOtpCode">
              {{ otpCountdown>0 ? `重新获取(${otpCountdown}s)` : '获取验证码' }}
            </el-button>
          </el-form-item>
          <div class="form-actions">
            <el-link type="primary" class="fixed-link" @click="switchMode('password')"><span class="link-label">改用账号密码</span></el-link>
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
                <div class="meta">{{ item.provider }} · {{ item.enabled ? '启用' : '停用' }}</div>
              </div>
            </div>
            <el-empty v-if="!connectors.length && !oauthLoading" description="暂无可用的 OAuth 连接器" />
          </div>
          <div class="form-actions">
            <el-link type="primary" @click="switchMode('password')">返回账号密码</el-link>
          </div>
        </div>

        <div v-if="loginMode!=='oauth' && connectors.length" class="oauth-inline">
          <span class="inline-label">快捷使用 OAuth：</span>
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
        callback(new Error('密码长度至少 6 位'))
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
        username: [{ required: true, trigger: 'submit', message: '请输入用户名' }],
        password: [{ required: true, trigger: 'submit', validator: validatePassword }]
      },
      otpRules: {
        mail: [{ required: true, trigger: 'submit', message: '请输入邮箱' }],
        code: [{ required: true, trigger: 'submit', message: '请输入验证码', min: 6, max: 6 }]
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
      return '选择一个连接器完成 OAuth 授权并登录。'
    },
    otpButtonText() {
      return this.isResetFlow ? '重置密码并登录' : '登录'
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
          this.$message.error('公钥无效，登录已中断')
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
        this.$message.warning('请先填写邮箱')
        return
      }
      this.$refs.otpForm.validateField('mail', async valid => {
        if (valid) return
        try {
          await sendLoginOtp({ mail: this.otpForm.mail })
          this.$message.success('验证码已发送，请查收邮箱')
          this.startCountdown()
        } catch (e) {
          this.$message.error(e?.message || '发送失败')
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
            this.$message.error('登录失败，未获取到 token')
            this.loading = false
            return
          }
          this.applyToken(token)
          if (this.isResetFlow && data?.newPassword) {
            // 保存新密码到 sessionStorage，进入主界面后弹窗提示
            sessionStorage.setItem('generatedPassword', data.newPassword)
          }
          this.afterLogin()
        } catch (e) {
          this.$message.error(e?.message || '登录失败')
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
          this.$message.error('未获取到授权地址')
          this.oauthLoading = false
          return
        }
        this.pendingStates[data.state] = { connectorId: connector.id }
        const features = 'width=520,height=640,menubar=no,toolbar=no'
        this.oauthWindow = window.open(data.authUrl, '_blank', features)
        if (!this.oauthWindow) {
          this.$message.error('请允许浏览器弹出窗口后重试')
        }
      } catch (e) {
        this.$message.error(e?.message || '启动授权失败')
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
        // 如果是新创建用户，后端会返回生成的随机密码
        if (payload.generatedPassword) {
          sessionStorage.setItem('generatedPassword', payload.generatedPassword)
        }
        this.$message.success('OAuth 登录成功')
        this.afterLogin()
      } else {
        this.$message.warning('未获取到登录凭证，请重试')
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
