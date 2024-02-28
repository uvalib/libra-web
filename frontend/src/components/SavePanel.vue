<template>
   <Panel header="Save Work" class="save-panel">
      <Fieldset legend="Requirements">
         <div class="requirement">
            <i v-if="props.described" class="done pi pi-check"></i>
            <i v-else class="not-done pi pi-exclamation-circle"></i>
            <span>Describe your work</span>
         </div>
         <div class="requirement">
            <i v-if="props.files" class="done pi pi-check"></i>
            <i v-else class="not-done pi pi-exclamation-circle"></i>
            <span>Add files</span>
         </div>
         <div class="help">
            <span v-if="type=='etd'">View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</span>
            <span v-else>View the <a href="https://www.library.virginia.edu/libra/open/oc-checklist" target="_blank">Libra Open Checklist</a> for help.</span>
         </div>
      </Fieldset>
      <Fieldset legend="Visibility">
         <div v-for="v in visibilityOptions" :key="v.value" class="visibility-opt">
            <RadioButton v-model="visibility" :inputId="v.value"  :value="v.value"  class="visibility"/>
            <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
         </div>
      </Fieldset>
      <div class="agree">
         <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
         <label for="agree-cb">
            By saving this work, I agree to the
            <a href="https://www.library.virginia.edu/libra/open/libra-deposit-license" target="_blank">Libra Deposit Agreement</a>
         </label>
      </div>
      <div class="button-bar">
         <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
         <Button label="Submit" @click="emit('submit', visibility)" :disabled="!canSubmit"/>
      </div>
   </Panel>
</template>

<script setup>
import { ref, computed } from 'vue'
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Panel from 'primevue/panel'
import { useSystemStore } from "@/stores/system"

const emit = defineEmits( ['submit', 'cancel'])
const props = defineProps({
   type: {
      type: String,
      required: true,
      validator(value) {
         return ['oa', 'etd'].includes(value)
      },
   },
   files: {
      type: Boolean,
      required: true
   },
   described: {
      type: Boolean,
      required: true
   }
})

const system = useSystemStore()
const visibility = ref("")
const agree = ref(false)

const visibilityOptions = computed( () => {
   if ( props.type == 'oa') {
      return system.oaVisibility
   }
   return system.etdVisibility
})

const canSubmit = computed(() =>{
   if (props.described == false ) return false
   return agree.value == true && visibility.value != "" && props.files
})
</script>

<style lang="scss" scoped>
.save-panel {
   .help {
      font-size: 0.9em;
      margin-top:15px;
   }
   .requirement {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: center;
      i {
         display: inline-block;
         margin-right: 10px;
         font-size: 1.25rem;
      }
      .not-done {
         color: var(--uvalib-red-darker);
      }
      .done {
         color: var(--uvalib-green-dark);
      }
   }
   .requirement:first-of-type {
      margin-bottom: 5px;
   }
   .visibility-opt {
      margin: 5px 0;
      label.visibility {
         font-size: 0.85em;
         margin-left: 10px;
         border-radius: 5px;
         padding: 2px 10px;
         color: white;
      }
      .visibility.open {
         background-color: var(--uvalib-green-dark);
      }
      .visibility.authenticated {
         background-color: var(--uvalib-brand-orange);
      }
      .visibility.embargo {
         background-color: var(--uvalib-blue-alt);
      }
      .visibility.restricted {
         background-color: var(--uvalib-red-darker);
      }
   }
   .agree {
      margin: 25px 0;
      label {
         margin-left: 10px;
      }
   }
   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: center;
      align-items: flex-end;
      button {
         margin-left: 10px;
      }
   };
}
</style>