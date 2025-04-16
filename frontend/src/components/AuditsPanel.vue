<template>
   <Button @click="show" label="Audit Log" severity="secondary" class="view-audit" />
   <Dialog v-model:visible="isOpen" :modal="true" header="Audit Log"
      style="width:90%; max-height:90%" position="top"
      :blockScroll="true" :maximizable="true"
   >
      <WaitSpinner v-if="auditStore.working" :overlay="false" message="Loading Audit History..." />
      <template v-if="auditStore.error">
         <p class="error">
            The audit log for this work is currently unavailable
            <span>{{ auditStore.error }}</span>
         </p>
      </template>

      <DataTable v-else :value="auditStore.audits" tableStyle="min-width: 20rem" stripedRows showGridlines
         :lazy="false" :paginator="true" :rows="10" :rowsPerPageOptions="[10, 20, 30]"
         paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
         currentPageReportTemplate="{first} - {last} of {totalRecords}" scrollable scrollHeight="600px"
      >
         <Column header="At" class="nowrap">
            <template #body="{ data }">{{ $formatDateTime(data.eventTime) }}</template>
         </Column>
         <Column field="who" header="Who"></Column>
         <Column field="fieldName" header="Field"></Column>
         <Column field="before" header="Before"></Column>
         <Column field="after" header="After"></Column>
      </DataTable>
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
   namespace: {
      type: String,
      required: true
   }
})

const show = (() => {
   auditStore.getAudits(props.workID, props.namespace)
   isOpen.value = true
})
</script>

<style scoped>
:deep(td.nowrap),  :deep(th){
   white-space: nowrap;
}
.view-audit {
   font-size: 0.9em;
   padding: 4px 12px;
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