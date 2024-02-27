<template>
   <div class="form">
      <Panel header="Save Work" class="sidebar">
         <Fieldset legend="Requirements">
            <div class="requirement">
               <i v-if="workDescribed" class="done pi pi-check"></i>
               <i v-else class="not-done pi pi-exclamation-circle"></i>
               <span>Describe your work</span>
            </div>
            <div class="requirement">
               <i v-if="data.files.length > 0" class="done pi pi-check"></i>
               <i v-else class="not-done pi pi-exclamation-circle"></i>
               <span>Add files</span>
            </div>
            <div class="help">
               View the <a href="https://www.library.virginia.edu/libra/open/oc-checklist" target="_blank">Libra Open Checklist</a> for help.
            </div>
         </Fieldset>
         <Fieldset legend="Visibility">
            <div v-for="v in system.oaVisibility" :key="v.value" class="visibility-opt">
               <RadioButton v-model="data.visibility" :inputId="v.value"  :value="v.value"  class="visibility"/>
               <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
            </div>
         </Fieldset>
         <div class="agree">
            <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
            <label for="agree-cb">
               By saving this work, I agree to the
               <a href="https://www.library.virginia.edu/libra/open/libra-deposit-license" target="_blank">Libra Deposit Agreement</a>
            </label>
         </div>
         <div class="button-bar">
            <Button severity="secondary" label="Cancel" @click="cancelClicked" :disabled="repository.working"/>
            <Button label="Submit" @click="submitClicked" :disabled="!canSubmit"/>
         </div>
      </Panel>

      <Panel header="Add new work">
         <FormKit ref="oaForm" type="form" :actions="false" @submit="submitHandler">
            <FormKit type="select" label="Resource Type" v-model="data.resourceType"
               placeholder="Select a resource type"  outer-class="first"
               :options="system.oaResourceTypes" validation="required"
            />
            <FormKit label="Title" type="text" v-model="data.title" validation="required"/>

            <FormKit v-model="data.authors" type="list" dynamic #default="{ items }">
               <label class="libra-form-label">Authors</label>
               <p class="note controls">
                  <span>The main researchers involved in producing the work, or the authors of the publication, in priority order.</span>
                  <Button label="Add Author" @click="addAuthor"/>
               </p>
               <div class="authors">
                  <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                     <div class="author" :class="{border: index != data.authors.length-1}">
                        <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove author"
                           :disabled="data.authors.length == 1" @click="removeAuthor(index)"/>
                        <div class="two-col">
                           <div class="search-field">
                              <FormKit type="text" name="computeID" label="Computing ID"/>
                              <Button class="check" icon="pi pi-search" severity="secondary" @click="checkAuthorID(index)"/>
                           </div>
                           <span class="sep"></span>
                           <p class="note inline">Lookup a UVA Computing ID to automatically fill the remaining fields for this person.</p>
                        </div>
                        <p v-if="data.authors[index].msg != ''" class="err">{{ data.authors[index].msg }}</p>
                        <div class="two-col">
                           <FormKit type="text" name="firstName" label="First Name"/>
                           <span class="sep"></span>
                           <FormKit type="text" name="lastName" label="Last Name"/>
                        </div>
                        <div class="two-col">
                           <FormKit type="text" name="department" label="Department"/>
                           <span class="sep"></span>
                           <FormKit type="text" name="institution" label="Institution"/>
                        </div>
                     </div>
                  </FormKit>
               </div>
            </FormKit>

            <FormKit label="Abstract" type="textarea" v-model="data.abstract" rows="10" validation="required"/>

            <FormKit type="select" label="Rights" v-model="data.license"
               placeholder="Select rights"
               :options="system.oaLicenses" validation="required"
            />
            <p class="note">
               Libra lets you choose an open license when you post your work, and will prominently display the
               license you choose as part of the record for your work. See
               <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
               for option details.
            </p>

            <FormKit v-model="data.languages" type="list" dynamic #default="{ items }">
               <div v-for="(item, index) in items" :key="item" class="input-row">
                  <div class="input-wrap">
                     <FormKit type="select" :label="inputLabel('Language', index)" :index="index"
                        placeholder="Select a language" :options="system.languages"
                     />
                  </div>
                  <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove language"
                     :disabled="data.languages.length == 1" @click="removeLanguage(index)"/>
               </div>
            </FormKit>
            <p class="note controls">
               <span>The language of the work's content.</span>
               <Button label="Add Language" @click="addLanguage"/>
            </p>

            <FormKit v-model="data.keywords" type="list" dynamic #default="{ items }">
               <div v-for="(item, index) in items" :key="item" class="input-row">
                  <div class="input-wrap">
                     <FormKit :label="inputLabel('Keyword', index)" type="text" :index="index" />
                  </div>
                  <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove keyword"
                     :disabled="data.keywords.length == 1" @click="removeKeyword(index)"/>
               </div>
            </FormKit>
            <p class="note controls">
               <span>Add one keyword or keyword phrase per line.</span>
               <Button label="Add Keyword" @click="addKeyword"/>
            </p>

            <FormKit v-model="data.contributors" type="list" dynamic #default="{ items }">
               <label class="libra-form-label">Contributors</label>
               <p class="note controls">
                  <span>The person(s) responsible for contributing to the development of the resource, such as editor or producer (not an author).</span>
                  <Button label="Add Contributor" @click="addContributor"/>
               </p>
               <div class="authors">
                  <FormKit type="group" v-for="(item, index) in items" :key="item" :index="index">
                     <div class="author" :class="{border: index != data.contributors.length-1}">
                        <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove author"
                           :disabled="data.contributors.length == 1" @click="removeContributor(index)"/>
                        <div class="two-col">
                           <div class="search-field">
                              <FormKit type="text" name="computeID" label="Computing ID"/>
                              <Button class="check" icon="pi pi-search" severity="secondary" @click="checkContributorID(index)"/>
                           </div>
                           <span class="sep"></span>
                           <p class="note inline">Lookup a UVA Computing ID to automatically fill the remaining fields for this person.</p>
                        </div>
                        <p v-if="data.contributors[index].msg != ''" class="err">{{ data.contributors[index].msg }}</p>
                        <div class="two-col">
                           <FormKit type="text" name="firstName" label="First Name"/>
                           <span class="sep"></span>
                           <FormKit type="text" name="lastName" label="Last Name"/>
                        </div>
                        <div class="two-col">
                           <FormKit type="text" name="department" label="Department"/>
                           <span class="sep"></span>
                           <FormKit type="text" name="institution" label="Institution"/>
                        </div>
                     </div>
                  </FormKit>
               </div>
            </FormKit>

            <FormKit label="Publisher" type="text" v-model="data.publisher" validation="required"/>
            <p class="note">
               Libra lets you choose an open license when you post your work, and will prominently display the
               license you choose as part of the record for your work. See
               <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
               for option details.
            </p>
            <FormKit label="Source citation" type="text" v-model="data.citation"/>
            <p class="note">The bibliographic citation of the work that reflects where it was originally published.</p>
            <FormKit label="Published date" type="text" v-model="data.pubDate"/>

            <FormKit v-model="data.relatedURLs" type="list" dynamic #default="{ items }">
               <div v-for="(item, index) in items" :key="item" class="input-row">
                  <div class="input-wrap">
                     <FormKit :label="inputLabel('Related URL', index)" type="text" :index="index" />
                  </div>
                  <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove url"
                     :disabled="data.relatedURLs.length == 1" @click="removeURL(index)"/>
               </div>
            </FormKit>
            <p class="note controls">
               <span>Links to another version, another location with the file, website or other specific content (audio, video, PDF document) related to the work.</span>
               <Button label="Add URL" @click="addURL"/>
            </p>

            <FormKit v-model="data.sponsors" type="list" dynamic #default="{ items }">
               <div v-for="(item, index) in items" :key="item" class="input-row">
                  <div class="input-wrap">
                     <FormKit :label="inputLabel('Sponsoring Agency', index)" type="text" :index="index" />
                  </div>
                  <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove agency"
                     :disabled="data.sponsors.length == 1" @click="removeAgency(index)"/>
               </div>
            </FormKit>
            <p class="note controls">
               <Button label="Add Agency" @click="addAgency"/>
            </p>

            <FormKit label="Notes" type="textarea" v-model="data.notes" rows="10"/>

            <label class="libra-form-label">Files</label>
            <FileUpload name="file" :url="`/api/upload/${repository.depositToken}`"
               @upload="filesUploaded($event)" @before-send="uploadRequested($event)"
               @error="uploadFailed($event)" @removeUploadedFile="fileRemoved($event)"
               :multiple="true" :withCredentials="true" :auto="true"
               :showUploadButton="false" :showCancelButton="false">
               <template #empty>
                  <p>Click Choose or drag and drop files to upload. Uploaded files will be attached to the work upon submission.</p>
               </template>
            </FileUpload>

         </FormKit>
      </Panel>
   </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useRepositoryStore } from "@/stores/repository"
import FileUpload from 'primevue/fileupload'
import Checkbox from 'primevue/checkbox'
import Fieldset from 'primevue/fieldset'
import RadioButton from 'primevue/radiobutton'
import Panel from 'primevue/panel'
import axios from 'axios'
import { useRouter } from 'vue-router'

const router = useRouter()

const system = useSystemStore()
const user = useUserStore()
const repository = useRepositoryStore()

const oaForm = ref(null)
const data = ref({
   resourceType: "Book",
   title: "title",
   authors: [],
   abstract: "ABS",
   license: "1",
   languages: ["English"],
   keywords: ["key1"],
   contributors: [{computeID: "", firstName: "", lastName: "", department: "", institution: "", msg: ""}],
   publisher: "University of Virginia",
   citation: "fake citation",
   pubDate: "1980",
   relatedURLs: ["fake_url"],
   sponsors: ["sponsor"],
   notes: "note text",
   files: [],
   visibility: ""
})
const agree = ref(false)

const canSubmit = computed( () => {
   if (workDescribed.value == false ) return false
   return agree.value == true && repository.working == false && data.value.visibility != "" && data.value.files.length > 0
})
const workDescribed = computed( () => {
   if ( oaForm.value ) {
      return oaForm.value.node.context.state.valid
   }
   return false
})
onMounted( async () => {
   if ( user.isSignedIn) {
      data.value.authors.push({
         computeID: user.computeID, firstName: user.firstName,
         lastName: user.lastName, department: user.department[0], institution: "University of Virginia", msg: ""
      })
   } else {
      data.value.authors.push({computeID: "", firstName: "", lastName: "", department: "", institution: "", msg: ""})
   }
   await repository.getDepositToken()
})

const uploadRequested = ( (request) => {
   console.log(request)
   request.xhr.setRequestHeader('Authorization', 'Bearer ' + user.jwt)
   return request
})

const fileRemoved = ( event => {
   repository.removeFile( event.file.name )
})
const filesUploaded = ( (event) => {
   event.files.forEach( uf => data.value.files.push( uf.name ))
})
const uploadFailed = ( (err) => {
   console.log(err)
})

const inputLabel = ( (lbl, idx) => {
   if (idx==0) return lbl
   return null
})
const addAuthor = ( () => {
   data.value.authors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})
const removeAuthor = ((idx)=> {
   data.value.authors.splice(idx,1)
})
const checkAuthorID = ((idx) => {
   let cID = data.value.authors[idx].computeID
   data.value.authors[idx].msg = ""
   if (cID.lenth <3) return
   axios.get(`/api/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      data.value.authors[idx] = auth
   }).catch( () => {
      data.value.authors[idx].msg = cID+" is not a valid computing ID"
   })
})
const addContributor = ( () => {
   data.value.contributors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})
const removeContributor = ((idx)=> {
   data.value.contributors.splice(idx,1)
})
const checkContributorID = ((idx) => {
   let cID = data.value.contributors[idx].computeID
   data.value.contributors[idx].msg = ""
   if (cID.lenth <3) return
   axios.get(`/api/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      data.value.contributors[idx] = auth
   }).catch( () => {
      data.value.contributors[idx].msg = cID+" is not a valid computing ID"
   })
})
const removeKeyword = ((idx)=> {
   data.value.keywords.splice(idx,1)
})
const addKeyword = ( () => {
   data.value.keywords.push("")
})
const removeLanguage = ((idx)=> {
   data.value.languages.splice(idx,1)
})
const addLanguage = ( () => {
   data.value.languages.push("")
})
const removeURL = ((idx)=> {
   data.value.relatedURLs.splice(idx,1)
})
const addURL = ( () => {
   data.value.relatedURLs.push("")
})
const removeAgency = ((idx)=> {
   data.value.sponsors.splice(idx,1)
})
const addAgency = ( () => {
   data.value.sponsors.push("")
})

const submitClicked = ( () => {
   const node = oaForm.value.node
   node.submit()
})
const submitHandler = ( () => {
   repository.depositOA( data.value )
})
const cancelClicked = (() => {
   repository.cancel()
   router.push("/oa")

})
</script>

<style lang="scss" scoped>
.form {
   margin: 50px auto;
   min-height: 600px;
   text-align: left;
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-start;
   align-items: flex-start;
   padding: 0 50px;
   .formkit-form {
      padding: 10px 20px 20px 20px;
   }

   .sidebar {
      width: 300px;
      margin-right: 25px;
      .help {
         font-size: 0.9em;
         margin-top:15px;
      }
      .requirement {
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: center;
         i {
            display: inline-block;
            margin-right: 10px;
            font-size: 1.25rem;
         }
         .not-done {
            color: var(--uvalib-red-darker);
         }
         .done {
            color: var(--uvalib-green-dark);
         }
      }
      .requirement:first-of-type {
         margin-bottom: 5px;
      }
      .visibility-opt {
         margin: 5px 0;
         label.visibility {
            font-size: 0.85em;
            margin-left: 10px;
            border-radius: 5px;
            padding: 2px 10px;
            color: white;
         }
         .visibility.open {
            background-color: var(--uvalib-green-dark);
         }
         .visibility.authenticated {
            background-color: var(--uvalib-brand-orange);
         }
         .visibility.embargo {
            background-color: var(--uvalib-blue-alt);
         }
         .visibility.restricted {
            background-color: var(--uvalib-red-darker);
         }
      }
      .agree {
         margin: 25px 0;
         label {
            margin-left: 10px;
         }
      }
      .button-bar {
         display: flex;
         flex-flow: row nowrap;
         justify-content: center;
         align-items: flex-end;
         button {
            margin-left: 10px;
         }
      };
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

   .authors {
      border: 1px solid var(--uvalib-grey-light);
      border-radius: 5px;
      padding: 0;
      .author {
         position: relative;
         padding: 0 25px 25px 25px;
         .remove {
            position: absolute;
            right: 5px;
            top: 5px;
         }
      }
      .author.border {
         padding-bottom: 25px;
         border-bottom: 1px solid var(--uvalib-grey-light);
      }
   }
   .two-col {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      p.note.inline {
         max-width: 50%;
         padding: 0;
         margin: 0;
         position: relative;
         top: -5px;
      }
      .sep {
         display: block;
         width: 15px;
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
      .remove {
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
   .note.controls {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      button {
         font-size: 0.9em;
         padding: 4px 10px;
         white-space: nowrap;
         margin-left: auto;
      }
   }
}
</style>