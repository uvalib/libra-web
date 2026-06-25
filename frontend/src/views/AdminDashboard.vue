<template>
   <div class="admin">
      <h1>
         <span>Admin Dashboard</span>
         <Button severity="secondary" size="small" icon="pi pi-pen-to-square" label="Supported Mime Types" @click="editMimeTypesClicked()"/>
      </h1>
      <WaitSpinner v-if="admin.working" :overlay="true" message="<div>Please wait...</div><p>Loading works</p>" />

      <div class="search">
         <div class="search-input">
            <IconField iconPosition="left" class="query">
               <InputIcon class="pi pi-search" />
               <InputText v-model="admin.query" @keypress="searchKeyPressed($event)" fluid aria-label="search works" id="admin-search"/>
            </IconField>
            <Button label="Search" @click="admin.search()" :loading="admin.working" :disabled="admin.working"/>
         </div>
         <Button severity="secondary" label="Reset Search" @click="admin.resetSearch()"/>
         <Button severity="secondary" label="Export" @click="admin.exportCSV()"
            :disabled="admin.total == 0 || admin.working || admin.searchCompleted == false|| admin.total >= system.maxSearchHits" :loading="admin.working"
         />
      </div>

      <Fieldset legend="Filter Results">
         <div class="filter-content">
            <div class="row">
               <div class="filter">
                  <label for="status-filter">Status:</label>
                  <Select v-model="admin.statusFilter" :options="publishOpts" optionLabel="label" optionValue="value" id="status-filter" @update:modelValue="statusChanged"/>
               </div>
               <div class="filter">
                  <label for="src-filter">Source:</label>
                  <Select id="src-filter" v-model="admin.sourceFilter" :options="sourceOpts" optionLabel="label" optionValue="value" />    
               </div>
            </div>
            <div class="row">
               <div class="filter">
                  <label>Date Created:</label>
                  <div class="range">
                     <DatePicker v-model="admin.createdFilter.from" :maxDate="new Date()" dateFormat="yy-mm-dd" placeholder="Created from" 
                        showIcon :showClear="true" @update:modelValue="admin.createdFilter.to = admin.createdFilter.from"
                     />
                     <span>to</span>
                     <DatePicker v-model="admin.createdFilter.to" :maxDate="new Date()" :minDate="admin.createdFilter.from" 
                        dateFormat="yy-mm-dd" placeholder="Created to" showIcon :showClear="true"
                     />
                  </div>
               </div>
    
               <div class="filter">
                  <label>Date Published:</label>
                  <div class="range">
                     <DatePicker v-model="admin.publishedFilter.from" :maxDate="new Date()" dateFormat="yy-mm-dd" placeholder="Published from" 
                        showIcon :showClear="true" @update:modelValue="admin.publishedFilter.to = admin.publishedFilter.from"
                     />
                     <span>to</span>
                     <DatePicker v-model="admin.publishedFilter.to" :maxDate="new Date()" :minDate="admin.publishedFilter.from" dateFormat="yy-mm-dd" 
                        placeholder="Published to" showIcon :showClear="true"
                     />
                  </div>
               </div>
            </div>
            <div class="footer">
               <span><b>NOTE</b>: For date ranges both 'to' and 'from' are required.</span>
               <Button label="Apply Filter" icon="pi pi-filter" @click="applyFilter" style="margin-left: auto;"/>
            </div>
         </div>
      </Fieldset>

      <h2 v-if="admin.searchCompleted == false">
         Recent Activity
      </h2>
      <DataTable :value="admin.hits" ref="adminHits" dataKey="id"
            stripedRows showGridlines responsiveLayout="scroll"
            :lazy="true" :paginator="true" 
            @page="onPage($event)"  paginatorPosition="both"
            :first="admin.offset" :rows="admin.limit" :totalRecords="admin.total" :rowsPerPageOptions="[25,50,100]"
            paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
            currentPageReportTemplate="{first} - {last} of {totalRecords}"
            :loading="admin.working" removableSort @sort="onSort($event)"  :sortField="admin.sortField"
      >
         <template #header v-if="admin.total >= system.maxSearchHits">
            <div class="cap-note"><i class="pi pi-exclamation-triangle"></i>Results are capped at {{ system.maxSearchHits }} hits. Please narrow your search.</div>
         </template>
         <template #empty>
            <div v-if="admin.searchCompleted" class="none">No matching works found for {{ admin.query }}</div>
            <div v-else class="none">Enter a search query to see matching works</div>
         </template>
         <template #loading>Loading works. Please wait. </template>
         <Column field="source" header="Source" style="width:120px">
            <template #body="slotProps">
               <div v-if="slotProps.data.source=='sis'" style="text-transform: uppercase;">{{ slotProps.data.source }}</div>
               <div v-else-if="slotProps.data.source=='libra-oa'" style="white-space: nowrap;">Libra-OA</div>
               <div v-else style="text-transform: capitalize;">{{ slotProps.data.source }}</div>
            </template>
         </Column>
         <Column field="id" header="ID" class="nowrap" style="width:230px"/>
         <Column field="created" header="Created" dataType="date" sortable class="nowrap">
            <template #body="slotProps">{{ $formatDateTime(slotProps.data.created)}}</template>
         </Column>
         <Column field="modified" header="Modified" sortable class="nowrap">
            <template #body="slotProps">
               <div v-if="slotProps.data.modified">{{ $formatDateTime(slotProps.data.modified) }}</div>
               <div v-else class="na">N/A</div>
            </template>
         </Column>
         <Column field="published" header="Published" sortable class="nowrap">
            <template #body="slotProps">
               <div v-if="slotProps.data.published">{{ $formatDateTime(slotProps.data.published) }}</div>
               <div v-else class="na">N/A</div>
            </template>
         </Column>
         <Column field="author" header="Author">
            <template #body="slotProps">
               {{ slotProps.data.author.lastName }}, {{ slotProps.data.author.firstName }}
            </template>
         </Column>
         <Column field="title" header="Title" sortable>
            <template #body="slotProps">
               <span v-if="slotProps.data.title" :id="slotProps.data.id" v-html="slotProps.data.title"></span>
               <span v-else class="na">Undefined</span>
            </template>
         </Column>
         <Column header="Actions" style="width:75px">
            <template #body="slotProps">
               <div  class="acts">
                  <Button asChild v-slot="btnProps" severity="secondary" size="small">
                     <RouterLink :to="`/admin/etd/${slotProps.data.id}`" :class="btnProps.class" :aria-describedby="slotProps.data.id">
                        Edit Work
                     </RouterLink>
                  </Button>
                  <AuditsPanel :workID="slotProps.data.id" :workTitle="slotProps.data.title"/>
                  <Button v-if="slotProps.data.author.computeID" :label="`Become User ${slotProps.data.author.computeID}`" severity="secondary"
                     size="small" @click="admin.becomeUser(slotProps.data.author.computeID)"
                  />
                  <a v-if="slotProps.data.published" class="public-view" target="_blank" aria-describedby="new-window"
                     :href="`./public_view/${slotProps.data.id}`">
                     <span>Public View</span>
                     <i class="pi pi-external-link"/>
                  </a>
               </div>
            </template>
         </Column>
      </DataTable>
   </div>
   <Dialog v-model:visible="editMime" :modal="true" style="max-width: 450px;" header="Supported Mime Types" @hide="editMime=false">
      <div class="add-mime">
         <InputText v-model="newMime" fluid />
         <Button label="Add" icon="pi pi-plus" @click="addMimeTypeClicked()" />
      </div>
      <div class="types">
         <Chip v-for="mime in system.mimeTypes" :label="mime" removable @remove="removeMimeType(mime)"/>
      </div>
      <template #footer>
         <Button label="Cancel" severity="secondary" @click="editMime=false"/>
         <Button label="Update" autofocus  @click="updateMimeTypes()"/>
      </template>
   </Dialog>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useAdminStore } from "@/stores/admin"
import { useSystemStore } from "@/stores/system"
import Fieldset from 'primevue/fieldset'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import AuditsPanel from '@/components/AuditsPanel.vue'
import DatePicker from 'primevue/datepicker'
import Dialog from 'primevue/dialog'
import Chip from 'primevue/chip'
import WaitSpinner from '@/components/WaitSpinner.vue'

const admin = useAdminStore()
const system = useSystemStore()

const newMime = ref("")
const editMime = ref(false)

const publishOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "Draft", value: "draft"}, {label: "Published", value: "published"} ]
})
const sourceOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "SIS", value: "sis"}, {label: "Optional", value: "optional"} ]
})

onMounted( () => {
   if (admin.searchCompleted == false) {
      admin.getRecentActivity()
   }
})

const editMimeTypesClicked = (() => {
   system.getMimeTypes()
   editMime.value=true
})

const removeMimeType = ((t)=> {
    system.mimeTypes =  system.mimeTypes.filter( mt => mt != t) 
})

const addMimeTypeClicked = (() => {
   system.mimeTypes.push( newMime.value )
   newMime.value = ""
})

const updateMimeTypes = (()=> {
   admin.updateMimeTypes( system.mimeTypes )
   editMime.value = false
})

const statusChanged = (() => {
   if (admin.statusFilter == 'draft' ) {
      admin.publishedFilter.from = null
      admin.publishedFilter.to = null
   }
})

const applyFilter = (() => {
   if (admin.isCreatedFilterValid == false || admin.isPublishedFilterValid == false) {
      system.toastError("Incomplete date rage", "Please specify both to and from dates in the date filter")
      return
   }
   admin.search()
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
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .admin {
       width: 90%;
      .search {
         gap: 1rem;
         justify-content: flex-start;
      }
   }
}
@media only screen and (max-width: 768px) {
   .admin {
      width: 95%;
      gap: 10px;
      .search {
         gap: 10px;
         justify-content: flex-end;
      }
   }
}

.admin {
   margin: 0 auto 50px;
   min-height: 600px;
   text-align: left;
   display: flex;
   flex-direction: column;
   gap: 20px;

   h1 {
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      gap: 10px;
      padding-bottom: 0;
   }

   div.filter-content {
      display: flex;
      flex-direction: column;
      gap: 20px;
      padding-top: 10px;
      .row {
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-start;
         align-items: baseline;
         gap: 25px;
      }
      .filter, .range {
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-start;
         align-items: baseline;
         gap: 10px;
         label {
            white-space: nowrap;;
         }
      }
      .footer {
         display: flex;
         flex-flow: row wrap;
         justify-content: space-between;
         align-items: baseline;
      }
   }

   h2 {
      margin: 0;
   }

   
   a.public-view {
      display: flex;
      flex-flow: row nowrap;
      gap: 5px;
      align-items: center;
      justify-content: center;
      margin-top: 10px;
   }

   .cap-note {
      text-align: center;
      display: flex;
      flex-flow: row nowrap;
      gap: 10px;
      justify-content: center;
      align-items: center;
      background-color: $uva-red-A;
      color: white;
      font-weight: bold;
      padding: 5px;
      border-radius: 0.3rem;
      i {
         font-weight: bold;
         font-size: 1.25rem
      }
   }

   .search {
      display: flex;
      flex-flow: row wrap;
      align-items: center;
      justify-content: flex-start;
      margin-top: 20px;
      label {
         display: flex;
         flex-flow: row wrap;
         gap: 10px;
         justify-content: flex-start;
         align-items: baseline;
      }
      .p-select {
         width: 130px;
         text-align: center;
      }
      .search-input {
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: center;
         gap: 0;
         flex-grow: 1;
         input {
            border-radius: 0.3rem 0 0 0.3rem;
            border-right:0;
         }
         button {
            border-radius:  0 0.3rem 0.3rem 0;
         }
         .query {
            flex-grow: 1;
         }
      }
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
      gap: 0.5rem;
   }
}
.add-mime {
   display: flex;
   flex-flow: row nowrap;
   gap: 10px;
   margin-bottom: 15px;
}
.types {
   display: flex;
   flex-flow: row wrap;
   gap: 10px;
}
</style>