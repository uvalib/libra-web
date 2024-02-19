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

// Primevue setup
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'

app.use(PrimeVue, { ripple: true })
app.use(ConfirmationService)
app.use(ToastService)

app.component("Button", Button)
app.component("ConfirmDialog", ConfirmDialog)

app.mount('#app')
