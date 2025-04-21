<template>
   <div class="scroll-body"  id="etd-form-layout">
      <div class="sidebar-col" :class="{admin: adminEdit}" v-if="etdRepo.working==false">
         <AdminPanel v-if="adminEdit"
            :source="etdRepo.source" :identifier="etdRepo.work.id" :depositor="etdRepo.depositor" :created="etdRepo.createdAt"
            :modified="etdRepo.modifiedAt" :published="etdRepo.publishedAt" :visibility="etdRepo.visibility"
            :embargoEndDate="etdRepo.embargoReleaseDate" :embargoEndVisibility="etdRepo.embargoReleaseVisibility"
            :degree="etdRepo.work.degree" :program="etdRepo.work.program" :notes="etdRepo.work.adminNotes"
            ref="savepanel" @cancel="cancelClicked" @delete="router.back()" @save="adminSaveCliced"
         />
         <SavePanel v-else
            :described="workDescribed" :files="etdRepo.work.files.length > 0 || etdRepo.pendingFileAdd.length > 0"
            :visibility="etdRepo.visibility" :releaseDate="etdRepo.embargoReleaseDate" :releaseVisibility="etdRepo.embargoReleaseVisibility"
            :draft="etdRepo.isDraft" :degree="etdRepo.work.degree" :program="etdRepo.work.program"
            ref="savepanel" @save="saveClicked" @cancel="cancelClicked"
         />
      </div>
      <div class="content-col">
         <Panel class="main-form">
            <template #header>
               <div class="work-header">
                  <span>LibraETD Work</span>
                  <span v-if="adminEdit==false && etdRepo.isDraft" class="draft">DRAFT</span>
               </div>
            </template>
            <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
            <FormKit v-else ref="etdForm" type="form" :actions="false" @submit="submitHandler">
               <div v-if="adminEdit==false" class="two-col margin-bottom">
                  <table class="readonly">
                     <tbody>
                        <tr>
                           <td class="label">Institution:</td>
                           <td>{{ etdRepo.work.author.institution  }}</td>
                        </tr>
                        <tr>
                           <td class="label">Program:</td>
                           <td>{{ etdRepo.work.program  }}</td>
                        </tr>
                        <tr>
                           <td class="label">Degree:</td>
                           <td>{{ etdRepo.work.degree }}</td>
                        </tr>
                        </tbody>
                  </table>
                     <table class="readonly">
                        <tbody>
                           <tr>
                              <td class="label">Date Created:</td>
                              <td>{{ $formatDate(etdRepo.createdAt) }}</td>
                           </tr>
                           <tr v-if="etdRepo.isDraft==false">
                              <td class="label">Date Published:</td>
                              <td>{{ $formatDate(etdRepo.publishedAt) }}</td>
                           </tr>
                           <tr>
                              <td></td><td><AuditsPanel v-if="etdRepo.working==false" :workID="etdRepo.work.id"/></td>
                           </tr>
                        </tbody>
                     </table>
               </div>

               <FormKit label="Title" type="text" v-model="etdRepo.work.title" validation="required" outer-class="first"/>

               <Panel>
                  <template #header>Author</template>
                  <FormKit type="group" v-model="etdRepo.work.author">
                     <div class="author">
                        <div class="two-col">
                           <FormKit type="text" name="firstName" label="First Name" outer-class="first" validation="required"/>
                           <FormKit type="text" name="lastName" label="Last Name"  outer-class="first" validation="required"/>
                        </div>
                     </div>
                  </FormKit>
               </Panel>

               <Panel>
                  <template #header>
                     <div>Advisors<span class='libra-required'><span class='star'>*</span>(required)</span></div>
                  </template>
                  <FormKit v-model="etdRepo.work.advisors" type="list" dynamic #default="{ items }" validation="required">
                     <div v-if="etdRepo.work.advisors.length == 0" class="no-advisor">
                        <div class="note">One or more advisors is required</div>
                        <Button label="Add Advisor" size="small" @click="addAdvisor"/>
                     </div>
                     <div v-else class="advisors">
                        <div class="note">Lookup a UVA Computing ID to automatically fill the remaining fields for this advisor.</div>
                        <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                           <div class="advisor">
                              <div class="id-field">
                                 <div class="search-field"  :id="`advisor-${index+1}`">
                                    <FormKit type="text" name="computeID" label="Computing ID"/>
                                    <Button class="check" icon="pi pi-search" severity="secondary" @click="checkAdvisorID(index)"/>
                                 </div>
                                 <Button v-if="index > 0" icon="pi pi-trash" severity="danger" aria-label="remove advisor" @click="removeAdvisor(index)"/>
                              </div>
                              <p v-if="etdRepo.work.advisors[index].msg" class="err">{{ etdRepo.work.advisors[index].msg }}</p>
                              <div class="two-col">
                                 <FormKit type="text" name="firstName" label="First Name"/>
                                 <FormKit type="text" name="lastName" label="Last Name"/>
                              </div>
                              <div class="two-col">
                                 <FormKit type="text" name="department" label="Department"/>
                                 <FormKit type="text" name="institution" label="Institution"/>
                              </div>
                           </div>
                           <div class="acts">
                              <Button label="Add Advisor" size="small" @click="addAdvisor"/>
                           </div>
                        </FormKit>
                     </div>
                  </FormKit>
               </Panel>

               <FormKit label="Abstract" type="textarea" v-model="etdRepo.work.abstract" rows="10" validation="required"/>

               <FormKit type="select" label="Rights" v-model="etdRepo.licenseID"
                  placeholder="Select rights"
                  :options="system.userLicenses" validation="required"
               />
               <p class="note">
                  Libra lets you choose an open license when you post your work, and will prominently display the
                  license you choose as part of the record for your work. See
                  <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
                  for option details.
               </p>

               <FormKit v-model="etdRepo.work.keywords" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`keyword-${index+1}`">
                        <FormKit :label="inputLabel('Keywords', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || etdRepo.work.keywords.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add keyword" @click="addKeyword"/>
                     <Button v-if="index > 0 || index == 0 && etdRepo.work.keywords[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove keyword"  @click="removeKeyword(index)"/>
                  </div>
                  <p class="note">Add one keyword or keyword phrase per line.</p>
               </FormKit>

               <FormKit type="select" label="Language" v-model="etdRepo.work.language"
                  placeholder="Select a language" :options="system.languages"/>

               <FormKit v-model="etdRepo.work.relatedURLs" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`url-${index+1}`">
                        <FormKit :label="inputLabel('Related Link(s)', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || etdRepo.work.relatedURLs.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add url" @click="addURL"/>
                     <Button v-if="index > 0 || index == 0 && etdRepo.work.relatedURLs[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove url"  @click="removeURL(index)"/>
                  </div>
                  <p class="note">A link to a website or other specific content (audio, video, PDF document) related to the work.</p>
               </FormKit>

               <FormKit v-model="etdRepo.work.sponsors" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`agency-${index+1}`">
                        <FormKit :label="inputLabel('Sponsoring Agency', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || etdRepo.work.sponsors.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add agency" @click="addAgency"/>
                     <Button v-if="index > 0 || index == 0 && etdRepo.work.sponsors[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove agency"  @click="removeAgency(index)"/>
                  </div>
               </FormKit>

               <FormKit label="Notes" type="textarea" v-model="etdRepo.work.notes" rows="10"/>

               <template v-if="etdRepo.work.files.length > 0">
                  <label class="libra-form-label">Previously Uploaded Files</label>
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
               </template>
               <label for="file" class="libra-form-label">Files</label>
               <FileUpload name="file" :url="`/api/upload/${etdRepo.work.id}`"
                  @upload="fileUploaded($event)" @before-send="uploadRequested($event)"
                  @removeUploadedFile="fileRemoved($event)"
                  :multiple="true" :withCredentials="true" :auto="true"
                  :showUploadButton="false" :showCancelButton="false">
                  <template #empty>
                     <p>Click Choose or drag and drop files to upload. Uploaded files will be attached to the work upon submission.</p>
                  </template>
               </FileUpload>

            </FormKit>
         </Panel>
      </div>
   </div>
</template>

<script setup>
import { ref, onBeforeMount, computed, nextTick } from 'vue'
import AdminPanel from "@/components/AdminPanel.vue"
import AuditsPanel from '@/components/AuditsPanel.vue'
import SavePanel from "@/components/SavePanel.vue"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useETDStore } from "@/stores/etd"
import FileUpload from 'primevue/fileupload'
import Panel from 'primevue/panel'
import WaitSpinner from "@/components/WaitSpinner.vue"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useConfirm } from "primevue/useconfirm"
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { usePinnable } from '@/composables/pin'

const confirm = useConfirm()
const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const etdRepo = useETDStore()
const etdForm = ref(null)
const savepanel = ref(null)

const adminEdit = computed( () => {
   return route.path.includes("/admin")
})

const workDescribed = computed( () => {
   if ( etdForm.value ) {
     return (etdForm.value.node.context.state.valid && etdRepo.hasAdvisor )
   }
   return false
})

onBeforeMount( async () => {
   document.title = "LibraETD"
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   await etdRepo.getWork( route.params.id )
   usePinnable("save-panel", "etd-form-layout", "libra-footer")
})

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

const inputLabel = ( (lbl, idx) => {
   if (idx==0) return lbl
   return null
})
const addAdvisor = ( () => {
   etdRepo.work.advisors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
   focusNewEntry("advisor", etdRepo.work.advisors.length, "input")
})
const removeAdvisor = ((idx)=> {
   etdRepo.work.advisors.splice(idx,1)
})
const checkAdvisorID = ((idx) => {
   let cID = etdRepo.work.advisors[idx].computeID
   etdRepo.work.advisors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      etdRepo.work.advisors[idx] = auth
   }).catch( () => {
      etdRepo.work.advisors[idx].msg = cID+" is not a valid computing ID"
   })
})
const removeKeyword = ((idx)=> {
   etdRepo.removeKeyword(idx)
})
const addKeyword = ( () => {
   etdRepo.work.keywords.push("")
   focusNewEntry("keyword", etdRepo.work.keywords.length, "input")
})
const removeURL = ((idx)=> {
   etdRepo.removeURL(idx)
})
const addURL = ( () => {
   etdRepo.work.relatedURLs.push("")
   focusNewEntry("url", etdRepo.work.relatedURLs.length, "input")
})
const removeAgency = ((idx)=> {
   etdRepo.removeAgency(idx)
})
const addAgency = ( () => {
   etdRepo.work.sponsors.push("")
   focusNewEntry("agency", etdRepo.work.sponsors.length, "input")
})

const focusNewEntry = (( name, length, type) => {
   nextTick( () => {
      const tgtInput = document.querySelector(`#${name}-${length} ${type}`)
      if (tgtInput) {
         tgtInput.focus()
      }
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
const saveClicked = ( async (visibility, releaseDate, releaseVisibility) => {
   updateWorkModel(visibility, releaseDate, releaseVisibility)
   etdForm.value.node.submit()
})
const updateWorkModel = (( visibility, releaseDate, releaseVisibility ) => {
   etdRepo.visibility = visibility
   etdRepo.embargoReleaseDate =  releaseDate
   etdRepo.embargoReleaseVisibility =  releaseVisibility
   let license = system.licenseDetail(etdRepo.licenseID)
   etdRepo.work.license = license.label
   etdRepo.work.licenseURL = license.url
})

const submitHandler = ( async () => {
   await etdRepo.update( )
   if ( system.showError == false ) {
      router.push("/etd")
   }
})

const adminSaveCliced = ( async(data) => {
   etdRepo.visibility = data.visibility
   etdRepo.embargoReleaseDate = data.embargoEndDate
   etdRepo.embargoReleaseVisibility = data.embargoEndVisibility
   etdRepo.work.program = data.program
   etdRepo.work.degree = data.degree
   etdRepo.work.adminNotes = data.adminNotes
   await etdRepo.update( )
   if ( system.showError == false ) {
      router.back()
   }
})

const cancelClicked = (() => {
   etdRepo.cancelEdit()
   router.back()
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .scroll-body {
      padding: 25px;
      text-align: left;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-start;
      gap: 25px;
   }
   .sidebar-col {
      width: 400px;
   }
   .sidebar-col.admin {
      width: 470px;
   }
   .main-form {
      margin-bottom: 20px;
   }
}
@media only screen and (max-width: 768px) {
   .scroll-body {
      padding: 10px;
      text-align: left;
      display: flex;
      flex-flow: column-reverse;
      gap: 0;
   }
   .audit {
      display: none;
   }
   .main-form {
      margin-bottom: 15px;
   }
   .sidebar-col {
      width: 100%;
   }
}

.scroll-body {
   .work-header {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      width:100%;
      .draft {
         font-weight: bold;
      }
   }
}

.content-col {
   display: flex;
   flex-direction: column;

   .acts {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 0.5rem;
      width: fit-content;
   }
   table {
      td.label {
         font-weight: bold;
         text-align: right;
         padding-right: 10px;
      }
   }
   .two-col {
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      align-items: flex-start;
      gap: 25px;
      div.formkit-outer {
         flex-grow: 1;
      }
   }

   .no-advisor {
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      gap: 0.3rem;
   }

   .advisors {
      display: flex;
      flex-direction: column;
      justify-content: flex-start;
      align-items: stretch;
      gap: 15px;
      .advisor {
         display: flex;
         flex-direction: column;
         justify-content: flex-start;
         align-items: stretch;
         gap: 15px;
         border-top: 1px solid $uva-grey-100;
         padding-top: 15px;
         .id-field {
            display: flex;
            flex-flow: row nowrap;
            align-items: baseline;
            justify-content: space-between;
            .search-field {
               display: flex;
               flex-flow: row nowrap;
               align-items: flex-end;
               gap: 0.5rem;
            }
         }
      }
   }
   .err {
      padding: 0;
      margin: 2px 0 0 0;
      color: $uva-red-A;
      font-style: italic;
   }
   .input-row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      gap: 5px;
      .input-wrap {
         flex-grow: 1;
      }
   }
   .note {
      font-style: italic;
      color: $uva-grey;
      margin-top: 0;
   }
}
</style>