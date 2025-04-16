<template>
   <Panel header="Admin Info" class="admin-panel">
      <table>
         <tbody>
            <tr>
               <td class="label">Identifier:</td>
               <td>{{ props.identifier }}</td>
            </tr>
            <tr>
               <td class="label">Depositor:</td>
               <td>{{ props.depositor }}</td>
            </tr>
            <tr>
               <td class="label">Created:</td>
               <td>{{ $formatDateTime(props.created) }}</td>
            </tr>
            <tr v-if="props.modified">
               <td class="label">Modified:</td>
               <td>{{ $formatDateTime(props.modified) }}</td>
            </tr>
            <tr v-if="props.published">
               <td class="label">Published:</td>
               <td>{{ $formatDateTime(props.published) }}</td>
            </tr>
            <template v-if="props.type == 'etd'">
               <tr>
                  <td class="label">Program:</td>
                  <td>
                     <div class="picker"  v-if="editField!='program'">
                        <span>{{ program }}</span>
                        <Button icon="pi pi-file-edit" severity="secondary" rounded aria-label="edit program" @click="editField='program'"/>
                     </div>
                     <div class="picker" v-else>
                        <Select v-model="program" :options="programs" :autoOptionFocus="true" @update:modelValue="editField=''"/>
                        <Button icon="pi pi-times" severity="secondary" rounded aria-label="cancel edit program" @click="editField=''"/>
                     </div>
                  </td>
               </tr>
               <tr>
                  <td class="label">Degree:</td>
                  <td>
                     <div class="picker" v-if="editField!='degree'">
                        <span>{{ degree }}</span>
                        <Button icon="pi pi-file-edit" severity="secondary" rounded aria-label="edit degree" @click="editField='degree'"/>
                     </div>
                     <div class="picker" v-else>
                        <Select v-model="degree" :options="degrees" :autoOptionFocus="true" @update:modelValue="editField=''"/>
                        <Button icon="pi pi-times" severity="secondary" rounded aria-label="cancel edit degree" @click="editField=''"/>
                     </div>
                  </td>
               </tr>
            </template>
            <tr>
               <td class="label">Visibility:</td>
               <td>
                  <Select v-model="visibility" :options="visibilityOpts" optionLabel="label" optionValue="value" @change="visibilityChanged()"/>
               </td>
            </tr>
            <template v-if="showEmbargoSettings">
               <tr>
                  <td class="label">End Date:</td>
                  <td class="embargo">
                     <span v-if="embargoEndDate">{{ $formatDate(embargoEndDate) }}</span>
                     <span v-else>Never</span>
                     <DatePickerDialog :type="props.type" :endDate="embargoEndDate" :admin="true"
                        :visibility="visibility" @picked="endDatePicked" :degree="degree" :program="program"/>
                  </td>
               </tr>
               <tr>
                  <td class="label">End Visibility:</td>
                  <td>
                     <span v-if="props.type=='etd'">{{ system.visibilityLabel('etd',embargoEndVisibility) }}</span>
                     <Select v-else v-model="embargoEndVisibility" :options="endOpts" optionLabel="label" optionValue="value"/>
                  </td>
               </tr>
            </template>
         </tbody>
      </table>
      <div class="notes">
         <label for="admin-notes">Admin Notes</label>
         <Textarea id="admin-notes" v-model="adminNotes" rows="5" />
      </div>

      <div class="button-bar">
         <Button v-if="props.published" label="Unpublish" severity="warning" icon="pi pi-eye-slash" @click="unpublishWorkClicked()"/>
         <Button v-else label="Delete" severity="danger" icon="pi pi-trash" @click="deleteWorkClicked()"/>
         <span>
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Save" @click="saveClicked()" />
         </span>
      </div>
   </Panel>
</template>

<script setup>
import Panel from 'primevue/panel'
import { ref, onMounted, computed } from 'vue'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'
import DatePickerDialog from "@/components/DatePickerDialog.vue"
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import { useConfirm } from "primevue/useconfirm"

const confirm = useConfirm()
const admin = useAdminStore()
const system = useSystemStore()
const adminNotes = ref("")
const degree = ref("")
const program = ref("")
const visibility = ref("")
const embargoEndDate = ref("")
const embargoEndVisibility = ref("")
const endOpts = ref([{label: "Worldwide", value: "open"}, {label: "UVA Only", value: "uva"}])
const editField = ref("")

const emit = defineEmits( ['save', 'cancel', 'delete'])
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
   source: {
      type: String,
      default: "",
   },
   depositor: {
      type: String,
      required: true
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
   },
   visibility: {
      type: String,
      required: true
   },
   degree: {
      type: String,
      default: "",
   },
   program: {
      type: String,
      default: "",
   },
   embargoEndDate: {
      type: String,
      default: null
   },
   embargoEndVisibility: {
      type: String,
      default: null
   },
   notes: {
      type: String,
      default: null
   }
})

const programs = computed( () =>{
   if (props.source == "sis") return system.sisPrograms
   return system.optPrograms
})

const degrees = computed( () =>{
   if (props.source == "sis") return system.sisDegrees
   return system.optDegrees
})

const visibilityOpts = computed( () => {
   if (props.type == "oa") return system.oaVisibility

   // admins get the notmal ETD visibility options plus embargo
   // note: copy the array with slice to avoid updating the data in the system store
   let etdVis = system.etdVisibility.slice()
   etdVis.push({
      "label": "Embargo",
      "value": "embargo",
      "oa": false,
      "etd": true}
   )
   return etdVis
})

const showEmbargoSettings = computed( () => {
   if ( visibility.value == 'embargo' ) return true
   if ( props.type == 'etd') return visibility.value == 'uva'
   return false
})

onMounted( () => {
   adminNotes.value = props.notes
   degree.value = props.degree
   program.value = props.program
   visibility.value = props.visibility
   embargoEndDate.value = props.embargoEndDate
   embargoEndVisibility.value = props.embargoEndVisibility
})

const visibilityChanged = (() => {
   if ( (props.type == "etd" && visibility.value == "uva") || visibility.value == "embargo") {
      embargoEndVisibility.value = "open"
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      embargoEndDate.value = endDate.toJSON()
   }
})

const endDatePicked = ( (newDate) => {
   embargoEndDate.value = newDate
})

const unpublishWorkClicked = ( () => {
   confirm.require({
      message: "Unpublish this work? It will no longer be visible to UVA or worldwide users. Are you sure?",
      header: 'Confirm Work Unpublish',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.unpublish(props.type, props.identifier)
      },
   })
})

const deleteWorkClicked = ( () => {
   confirm.require({
      message: "Delete this work? All data will be lost. This cannot be reversed. Are you sure?",
      header: 'Confirm Work Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.delete(props.type, props.identifier)
         emit('delete')
      },
   })
})

const saveClicked = (() => {
   let changes = {
      adminNotes: adminNotes.value,
      visibility: visibility.value,
      embargoEndDate: embargoEndDate.value,
      embargoEndVisibility: embargoEndVisibility.value
   }
   if ( props.type == "etd") {
      changes.degree = degree.value
      changes.program = program.value
   }
   emit("save", changes)
})
</script>

<style lang="scss" scoped>
.admin-panel {
   background: white;
   :deep(.p-panel-title) {
      font-weight: normal;
   }
   .picker {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-start;
      width: 210px;
      :deep(.p-Select ) {
         width: 170px;
      }
      button {
         margin-left: 5px;
         height: 30px;
         min-width:30px;
         max-width: 30px;
         padding: 0;
         border-radius: 20px;
         border-color: $uva-grey-100;
         color: $uva-grey;
      }
   }
   label {
      font-size: 0.9em;
      font-weight: bold;
      display: block;
      margin: 10px 0 5px 0;
   }
   table {
      font-size: 0.9em;
      td {
         padding: 5px 0;
      }
      td.label {
         font-weight: bold;
         text-align: right;
         padding-right: 10px;
         white-space: nowrap;
         vertical-align: text-top;
         width: 110px;
      }
      td.embargo {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
      }
   }
   .notes {
      display: flex;
      flex-direction: column;
      gap: 0.3rem;
   }
   .button-bar {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      margin-top: 15px;
      gap: 0.5rem;
      span {
         display: flex;
         flex-flow: row nowrap;
         gap: 0.5rem;
      }
   };
}
</style>