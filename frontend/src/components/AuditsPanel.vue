<template>
  <Panel header="Audit History" class="audit-panel">
    <WaitSpinner v-if="auditStore.working" :overlay="false" message="Loading Audit History..." />
    <p v-if="auditStore.error" class="error">Currently Unavailable</p>

    <DataTable v-else :value="auditStore.audits" tableStyle="min-width: 20rem" size="small" stripedRows>
      <Column header="At">
        <template #body="{ data }">{{ $formatDateTime(data.eventTime) }}</template>
      </Column>
      <Column field="who" header="Who"></Column>
      <Column field="fieldName" header="Field"></Column>
      <Column field="before" header="Before"></Column>
      <Column field="after" header="After"></Column>
    </DataTable>
  </Panel>
</template>

<script setup>
import Panel from 'primevue/panel'
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useAuditStore } from '@/stores/audit'
import { onMounted } from 'vue'

const auditStore = useAuditStore()

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
onMounted( async () => {
  await auditStore.getAudits( props.workID, props.namespace )
})

</script>
<style scoped>
</style>