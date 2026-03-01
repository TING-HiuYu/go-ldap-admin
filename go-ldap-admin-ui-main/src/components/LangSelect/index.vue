<template>
  <el-dropdown trigger="click" @command="handleCommand">
    <span class="lang-trigger">
      <svg-icon icon-class="language" class="lang-icon" />
      <span v-if="showLabel" class="lang-label">{{ currentLabel }}</span>
    </span>
    <el-dropdown-menu slot="dropdown" class="lang-dropdown-menu">
      <el-dropdown-item v-for="lang in languages" :key="lang.code" :command="lang.code" :class="{ 'is-active': language === lang.code }">
        {{ lang.label }}
      </el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script>
const supportedLanguages = [
  { code: 'zh', label: '中文' },
  { code: 'en', label: 'English' },
  { code: 'ja', label: '日本語' },
  { code: 'de', label: 'Deutsch' },
  { code: 'fr', label: 'Français' },
  { code: 'es', label: 'Español' },
  { code: 'ru', label: 'Русский' },
  { code: 'ar', label: 'العربية' }
]

export default {
  name: 'LangSelect',
  props: {
    showLabel: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      languages: supportedLanguages
    }
  },
  computed: {
    language() {
      return this.$i18n.locale
    },
    currentLabel() {
      const found = supportedLanguages.find(l => l.code === this.language)
      return found ? found.label : '中文'
    }
  },
  methods: {
    handleCommand(lang) {
      localStorage.setItem('locale', lang)
      this.$i18n.locale = lang

      this.$message({
        message: this.$t('langSwitch.switchSuccess'),
        type: 'success',
        duration: 1500
      })
    }
  }
}
</script>

<style scoped>
.lang-trigger {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: inherit;
  font-size: inherit;
}
.lang-icon {
  width: 1.15em;
  height: 1.15em;
  vertical-align: middle;
}
.lang-label {
  font-size: 13px;
}
.is-active {
  color: #409eff;
  font-weight: bold;
}
.lang-dropdown-menu {
  max-height: 320px;
  overflow-y: auto;
}
</style>
