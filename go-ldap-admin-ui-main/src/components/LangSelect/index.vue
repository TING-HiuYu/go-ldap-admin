<template>
  <el-dropdown trigger="click" @command="handleCommand">
    <span class="lang-trigger">
      <i class="el-icon-s-flag" />
      <span v-if="showLabel" class="lang-label">{{ currentLabel }}</span>
    </span>
    <el-dropdown-menu slot="dropdown">
      <el-dropdown-item command="en" :class="{ 'is-active': language === 'en' }">English</el-dropdown-item>
      <el-dropdown-item command="zh" :class="{ 'is-active': language === 'zh' }">中文</el-dropdown-item>
    </el-dropdown-menu>
  </el-dropdown>
</template>

<script>
export default {
  name: 'LangSelect',
  props: {
    showLabel: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    language() {
      return this.$i18n.locale
    },
    currentLabel() {
      return this.language === 'zh' ? '中文' : 'English'
    }
  },
  methods: {
    handleCommand(lang) {
      this.$i18n.locale = lang
      localStorage.setItem('locale', lang)
      this.$message({
        message: lang === 'zh' ? '切换语言成功' : 'Language switched',
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
.lang-label {
  font-size: 13px;
}
.is-active {
  color: #409eff;
  font-weight: bold;
}
</style>
