<template>
   <div class="scroll-body">
      <div class="form" id="oa-form-layout">
         <div class="sidebar-col" :class="{admin: adminEdit}" v-if="oaRepo.working==false">
            <AdminPanel v-if="adminEdit"
               type="oa"  :identifier="oaRepo.work.id" :created="oaRepo.createdAt" :depositor="oaRepo.depositor"
               :modified="oaRepo.modifiedAt" :published="oaRepo.publishedAt" :visibility="oaRepo.visibility"
               :embargoEndDate="oaRepo.embargoReleaseDate" :embargoEndVisibility="oaRepo.embargoReleaseVisibility"
               :notes="oaRepo.work.adminNotes" ref="savepanel" @cancel="cancelClicked" @delete="router.back()" @save="adminSaveCliced"
            />
            <SavePanel v-else
               type="oa" :create="isNewSubmission" :described="workDescribed" :files="oaRepo.work.files.length > 0 || oaRepo.pendingFileAdd.length > 0"
               :visibility="oaRepo.visibility" :releaseDate="oaRepo.embargoReleaseDate" :releaseVisibility="oaRepo.embargoReleaseVisibility"
               :draft="oaRepo.isDraft" @submit="submitClicked" @cancel="cancelClicked" ref="savepanel"
            />
         </div>

         <Panel class="main-form">
            <template #header>
               <div class="work-header">
                  <span>{{ panelTitle }}</span>
                  <template v-if="adminEdit==false">
                     <span v-if="oaRepo.isDraft" class="visibility draft">DRAFT</span>
                     <span v-else><b>Published</b>: {{ $formatDate(oaRepo.publishedAt) }}</span>
                  </template>
                  <AuditsPanel v-if="oaRepo.working==false && isNewSubmission==false" :workID="oaRepo.work.id" :namespace="system.oaNamespace" />
               </div>
            </template>
            <WaitSpinner v-if="oaRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
            <FormKit v-else ref="oaForm" type="form" :actions="false" @submit="submitHandler">
               <FormKit type="select" label="Resource Type" v-model="oaRepo.work.resourceType"
                  placeholder="Select a resource type"  outer-class="first"
                  :options="system.oaResourceTypes" validation="required"
               />
               <FormKit label="Title" type="text" v-model="oaRepo.work.title" validation="required"/>

               <Panel class="sub-panel">
                  <template #header>
                     <span class="hdr">
                        <div>Authors</div>
                        <span>
                           <Button label="Add Author" @click="addAuthor"/>
                        </span>
                     </span>
                  </template>
                  <FormKit v-model="oaRepo.work.authors" type="list" dynamic #default="{ items }">
                     <p class="note">The main researchers involved in producing the work, or the authors of the publication, in priority order.</p>
                     <p class="note">Lookup a UVA Computing ID to automatically fill the remaining fields for this person.</p>
                     <div class="authors">
                        <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                           <div class="author">
                              <div class="titlebar">
                                 <span>Author {{ index+1 }}</span>
                                 <span class="buttons">
                                    <Button :disabled="isMoveDisabled(index, 'down')" icon="pi pi-angle-down"
                                    severity="secondary" aria-label="author down" @click="moveAuthor(index, 'down')"/>
                                    <Button :disabled="isMoveDisabled(index, 'down')" icon="pi pi-angle-double-down"
                                       severity="secondary" aria-label="author last" @click="moveAuthor(index, 'last')"/>
                                    <Button :disabled="isMoveDisabled(index, 'up')" icon="pi pi-angle-up"
                                       severity="secondary" aria-label="author up" @click="moveAuthor(index, 'up')"/>
                                    <Button :disabled="isMoveDisabled(index, 'up')" icon="pi pi-angle-double-up"
                                       severity="secondary" aria-label="author first" @click="moveAuthor(index, 'first')"/>
                                    <Button :disabled="oaRepo.work.authors.length==1" icon="pi pi-trash"
                                       severity="danger" aria-label="remove author" @click="removeAuthor(index)"/>
                                 </span>
                              </div>
                              <div class="fields">
                                 <div class="id-field">
                                    <div class="search-field" :id="`author-${index+1}`">
                                       <FormKit type="text" name="computeID" label="Computing ID" outer-class="first"/>
                                       <Button class="check" icon="pi pi-search" severity="secondary" @click="checkAuthorID(index)"/>
                                    </div>
                                 </div>
                                 <p v-if="oaRepo.work.authors[index].msg != ''" class="err">{{ oaRepo.work.authors[index].msg }}</p>
                                 <div class="two-col">
                                    <FormKit type="text" name="firstName" label="First Name"/>
                                    <FormKit type="text" name="lastName" label="Last Name"/>
                                 </div>
                                 <div class="two-col">
                                    <FormKit type="text" name="department" label="Department"/>
                                    <FormKit type="text" name="institution" label="Institution"/>
                                 </div>
                              </div>
                           </div>
                        </FormKit>
                     </div>
                  </FormKit>
               </Panel>

               <FormKit label="Abstract" type="textarea" v-model="oaRepo.work.abstract" rows="10" validation="required"/>

               <FormKit type="select" label="Rights" v-model="oaRepo.licenseID"
                  placeholder="Select rights"
                  :options="system.oaLicenses" validation="required"
               />
               <p class="note">
                  Libra lets you choose an open license when you post your work, and will prominently display the
                  license you choose as part of the record for your work. See
                  <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
                  for option details.
               </p>

               <FormKit v-model="oaRepo.work.languages" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap"  :id="`language-${index+1}`">
                        <FormKit type="select" :label="inputLabel('Language', index)" :index="index"
                           placeholder="Select a language" :options="system.languages"
                        />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.languages.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add language" @click="addLanguage"/>
                     <Button v-if="index > 0 || index == 0 && oaRepo.work.languages[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove language" @click="removeLanguage(index)"/>
                  </div>
                  <p class="note">The language of the work's content.</p>
               </FormKit>


               <FormKit v-model="oaRepo.work.keywords" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`keyword-${index+1}`">
                        <FormKit :label="inputLabel('Keywords', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.keywords.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add keyword" @click="addKeyword"/>
                     <Button v-if="index > 0 || index == 0 && oaRepo.work.keywords[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove keyword"  @click="removeKeyword(index)"/>
                  </div>
                  <p class="note">Add one keyword or keyword phrase per line.</p>
               </FormKit>

               <Panel class="sub-panel">
                  <template #header>
                     <span class="hdr">
                        <div>Contributors</div>
                        <Button label="Add Contributor" @click="addContributor"/>
                     </span>
                  </template>
                  <FormKit v-model="oaRepo.work.contributors" type="list" dynamic #default="{ items }">
                     <p class="note">The person(s) responsible for contributing to the development of the resource, such as editor or producer (not an author).</p>
                     <div class="contributors">
                        <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                           <div class="author">
                              <div class="id-field">
                                 <div class="search-field" :id="`contributor-${index+1}`">
                                    <FormKit type="text" name="computeID" label="Computing ID"/>
                                    <Button class="check" icon="pi pi-search" severity="secondary" @click="checkContributorID(index)"/>
                                 </div>
                                 <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove contributor" @click="removeContributor(index)"/>
                              </div>
                              <p v-if="oaRepo.work.contributors[index].msg != ''" class="err">{{ oaRepo.work.contributors[index].msg }}</p>
                              <div class="two-col">
                                 <FormKit type="text" name="firstName" label="First Name"/>
                                 <FormKit type="text" name="lastName" label="Last Name"/>
                              </div>
                              <div class="two-col">
                                 <FormKit type="text" name="department" label="Department"/>
                                 <FormKit type="text" name="institution" label="Institution"/>
                              </div>
                           </div>
                        </FormKit>
                     </div>
                  </FormKit>
               </Panel>

               <FormKit label="Publisher" type="text" v-model="oaRepo.work.publisher" validation="required"/>
               <p class="note">
                  The original publisher of the work. If there is no original publisher, leave "University of Virginia" in this field.
               </p>
               <FormKit label="Source citation" type="text" v-model="oaRepo.work.citation"/>
               <p class="note">The bibliographic citation of the work that reflects where it was originally published.</p>
               <FormKit label="Published date" type="text" v-model="oaRepo.work.pubDate"/>

               <FormKit v-model="oaRepo.work.relatedURLs" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`url-${index+1}`">
                        <FormKit :label="inputLabel('Related URL', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.relatedURLs.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add url" @click="addURL"/>
                     <Button v-if="index > 0 || index == 0 && oaRepo.work.relatedURLs[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove url"  @click="removeURL(index)"/>
                  </div>
                  <p class="note">Links to another version, another location with the file, website or other specific content (audio, video, PDF document) related to the work.</p>
               </FormKit>

               <FormKit v-model="oaRepo.work.sponsors" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap" :id="`agency-${index+1}`">
                        <FormKit :label="inputLabel('Sponsoring Agency', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.sponsors.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add agency" @click="addAgency"/>
                     <Button v-if="index > 0 || index == 0 && oaRepo.work.sponsors[0] != ''"
                        class="remove" icon="pi pi-trash" severity="danger" aria-label="remove agency"  @click="removeAgency(index)"/>
                  </div>
               </FormKit>

               <FormKit label="Notes" type="textarea" v-model="oaRepo.work.notes" rows="10"/>

               <template v-if="isNewSubmission==false">
                  <label class="libra-form-label">Previously Uploaded Files</label>
                  <DataTable :value="oaRepo.work.files" ref="oaFiles" dataKey="id"
                        stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
                        :lazy="false" :paginator="true" :alwaysShowPaginator="false"
                        :rows="30" :totalRecords="oaRepo.work.files.length"
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
                           <Button class="action" icon="pi pi-trash" label="Delete" severity="danger" text @click="deleteFileClicked(slotProps.data.name)"/>
                           <Button class="action" icon="pi pi-cloud-download" label="Download" severity="secondary" text @click="downloadFileClicked(slotProps.data.name)"/>
                        </template>
                     </Column>
                  </DataTable>
               </template>
               <label class="libra-form-label">Files</label>
               <FileUpload name="file" :url="`/api/upload/${uploadToken}`"
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
import SavePanel from "@/components/SavePanel.vue"
import AuditsPanel from "@/components/AuditsPanel.vue"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useOAStore } from "@/stores/oa"
import FileUpload from 'primevue/fileupload'
import Panel from 'primevue/panel'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { usePinnable } from '@/composables/pin'
import WaitSpinner from "@/components/WaitSpinner.vue"
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useConfirm } from "primevue/useconfirm"

usePinnable("user-header", "scroll-body", ( (isPinned) => {
   const formEle = document.getElementById("oa-form-layout")
   const compStyles = window.getComputedStyle(formEle)
   const flowType = compStyles.getPropertyValue("flex-flow")

   // in mobile mode, the panel is at the bottom of the screen and doesn't need to be pinned
   // when this is the case, the flex-flow will be "column-reverse".
   if ( flowType.indexOf("column") == -1) {
      let panelEle = savepanel.value.$el
      if ( isPinned ) {
         panelEle.style.top = `85px` // HACK: top padding + height of user toolbar
         panelEle.style.width = `${panelEle.getBoundingClientRect().width}px`
         // FIXME this breaks with the admin panel - it is wider. the sidebar-col collapses
         panelEle.classList.add("pinned")
      } else {
         panelEle.classList.remove("pinned")
      }
   }
}))

const confirm = useConfirm()
const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const oaRepo = useOAStore()
const oaForm = ref(null)
const savepanel = ref(null)
const panelTitle = ref("Add New LibraOpen Work")

const adminEdit = computed( () => {
   return route.path.includes("/admin")
})

const workDescribed = computed( () => {
   if ( oaForm.value ) {
     return oaForm.value.node.context.state.valid
   }
   return false
})

const isNewSubmission = computed(() => {
   return route.params.id == "new"
})

const uploadToken = computed( () => {
   if ( isNewSubmission.value) {
      return oaRepo.depositToken
   }
   return oaRepo.work.id
})

onBeforeMount( async () => {
   document.title = "LibraOpen"
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   if ( adminEdit.value) {
      panelTitle.value = "LibraOpen Work"
      await oaRepo.getWork( route.params.id )
   } else if ( isNewSubmission.value == false) {
      panelTitle.value = "Edit LibraOpen Work"
      await oaRepo.getWork( route.params.id )
   } else {
      await oaRepo.initDeposit(user.computeID, user.firstName, user.lastName, user.department[0])
   }
})

const isMoveDisabled = ( (index, direction) => {
   let authCnt = oaRepo.work.authors.length
   if ( authCnt == 1 ) return true
   if ( index == 0 && direction == "up") return true
   if ( index == (authCnt-1) && direction == "down") return true
   return false
})
const moveAuthor = ( (index, direction) => {
   let tgtAuth = oaRepo.work.authors[index]
   if (direction == "down" ) {
      let auth2 =  oaRepo.work.authors[index+1]
      oaRepo.work.authors[index] = auth2
      oaRepo.work.authors[index+1] = tgtAuth
   } else if (direction == "last" ) {
      oaRepo.work.authors.splice(index,1)
      oaRepo.work.authors.push(tgtAuth)
   } else if (direction == "up" ) {
      let auth2 =  oaRepo.work.authors[index-1]
      oaRepo.work.authors[index] = auth2
      oaRepo.work.authors[index-1] = tgtAuth
   } else if (direction == "first" ) {
       oaRepo.work.authors.splice(index,1)
      oaRepo.work.authors.unshift(tgtAuth)
   }
})

const uploadRequested = ( (request) => {
   request.xhr.setRequestHeader('Authorization', 'Bearer ' + user.jwt)
   return request
})

const fileRemoved = ( event => {
   oaRepo.removeFile( event.file.name )
})
const fileUploaded = ( (event) => {
   event.files.forEach( f => {
      oaRepo.addFile( f.name )
   })
})

const inputLabel = ( (lbl, idx) => {
   if (idx==0) return lbl
   return null
})
const addAuthor = ( () => {
   oaRepo.work.authors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
   focusNewEntry("author", oaRepo.work.authors.length, "input")
})
const removeAuthor = ((idx)=> {
   oaRepo.work.authors.splice(idx,1)
})
const checkAuthorID = ((idx) => {
   let cID = oaRepo.work.authors[idx].computeID
   oaRepo.work.authors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      oaRepo.work.authors[idx] = auth
   }).catch( () => {
      oaRepo.work.authors[idx].msg = cID+" is not a valid computing ID"
   })
})
const addContributor = ( () => {
   oaRepo.work.contributors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
   focusNewEntry("contributor", oaRepo.work.contributors.length, "input")
})
const removeContributor = ((idx)=> {
   oaRepo.removeContributor(idx)
})
const checkContributorID = ((idx) => {
   let cID = oaRepo.work.contributors[idx].computeID
   oaRepo.work.contributors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      oaRepo.work.contributors[idx] = auth
   }).catch( () => {
      oaRepo.work.contributors[idx].msg = cID+" is not a valid computing ID"
   })
})
const removeKeyword = ((idx)=> {
   oaRepo.removeKeyword(idx)
})
const addKeyword = ( () => {
   oaRepo.work.keywords.push("")
   focusNewEntry("keyword", oaRepo.work.keywords.length, "input")
})
const removeLanguage = ((idx)=> {
   oaRepo.removeLanguage(idx)
})
const addLanguage = ( () => {
   oaRepo.work.languages.push("")
   focusNewEntry("language", oaRepo.work.languages.length, "select")
})
const removeURL = ((idx)=> {
   oaRepo.removeURL(idx)
})
const addURL = ( () => {
   oaRepo.work.relatedURLs.push("")
   focusNewEntry("url", oaRepo.work.relatedURLs.length, "input")
})
const removeAgency = ((idx)=> {
   oaRepo.removeAgency(idx)
})
const addAgency = ( () => {
   oaRepo.work.sponsors.push("")
   focusNewEntry("agency", oaRepo.work.sponsors.length, "input")
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
         oaRepo.removeFile(name)
      },
   })
})
const downloadFileClicked = ( (name) => {
   oaRepo.downloadFile(name)
})
const submitClicked = ( (visibility, releaseDate, releaseVisibility ) => {
   // update work submission with other details from items that are not directly part of the metadata record
   oaRepo.visibility = visibility
   oaRepo.embargoReleaseDate =  releaseDate
   oaRepo.embargoReleaseVisibility =  releaseVisibility

   let license = system.licenseDetail("oa", oaRepo.licenseID)
   oaRepo.work.license = license.label
   oaRepo.work.licenseURL = license.url

   // Get the forkmit root node and manually call submit to trigger the automated form validations
   // if the validations are successful, formkit submitHandler will be called and submit the updates
   const node = oaForm.value.node
   node.submit()
})

const submitHandler = ( async () => {
   if ( isNewSubmission.value ) {
      await oaRepo.deposit( )
   } else {
      await oaRepo.update( )
   }
   if ( system.showError == false ) {
      router.push("/oa")
   }
})

const adminSaveCliced = ( async(data) => {
   oaRepo.visibility = data.visibility
   oaRepo.embargoReleaseDate = data.embargoEndDate
   oaRepo.embargoReleaseVisibility = data.embargoEndVisibility
   oaRepo.work.adminNotes = data.adminNotes
   await oaRepo.update( )
   if ( system.showError == false ) {
      router.back()
   }
})


const cancelClicked = (() => {
   if ( isNewSubmission.value) {
      oaRepo.cancelCreate()
   } else {
      oaRepo.cancelEdit()
   }
   router.back()

})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .scroll-body {
      padding: 25px;
   }
   .sidebar-col {
      width: 400px;
      margin-right: 25px;
   }
   .sidebar-col.admin {
      width: 450px;
      margin-right: 25px;
   }
   .main-form {
      margin-bottom: 100px;
   }
   .form {
      text-align: left;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-start;
   }
}
@media only screen and (max-width: 768px) {
   .scroll-body {
      padding: 10px;
   }
   .main-form {
      margin-bottom: 15px;
   }
   .sidebar-col {
      width: 100%;
   }
   :deep(.p-panel-content) {
      padding: 10px;
   }
   .form {
      text-align: left;
      display: flex;
      flex-flow: column-reverse;
   }
}

.scroll-body {
   display: block;
   position: relative;
}

.action {
   margin-right: 15px;
}

.work-header {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   align-items: center;
   width: 100%;
}

.form {
   text-align: left;

   .sub-panel {
      margin-top: 25px;
      :deep(.p-panel-header) {
         background-color: #fcfcfc;
         padding: 10px;
      }
      .id-field {
         display: flex;
         flex-flow: row nowrap;
         align-items: baseline;
         justify-content: space-between;
      }
      .note {
         padding: 0;
         margin-bottom: 0px;
      }
      .hdr {
         width: 100%;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
         button {
            font-size: 0.8em;
            padding: 5px 10px;
            margin-right: 5px;
         }
      }
      .contributors {
         padding: 0;
         .author {
            position: relative;
            border-top: 1px solid var(--uvalib-grey-light);
            margin-top: 20px;
         }
      }
      .authors {
         padding: 0;
         .author {
            position: relative;
            border: 1px solid var(--uvalib-grey-light);
            border-radius: 5px;
            margin-top: 20px;
            .titlebar {
               margin: 0;
               background: #fcfcfc;
               padding: 4px 8px;
               border-radius: 5px 5px 0 0;
               border-bottom: 1px solid var(--uvalib-grey-light);
               display: flex;
               flex-flow: row wrap;
               justify-content: space-between;
               align-items: center;
               button {
                  margin-left: 5px;
                  border-radius: 20px;
               }
            }
            .fields {
               padding: 10px;
            }
         }
      }
   }

   .err {
      padding: 0;
      margin: 2px 0 0 0;
      color: var(--uvalib-red-emergency);
      font-style: italic;
   }

   .search-field {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      button {
         font-size: 0.8em;
         padding: 7px;
         margin-bottom: 0.3em;
         margin-left: 5px;
      }
   }

   .two-col {
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-items: flex-end;
      .formkit-outer:first-of-type {
         margin-right: 15px;
      }
      div.formkit-outer {
         flex-grow: 1;
      }
   }

   .input-row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      .remove, .add {
         padding: 6.25px 15px;
         margin-bottom: 0.3em;
         border: 0;
         margin-left: 5px;
      }
      .input-wrap {
         flex-grow: 1;
      }
   }
   .note {
      font-size: 0.85em;
      font-style: italic;
      color: var(--uvalib-grey);
      margin-top: 0;
      padding-top: 5px;
   }
}
</style>@/stores/oa