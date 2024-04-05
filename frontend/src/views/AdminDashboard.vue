<template>
   <div class="admin">
      <Panel>
         <template #header>
            <span class="hdr">
               <div>Libra Admin Dashboard</div>
            </span>
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
   document.title = "Libra Admin"
   searchStore.search("oa", user.computeID)
})

const editWorkClicked = ( (id) => {
   let url = `/oa/${id}`
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
.admin {
   width: 95%;
   margin: 2% auto;
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