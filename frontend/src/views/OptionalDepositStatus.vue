<template>
   <div class="status">
      <h1>Optional Deposit Status</h1>
      <DataTable :value="registrar.deposits" dataKey="id" v-model:expandedRows="expandedRows"
            stripedRows showGridlines responsiveLayout="scroll"
            :lazy="true" :paginator="true" :alwaysShowPaginator="true"
            @page="onPage($event)"  paginatorPosition="both"
            :first="registrar.offset" :rows="registrar.limit" :rowsPerPageOptions="[25, 50, 100]"
            :totalRecords="registrar.total" paginatorTemplate="PrevPageLink CurrentPageReport NextPageLink RowsPerPageDropdown"
            currentPageReportTemplate="{first} - {last} of {totalRecords}"
            :loading="registrar.working" @sort="onSort($event)" :sortField="registrar.sortField" :sortOrder="sortOrder"
            filterDisplay="row" v-model:filters="filters"
            @update:filters="onFilter($event)"
      >
            <template #paginatorstart>
               <Button severity="secondary" size="small" label="Clear Filters" @click="resetFilters" :disabled="!hasFilters"/>
            </template>
            <template #paginatorend>
                <div class="expand-ctls">
                    <Button severity="secondary" size="small" icon="pi pi-plus" label="Expand All" @click="expandAll" />
                    <Button severity="secondary" size="small" icon="pi pi-minus" label="Collapse All" @click="collapseAll" />
                </div>
            </template>
         <template #empty>
            <div class="err">{{ registrar.depositSearchMessage }}</div>
         </template>
         <Column expander style="width: 5rem" />
         <Column field="registrar" header="Registrar" filterField="registrar" :showFilterMenu="false" sortable>
            <template #filter="{ filterModel, filterCallback }">
               <InputGroup>
                  <InputText v-model="filterModel.value" type="text" @update:modelValue="filterCallback()" placeholder="Submitted by" />
                  <InputGroupAddon>
                     <Button severity="secondary" text icon="pi pi-times" :disabled="clearRegistrarDisabled" @click="clearRegistrarFilter"/>
                  </InputGroupAddon> 
               </InputGroup>
            </template>
         </Column>
         <Column field="submittedAt" header="Submitted" sortable class="nowrap" sortField="submitted_at" filterField="submitted_at" :showFilterMenu="false">
            <template #body="slotProps">{{ $formatDateTime(slotProps.data.submittedAt) }}</template>
            <template #filter="{ filterModel, filterCallback }">
               <DatePicker v-model="filterModel.value" dateFormat="yy-mm-dd" placeholder="Submitted after" 
                  @update:modelValue="filterCallback()" :showIcon="true" updateModelType="string" :showClear="true"
               />
            </template>
         </Column>
         <Column field="program" header="Program" filterField="program" :showFilterMenu="false" sortable>
            <template #filter="{ filterModel, filterCallback }">
               <Select v-model="filterModel.value" fluid placeholder="Submitted program" showClear filter
                  :options="system.programs" optionLabel="program" optionValue="program" @change="filterCallback()" 
               />
            </template>
         </Column>
         <Column field="degree" header="Degree" filterField="degree" :showFilterMenu="false" sortable>
            <template #filter="{ filterModel, filterCallback }">
               <Select v-model="filterModel.value" fluid placeholder="Submitted degree" showClear filter
                  :options="system.degrees" optionLabel="degree" optionValue="degree" @change="filterCallback()" 
               />
            </template>
         </Column>
         <Column field="students" header="Students">
             <template #body="slotProps">{{ slotProps.data.students.length }}</template>
         </Column>
         <template #expansion="slotProps">
            <div class="registered-students">
               <div class="student-title">Registered Students</div>
               <DataTable :value="slotProps.data.students" class="nested" size="small" >
                  <Column field="computeID" header="Computing ID"></Column>
                  <Column field="workID" header="Work ID"></Column>
                  <Column field="completedAt" header="Completed" class="nowrap">
                     <template #body="slotProps">
                        <div v-if="slotProps.data.completedAt">{{ $formatDateTime(slotProps.data.completedAt) }}</div>
                        <div v-else class="none">No</div>
                     </template>
                  </Column>
               </DataTable>
            </div>
         </template>
      </DataTable>
   </div>
</template>

<script setup>
import { onMounted, computed, ref } from 'vue'
import { useRegistrarStore } from "@/stores/registrar"
import { useSystemStore } from "@/stores/system"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import InputText from 'primevue/inputtext'
import DatePicker from 'primevue/datepicker'
import Select from 'primevue/select'
import { FilterMatchMode } from '@primevue/core/api'

const registrar = useRegistrarStore()
const system = useSystemStore()

const expandedRows = ref({})
const filters = ref({
   registrar: { value: null, matchMode: FilterMatchMode.STARTS_WITH },
   submitted_at: { value: null, matchMode: FilterMatchMode.DATE_AFTER },
   program: { value: null, matchMode: FilterMatchMode.EQUALS },
   degree: { value: null, matchMode: FilterMatchMode.EQUALS },
})

const sortOrder = computed(() => {
   if (registrar.sortOrder == "asc") {
      return 1
   }  
   return -1
})

const hasFilters = computed(() => {
   return ( filters.value.registrar.value ||  filters.value.submitted_at.value || 
        filters.value.program.value ||  filters.value.degree.value)  
})

const clearRegistrarDisabled = computed(() =>{
   return (!filters.value.registrar.value)  
})

onMounted( () => {
   // default search by 25 m0st recent registrations
   registrar.optDepositStatusSearch()
})

const expandAll = () => {
   // reduse the array of deposits into an object with field ${id}: true for each deposits identifier
   expandedRows.value = registrar.deposits.reduce((accumulator, d) => (accumulator[d.id] = true) && accumulator, {})
}

const collapseAll = () => {
   expandedRows.value = {}
}

const onPage = ((event) => {
   registrar.offset = event.first
   registrar.optDepositStatusSearch()
})

const clearRegistrarFilter = (() => {
   filters.value.registrar.value=null
   delete registrar.filters['registrar']
   registrar.optDepositStatusSearch()
})

const resetFilters = (() => {
   filters.value = {
      registrar: { value: null, matchMode: FilterMatchMode.STARTS_WITH },
      submitted_at: { value: null, matchMode: FilterMatchMode.DATE_AFTER },
      program: { value: null, matchMode: FilterMatchMode.EQUALS },
      degree: { value: null, matchMode: FilterMatchMode.EQUALS },
   } 
   registrar.filters = {}
   registrar.optDepositStatusSearch()
})

const onFilter = ((filter) => {
   registrar.filters = {}
   if ( filter.registrar.value ) {
      registrar.filters['registrar'] = filter.registrar.value
   }
   if ( filter.degree.value ) {
      registrar.filters['degree'] = filter.degree.value
   }
   if ( filter.program.value ) {
      registrar.filters['program'] = filter.program.value
   }
   if ( filter.submitted_at.value ) {
      registrar.filters['submitted_at'] = filter.submitted_at.value
   }
   registrar.optDepositStatusSearch()
})

const onSort = ((event) => {
   console.log("ORDER "+event.sortOrder)
   registrar.sortField = ""
   registrar.sortOrder = ""
   if (event.sortOrder == 1) {
      registrar.sortField = event.sortField
      registrar.sortOrder = "asc"

   } else if (event.sortOrder == -1) {
      registrar.sortField = event.sortField
      registrar.sortOrder = "desc"
   }
   registrar.optDepositStatusSearch()
})

</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .status {
      margin: 0 50px 50px;
   }
}
@media only screen and (max-width: 768px) {
   .status {
      margin: 0 10px;
   }
}
.status {
   display: flex;
   flex-direction: column;
   gap: 20px;
   min-height: 600px;
   text-align: left;
   .expand-ctls {
      display: flex;
      flex: flow wrap;
      gap: 10px;
   }
}
.registered-students {
   margin-left: 4rem;
   .student-title {
      font-weight: bold;
      margin-bottom: 10px;
   }
   :deep(th.p-datatable-header-cell) {
      background-color: white;
   }
}
:deep(thead.p-datatable-thead) {
   tr:last-of-type {
      th {
         background-color: white ;
         font-size: 0.8em;
         padding: 10px;
         input {
            width: 100%;
         }
      }
   }
}
</style>