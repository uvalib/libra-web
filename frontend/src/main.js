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
   // plugins: [addRequiredNotePlugin],
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
   var showRequired = true
   node.on('created', () => {
      if (node.config.disableRequiredDecoration == true) {
         showRequired = false
      }
      const schemaFn = node.props.definition.schema
      node.props.definition.schema = (sectionsSchema = {}) => {
         const isRequired = node.props.parsedRules.some(rule => rule.name === 'required')

         if (isRequired && showRequired) {
            // this input has the required rule so we modify
            // the schema to add an astrics to the label.
            sectionsSchema.label = {
               attrs: {
                  innerHTML: `<span class="req-label">${node.props.label}</span><span class="req">required</span>`
               },
               children: null//['$label', '*']
            }
         }
         return schemaFn(sectionsSchema)
      }
   })
}
