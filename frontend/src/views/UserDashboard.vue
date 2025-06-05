<template>
   <div class="dashboard">
      <h2>My Active Theses</h2>
      <WaitSpinner v-if="user.working" :overlay="true" message="<div>Please wait...</div><p>Searching for active theses</p>" />
      <div class="content" v-else>
         <Card>
            <template #title>{{ user.displayName }}</template>
            <template #content>
               <div class="email">
                  <a class="ext-link" :href="`mailto:${user.email}`"><i class="pi pi-envelope"></i>{{ user.email }}</a>
               </div>
               <div class="orcid">
                  <template  v-if="user.orcid.id.length == 0">
                     <Button as="a" severity="secondary" :href="system.orcidURL" target="_blank" variant="outlined">
                        <img class="orcid-img" src="@/assets/orcid_id.svg"/>
                        <span>Register or connect your ORCID ID</span>
                     </Button>
                  </template>
                  <template v-else>
                     <Button as="a" severity="secondary" :href="system.orcidURL" target="_blank" variant="outlined">
                        <img class="orcid-img" src="@/assets/orcid_id.svg"/>
                        <span>Manage your ORCID ID</span>
                     </Button>
                     <Button as="a" severity="secondary" :href="user.orcid.uri" target="_blank" variant="outlined">
                        <img class="orcid-img" src="@/assets/orcid_id.svg"/>
                        <span>{{ user.orcid.id }}</span>
                     </Button>
                  </template>
                  <a class="ext-link" href="https://orcid.org/faq-page" target="_blank">Learn more about ORCID<i class="pi pi-external-link"></i></a>
               </div>
            </template>
         </Card>
         <div class="theses">
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
                           @click="previewWorkClicked(slotProps.data.id)"/>
                        <Button v-else class="action" label="Edit Thesis" severity="secondary"
                           @click="editWorkClicked(slotProps.data.id)"/>
                     </div>
                  </template>
               </Column>
            </DataTable>
            <div class="help">
               View
               <a class="ext-link" target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist<i class="pi pi-external-link"></i></a>
               for help.
            </div>
         </div>
      </div>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"
import Card from 'primevue/card'
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
@media only screen and (min-width: 768px) {
   .dashboard {
      width: 90%;
      .content {
         display: flex;
         flex-flow: row wrap;
         align-items: flex-start;
         gap: 25px;
      }
      .theses {
         max-width: 70%;
      }
      .p-card {
         max-width: 380px;
      }
   }
}
@media only screen and (max-width: 768px) {
   .dashboard {
      width: 95%;
      .content {
         display: flex;
         flex-direction: column;
         gap: 25px;
      }
   }
}
.orcid {
   display: flex;
   flex-direction: column;
   align-items: flex-start;
   gap: 10px;
   margin: 20px 0;
   .orcid-img {
      width: 30px;
   }
}
a.ext-link {
   display: flex;
   flex-flow: row nowrap;
   gap: 5px;
}

.dashboard {
   margin: 0 auto 50px auto;
   text-align: left;
   .help {
      display: flex;
      flex-flow: row wrap;
      margin: 25px 0 10px 0;
      gap: 8px;
   }
   .none {
      color: $uva-grey-A;
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
</style>