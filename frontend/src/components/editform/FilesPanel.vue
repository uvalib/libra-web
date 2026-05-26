<template>
   <Panel header="Files" toggleable pt:title:id="files-panel" pt:contentContainer:aria-labelledby="files-panel">
      <template #icons>
         <span v-if="etdRepo.hasFiles" class="complete">
            <i class="pi pi-check-circle"></i>
            <span>Complete</span>
         </span>
         <span v-else class="incomplete">
            <i class="pi pi-exclamation-circle"></i>
            <span>Incomplete</span>
         </span>
      </template>

      <div class="files">
         <div class="section" v-if="etdRepo.work.files.length > 0">
            <div>Previously Uploaded Files</div>
            <div class="uploaded">
               <Card v-for="(file) in etdRepo.work.files">
                  <template #title>{{ file.name }}</template>
                  <template #subtitle>Uploaded on {{ $formatDateTime(file.createdAt) }}</template>
                  <template #footer>
                     <div class="acts" v-if="rename != file.name">
                        <Button size="small" icon="pi pi-times" label="Delete" severity="danger" @click="deleteFileClicked(file.name)"/>
                        <Button size="small" icon="pi pi-cloud-download"  label="Download" severity="secondary"
                           @click="etdRepo.downloadFile(file.name)" :loading="etdRepo.downloading == file.name"
                        />
                        <Button size="small" icon="pi pi-file-edit" label="Rename" severity="secondary" @click="renameClicked(file.name)"/>
                        <Button v-if="user.isAdmin" size="small" icon="pi pi-refresh" label="Replace" severity="secondary" @click="replaceClicked(file.name)"/>
                     </div>
                     <div class="rename" v-else>
                        <span class="rename-entry">
                           <InputText v-model="newName" placeholder="New Name" autofocus @keyup.enter="doRename()" v-keyfilter="/([0-9])|([a-z])|([A-Z])|_|-/"/>
                           <span v-if="newNameExt">.{{ newNameExt }}</span>
                        </span>
                        <Button size="small" icon="pi pi-times" rounded severity="secondary" aria-label="cancel" @click="rename=false"/>
                        <Button size="small" icon="pi pi-check" rounded severity="secondary" aria-label="rename" @click="doRename()" :disabled="newName.length < 3"/>
                     </div>  
                  </template>
               </Card>
            </div>
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
            <Divider />
            <div>Pending Uploads</div>
            <ul class="pending">
               <li v-for="fn in etdRepo.pendingFileAdd" class="pending-file">
                  <Button size="small" rounded icon="pi pi-times" aria-label="Delete" severity="danger" @click="deleteFileClicked( fn )"/>
                  <span class="filename">{{ fn }}</span>
               </li>
            </ul>
            <div>These files will be added to your thesis when it is saved.</div>
         </div>
      </div>
   </Panel>
   
   <Dialog v-model:visible="showReplace" :modal="true" :header="`Replace ${replaceFile}`" @hide="showReplace=false">
      <FileUpload ref="replace" name="replacment" chooseLabel="Select a replacement file"
         :customUpload="true" mode="basic"
         @uploader="startReplacementUpload($event)"
         :withCredentials="true" :auto="false"
         :showUploadButton="false" :showCancelButton="false"
         :accept="fileTypesAccepted"
      />
      <template #footer>
         <Button label="Cancel" severity="secondary" @click="showReplace=false"/>
         <Button label="Replace" autofocus  @click="doReplace()"/>
      </template>
   </Dialog>

</template>

<script setup>
import FileUpload from 'primevue/fileupload'
import Panel from 'primevue/panel'
import Card from 'primevue/card'
import InputText from 'primevue/inputtext'
import Divider from 'primevue/divider'
import Dialog from 'primevue/dialog'
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
const newNameExt = ref("")

const replace = ref()
const replaceFile = ref("")
const showReplace = ref(false)

const fileTypesAccepted = computed( () => {
   if (!user.isAdmin) {
      return system.mimeTypes
   }
   return null
})

const replaceClicked = (( origFileName ) => {
   replaceFile.value = origFileName
   showReplace.value = true
})

const doReplace = (() => {
   replace.value.upload()
})

const startReplacementUpload = (( event ) => {
   const file = event.files[0]
   let formData = new FormData()
   formData.append('file', file, replaceFile.value)
   axios.post(`/api/admin/works/${etdRepo.work.id}/files/${replaceFile.value}/replace`, formData, {
      headers: {
         'Content-Type': 'multipart/form-data',
      }
   }).then(() => {
      showReplace.value = false
      replaceFile.value = ""
   }).catch((error) => {
      system.toastError("Upload failed", error)
   })
})

const startUpload = ( (event) => {
   const file = event.files[0]
   if ( file.name.indexOf(".") == -1) {
      system.toastError("Filename Error", "An extension is required for any files you upload")
      return
   }
   let formData = new FormData()

   // find the highest index prefix on all current, add, del files. Make the
   // new filename prefix be 1 greater than this numbner
   let fileIdx = 0
   let fileNames = etdRepo.work.files.map( f => f.name)
   fileNames.concat(etdRepo.pendingFileAdd, etdRepo.pendingFileDel ).forEach( fn => {
       let bits = fn.split("_")
       if (bits.length > 0) {
         let id =  parseInt(bits[0],10)
         if (id > fileIdx) {
            fileIdx = id
         }
      }
   })
   fileIdx++

   const today = new Date()
   const year = today.getFullYear()
   const degree = etdRepo.work.degree.split(" (")[0]
   const ext = file.name.split('.').pop()
   const newFileName = `${fileIdx}_${user.lastName}_${user.firstName}_${year}_${degree}.${ext}`
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

const renameClicked = ( (name) => {
   rename.value = name
   newName.value = ""
   newNameExt.value = ""
   if ( name.indexOf(".") > -1) {
      newNameExt.value = name.split(".").pop()
   }
})

const doRename = ( () => {
   let fulNewlName = newName.value
   if (newNameExt.value != "" ) {
      fulNewlName += "."+newNameExt.value
   }
   etdRepo.renameFile(rename.value, fulNewlName)
   newName.value = ""
   newNameExt.value = ""
   rename.value = ""
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
</script>

<style lang="scss" scoped>
ul.pending {
   list-style: none;
   margin: 0 0 15px 0;
   padding: 0 0 0 5px;
   li {
      display: flex;
      flex-flow: row nowrap;
      gap: 10px;
      justify-content: flex-start;
      align-items: center;
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
   gap: 10px;
}
.files {
   display: flex;
   flex-direction: column;
   gap: 15px;
   .uploaded {
      display: flex;
      flex-direction: column;
      gap: 10px;
   }
   .acts {
      display: flex;
      flex-flow: row wrap;
      gap: 10px;
   }
   .rename {
      display: flex;
      flex-flow: row wrap;
      gap: 10px;
      .rename-entry {
         flex-grow: 1;
         display: flex;
         flex-flow: row nowrap;
         align-items: center;
         input {
            margin-right: 5px;
            flex-grow: 1;
         }
      }
   }
}
</style>