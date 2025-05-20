<template>
   <div class="dashboard">
      <h1>My Active Theses</h1>
      <WaitSpinner v-if="user.working" :overlay="true" message="<div>Please wait...</div><p>Searching for active theses</p>" />
      <template v-else>
         <div class="help">View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</div>
         <div  v-if="user.theses.length == 0" class="none">You have no active theses</div>
         <DataTable v-else :value="user.theses" ref="etdWorks" dataKey="id"
            stripedRows showGridlines responsiveLayout="scroll"
            :lazy="false" :paginator="false"
         >
            <Column field="title" header="Title">
               <template #body="slotProps">
                  <span v-if="slotProps.data.title">{{ slotProps.data.title }}</span>
                  <span v-else class="none">Undefined</span>
               </template>
            </Column>
             <Column field="created" header="Date Uploaded" >
               <template #body="slotProps">{{ $formatDateTime(slotProps.data.created)}}</template>
            </Column>
            <Column header="Visibility" >
               <template #body="slotProps">
                  <div class="tag">
                     <span class="visibility" :class="slotProps.data.visibility">{{ system.visibilityLabel("etd", slotProps.data.visibility) }}</span>
                  </div>
               </template>
            </Column>
            <Column field="published" header="Date Published" >
               <template #body="slotProps">
                  <span v-if="slotProps.data.published">{{ $formatDateTime(slotProps.data.published)}}</span>
                  <div v-else class="tag">
                     <span class="visibility draft">Draft</span>
                  </div>
               </template>
            </Column>
            <Column header="Actions" style="width:175px;">
               <template #body="slotProps">
                  <div  class="acts">
                     <Button v-if="slotProps.data.published" class="action" label="Public View" severity="secondary"
                        size="small" @click="previewWorkClicked(slotProps.data.id)"/>
                     <Button v-else class="action" label="Edit Thesis" severity="secondary"
                        size="small" @click="editWorkClicked(slotProps.data.id)"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </template>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"

import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import WaitSpinner from "@/components/WaitSpinner.vue"

const router = useRouter()
const user = useUserStore()
const system = useSystemStore()

onMounted( () => {
   user.getTheses()
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
   margin: 0 auto 50px auto;
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
      align-items: stretch;
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