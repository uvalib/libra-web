<template>
   <Panel header="Admin Info" class="admin-panel">
      <table>
         <tr>
            <td class="label">Identifier:</td>
            <td>{{ props.identifier }}</td>
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
               <td class="label">Plan/Program:</td>
               <td><Dropdown v-model="department" :options="system.departments" /></td>
            </tr>
            <tr>
               <td class="label">Degree:</td>
               <td><Dropdown v-model="degree" :options="system.degrees" /></td>
            </tr>
            <tr>
               <td class="label">Visibility:</td>
               <td>
                  <Dropdown v-model="visibility" :options="visibilityOpts" optionLabel="label" optionValue="value" @change="visibilityChanged()"/>
               </td>
            </tr>
            <template v-if="visibility == 'uva'">
               <tr>
                  <td class="label">Embargo End:</td>
                  <td class="embargo">
                     <span v-if="embargoEndDate">{{ $formatDate(embargoEndDate) }}</span>
                     <span v-else>Never</span>
                     <Button label="Change" severity="secondary" @click="showPickEnd = true"/>
                  </td>
               </tr>
               <tr>
                  <td class="label">End Visibility:</td>
                  <td>{{ system.visibilityLabel('etd',embargoEndVisibility) }}</td>
               </tr>
            </template>
         </template>
         <template v-else>
            <tr>
               <td class="label">Visibility:</td>
               <td>
                  <Dropdown v-model="visibility" :options="visibilityOpts" optionLabel="label" optionValue="value" @chage="visibilityChanged()"/>
               </td>
               <template v-if="visibility == 'embargo'">
                  <tr>
                     <td class="label">Embargo End:</td>
                     <td class="embargo">
                        {{ $formatDate(embargoEndDate) }}
                        <Button label="Change" severity="secondary" @click="showPickEnd = true"/>
                     </td>
                  </tr>
                  <tr>
                     <td class="label">End Visibility:</td>
                     <td><Dropdown v-model="embargoEndVisibility" :options="endOpts" optionLabel="label" optionValue="value"/></td>
                  </tr>
               </template>
            </tr>
         </template>
      </table>
      <FloatLabel>
         <Textarea v-model="adminNotes" rows="5" />
         <label>Admin Notes</label>
      </FloatLabel>

      <div class="button-bar">
         <Button v-if="props.published" label="Unpublish" severity="warning" icon="pi pi-eye-slash" @click="unpublishWorkClicked()"/>
         <Button v-else label="Delete" severity="danger" icon="pi pi-trash" @click="deleteWorkClicked()"/>
         <span>
            <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
            <Button label="Save" @click="saveClicked()" />
         </span>
      </div>
   </Panel>
   <Dialog v-model:visible="showPickEnd" :modal="true" header="Set Embargo End Date" style="width:fit-content" position="top">
      <div class="embargo-date">
         <p>Use one of the quick helper buttons or pick a custom end date</p>
         <div class="datepick">
            <Calendar v-model="embargoEndDate" inline showWeek />
            <div class="helpers">
               <Button label="6 Months" severity="secondary" @click="setEmbargoEndDate(6,'month')"/>
               <Button label="1 Year" severity="secondary" @click="setEmbargoEndDate(1,'year')"/>
               <Button label="2 Years" severity="secondary" @click="setEmbargoEndDate(2,'year')"/>
               <Button label="5 Years" severity="secondary" @click="setEmbargoEndDate(5,'year')"/>
               <Button label="10 Years" severity="secondary" @click="setEmbargoEndDate(10,'year')"/>
               <Button label="Forever" severity="secondary" @click="embargoEndDate = null"/>
            </div>
         </div>
         <div class="controls">
            <span v-if="embargoEndDate" ><b>Embargo end date</b>: {{ $formatDate(embargoEndDate) }}</span>
            <span v-else>Embargo does not expire</span>
            <Button label="OK" @click="showPickEnd=false"/>
         </div>
      </div>
   </Dialog>
</template>

<script setup>
import Panel from 'primevue/panel'
import { ref, onMounted, computed } from 'vue'
import Textarea from 'primevue/textarea'
import FloatLabel from 'primevue/floatlabel'
import Dropdown from 'primevue/dropdown'
import Dialog from 'primevue/dialog'
import Calendar from 'primevue/calendar'
import { useSystemStore } from "@/stores/system"
import { useAdminStore } from "@/stores/admin"
import dayjs from 'dayjs'
import { useConfirm } from "primevue/useconfirm"

const confirm = useConfirm()
const admin = useAdminStore()
const system = useSystemStore()
const adminNotes = ref("")
const degree = ref("")
const department = ref("")
const visibility = ref("")
const embargoEndDate = ref(null)
const embargoEndVisibility = ref("")
const showPickEnd = ref(false)
const endOpts = ref([{name: "Worldwide", code: "open"}, {name: "UVA Only", code: "uva"}])

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
   },
   visibility: {
      type: String,
      required: true
   },
   degree: {
      type: String,
      default: "",
   },
   department: {
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
   }
})

const visibilityOpts = computed( () => {
   if (props.type == "oa") return system.oaVisibility
   return system.etdVisibility
})

onMounted( () => {
   degree.value = props.degree
   department.value = props.department
   visibility.value = props.visibility
   embargoEndDate.value = dayjs(props.embargoEndDate).toDate()
   embargoEndVisibility.value = props.embargoEndVisibility
})

const visibilityChanged = (() => {
   if ( (props.type == "etd" && visibility.value == "uva") || (props.type == "oa" && visibility.value == "embargo")) {
      embargoEndVisibility.value = "open"
      setEmbargoEndDate(6, "month")
   }
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
      },
   })
})

const setEmbargoEndDate = ((count, type) => {
   let endDate = new Date()
   if (type=="month") {
      endDate.setMonth( endDate.getMonth() + count)
   } else {
      endDate.setFullYear( endDate.getFullYear() + count)
   }
   embargoEndDate.value = endDate
})
const saveClicked = (() => {

})

</script>

<style lang="scss" scoped>
.embargo-date {
   p {
      margin:0 0 15px 0;
      padding:0;
      text-align: center;
   }
   .datepick {
      display: flex;
      flex-flow: row nowrap;
      .helpers {
         display: flex;
         flex-direction: column;
         margin-left: 15px;
         button {
            margin-bottom: 5px;
            font-size: 0.85em;
            padding: 4px 10px;
         }
      }
   }
   .controls {
      margin-top:10px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
   }
   :deep(.p-datepicker-today) {
      span {
         background: white;
         border: 1px solid var(--uvalib-grey-light);
      }
   }
}
.admin-panel {
   :deep(.p-panel-title) {
      font-weight: normal;
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
      }
      td.embargo {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
         button {
            font-size: 0.8em;
            padding: 3px 10px;
         }
      }
      :deep(.p-dropdown ) {
         width: 300px;
         .p-dropdown-label {
            font-size: 0.8em;
            padding: 4px 8px;
         }
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
      justify-content: space-between;
      align-items: center;
      margin-top: 15px;
      button {
         font-size: 0.85em;
         padding: 5px 10px;
         margin-left: 5px;
      }
   };
}
</style>