<template>
   <div class="dashboard">
      <h1>My Active Theses</h1>
      <WaitSpinner v-if="searchStore.working" :overlay="true" message="<div>Please wait...</div><p>Searching for active theses</p>" />
      <template v-else>
         <div class="help">View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</div>
         <div  v-if="searchStore.hits.length == 0" class="none">You have no active theses</div>
         <DataTable v-else :value="searchStore.hits" ref="etdWorks" dataKey="id"
            stripedRows showGridlines responsiveLayout="scroll"
            :lazy="false" :paginator="true" :alwaysShowPaginator="false"
            :rows="30" :totalRecords="searchStore.hits.length"
            paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
            :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
            currentPageReportTemplate="{first} - {last} of {totalRecords}"
         >
            <Column field="title" header="Title">
               <template #body="slotProps">
                  <span v-if="slotProps.data.title">{{ slotProps.data.title }}</span>
                  <span v-else class="none">Undefined</span>
               </template>
            </Column>
            <Column field="createdAt" header="Date Uploaded" >
               <template #body="slotProps">{{ $formatDate(slotProps.data.createdAt)}}</template>
            </Column>
            <Column header="ORCID Status"/>
            <Column header="Visibility" >
               <template #body="slotProps">
                  <div class="tag">
                     <span class="visibility" :class="slotProps.data.visibility">{{ system.visibilityLabel("etd", slotProps.data.visibility) }}</span>
                  </div>
               </template>
            </Column>
            <Column field="publishedAt" header="Date Published" >
               <template #body="slotProps">
                  <span v-if="slotProps.data.publishedAt">{{ $formatDate(slotProps.data.publishedAt)}}</span>
                  <div v-else class="tag">
                     <span class="visibility draft">Draft</span>
                  </div>
               </template>
            </Column>
            <Column header="Actions" style="width:175px;">
               <template #body="slotProps">
                  <div  class="acts">
                     <template v-if="slotProps.data.publishedAt">
                        <Button class="action" icon="pi pi-eye" label="Public View" severity="secondary"
                           size="small" @click="previewWorkClicked(slotProps.data.id)"/>
                     </template>
                     <template v-else>
                        <Button class="action" icon="pi pi-file-edit" label="Edit Thesis" severity="secondary"
                           size="small" @click="editWorkClicked(slotProps.data.id)"/>
                        <Button class="action" v-if="slotProps.data.title" icon="pi pi-check" label="Preview / Submit"
                           size="small" @click="previewWorkClicked(slotProps.data.id)"/>
                     </template>
                  </div>
               </template>
            </Column>
         </DataTable>
      </template>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onBeforeMount } from 'vue'
import { useSearchStore } from "@/stores/search"
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"

import Panel from 'primevue/panel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import WaitSpinner from "@/components/WaitSpinner.vue"

const router = useRouter()
const searchStore = useSearchStore()
const user = useUserStore()
const system = useSystemStore()

onBeforeMount( () => {
   searchStore.search(user.computeID)
})

const editWorkClicked = ( (id) => {
   let url = `/etd/${id}`
   router.push(url)
})

const previewWorkClicked = ( (id) => {
   let url = `/public/etd/${id}`
   router.push(url)
})
</script>

<style lang="scss" scoped>
.dashboard {
   margin: 0 auto;
   min-height: 600px;
   text-align: left;
   .help {
      margin: 1.5rem 0;
   }
   .none {
      color: $uva-grey-50;
      font-style: italic;
   }
   .tag {
      display: flex;
      flex-direction: column;
   }
   .acts {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      justify-content: flex-start;
      gap: 0.5rem;
      width:max-content;
   }
}
@media only screen and (min-width: 768px) {
   .dashboard {
      width: 80%;
   }
}
@media only screen and (max-width: 768px) {
   .dashboard {
      width: 95%;
   }
}
</style>