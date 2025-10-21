<template>
   <div class="admin">
      <h1>Admin Dashboard</h1>
      <div class="search-setup">
         <div class="search">
            <label for="admin-search">Search for works:</label>
            <IconField iconPosition="left" class="query">
               <InputIcon class="pi pi-search" />
               <InputText v-model="admin.query" @keypress="searchKeyPressed($event)" fluid aria-label="search works" id="admin-search"/>
            </IconField>
            <Button label="Search" @click="admin.search()" :loading="admin.working"/>
            <Button severity="secondary" label="Reset Search" @click="resetSearch"/>
         </div>
         <div class="search-filter">
            <label for="status-filter">Publication Status:</label>
            <select v-model="admin.statusFilter" id="status-filter">
               <option v-for="o in publishOpts" :value="o.value">{{ o.label }}</option>
            </select>
            <div class="pub-dates"  v-if="admin.statusFilter=='published'">
               <label for="from-picker">From:</label>
               <InputMask  input-id="from-picker" v-model="admin.fromDate" mask="9999-99-99" placeholder="yyyy-mm-dd" />
               <label for="to-picker">To: </label>
               <InputMask input-id="to-picker" v-model="admin.toDate" mask="9999-99-99" placeholder="yyyy-mm-dd" />
            </div>
            <label for="source-filter">Source:</label>
            <select v-model="admin.sourceFilter" id="source-filter">
               <option v-for="o in sourceOpts" :value="o.value">{{ o.label }}</option>
            </select>
            <div class="filter-acts">
               <Button severity="secondary" label="Apply Filters" @click="admin.search()" :loading="admin.working"/>
               <Button severity="secondary" label="Export" @click="exportClicked()"
                  :disabled="admin.total == 0 || admin.total >= 1000" :loading="admin.working"/>
            </div>
         </div>
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
            <div v-else class="none">Enter a search query to see matching works</div>
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
         <Column field="published" header="Published" sortable class="nowrap">
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
               <span v-if="slotProps.data.title" :id="slotProps.data.id" v-html="slotProps.data.title"></span>
               <span v-else class="na">Undefined</span>
            </template>
         </Column>
         <Column header="Actions">
            <template #body="slotProps">
               <div  class="acts">
                  <Button asChild v-slot="btnProps" severity="secondary" size="small">
                     <RouterLink :to="`/admin/etd/${slotProps.data.id}`" :class="btnProps.class" :aria-describedby="slotProps.data.id">
                        Edit Work
                     </RouterLink>
                  </Button>
                  <Button v-if="slotProps.data.published" asChild v-slot="btnProps" severity="secondary" size="small">
                     <RouterLink :to="`/public_view/${slotProps.data.id}`" :class="btnProps.class" :aria-describedby="slotProps.data.id">
                        <span>Public View</span>
                     </RouterLink>
                  </Button>
                  <AuditsPanel :workID="slotProps.data.id" :workTitle="slotProps.data.title"/>
                  <Button v-if="slotProps.data.author.computeID" :label="`Become User ${slotProps.data.author.computeID}`" severity="secondary"
                     size="small" @click="becomeUser(slotProps.data.author.computeID)"
                  />
               </div>
            </template>
         </Column>
      </DataTable>
   </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useAdminStore } from "@/stores/admin"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import InputMask from 'primevue/inputmask'
import AuditsPanel from '@/components/AuditsPanel.vue'
import { useHead } from '@unhead/vue'

const admin = useAdminStore()
const fromDate = ref()
const toDate = ref()

const publishOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "Draft", value: "draft"}, {label: "Published", value: "published"} ]
})
const sourceOpts = computed(() => {
   return[ {label: "Any", value: "any"}, {label: "SIS", value: "sis"}, {label: "Optional", value: "optional"} ]
})

useHead({
   title: 'Libra Admin Dashboard'
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

const resetSearch = (() => {
   toDate.value = null
   fromDate.value = null
   admin.resetSearch()
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

const exportClicked = (() => {
   admin.exportCSV()
})

const becomeUser = ((computeID) => {
   admin.becomeUser( computeID )
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .admin {
       width: 90%;
      .search-setup {
         gap: 1rem;
         .search {
            gap: 1rem;
            justify-content: flex-start;
         }
         .search-filter {
            flex-flow: row wrap;
            justify-content: flex-start;
            align-items: center;
            gap: 1rem;
            .pub-dates {
               display: flex;
               flex-flow: row wrap;
               justify-content: flex-end;
               align-items: center;
               gap: 1rem;
            }
         }
      }
   }
}
@media only screen and (max-width: 768px) {
   .admin {
      width: 95%;
      .search-setup {
         gap: 10px;
         .search {
            gap: 10px;
            justify-content: flex-end;
         }
         .search-filter {
            display: flex;
            flex-direction: column;
            gap: 10px;

            .pub-dates {
               display: flex;
               flex-direction: column;
               gap: 10px;
               .p-datepicker {
                  flex-grow: 1;
                  :deep(input.p-inputtext) {
                     width: 100%;
                  }
               }
            }
         }
      }
   }
}
.admin {
   margin: 0 auto 50px;
   min-height: 600px;
   text-align: left;
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

   .p-datatable {
      margin-top: 20px;
   }

   .search-setup {
      display: flex;
      flex-direction: column;
      .search {
         display: flex;
         flex-flow: row wrap;
         align-items: center;
         margin-top: 20px;
         .query {
            flex-grow: 1;
         }
      }
      .search-filter {
         display: flex;
         .filter-acts {
            margin-left: auto;
            display: flex;
            flex-flow: row wrap;
            gap: 1rem;
         }
         .pub-dates {
            display: flex;
         }
      }
   }

   .panel-header {
      font-weight: bold;
      display: flex;
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
      gap: 0.5rem;
   }
}
</style>