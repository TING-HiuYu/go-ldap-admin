<template>
  <el-dropdown trigger="click" @command="handleCommand">
    <span class="lang-trigger">
      <i class="el-icon-s-flag" />
      <span v-if="showLabel" class="lang-label">{{ currentLabel }}</span>
    </span>
    <el-dropdown-menu slot="dropdown" class="lang-dropdown-menu">
      <el-dropdown-item command="en" :class="{ 'is-active': language === 'en' }">English</el-dropdown-item>
      <el-dropdown-item command="zh" :class="{ 'is-active': language === 'zh' }">中文</el-dropdown-item>
      <el-dropdown-item v-for="lang in autoLanguages" :key="lang.code" :command="lang.code" :class="{ 'is-active': language === lang.code }" divided-first>
        {{ lang.label }}
      </el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script>
import { autoTranslateLanguages, isAutoTranslateLang, switchAutoTranslate, resetAutoTranslate } from '@/utils/translatejs'

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
      autoLanguages: autoTranslateLanguages
    }
  },
  computed: {
    language() {
      return localStorage.getItem('locale') || this.$i18n.locale
    },
    currentLabel() {
      if (this.language === 'zh') return '中文'
      if (this.language === 'en') return 'English'
      const found = autoTranslateLanguages.find(l => l.code === this.language)
      return found ? found.label : 'English'
    }
  },
  methods: {
    handleCommand(lang) {
      localStorage.setItem('locale', lang)

      if (isAutoTranslateLang(lang)) {
        // For auto-translated languages, keep vue-i18n on English (source)
        // and let translate.js handle the DOM translation
        this.$i18n.locale = 'en'
        switchAutoTranslate(lang)
      } else {
        // For manually maintained locales (en/zh), reset translate.js
        // and switch vue-i18n locale
        resetAutoTranslate()
        this.$i18n.locale = lang
      }

      this.$message({
        message: this.$t('langSwitch.switchSuccess'),
        type: 'success',
        duration: 1500
      })

      // Force re-render to update computed label
      this.$forceUpdate()
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
