<template>
   <div class="files">
      <div class="section" v-if="etdRepo.work.files.length > 0">
         <div>Previously Uploaded Files</div>
         <DataTable :value="etdRepo.work.files" ref="etdFiles" dataKey="id"
               stripedRows showGridlines size="small"
               :lazy="false" :paginator="true" :alwaysShowPaginator="false"
               :rows="30" :totalRecords="etdRepo.work.files.length"
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               :rowsPerPageOptions="[30,50,100]" paginatorPosition="top"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
            >
            <Column field="name" header="Name" />
            <Column field="createdAt" header="Date Uploaded" >
               <template #body="slotProps">{{ $formatDate(slotProps.data.createdAt)}}</template>
            </Column>
            <Column  header="Actions" >
               <template #body="slotProps">
                  <div class="acts">
                     <Button class="action" icon="pi pi-trash" label="Delete" severity="danger" size="small" @click="deleteFileClicked(slotProps.data.name)"/>
                     <Button class="action" icon="pi pi-cloud-download" label="Download" severity="secondary" size="small" @click="downloadFileClicked(slotProps.data.name)"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </div>

      <div class="section">
         <label class="libra-form-label">
            Upload Files
            <FileUpload name="file" :url="`/api/upload/${etdRepo.work.id}`"
               @upload="fileUploaded($event)" @before-send="uploadRequested($event)"
               @removeUploadedFile="fileRemoved($event)" :previewWidth="0"
               :multiple="true" :withCredentials="true" :auto="true"
               :showUploadButton="false" :showCancelButton="false">
               <template #empty>
                  <p>Click Choose or drag and drop files to upload. Uploaded files will be attached to the work upon submission.</p>
               </template>
            </FileUpload>
         </label>
      </div>
   </div>
</template>

<script setup>
import FileUpload from 'primevue/fileupload'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useETDStore } from "@/stores/etd"
import { useUserStore } from "@/stores/user"
import { useConfirm } from "primevue/useconfirm"

const etdRepo = useETDStore()
const user = useUserStore()
const confirm = useConfirm()

const uploadRequested = ( (request) => {
   request.xhr.setRequestHeader('Authorization', 'Bearer ' + user.jwt)
   return request
})

const fileRemoved = ( event => {
   etdRepo.removeFile( event.file.name )
})
const fileUploaded = ( (event) => {
   event.files.forEach( f => {
      etdRepo.addFile( f.name )
   })
})

const deleteFileClicked = ( (name) => {
   confirm.require({
      message: `Delete file ${name}?`,
      header: 'Confirm Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         etdRepo.removeFile(name)
      },
   })
})

const downloadFileClicked = ( (name) => {
   etdRepo.downloadFile(name)
})
</script>

<style lang="scss" scoped>
:deep(.p-fileupload-header) {
   border-bottom: 1px solid $uva-grey-200;
   margin-bottom: 15px;
}
:deep(.p-fileupload-content) {
   .p-fileupload-file-list {
      gap: 15px;
      .p-fileupload-file {
         padding: 0 0 10px 0;
         gap: 30px;
         img.p-fileupload-file-thumbnail {
            display: none;
         }
         .p-badge-success.p-fileupload-file-badge {
            background: $uva-green-A;
            border-radius: 0;
         }
      }
   }
}
.p-fileupload {
   margin-top: 5px;
}
.section {
   display: flex;
   flex-direction: column;
   gap: 5px;
}
.files {
   display: flex;
   flex-direction: column;
   gap: 25px;
   label {
      display: block;
      margin-bottom: 5px;
   }
   .acts {
      display: flex;
      flex-flow: row wrap;
      gap: 10px;
   }
}
</style>