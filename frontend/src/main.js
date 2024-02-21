import { createApp, markRaw } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)

const pinia = createPinia()
pinia.use(({ store }) => {
   // all stores can access router with this.router
   store.router = markRaw(router)
})

app.use(pinia)
app.use(router)

// Styles
import '@fortawesome/fontawesome-free/css/all.css'
import './assets/styles/main.scss'
import './assets/styles/uva-colors.css'
import './assets/styles/styleoverrides.scss'
import './assets/styles/forms.scss'

// Primevue setup
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'
import 'primevue/resources/themes/saga-blue/theme.css'

app.use(PrimeVue, { ripple: true })
app.use(ConfirmationService)
app.use(ToastService)

app.component("Button", Button)
app.component("ConfirmDialog", ConfirmDialog)

// FormKit
import { plugin, defaultConfig } from '@formkit/vue'

const fkCfg = defaultConfig({
   config: {
      classes: {
         input: '$reset v4-form-input',
         label: '$reset v4-form-label',
         messages: '$reset v4-form-invalid',
         help: '$reset v4-form-help',
      },
      incompleteMessage: false,
      validationVisibility: 'submit'
   }
})
app.use(plugin, fkCfg)

app.mount('#app')
