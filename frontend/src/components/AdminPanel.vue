<template>
   <Panel header="Admin Info" class="admin-panel">
      <div class="admin-content">
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
               <tr>
                  <td class="label">Program:</td>
                  <td><Select v-model="program" :options="programs"/></td>
               </tr>
               <tr>
                  <td class="label">Degree:</td>
                  <td><Select v-model="degree" :options="degrees"/></td>
               </tr>
               <tr>
                  <td class="label">Visibility:</td>
                  <td>
                     <Select v-model="visibility" :options="system.visibility" optionLabel="label" optionValue="value" @change="visibilityChanged()"/>
                  </td>
               </tr>
               <template v-if="showEmbargoSettings">
                  <tr>
                     <td class="label">End Date:</td>
                     <td class="embargo">
                        <span v-if="embargoEndDate">{{ $formatDate(embargoEndDate) }}</span>
                        <span v-else>Never</span>
                        <DatePickerDialog :endDate="embargoEndDate" :admin="true"
                           :visibility="visibility" @picked="endDatePicked" :degree="degree" :program="program"/>
                     </td>
                  </tr>
                  <tr>
                     <td class="label">End Visibility:</td>
                     <td>
                        <span>{{ system.visibilityLabel('etd',embargoEndVisibility) }}</span>
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

const emit = defineEmits( ['save', 'cancel', 'delete'])

const props = defineProps({
   identifier: {
      type: String,
      required: true,
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

const showEmbargoSettings = computed( () => {
   if ( visibility.value == 'embargo' ) return true
   return visibility.value == 'uva'
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
   if ( visibility.value == "uva" || visibility.value == "embargo") {
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
         admin.unpublish(props.identifier)
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
         admin.delete(props.identifier)
         emit('delete')
      },
   })
})

const saveClicked = (() => {
   let changes = {
      adminNotes: adminNotes.value,
      visibility: visibility.value,
      embargoEndDate: embargoEndDate.value,
      embargoEndVisibility: embargoEndVisibility.value,
      degree: degree.value,
      program: program.value,
   }
   emit("save", changes)
})
</script>

<style lang="scss" scoped>
.admin-panel {
   .admin-content {
      display: flex;
      flex-direction: column;
      gap: 25px;
   }
   table {
      font-size: 0.9em;
      td {
         padding: 5px 0;
         .p-select {
            width: 100%;
         }
      }
      td.label {
         font-weight: bold;
         text-align: right;
         padding-right: 10px;
         white-space: nowrap;
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
      gap: 0.5rem;
      span {
         display: flex;
         flex-flow: row nowrap;
         gap: 0.5rem;
      }
   };
}
</style>