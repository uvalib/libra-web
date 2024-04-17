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
               <InputText v-model="computeID" placeholder="Compute ID" @keypress="searchKeyPressed($event)"/>
               <Button class="search" icon="pi pi-search" severity="secondary" @click="searchClicked()"/>
            </IconField>
         </Fieldset>
         <DataTable :value="admin.hits" ref="adminHits" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
               :lazy="false" :paginator="true" :alwaysShowPaginator="false"
               :rows="30" :totalRecords="admin.hits.length"
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
               :loading="admin.working"
         >
            <template #empty>
               <div v-if="computeID" class="none">No matching works found for {{ computeID }}</div>
               <div v-else class="none">Search for works by compute ID</div>
            </template>
            <Column field="namespace" header="Source">
               <template #body="slotProps">
                  <div>{{ system.namespaceLabel(slotProps.data.namespace) }}</div>
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
            <Column field="computeID" header="Author" sortable style="width: 275px"/>
            <Column field="title" header="Title" sortable>
               <template #body="slotProps">
                  <span v-if="slotProps.data.title">{{ slotProps.data.title }}</span>
                  <span v-else class="na">Undefined</span>
               </template>
            </Column>
            <Column header="Actions" style="width:110px">
               <template #body="slotProps">
                  <div  class="acts">
                     <Button class="action" icon="pi pi-file-edit" label="Edit" severity="primary" @click="editWorkClicked(slotProps.data)"/>
                     <Button v-if="!slotProps.data.publishedAt" class="action"
                        icon="pi pi-trash" label="Delete" severity="danger" @click="deleteWorkClicked(slotProps.data)"/>
                     <Button v-else class="action" icon="pi pi-eye-slash" label="Unpublish" severity="warning" @click="unpublishWorkClicked(slotProps.data)"/>
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
import Fieldset from 'primevue/fieldset'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import { useConfirm } from "primevue/useconfirm"

const router = useRouter()
const admin = useAdminStore()
const system = useSystemStore()
const confirm = useConfirm()

const computeID = ref("")

onBeforeMount( () => {
   document.title = "Libra Admin"
})

const searchKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      admin.search(computeID.value)
   }
})

const searchClicked = (() => {
   admin.search(computeID.value)
})

const editWorkClicked = ( (work) => {
   let url = `/admin/${work.type}/${work.id}`
   router.push(url)
})

const unpublishWorkClicked = ( (work) => {
   confirm.require({
      message: "Unpublish this work? It will no longer be visible to UVA or worldwide users. Are you sure?",
      header: 'Confirm Work Unpublish',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.unpublish(work.type, work.id)
      },
   })
})

const deleteWorkClicked = ( (work) => {
   confirm.require({
      message: "Delete this work? All data will be lost. This cannot be reversed. Are you sure?",
      header: 'Confirm Work Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.delete(work.type, work.id)
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

   button.search {
      margin-left: 5px;
   }

   .source {
      margin-top: 5px;
      font-size: 0.85em;
      color: var(--uvalib-grey);
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
      color: var(--uvalib-grey-light);
      font-style: italic;
      padding: 20px;
   }
   .na {
      color: var(--uvalib-grey-light);
      font-style: italic;
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