<template>
   <div class="admin">
      <Panel>
         <template #header>
            <div class="panel-header">
               <span>Libra Admin Dashboard</span>
               <DepositRegistrationDialog />
            </div>
         </template>

         <IconField iconPosition="left">
            <InputIcon class="pi pi-search" />
            <InputText v-model="queryString" placeholder="Compute ID" @keypress="searchKeyPressed($event)" fluid/>
         </IconField>

         <DataTable :value="admin.hits" ref="adminHits" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll"
               :lazy="false" :paginator="true" :alwaysShowPaginator="false"
               :rows="30" :totalRecords="admin.hits.length"
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
               :loading="admin.working"
         >
            <template #empty>
               <div v-if="queryString" class="none">No matching works found for {{ queryString }}</div>
               <div v-else class="none">Search for works by compute ID</div>
            </template>
            <Column field="namespace" header="Source">
               <template #body="slotProps">
                  <div>{{ system.namespace.label }}</div>
                  <div v-if="slotProps.data.source" class="source">( {{ slotProps.data.source }} )</div>
               </template>
            </Column>
            <Column field="createdAt" header="Created" sortable class="nowrap">
               <template #body="slotProps">{{ $formatDateTime(slotProps.data.createdAt)}}</template>
            </Column>
            <Column field="modifiedAt" header="Modified" sortable class="nowrap">
               <template #body="slotProps">
                  <div v-if="slotProps.data.modifiedAt">{{ $formatDateTime(slotProps.data.modifiedAt) }}</div>
                  <div v-else class="na">N/A</div>
               </template>
            </Column>
            <Column field="publishedAt" header="Published" sortable class="nowrap">
               <template #body="slotProps">
                  <div v-if="slotProps.data.publishedAt">{{ $formatDateTime(slotProps.data.publishedAt) }}</div>
                  <div v-else class="na">N/A</div>
               </template>
            </Column>
            <Column field="id" header="ID" sortable class="nowrap"/>
            <Column field="author" header="Author" sortable style="width: 275px">
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
                     <Button class="action" label="Edit" severity="primary" size="small" @click="editWorkClicked(slotProps.data.id)"/>
                     <Button v-if="slotProps.data.publishedAt" class="action" label="Public View" severity="secondary"
                        size="small" @click="viewWorkClicked(slotProps.data.id)"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </Panel>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onBeforeMount, ref } from 'vue'
import { useAdminStore } from "@/stores/admin"
import { useSystemStore } from "@/stores/system"
import Panel from 'primevue/panel'
import DepositRegistrationDialog from "@/components/DepositRegistrationDialog.vue"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'

const router = useRouter()
const admin = useAdminStore()
const system = useSystemStore()

const queryString = ref("")

onBeforeMount( () => {
   document.title = "Libra Admin"
})

const searchKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      admin.search(queryString.value)
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
</script>

<style lang="scss" scoped>
.admin {
   width: 95%;
   margin: 2% auto;
   min-height: 600px;
   text-align: left;

   .p-datatable {
      margin-top: 20px;
   }

   .source {
      margin-top: 5px;
      font-size: 0.85em;
      color: $uva-grey;
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
      color: $uva-grey-100;
      font-style: italic;
      padding: 20px;
   }
   .na {
      color: $uva-grey-100;
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