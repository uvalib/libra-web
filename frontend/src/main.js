import { createApp, markRaw } from 'vue'
import { createPinia } from 'pinia'
import formatDatePlugin from './plugins/formatdate'

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
app.use(formatDatePlugin)

// Styles
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
import 'primeicons/primeicons.css'

app.use(PrimeVue, { ripple: true })
app.use(ConfirmationService)
app.use(ToastService)

app.component("Button", Button)
app.component("ConfirmDialog", ConfirmDialog)

// FormKit
import { plugin, defaultConfig } from '@formkit/vue'

const fkCfg = defaultConfig({
   plugins: [addRequiredNotePlugin],
   config: {
      classes: {
         input: '$reset libra-form-input',
         label: '$reset libra-form-label',
         messages: '$reset libra-form-invalid',
         help: '$reset libra-form-help',
      },
      incompleteMessage: false,
      validationVisibility: 'submit'
   }
})
app.use(plugin, fkCfg)

app.mount('#app')

// Plugins for formkit -------

function addRequiredNotePlugin(node) {
   node.on('created', () => {
      const schemaFn = node.props.definition.schema
      node.props.definition.schema = (sectionsSchema = {}) => {
         sectionsSchema['label'] = {
            children: ['$label', {
               $el: 'span',
               if: '$state.required',
               attrs: {
                  class: 'req-field',
               },
               children: ['Required']
            }]
         }
         return schemaFn(sectionsSchema)
      }
   })
}
