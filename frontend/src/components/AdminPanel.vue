<template>
   <Panel header="Admin Info" class="admin-panel">
      <table>
         <tr>
            <td class="label">Identifier:</td>
            <td>{{ props.identifier }}</td>
         </tr>
         <tr>
            <td class="label">Date Created:</td>
            <td>{{ $formatDateTime(props.created) }}</td>
         </tr>
         <tr v-if="props.modified">
            <td class="label">Date Modified:</td>
            <td>{{ $formatDateTime(props.modified) }}</td>
         </tr>
         <tr v-if="props.published">
            <td class="label">Date Published:</td>
            <td>{{ $formatDateTime(props.published) }}</td>
         </tr>
      </table>
      <FloatLabel>
         <Textarea v-model="adminNotes" rows="5" />
         <label>Admin Notes</label>
      </FloatLabel>

      <div class="button-bar">
         <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
         <Button label="Save" @click="saveClicked()" />
      </div>
   </Panel>
</template>

<script setup>
import Panel from 'primevue/panel'
import { ref } from 'vue'
import Textarea from 'primevue/textarea'
import FloatLabel from 'primevue/floatlabel'

const adminNotes = ref("")

const emit = defineEmits( ['submit', 'cancel'])
const props = defineProps({
   identifier: {
      type: String,
      required: true,
   },
   type: {
      type: String,
      required: true,
      validator(value) {
         return ['oa', 'etd'].includes(value)
      },
   },
   created: {
      type: String,
      required: true
   },
   modified: {
      type: String,
      default: null
   },
   published: {
      type: String,
      default: null
   }

})

const saveClicked = (() => {

})

</script>

<style lang="scss" scoped>
.admin-panel {
   :deep(.p-panel-title) {
      font-weight: normal;
   }
   table {
      font-size: 0.9em;
      td.label {
         font-weight: bold;
         text-align: right;
         padding-right: 10px;
      }
   }
   .p-float-label {
      margin-top: 15px;
   }
   .p-inputtextarea {
      width: 100%;
   }
   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-items: stretch;
      margin-top: 15px;
      button {
         font-size: 0.85em;
         padding: 5px 10px;
         margin-left: 5px;
      }
   };
}
</style>