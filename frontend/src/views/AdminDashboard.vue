<template>
   <div class="admin">
      <Panel>
         <template #header>
            <div class="panel-header">
               <span>Libra Admin Dashboard</span>
               <DepositRegistrationDialog />
            </div>
         </template>

         <div class="search">
            <IconField iconPosition="left" class="query">
               <InputIcon class="pi pi-search" />
               <InputText v-model="admin.query" @keypress="searchKeyPressed($event)" fluid aria-label="search works"/>
            </IconField>
            <label>
               Publication Status:
               <Select v-model="admin.statusFilter" :options="publishOpts" optionLabel="label" optionValue="value" @update:modelValue="admin.search()"/>
            </label>
            <label>
               Source:
               <Select v-model="admin.sourceFilter" :options="sourceOpts" optionLabel="label" optionValue="value" @update:modelValue="admin.search()"/>
            </label>
            <Button severity="secondary" label="Reset Search" @click="admin.resetSearch"/>
         </div>

         <DataTable :value="admin.hits" ref="adminHits" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll"
               :lazy="true" :paginator="true" :alwaysShowPaginator="false"
               @page="onPage($event)"  paginatorPosition="both"
               :first="admin.offset" :rows="admin.limit" :totalRecords="admin.total"
               paginatorTemplate="PrevPageLink CurrentPageReport NextPageLink"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
               :loading="admin.working" removableSort @sort="onSort($event)"  :sortField="admin.sortField"
         >
            <template #header v-if="admin.total == 1000">
               <div class="cap-note"><i class="pi pi-exclamation-triangle"></i>Results are capped at 1000 hits. Please narrow your search.</div>
            </template>
            <template #empty>
               <div v-if="admin.searchCompleted" class="none">No matching works found for {{ admin.query }}</div>
               <div v-else class="none">Search for works</div>
            </template>
            <Column field="source" header="Source">
               <template #body="slotProps">
                  <div v-if="slotProps.data.source=='sis'" style="text-transform: uppercase;">{{ slotProps.data.source }}</div>
                  <div v-else-if="slotProps.data.source=='libra-oa'" style="white-space: nowrap;">Libra-OA</div>
                  <div v-else style="text-transform: capitalize;">{{ slotProps.data.source }}</div>
               </template>
            </Column>
            <Column field="id" header="ID" class="nowrap"/>
            <Column field="created" header="Created" sortable class="nowrap">
               <template #body="slotProps">{{ $formatDateTime(slotProps.data.created)}}</template>
            </Column>
            <Column field="modified" header="Modified" sortable class="nowrap">
               <template #body="slotProps">
                  <div v-if="slotProps.data.modified">{{ $formatDateTime(slotProps.data.modified) }}</div>
                  <div v-else class="na">N/A</div>
               </template>
            </Column>
            <Column field="published" header="Published" class="nowrap">
               <template #body="slotProps">
                  <div v-if="slotProps.data.published">{{ $formatDateTime(slotProps.data.published) }}</div>
                  <div v-else class="na">N/A</div>
               </template>
            </Column>
            <Column field="author" header="Author" style="width: 275px">
               <template #body="slotProps">
                  {{ slotProps.data.author.lastName }}, {{ slotProps.data.author.firstName }}
               </template>
            </Column>
            <Column field="title" header="Title" sortable>
               <template #body="slotProps">
                  <span v-if="slotProps.data.title">{{ slotProps.data.title }}</span>
                  <span v-else class="na">Undefined</span>
               </template>
            </Column>
            <Column header="Actions" style="width:110px">
               <template #body="slotProps">
                  <div  class="acts">
                     <Button v-if="slotProps.data.author.computeID" label="Become User" severity="secondary"
                        size="small" @click="becomeUser(slotProps.data.author.computeID)"
                     />
                     <Button label="Edit" severity="primary" size="small" @click="editWorkClicked(slotProps.data.id)"/>
                     <Button v-if="slotProps.data.published" label="Public View" severity="info"
                        size="small" @click="viewWorkClicked(slotProps.data.id)"
                     />
                     <AuditsPanel :workID="slotProps.data.id"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </Panel>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onBeforeMount, computed } from 'vue'
import { useAdminStore } from "@/stores/admin"
import { useSystemStore } from "@/stores/system"
import Panel from 'primevue/panel'
import DepositRegistrationDialog from "@/components/DepositRegistrationDialog.vue"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import AuditsPanel from '@/components/AuditsPanel.vue'

const router = useRouter()
const admin = useAdminStore()
const system = useSystemStore()

const publishOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "Draft", value: "draft"}, {label: "Published", value: "published"} ]
})
const sourceOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "SIS", value: "sis"}, {label: "Optional", value: "optional"} ]
})

onBeforeMount( () => {
   document.title = "Libra Admin"
})

const onPage = ((event) => {
   admin.offset = event.first
   admin.search()
})

const onSort = ((event) => {
   admin.sortField = ""
   admin.sortOrder = ""
   if (event.sortOrder == 1) {
      admin.sortField = event.sortField
      admin.sortOrder = "asc"

   } else if (event.sortOrder == -1) {
      admin.sortField = event.sortField
      admin.sortOrder = "desc"
   }
   admin.search()
})

const searchKeyPressed = ((event) => {
   admin.sortField = ""
   admin.sortOrder = ""
   admin.offset = 0
   admin.searchCompleted = false
   if (event.keyCode == 13) {
      admin.search()
   }
})

const editWorkClicked = ( (id) => {
   let url = `/admin/etd/${id}`
   router.push(url)
})

const viewWorkClicked = ( (id) => {
   let url = `/public/etd/${id}`
   router.push(url)
})

const becomeUser = ((computeID) => {
   admin.becomeUser( computeID )
})
</script>

<style lang="scss" scoped>
.admin {
   width: 95%;
   margin: 2% auto;
   min-height: 600px;
   text-align: left;
   .cap-note {
      text-align: center;
      display: flex;
      flex-flow: row nowrap;
      gap: 10px;
      justify-content: center;
      align-items: center;
      i {
         font-size: 1.25rem
      }
   }

   .p-datatable {
      margin-top: 20px;
   }

   .search {
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-items: center;
      gap: 1rem;
      .query {
         flex-grow: 1;
      }
      .p-select {
         margin-left: 5px;
      }
   }
   .panel-header {
      font-weight: bold;
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      width: 100%;
   }

   .none {
      text-align: center;
      font-size: 1.25em;
      color: $uva-grey-A;
      font-style: italic;
      padding: 20px;
   }
   .na {
      color: $uva-grey-A;
      font-style: italic;
   }
   .acts {
      display: flex;
      flex-direction: column;
      align-items: stretch;
      justify-content: flex-start;
      gap: 0.3rem;
   }
}
</style>