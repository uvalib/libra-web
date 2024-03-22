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
         <label v-if="type=='oa'" for="agree-cb">
            By saving this work, I agree to the
            <a href="https://www.library.virginia.edu/libra/open/libra-deposit-license" target="_blank">Libra Deposit Agreement</a>
         </label>
         <label v-else for="agree-cb">
            I have read and agree to the
            <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank">Libra Deposit License</a>,
            including discussing my deposit access options with my faculty advisor.
         </label>

      </div>
      <div class="button-bar">
         <template v-if="props.mode=='create'">
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Submit" @click="emit('submit', visibility)" :disabled="!canSubmit"/>
         </template>
         <template v-else>
            <Button severity="secondary" label="Cancel Edit" @click="emit('cancel')"/>
            <Button label="Save and Exit" @click="emit('saveExit', visibility)" :disabled="!canSubmit"/>
            <Button label="Save and Continue" @click="emit('saveContinue', visibility)" :disabled="!canSubmit"/>
         </template>
      </div>
   </Panel>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Panel from 'primevue/panel'
import { useSystemStore } from "@/stores/system"

const emit = defineEmits( ['submit', 'cancel', 'saveExit', 'saveContinue'])
const props = defineProps({
   type: {
      type: String,
      required: true,
      validator(value) {
         return ['oa', 'etd'].includes(value)
      },
   },
   mode: {
      type: String,
      required: true,
      validator(value) {
         return ['create', 'edit'].includes(value)
      },
   },
   files: {
      type: Boolean,
      required: true
   },
   visibility: {
      type: String,
      required: true
   },
   described: {
      type: Boolean,
      required: true
   }
})

const system = useSystemStore()
const visibility = ref(props.visibility)
const agree = ref(false)

onMounted( () => {
   visibility.value = props.visibility
})

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
      div.visibility {
         padding: 0;
         margin-left: 0;
      }
   }
   .agree {
      display: flex;
      flex-direction: row;
      align-items: flex-start;
      margin: 25px 0;
      label {
         margin-left: 15px;
      }
   }
   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: center;
      align-items: stretch;
      button {
         font-size: 0.85em;
         padding: 5px 10px;
         margin-left: 5px;
      }
   };
}
</style>