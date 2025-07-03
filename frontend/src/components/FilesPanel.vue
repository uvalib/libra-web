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
            <Column field="createdAt" header="Uploaded" >
               <template #body="slotProps">{{ $formatDateTime(slotProps.data.createdAt)}}</template>
            </Column>
            <Column  header="Actions" >
               <template #body="slotProps">
                  <div class="acts" v-if="rename != slotProps.data.name">
                     <Button class="action" icon="pi pi-trash" label="Delete" severity="danger" size="small" @click="deleteFileClicked(slotProps.data.name)"/>
                     <Button v-if="slotProps.data.url.length==0" class="action" label="Request Download" severity="secondary" size="small" @click="downloadFileClicked(slotProps.data.name)"/>
                     <Button v-else as="a" icon="pi pi-cloud-download" label="Download" :href="slotProps.data.url" target="_blank" rel="noopener" />
                     <Button class="action" icon="pi pi-file-edit" label="Rename" severity="secondary" size="small" @click="rename=slotProps.data.name"/>
                  </div>
                  <div class="rename" v-else>
                     <InputText v-model="newName" placeholder="New Name"/>
                     <Button class="action" icon="pi pi-times" rounded severity="secondary" aria-label="cancel" size="small" @click="rename=false"/>
                     <Button class="action" icon="pi pi-check" rounded severity="secondary" aria-label="rename" size="small" @click="rename=false" :disabled="newName.length < 4"/>
                  </div>
               </template>
            </Column>
         </DataTable>
      </div>

      <div class="section">
         <label class="libra-form-label">
            <FileUpload name="file" chooseLabel="Select a file to upload"
               :customUpload="true" mode="basic"
               @uploader="startUpload($event)"
               :withCredentials="true" :auto="true"
               :showUploadButton="false" :showCancelButton="false"
               :accept="fileTypesAccepted"
            />
         </label>
      </div>

      <div class="section" v-if="etdRepo.pendingFileAdd.length > 0">
         <div>Pending Uploads</div>
         <ul class="pending">
            <li v-for="fn in etdRepo.pendingFileAdd" class="pending-file">
               <span class="filename">{{ fn }}</span>
               <Button class="action" icon="pi pi-trash" label="Delete" severity="danger" size="small"  @click="etdRepo.removeFile( fn )"/>
            </li>
         </ul>
         <div>These files will be added to your thesis when it is saved.</div>
      </div>
   </div>
</template>

<script setup>
import FileUpload from 'primevue/fileupload'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import InputText from 'primevue/inputtext'
import { useETDStore } from "@/stores/etd"
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"
import { useConfirm } from "primevue/useconfirm"
import { computed, ref } from 'vue'
import axios from 'axios'

const etdRepo = useETDStore()
const user = useUserStore()
const system = useSystemStore()
const confirm = useConfirm()
const rename = ref("")
const newName = ref("")

const fileTypesAccepted = computed( () => {
   if (!user.isAdmin) {
      // CSV, GIF, HTM, HTML, JPEG, JPG, MOV, MP3, MP4, PDF, PNG, TIF, TIFF, TXT, XML
      return "text/csv, application/pdf, image/*, text/html, application/xml, text/plain, video/mp4, video/quicktime, audio/mp3"
   }
   return null
})

const startUpload = ( (event) => {
   const file = event.files[0]
   let formData = new FormData()
   const cnt = etdRepo.work.files.length + etdRepo.pendingFileAdd.length + 1
   const today = new Date()
   const year = today.getFullYear()
   const degree = etdRepo.work.degree.split(" (")[0]
   const ext = file.name.split('.').pop()
   const newFileName = `${cnt}_${user.lastName}_${user.firstName}_${year}_${degree}.${ext}`
   formData.append('file', file, newFileName)
   axios.post(`/api/upload/${etdRepo.work.id}`, formData, {
      headers: {
         'Content-Type': 'multipart/form-data',
      }
   }).then(() => {
      etdRepo.addFile( newFileName )
   }).catch((error) => {
      system.toastError("Upload failed", error)
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
ul.pending {
   list-style: none;
   margin: 0 0 15px 0;
   padding: 0 0 0 5px;
   li {
      margin: 10px 0;
   }
   .pending-file {
      .filename {
         margin-right: 10px;
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
   .rename {
      display: flex;
      flex-flow: row wrap;
      gap: 5px;
      input {
         flex-grow: 1;
      }
   }
}
</style>