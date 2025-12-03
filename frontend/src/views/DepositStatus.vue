<template>
   <div class="status">
      <h1>Deposit Status</h1>
      <div class="search">
         <label>Search by:</label>
         <Select v-model="queryType" aria-label="query type" :options="queries" option-label="name" option-value="id"/>
         <label>for:</label>
         <div class="search-input">
            <InputText v-model="query" @keypress="searchKeyPressed($event)" fluid
               aria-label="deposit query" id="admin-search" :placeholder="queryPlaceholder"
            />
            <Button label="Search" @click="admin.depositStatusSearch(queryType, query)" :loading="admin.working" :disabled="admin.working"/>
         </div>
      </div>
      <div class="hint">{{ searchHint }}</div>
      <div class="results" v-if="admin.deposits.length > 0 || admin.depositSearchMessage != ''">
         <DataTable :value="admin.deposits" ref="depositHits"
            stripedRows showGridlines responsiveLayout="scroll" removableSort
            :paginator="true" :alwaysShowPaginator="true" paginatorPosition="bottom"
            :rows="10" :rowsPerPageOptions="[10, 25, 50, 100]"
            paginatorTemplate="PrevPageLink CurrentPageReport NextPageLink RowsPerPageDropdown"
            currentPageReportTemplate="Showing {first} - {last} of {totalRecords} entries"
            :loading="admin.working"
         >
            <template #empty>
               <div class="err">{{ admin.depositSearchMessage }}</div>
            </template>
            <Column field="computeID" header="ID" class="nowrap"/>
            <Column field="fullName" header="Name" class="nowrap" sortable/>
            <Column field="receivedFromSIS" header="Received from SIS" class="nowrap" sortable/>
            <Column field="submittedToLibra" header="Submitted to Libra" class="nowrap" sortable/>
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
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useAdminStore } from "@/stores/admin"

const admin = useAdminStore()
const query = ref("")
const queryType = ref("cid")

const queries = computed( () =>{
   return [
      {id: "cid", name: "Compute ID"},
      {id: "created", name: "Create Date"},
      {id: "exported", name: "Export Date"}
   ]
})

const searchHint = computed( () => {
   if ( queryType.value == "created") {
      return "Query for deposits created on or after the selected date"
   }
   if ( queryType.value == "exported") {
      return "Query for deposits exported on or after the selected date"
   }
   return "Query for deposits by computing ID (complete or partial)"
})

const queryPlaceholder = computed( () => {
   if ( queryType.value == "created") {
      return "Enter create date..."
   }
   if ( queryType.value == "exported") {
      return "Enter exported date..."
   }
   return "Enter computing ID..."
})

const searchKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      admin.depositStatusSearch(queryType.value, query.value)
   }
})

const truncate = ((text) => {
   if (text.length <= 75) return text
   var trunc = text.substring(0, 75-1)
   var out = trunc.substring(0, trunc.lastIndexOf(' ')).trim()
   out += "..."
   return out
})

</script>

<style lang="scss" scoped>
.status {
   margin: 0 50px 50px;
   display: flex;
   flex-direction: column;
   gap: 20px;
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
         flex-grow: 1;
         input {
            border-radius: 0.3rem 0 0 0.3rem;
            border-right:0;
         }
         button {
            border-radius:  0 0.3rem 0.3rem 0;
         }
         .query {
            flex-grow: 1;
         }
      }
   }
   .hint {
      font-style: italic;
      font-size: 0.95em;
      color:  $uva-grey;
   }
   .err {
      margin: 20px;
      text-align: center;
      font-weight: bold;
      font-size: 1.2em;
   }
}
</style>