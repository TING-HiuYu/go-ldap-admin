import Vue from 'vue'
import VueI18n from 'vue-i18n'
import en from './en'
import zh from './zh'
import ja from './ja'
import de from './de'
import fr from './fr'
import es from './es'
import ru from './ru'
import ar from './ar'
import elementEnLocale from 'element-ui/lib/locale/lang/en'
import elementZhLocale from 'element-ui/lib/locale/lang/zh-CN'
import elementJaLocale from 'element-ui/lib/locale/lang/ja'
import elementDeLocale from 'element-ui/lib/locale/lang/de'
import elementFrLocale from 'element-ui/lib/locale/lang/fr'
import elementEsLocale from 'element-ui/lib/locale/lang/es'
import elementRuLocale from 'element-ui/lib/locale/lang/ru-RU'
import elementArLocale from 'element-ui/lib/locale/lang/ar'

Vue.use(VueI18n)

const supportedLocales = ['zh', 'en', 'ja', 'de', 'fr', 'es', 'ru', 'ar']

const messages = {
  en: { ...en, ...elementEnLocale },
  zh: { ...zh, ...elementZhLocale },
  ja: { ...ja, ...elementJaLocale },
  de: { ...de, ...elementDeLocale },
  fr: { ...fr, ...elementFrLocale },
  es: { ...es, ...elementEsLocale },
  ru: { ...ru, ...elementRuLocale },
  ar: { ...ar, ...elementArLocale }
}

/**
 * Detect the best locale from the browser / system language.
 * Falls back to 'zh' if nothing matches.
 */
function detectLocale() {
  const saved = localStorage.getItem('locale')
  if (saved && supportedLocales.includes(saved)) return saved

  // navigator.languages gives an ordered list of the user's preferred languages
  const browserLangs = navigator.languages || [navigator.language || navigator.userLanguage || '']
  for (const tag of browserLangs) {
    const code = tag.split('-')[0].toLowerCase()
    if (supportedLocales.includes(code)) return code
  }
  return 'zh' // default fallback
}

const locale = detectLocale()
localStorage.setItem('locale', locale)

const i18n = new VueI18n({
  locale,
  fallbackLocale: 'zh',
  messages
})

export default i18n
