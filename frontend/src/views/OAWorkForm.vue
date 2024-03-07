<template>
   <div class="scroll-body">
      <div class="form" id="oa-form-layout">
         <div class="sidebar-col">
            <SavePanel v-if="oaRepo.working==false"
               type="oa" mode="create" :described="workDescribed" :files="oaRepo.work.files.length > 0"
               @submit="submitClicked" @cancel="cancelClicked" ref="savepanel" :visibility="oaRepo.work.visibility"/>
         </div>

         <Panel :header="panelTitle" class="main-form">
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
                        <Button label="Add Author" @click="addAuthor"/>
                     </span>
                  </template>
                  <FormKit v-model="oaRepo.work.authors" type="list" dynamic #default="{ items }">
                     <p class="note">The main researchers involved in producing the work, or the authors of the publication, in priority order.</p>
                     <p class="note">Lookup a UVA Computing ID to automatically fill the remaining fields for this person.</p>
                     <div class="authors">
                        <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                           <div class="author">
                              <div class="id-field">
                                 <div class="search-field">
                                    <FormKit type="text" name="computeID" label="Computing ID"/>
                                    <Button class="check" icon="pi pi-search" severity="secondary" @click="checkAuthorID(index)"/>
                                 </div>
                                 <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove author" @click="removeAuthor(index)"/>
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
                        </FormKit>
                     </div>
                  </FormKit>
               </Panel>

               <FormKit label="Abstract" type="textarea" v-model="oaRepo.work.abstract" rows="10" validation="required"/>

               <FormKit type="select" label="Rights" v-model="oaRepo.work.license"
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
                     <div class="input-wrap">
                        <FormKit type="select" :label="inputLabel('Language', index)" :index="index"
                           placeholder="Select a language" :options="system.languages"
                        />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.languages.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add language" @click="addLanguage"/>
                     <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove language" @click="removeLanguage(index)"/>
                  </div>
                  <p class="note">The language of the work's content.</p>
               </FormKit>


               <FormKit v-model="oaRepo.work.keywords" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap">
                        <FormKit :label="inputLabel('Keywords', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.keywords.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add keyword" @click="addKeyword"/>
                     <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove keyword"  @click="removeKeyword(index)"/>
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
                     <div class="authors">
                        <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                           <div class="author">
                              <div class="id-field">
                                 <div class="search-field">
                                    <FormKit type="text" name="computeID" label="Computing ID"/>
                                    <Button class="check" icon="pi pi-search" severity="secondary" @click="checkContributorID(index)"/>
                                 </div>
                                 <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove contributor" @click="removeContributor(index)"/>
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
                  Libra lets you choose an open license when you post your work, and will prominently display the
                  license you choose as part of the record for your work. See
                  <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
                  for option details.
               </p>
               <FormKit label="Source citation" type="text" v-model="oaRepo.work.citation"/>
               <p class="note">The bibliographic citation of the work that reflects where it was originally published.</p>
               <FormKit label="Published date" type="text" v-model="oaRepo.work.pubDate"/>

               <FormKit v-model="oaRepo.work.relatedURLs" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap">
                        <FormKit :label="inputLabel('Related URL', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.relatedURLs.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add url" @click="addURL"/>
                     <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove url"  @click="removeURL(index)"/>
                  </div>
                  <p class="note">Links to another version, another location with the file, website or other specific content (audio, video, PDF document) related to the work.</p>
               </FormKit>

               <FormKit v-model="oaRepo.work.sponsors" type="list" dynamic #default="{ items }">
                  <div v-for="(item, index) in items" :key="item" class="input-row">
                     <div class="input-wrap">
                        <FormKit :label="inputLabel('Sponsoring Agency', index)" type="text" :index="index" />
                     </div>
                     <Button v-if="index > 0 || oaRepo.work.sponsors.length == 1" class="add" icon="pi pi-plus" severity="success" aria-label="add agency" @click="addAgency"/>
                     <Button v-if="index > 0" class="remove" icon="pi pi-trash" severity="danger" aria-label="remove agency"  @click="removeAgency(index)"/>
                  </div>
               </FormKit>

               <FormKit label="Notes" type="textarea" v-model="oaRepo.work.notes" rows="10"/>

               <label class="libra-form-label">Files</label>
               <FileUpload name="file" :url="`/api/upload/${oaRepo.depositToken}`"
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
import { ref, onBeforeMount, computed } from 'vue'
import SavePanel from "@/components/SavePanel.vue"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useOAStore } from "@/stores/oa"
import FileUpload from 'primevue/fileupload'
import Panel from 'primevue/panel'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { usePinnable } from '@/composables/pin'
import WaitSpinner from "@/components/WaitSpinner.vue"

usePinnable("user-header", "scroll-body", ( (isPinned) => {
   const formEle = document.getElementById("oa-form-layout")
   const compStyles = window.getComputedStyle(formEle)
   const flowType = compStyles.getPropertyValue("flex-flow")

   // in mobile mode, the panel is at the bottom of the screen and doesn't need to be pinned
   // when this is the case, the flex-flow will be "column-reverse".
   if ( flowType.indexOf("column") == -1) {
      let panelEle = savepanel.value.$el
      if ( isPinned ) {
         panelEle.style.top = `88px` // HACK: top padding + height of user toolbar
         panelEle.style.width = `${panelEle.getBoundingClientRect().width}px`
         panelEle.classList.add("pinned")
      } else {
         panelEle.classList.remove("pinned")
      }
   }
}))

const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const oaRepo = useOAStore()

const oaForm = ref(null)
const savepanel = ref(null)
const panelTitle = ref("Add New Work")

const workDescribed = computed( () => {
   if ( oaForm.value ) {
     return oaForm.value.node.context.state.valid
   }
   return false
})

onBeforeMount( async () => {
   document.title = "LibraOpen"
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }

   oaRepo.initSubmission(user.computeID, user.firstName, user.lastName, user.department[0])
   await oaRepo.getDepositToken()
   if ( route.params.id != "new") {
      panelTitle.value = "Edit Work"
      await oaRepo.getWork( route.params.id )
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
   oaRepo.addFile( event.files[0].name )
})

const inputLabel = ( (lbl, idx) => {
   if (idx==0) return lbl
   return null
})
const addAuthor = ( () => {
   oaRepo.work.authors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})
const removeAuthor = ((idx)=> {
   oaRepo.work.authors.splice(idx,1)
})
const checkAuthorID = ((idx) => {
   let cID = oaRepo.work.authors[idx].computeID
   oaRepo.work.authors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      data.value.authors[idx] = auth
   }).catch( () => {
      data.value.authors[idx].msg = cID+" is not a valid computing ID"
   })
})
const addContributor = ( () => {
   oaRepo.work.contributors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})
const removeContributor = ((idx)=> {
   oaRepo.work.contributors.splice(idx,1)
})
const checkContributorID = ((idx) => {
   let cID = data.value.contributors[idx].computeID
   data.value.contributors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      data.value.contributors[idx] = auth
   }).catch( () => {
      data.value.contributors[idx].msg = cID+" is not a valid computing ID"
   })
})
const removeKeyword = ((idx)=> {
   oaRepo.work.keywords.splice(idx,1)
})
const addKeyword = ( () => {
   oaRepo.work.keywords.push("")
})
const removeLanguage = ((idx)=> {
   oaRepo.work.languages.splice(idx,1)
})
const addLanguage = ( () => {
   oaRepo.work.languages.push("")
})
const removeURL = ((idx)=> {
   oaRepo.work.relatedURLs.splice(idx,1)
})
const addURL = ( () => {
   oaRepo.work.relatedURLs.push("")
})
const removeAgency = ((idx)=> {
   oaRepo.work.sponsors.splice(idx,1)
})
const addAgency = ( () => {
   oaRepo.work.sponsors.push("")
})

const submitClicked = ( (visibility) => {
   oaRepo.work.visibility = visibility
   const node = oaForm.value.node
   node.submit()
})
const submitHandler = ( async () => {
   await oaRepo.deposit( )
   router.push("/oa")
})
const cancelClicked = (() => {
   oaRepo.cancel()
   router.push("/oa")

})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .scroll-body {
      padding: 50px;
   }
   .sidebar-col {
      width: 400px;
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
         }
      }
      .authors {
         padding: 0;
         .author {
            position: relative;
            border-top: 1px solid var(--uvalib-grey-light);
            margin-top: 20px;
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