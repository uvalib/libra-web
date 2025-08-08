import { createApp, markRaw } from 'vue'
import { createPinia } from 'pinia'
import formatDatePlugin from './plugins/formatdate'
import formatDateTimePlugin from './plugins/formatdatetime'
import App from './App.vue'
import router from './router'
import { createHead } from '@unhead/vue/client'


const app = createApp(App)

app.use(router)
const head = createHead()
app.use(head)
app.use(formatDatePlugin)
app.use(formatDateTimePlugin)

// Primevue setup
import PrimeVue from 'primevue/config'
import UVA from './assets/theme/uva'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'
import 'primeicons/primeicons.css'
import KeyFilter from 'primevue/keyfilter'

app.use(PrimeVue, {
   theme: {
      preset: UVA,
      options: {
         prefix: 'p',
         darkModeSelector: '.libra-dark'
      }
   }
})

app.use(ConfirmationService)
app.use(ToastService)

app.component("Button", Button)
app.component("ConfirmDialog", ConfirmDialog)
app.directive('keyfilter', KeyFilter)

// Per some suggestions on vue / pinia git hub issue reports, create and add pinia support LAST
// and use the chained form of the setup. This to avid problems where the vuew dev tools fail to
// include pinia in the tools
app.use(createPinia().use( ({ store }) => {
   store.router = markRaw(router)
}))

app.mount('#app')
