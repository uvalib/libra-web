<template>
   <div class="dashboard">
      <Panel>
         <template #header>
            <span class="hdr">
               <div>My LibraOpen Works</div>
               <Button icon="pi pi-plus" label="Create new work" @click="createWorkClicked"/>
            </span>
         </template>
         <DataTable :value="searchStore.hits" ref="oaWorks" dataKey="id"
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
                  <span class="visibility" :class="slotProps.data.visibility">{{ slotProps.data.visibility }}</span>
               </template>
            </Column>
            <Column header="Actions">
               <template #body="slotProps">
                  <div  class="acts">
                     <Button class="action" icon="pi pi-file-edit" label="Edit Work" severity="secondary" text @click="editWorkClicked(slotProps.data.id)"/>
                     <Button class="action" icon="pi pi-eye" label="Public Preview" severity="secondary" text @click="previewWorkClicked(slotProps.data.id)"/>
                     <Button class="action" icon="pi pi-trash" label="Delete Work" severity="danger" text @click="deleteWorkClicked(slotProps.data.id)"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </Panel>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import { useSearchStore } from "@/stores/search"
import { useUserStore } from "@/stores/user"
import Panel from 'primevue/panel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'


const router = useRouter()
const searchStore = useSearchStore()
const user = useUserStore()

onMounted( () => {
   searchStore.search("oa", user.computeID)
})

const editWorkClicked = ( (id) => {

})
const previewWorkClicked = ( (id) => {

})
const deleteWorkClicked = ( (id) => {

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
   .acts {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      justify-content: flex-start;
      button.action {
         font-size: 0.9em;
         margin-bottom: 5px;
      }
   }
}
</style>