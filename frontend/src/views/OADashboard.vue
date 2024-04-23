<template>
   <div class="dashboard">
      <Panel>
         <template #header>
            <span class="hdr">
               <div>My LibraOpen Works</div>
               <Button icon="pi pi-plus" label="Create new work" @click="createWorkClicked"/>
            </span>
         </template>
         <WaitSpinner v-if="searchStore.working" :overlay="true" message="<div>Please wait...</div><p>Searching for active thesis</p>" />
         <template v-else>
            <div  v-if="searchStore.hits.length == 0" class="none">
               You have no active works
            </div>
            <DataTable v-else :value="searchStore.hits" ref="oaWorks" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
               :lazy="false" :paginator="true" :alwaysShowPaginator="false"
               :rows="30" :totalRecords="searchStore.hits.length"
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
            >
               <Column field="title" header="Title" />
               <Column field="createdAt" header="Date Uploaded" >
                  <template #body="slotProps">{{ $formatDate(slotProps.data.createdAt)}}</template>
               </Column>
               <Column header="ORCID Status"/>
               <Column field="visibility" header="Visibility" >
                  <template #body="slotProps">
                     <span class="visibility" :class="slotProps.data.visibility">{{ system.visibilityLabel("oa",slotProps.data.visibility) }}</span>
                  </template>
               </Column>
               <Column field="publishedAt" header="Date Published" >
                  <template #body="slotProps">
                     <span v-if="slotProps.data.publishedAt">{{ $formatDate(slotProps.data.publishedAt)}}</span>
                     <span v-else class="visibility draft">Draft</span>
                  </template>
               </Column>
               <Column header="Actions" style="max-width:130px">
                  <template #body="slotProps">
                     <div  class="acts">
                        <Button class="action" icon="pi pi-file-edit" label="Edit Work" severity="secondary" @click="editWorkClicked(slotProps.data.id)"/>
                        <template v-if="slotProps.data.publishedAt">
                           <Button class="action" icon="pi pi-eye" label="Public View" severity="secondary" @click="previewWorkClicked(slotProps.data.id)"/>
                        </template>
                        <template v-else>
                           <Button class="action" icon="pi pi-eye" label="Preview / Publish" @click="previewWorkClicked(slotProps.data.id)"/>
                           <Button class="action" icon="pi pi-trash" label="Delete Work" severity="danger" @click="deleteWorkClicked(slotProps.data.id)"/>
                        </template>
                     </div>
                  </template>
               </Column>
            </DataTable>
         </template>
      </Panel>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onBeforeMount } from 'vue'
import { useSearchStore } from "@/stores/search"
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"
import { useOAStore } from "@/stores/oa"
import Panel from 'primevue/panel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useConfirm } from "primevue/useconfirm"

const router = useRouter()
const searchStore = useSearchStore()
const user = useUserStore()
const system = useSystemStore()
const oaRepo = useOAStore()
const confirm = useConfirm()

onBeforeMount( () => {
   document.title = "LibraOpen"
   searchStore.search("oa", user.computeID)
})

const editWorkClicked = ( (id) => {
   let url = `/oa/${id}`
   router.push(url)
})
const previewWorkClicked = ( (id) => {
   let url = `/public/oa/${id}`
   router.push(url)
})
const deleteWorkClicked = ( (id) => {
   confirm.require({
      message: "Delete this work? All data will be lost. This cannot be reversed. Are you sure?",
      header: 'Confirm Work Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: async (  ) => {
         await oaRepo.deleteWork(id)
         if ( system.showError == false) {
            searchStore.removeDeletedWork(id)
         }
      },
   })
})

const createWorkClicked = (() => {
   router.push("/oa/new")
})
</script>

<style lang="scss" scoped>
.dashboard {
   width: 70%;
   margin: 50px auto;
   min-height: 600px;
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
   .none {
      color: var(--uvalib-grey-light);
      font-style: italic;
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