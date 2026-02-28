import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en'
import zh from './zh'
import elementEnLocale from 'element-ui/lib/locale/lang/en'
import elementZhLocale from 'element-ui/lib/locale/lang/zh-CN'
import { isAutoTranslateLang } from '@/utils/translatejs'

Vue.use(VueI18n)

const messages = {
  en: {
    ...en,
    ...elementEnLocale
  },
  zh: {
    ...zh,
    ...elementZhLocale
  }
}

// If the saved locale is an auto-translate language, fall back to 'en' for vue-i18n
// (translate.js will handle the actual DOM translation)
const savedLocale = localStorage.getItem('locale') || 'en'
const vueI18nLocale = isAutoTranslateLang(savedLocale) ? 'en' : savedLocale

const i18n = new VueI18n({
  locale: vueI18nLocale,
  fallbackLocale: 'zh',
  messages
})

export default i18n
