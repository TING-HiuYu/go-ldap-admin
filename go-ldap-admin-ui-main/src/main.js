import Vue from 'vue'

import Cookies from 'js-cookie'

import VueCompositionAPI from '@vue/composition-api'

import 'normalize.css/normalize.css' // a modern alternative to CSS resets

import Element from 'element-ui'
import './styles/element-variables.scss'

import '@/styles/index.scss' // global css

import App from './App'
import store from './store'
import router from './router'
import i18n from './lang'

import './icons' // icon
import './permission' // permission control
import './utils/error-log' // error log

import * as filters from './filters' // global filters
import { initTranslateJs, isAutoTranslateLang, switchAutoTranslate } from './utils/translatejs'

Vue.use(Element, {
  size: Cookies.get('size') || 'medium',
  i18n: (key, value) => i18n.t(key, value)
})

Vue.use(VueCompositionAPI)

// Register global utility filters
Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key])
})

Vue.config.productionTip = false

new Vue({
  el: '#app',
  router,
  store,
  i18n,
  mounted() {
    initTranslateJs()
    // If the saved locale is an auto-translate language, apply it
    const savedLocale = localStorage.getItem('locale')
    if (savedLocale && isAutoTranslateLang(savedLocale)) {
      switchAutoTranslate(savedLocale)
    }
  },
  render: h => h(App)
})
