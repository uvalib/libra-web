<template>
   <div class="status">
      <h1>Deposit Status</h1>
      <div class="search">
         <label>Search by:</label>
         <Select v-model="queryType" aria-label="query type selector" :options="queries" option-label="name" option-value="id"/>
         <label>for:</label>
         <div class="search-input">
            <InputText v-if="queryType=='cid'" v-model="query" @keypress="searchKeyPressed($event)"
               aria-label="compute id query string" placeholder="Compute ID"
            />
            <DatePicker v-else v-model="queryDate" dateFormat="yy-mm-dd" showIcon fluid iconDisplay="input" aria-label="date query"/>
            <Button label="Search" @click="searchClicked()" :loading="admin.working" :disabled="admin.working"/>
         </div>
      </div>
      <div class="hint">{{ searchHint }}</div>
      <div class="results">
         <DataTable :value="admin.deposits" v-model:filters="filters" ref="depositHits"
            stripedRows showGridlines responsiveLayout="scroll" removableSort
            :paginator="true" :alwaysShowPaginator="true" paginatorPosition="bottom"
            :rows="10" :rowsPerPageOptions="[10, 25, 50, 100]"
            paginatorTemplate="PrevPageLink CurrentPageReport NextPageLink RowsPerPageDropdown"
            currentPageReportTemplate="Showing {first} - {last} of {totalRecords} entries"
            :loading="admin.working"
         >
            <template #empty  v-if="admin.depositSearchMessage != ''">
               <div class="err">{{ admin.depositSearchMessage }}</div>
            </template>
            <template #header>
               <IconField iconPosition="left">
                  <InputIcon class="pi pi-search" />
                  <InputText v-model="filters['global'].value" placeholder="Search within results" fluid aria-label="search within results"/>
               </IconField>
            </template>
            <Column field="computeID" header="ID" class="nowrap"/>
            <Column field="fullName" header="Name" class="nowrap" sortable/>
            <Column field="receivedFromSIS" header="Received from SIS" class="nowrap" sortable/>
            <Column field="submittedToLibra" header="Deposited in Libra" class="nowrap" sortable/>
            <Column field="exportedToSIS" header="Exported to SIS" class="nowrap" sortable/>
            <Column field="title" header="Title">
               <template #body="slotProps">
                  {{ truncate(slotProps.data.title) }}
               </template>
            </Column>
         </DataTable>
      </div>
   </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
import InputText from 'primevue/inputtext'
import DatePicker from 'primevue/datepicker'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import { FilterMatchMode } from '@primevue/core/api'
import Column from 'primevue/column'
import { useAdminStore } from "@/stores/admin"
import dayjs from 'dayjs'

const admin = useAdminStore()
const query = ref("")
const queryDate = ref( new Date() )
const queryType = ref("cid")

const filters = ref({
    global: { value: null, matchMode: FilterMatchMode.CONTAINS }
})

const queries = computed( () =>{
   return [
      {id: "cid", name: "Compute ID"},
      {id: "created", name: "Received date"},
      {id: "exported", name: "Exported date"}
   ]
})

const searchHint = computed( () => {
   if ( queryType.value == "created") {
      return "Received from SIS on or after the selected date"
   }
   if ( queryType.value == "exported") {
      return "Exported to SIS on or after the selected date"
   }
   return "Query for deposits by computing ID (complete or partial)"
})

const searchKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      searchClicked()
   }
})

const truncate = ((text) => {
   if (text.length <= 75) return text
   var trunc = text.substring(0, 75-1)
   var out = trunc.substring(0, trunc.lastIndexOf(' ')).trim()
   out += "..."
   return out
})

const searchClicked = (() => {
   filters.value['global'].value = ""
   var queryStr = query.value
   if (queryType.value != "cid") {
      queryStr = dayjs(queryDate.value).format("YYYY-MM-DD")
   }
   admin.depositStatusSearch(queryType.value, queryStr)
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
   .search {
      display: flex;
      flex-flow: row wrap;
      align-items: center;
      gap: 10px;
      .search-input {
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: center;
         gap: 0;
         input {
            border-radius: 0.3rem 0 0 0.3rem;
            border-right:0;
         }
         button {
            border-radius:  0 0.3rem 0.3rem 0;
         }
      }
   }
   .hint {
      font-size: 0.95em;
   }
   .err {
      margin: 20px;
      text-align: center;
      font-weight: bold;
      font-size: 1.2em;
   }
}
</style>