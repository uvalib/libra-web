import { createApp, markRaw } from 'vue'
import { createPinia } from 'pinia'
import formatDatePlugin from './plugins/formatdate'
import formatDateTimePlugin from './plugins/formatdatetime'

import App from './App.vue'
import router from './router'

const app = createApp(App)


app.use(router)
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

// Per some suggestions on vue / pinia git hub issue reports, create and add pinia support LAST
// and use the chained form of the setup. This to avid problems where the vuew dev tools fail to
// include pinia in the tools
app.use(createPinia().use( ({ store }) => {
   store.router = markRaw(router)
}))

app.mount('#app')

// Plugins for formkit -------

function addRequiredNotePlugin(node) {
   if ( node.props.parsedRules.some(rule => rule.name !== 'required') ) return
   if ( ['button', 'submit', 'hidden', 'group', 'list', 'meta', 'radio', 'checkbox'].includes(node.props.type) ) return

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
