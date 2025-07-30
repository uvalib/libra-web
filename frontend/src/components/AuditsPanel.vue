<template>
   <Button @click="show" label="View Audit Log" severity="secondary" fluid size="small"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Audit Log"
      style="width:90%; max-height:90%" position="top"
      :blockScroll="true" :maximizable="true"
   >
      <div class="loading" v-if="auditStore.working">
         <WaitSpinner :overlay="false" message="Loading Audit History..." />
      </div>
      <div class="audit-panel" v-else>
         <template v-if="auditStore.error">
            <p class="error">
               The audit log for this work is currently unavailable
               <span>{{ auditStore.error }}</span>
            </p>
         </template>

         <DataTable v-else :value="auditStore.audits" stripedRows showGridlines
            paginator :rows="50" paginatorPosition="both" :alwaysShowPaginator="false"
            paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink"
            currentPageReportTemplate="{first} - {last} of {totalRecords}"
         >
            <Column header="At" class="nowrap">
               <template #body="{ data }">{{ $formatDateTime(data.eventTime) }}</template>
            </Column>
            <Column field="who" header="Who"/>
            <Column field="fieldName" header="Field"/>
            <Column field="before" header="Before"></Column>
            <Column field="after" header="After"></Column>
         </DataTable>
      </div>
   </Dialog>
</template>

<script setup>
import { ref } from 'vue'
import DataTable from 'primevue/datatable';
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useAuditStore } from '@/stores/audit'

const auditStore = useAuditStore()
const isOpen = ref(false)

const props = defineProps({
   workID: {
      type: String,
      required: true
   },
})

const show = (() => {
   auditStore.getAudits(props.workID)
   isOpen.value = true
})
</script>

<style scoped>
.loading {
   text-align: center;
   padding: 30px;
}
p.error {
   font-size: 1.4em;
   text-align: center;

   span {
      display: block;
      margin: 10px 0;
      font-size: 0.8em;
   }
}
</style>