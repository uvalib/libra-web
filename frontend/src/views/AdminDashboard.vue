<template>
   <div class="admin">
      <Panel>
         <template #header>
            <div class="panel-header">
               <span>Libra Admin Dashboard</span>
               <DepositRegistrationDialog />
            </div>
         </template>
         <Fieldset legend="Find Works By">
            <IconField iconPosition="left">
               <InputIcon class="pi pi-search" />
               <InputText v-model="admin.search.computeID" placeholder="Compute ID"  @keypress="searchKeyPressed($event)"/>
            </IconField>
         </Fieldset>
         <DataTable :value="admin.hits" ref="adminHits" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
               :lazy="false" :paginator="true" :alwaysShowPaginator="true"
               :rows="30" :totalRecords="admin.hits.length"
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
               :filters="admin.filters" :globalFilterFields="['title']"
               :loading="admin.working"
         >
            <template #empty>
               <div class="none">No matching works found</div>
            </template>
            <Column field="namespace" header="Source">
               <template #body="slotProps">
                  <div>{{ system.namespaceLabel(slotProps.data.namespace) }}</div>
                  <div v-if="slotProps.data.source"><{{ slotProps.data.source }}</div>
               </template>
            </Column>
            <Column field="dateCreated" header="Created" sortable class="nowrap">
               <template #body="slotProps">{{ $formatDate(slotProps.data.dateCreated)}}</template>
            </Column>
            <Column field="dateModified" header="Modified" sortable class="nowrap">
               <template #body="slotProps">{{ $formatDate(slotProps.data.dateModified)}}</template>
            </Column>
            <Column field="id" header="ID" sortable class="nowrap"/>
            <Column field="computeID" header="Author" sortable style="width: 275px"/>
            <Column field="title" header="Title" sortable />
            <Column header="Actions" style="max-width:50px">
               <template #body="slotProps">
                  <div  class="acts">
                     <Button class="action" icon="pi pi-file-edit" label="Edit" severity="primary" @click="editWorkClicked(slotProps.data.id)"/>
                     <Button class="action" icon="pi pi-eye" label="View" severity="secondary" @click="viewWorkClicked(slotProps.data.id)"/>
                     <Button class="action" v-if="!slotProps.data.datePublished"
                        icon="pi pi-trash" label="Delete" severity="danger" @click="deleteWorkClicked(slotProps.data.id)"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </Panel>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onBeforeMount } from 'vue'
import { useAdminStore } from "@/stores/admin"
import { useSystemStore } from "@/stores/system"
import Panel from 'primevue/panel'
import DepositRegistrationDialog from "@/components/DepositRegistrationDialog.vue"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Fieldset from 'primevue/fieldset'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import { useConfirm } from "primevue/useconfirm"

const router = useRouter()
const admin = useAdminStore()
const system = useSystemStore()
const confirm = useConfirm()

onBeforeMount( () => {
   document.title = "Libra Admin"
})

const searchKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      admin.search()
   }
})

const resetSearchClicked = (() => {
   admin.resetSearch()
})
const editWorkClicked = ( (id) => {
   let url = `/${admin.scope}/${id}`
   router.push(url)
})

const viewWorkClicked = ( (id) => {
   let url = `/public/${admin.scope}/${id}`
   router.push(url)
})

const deleteWorkClicked = ( (id) => {
   confirm.require({
      message: "Delete this work? All data will be lost. This cannot be reversed. Are you sure?",
      header: 'Confirm Work Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: async (  ) => {
         await admin.delete(admin.scope, id)
         if ( system.showError == false) {
            searchStore.removeDeletedWork(id)
         }
      },
   })
})
</script>

<style lang="scss" scoped>
.admin {
   width: 95%;
   margin: 2% auto;
   min-height: 600px;
   text-align: left;

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
      color: var(--uvalib-grey-light);
      font-style: italic;
      padding: 20px;
   }

   :deep(td.nowrap),  :deep(th){
      white-space: nowrap;
   }

   .hdr {
      width: 100%;
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      div {
         font-size: 1.25em;
         color: var(--uvalib-text);
      }
   }
   .type {
      font-size: 0.9em;
      padding: 4px 8px;
      border-radius: 20px;
      font-weight: bold;
   }
   .type.etd {
      background-color: var(--uvalib-blue-alt);
      color: white;
   }
   .type.oa {
      background-color: var(--uvalib-brand-orange);
      color: white;
   }
   .acts {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      justify-content: flex-start;
      button.action {
         font-size: 0.85em;
         margin-top: 10px;
         width: 100%;
         padding: 4px 8px;
      }
      button.action:first-of-type {
         margin-top: 0;
      }
   }
}
</style>