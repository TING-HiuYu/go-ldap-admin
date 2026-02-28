/**
 * translate.js integration for automatic translation of languages
 * beyond the manually maintained English and Chinese locale files.
 *
 * English and Chinese use vue-i18n manual translations.
 * All other languages are handled by translate.js (xnx3/translate)
 * which auto-translates the DOM content at runtime.
 */

// Languages supported via translate.js auto-translation.
// Each entry: { code, label (native name), translateJsLang (translate.js language identifier) }
export const autoTranslateLanguages = [
  { code: 'ja', label: '日本語', translateJsLang: 'japanese' },
  { code: 'ko', label: '한국어', translateJsLang: 'korean' },
  { code: 'fr', label: 'Français', translateJsLang: 'french' },
  { code: 'de', label: 'Deutsch', translateJsLang: 'german' },
  { code: 'es', label: 'Español', translateJsLang: 'spanish' },
  { code: 'pt', label: 'Português', translateJsLang: 'portuguese' },
  { code: 'ru', label: 'Русский', translateJsLang: 'russian' },
  { code: 'ar', label: 'العربية', translateJsLang: 'arabic' },
  { code: 'th', label: 'ไทย', translateJsLang: 'thai' },
  { code: 'vi', label: 'Tiếng Việt', translateJsLang: 'vietnamese' },
  { code: 'it', label: 'Italiano', translateJsLang: 'italian' },
  { code: 'hi', label: 'हिन्दी', translateJsLang: 'hindi' }
]

/**
 * Check whether a language code is handled by translate.js (not a manual locale).
 */
export function isAutoTranslateLang(code) {
  return autoTranslateLanguages.some(l => l.code === code)
}

/**
 * Get the translate.js language identifier for a given code.
 */
export function getTranslateJsLang(code) {
  const found = autoTranslateLanguages.find(l => l.code === code)
  return found ? found.translateJsLang : null
}

/**
 * Initialize translate.js (call once after app mount).
 * Hides the default language selector since we provide our own LangSelect.
 */
export function initTranslateJs() {
  if (typeof window.translate === 'undefined') {
    console.warn('[translate.js] Library not loaded from CDN.')
    return
  }
  const t = window.translate
  t.service.use('client.edge')
  t.language.setLocal('english')
  t.selectLanguageTag.show = false
  t.ignore.class.push('no-translate')
  t.listener.start()
}

/**
 * Switch to an auto-translated language via translate.js.
 * @param {string} langCode - one of the autoTranslateLanguages codes
 */
export function switchAutoTranslate(langCode) {
  if (typeof window.translate === 'undefined') return
  const tjLang = getTranslateJsLang(langCode)
  if (tjLang) {
    window.translate.changeLanguage(tjLang)
  }
}

/**
 * Reset translate.js back to the original (source) language.
 * Call this when switching back to a manually maintained locale (en/zh).
 */
export function resetAutoTranslate() {
  if (typeof window.translate === 'undefined') return
  window.translate.changeLanguage('english')
}
