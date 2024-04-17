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
                     {{ $formatDate(embargoEndDate) }}
                     <Button label="Change" severity="secondary" @click="showPickEnd = true"/>
                  </td>
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
         <Button severity="secondary" label="Cancel" @click="emit('cancel')"/>
         <Button label="Save" @click="saveClicked()" />
      </div>
   </Panel>
   <Dialog v-model:visible="showPickEnd" :modal="true" header="Set Embargo End Date">
   </Dialog>
</template>

<script setup>
import Panel from 'primevue/panel'
import { ref, onMounted, computed } from 'vue'
import Textarea from 'primevue/textarea'
import FloatLabel from 'primevue/floatlabel'
import Dropdown from 'primevue/dropdown'
import Dialog from 'primevue/dialog'
import { useSystemStore } from "@/stores/system"
import dayjs from 'dayjs'

const system = useSystemStore()
const adminNotes = ref("")
const degree = ref("")
const department = ref("")
const visibility = ref("")
const embargoEndDate = ref(null)
const embargoEndVisibility = ref("")
const showPickEnd = ref(false)
const endOpts = ref([{name: "Worldwide", code: "opem"}, {name: "UVA Only", code: "uva"}])

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
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      console.log(endDate)
      embargoEndDate.value = endDate
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