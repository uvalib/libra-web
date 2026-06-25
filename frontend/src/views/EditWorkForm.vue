<template>
   <div class="edit">
      <div class="work-head">
         <h1>
            <span>Libra ETD Work</span>
            <span v-if="etdRepo.isDraft && !etdRepo.error" class="draft">DRAFT</span>
         </h1>
         <div v-if="!etdRepo.error" class="help">
            View <a target="_blank" aria-describedby="new-window" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.
         </div>
      </div>
      <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
      <div v-else-if="etdRepo.error" class="error">
         <h2>Sorry, a system error has occurred!</h2>
         <div>{{ etdRepo.error }}</div>
      </div>
      <div v-else class="sections">
         <Panel header="Metadata" toggleable pt:title:id="metadata-panel" pt:contentContainer:aria-labelledby="metadata-panel">
            <ProgramPanel :admin="user.isAdmin" @changed="formChanged"/>
            <div class="fields">
               <div class="field" >
                  <LabeledInput label="Title" name="work.title" :required="true" @change="formChanged"
                     v-model="etdRepo.work.title" :readonly="etdRepo.source == 'sis' && user.isAdmin == false"
                  />
                  <Message v-if="errors.title" severity="error" size="small" variant="simple">{{ errors.title }}</Message>
               </div>
               <Fieldset legend="Author">
                  <div class="two-col">
                     <div class="field" >
                        <LabeledInput label="First Name" name="work.author.firstName" :required="true" v-model="etdRepo.work.author.firstName" @change="formChanged"/>
                        <Message v-if="errors.firstName" severity="error" size="small" variant="simple">{{ errors.firstName }}</Message>
                     </div>
                     <div class="field" >
                        <LabeledInput label="Last Name" name="work.author.lastName" :required="true" v-model="etdRepo.work.author.lastName" @change="formChanged"/>
                        <Message v-if="errors.lastName" severity="error" size="small" variant="simple">{{ errors.lastName }}</Message>
                     </div>
                  </div>
               </Fieldset>

               <AdvisorsPanel :errors="errors.advisors" @change="formChanged" /> 

               <div class="field" >
                  <LabeledInput label="Abstract" name="work.abstract" :required="true" v-model="etdRepo.work.abstract" type="textarea" @change="formChanged"/>
                  <Message v-if="errors.abstract" severity="error" size="small" variant="simple">{{ errors.abstract}}</Message>
               </div>

               <RepeatField title="Keywords" label="Keyword" name="keyword" @change="formChanged" help="Add one keyword or keyword phrase per line." v-model="etdRepo.work.keywords" />
               <LabeledInput label="Language" name="work.language" v-model="etdRepo.work.language" type="select" :options="system.languages" @change="formChanged"/>
               <RepeatField title="Related Links" label="Link" @change="formChanged" name="related"
                  help="A link to a website or other specific content (audio, video, PDF document) related to the work. Add one link per line."
                  v-model="etdRepo.work.relatedURLs"
               />
               <RepeatField title="Sponsoring Agencies" label="Agency" name="agency" @change="formChanged" v-model="etdRepo.work.sponsors" help="Add one agency per line."/>
               <LabeledInput label="Notes" name="work.notes" v-model="etdRepo.work.notes" type="textarea" @change="formChanged"/>
               <LabeledInput v-if="user.isAdmin" label="Admin Notes" name="work.adminNotes" v-model="etdRepo.work.adminNotes" type="textarea" @change="formChanged"/>
               <div class="field" >
                  <LabeledInput label="Rights" name="licenseID" :required="true" v-model="etdRepo.licenseID" type="select" :options="system.licenses" @change="formChanged"/>
                  <Message v-if="errors.licenseID" severity="error" size="small" variant="simple">{{ errors.licenseID }}</Message>
                  <div class="note">
                     Libra lets you choose an open license when you post your work, and will prominently display the
                     license you choose as part of the record for your work. See
                     <a href="https://creativecommons.org/share-your-work" target="_blank" aria-describedby="new-window">Choose a Creative Commons License</a>
                     for option details.
                  </div>
               </div>
            </div>
            <template #icons>
               <span v-if="metadataComplete" class="complete">
                  <i class="pi pi-check-circle"></i>
                  <span>Complete</span>
               </span>
               <span v-else class="incomplete">
                  <i class="pi pi-exclamation-circle"></i>
                  <span>Incomplete</span>
               </span>
            </template>
         </Panel>
         
         <FilesPanel />
         <VisibilityPanel :error="errors.visibility" @change="formChanged"/>
         
      </div>
      <div class="toolbar" v-if="!etdRepo.error && !etdRepo.working">
         <span class="group">
            <template v-if="user.isAdmin">
               <Button :disabled="!etdRepo.publishedAt" label="Unpublish" severity="danger" @click="unpublishClicked" />
               <Button :disabled="etdRepo.publishedAt" label="Delete" severity="danger" @click="deleteClicked" />
            </template>
            <Button label="Exit" severity="secondary" @click="exitClicked" />
         </span>
         <Message v-if="saveMessage" severity="success" variant="simple" size="large" :life="2000" @life-end="saveMessage=''">{{ saveMessage }}</Message>
         <span v-else class="last-saved">Last saved at {{ $formatDateTime(etdRepo.modifiedAt) }}</span>
         <span class="group">
            <Button :disabled="previewDisabled"  severity="success" @click="previewClicked" label="Review and submit" />
         </span>
      </div>
   </div>
</template>

<script setup>
import { ref, onBeforeMount, computed } from 'vue'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useETDStore } from "@/stores/etd"
import { useAdminStore } from "@/stores/admin"
import Panel from 'primevue/panel'
import Message from 'primevue/message'
import Fieldset from 'primevue/fieldset'
import { useConfirm } from "primevue/useconfirm"
import { useRouter, useRoute } from 'vue-router'
import { useIntervalFn } from '@vueuse/core'
import WaitSpinner from "@/components/WaitSpinner.vue"
import ProgramPanel from '@/components/ProgramPanel.vue'
import LabeledInput from '@/components/editform/LabeledInput.vue'
import RepeatField from '@/components/editform/RepeatField.vue'
import FilesPanel from '@/components/editform/FilesPanel.vue'
import VisibilityPanel from '@/components/editform/VisibilityPanel.vue'
import AdvisorsPanel from '@/components/editform/AdvisorsPanel.vue'

const confirm = useConfirm()
const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const etdRepo = useETDStore()
const admin = useAdminStore()

const isDirty = ref(false)
const metadataComplete = ref(false)
const lastChangedAt = ref( new Date().getTime() )
const saveMessage = ref("")

const errors = ref({
   title: "",
   firstName: "",
   lastName: "",
   abstract: "",
   licenseID: "",
   visibility: "",
   advisors: [] // array will contain { lastName: "", firstName: "" } for each adcisor
})

useIntervalFn(() => {
   const nowMs = new Date().getTime()
   const delta = nowMs - lastChangedAt.value
   if ( delta > 3000 && isDirty.value ) {
      // if no recent activity and form is dirty, save it
      saveChanges()
   }
}, 1000)

const formChanged = (() =>{
   lastChangedAt.value = new Date().getTime()   
   isDirty.value = true
   validate() 
})

onBeforeMount( async () => {
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   system.getMimeTypes()
   await etdRepo.getWork( route.params.id, "edit" )
   validate()
})

const previewDisabled = computed( () => {
   return ( isDirty.value || metadataComplete.value ==false || etdRepo.hasFiles==false || etdRepo.visibility == "")
})

const validate = (() => {
   metadataComplete.value = true
   errors.value.firstName = ""
   if ( etdRepo.work.author.firstName == "" ) {
      metadataComplete.value = false
      errors.value.firstName = "First name is required"
   } 
   errors.value.lastName = ""
   if ( etdRepo.work.author.lastName == "" ) {
      metadataComplete.value = false
      errors.value.lastName = "Last name is required"
   }
   errors.value.title = ""
   if ( etdRepo.work.title == "" ) {
      metadataComplete.value = false
      errors.value.title = "Title is required"
   } 

   errors.value.advisors = []
   if ( etdRepo.work.advisors.length == 0) {
      if ( user.isAdmin == false ) {
         // admins can override the advisor requirement
         metadataComplete.value = false
      }
   } else {
      etdRepo.work.advisors.forEach( (a) => {
         let advErr = {firstName: "", lastName: ""}
         if ( a.firstName == "") {
            metadataComplete.value = false
            advErr.firstName = 'Advisor first name is required'
         }
         if ( a.lastName == "") {
            metadataComplete.value = false
            advErr.lastName = 'Advisor last name is required'
         }
         errors.value.advisors.push( advErr )
      })
   }

   errors.value.abstract = ""
   if ( etdRepo.work.abstract == "" ) {
      metadataComplete.value = false
      errors.value.abstract = "Abstract is required"
   }

   let licID = parseInt(etdRepo.licenseID)
   errors.value.licenseID = ""
   if ( licID == 0 ) {
      metadataComplete.value = false
      errors.value.licenseID = "Rights are required"
   }

   errors.value.visibility = ""
   if ( !etdRepo.visibility || etdRepo.visibility == "") {
      errors.value.visibility = "Visibility is required"
   }
})

const previewClicked = (() => {
    router.push(`/preview/${etdRepo.work.id}`)
})

const exitClicked = (async () => {
   if ( isDirty.value ) {
      await saveChanges()
   }
   router.push(user.homePage)
})

const unpublishClicked = ( () => {
   confirm.require({
      message: "Unpublish this work? It will no longer be visible to UVA or worldwide users. Are you sure?",
      header: 'Confirm Work Unpublish',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.unpublish( etdRepo.work.id )
      },
   })
})

const deleteClicked = ( () => {
   confirm.require({
      message: "Delete this work? All data will be lost. This cannot be reversed. Are you sure?",
      header: 'Confirm Work Delete',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: (  ) => {
         admin.delete( etdRepo.work.id )
         router.push(user.homePage)
      },
   })
})

const saveChanges = ( async () => {
   saveMessage.value = "Saving changes..."
   let license = system.licenseDetail(etdRepo.licenseID)
   if (license) {
      etdRepo.work.license = license.label
      etdRepo.work.licenseURL = license.url
   }

   await etdRepo.update( )
   if ( system.showError == false ) {
      saveMessage.value = "Save complete"
      isDirty.value = false
      lastChangedAt.value = new Date().getTime()   
      validate()
   } else {
      return
   }
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .help {
      margin: 0 5% 20px 5%;
   }
   .sections, h1 {
      margin-left: 5%;
      margin-right: 5%;
      gap: 25px;
   }
}
@media only screen and (max-width: 768px) {
    h1, .help {
      margin: 15px;
   }
   .sections {
      margin-left: 2px;
      margin-right: 2px;
      gap: 15px;
   }
}
div.error {
   padding: 25px;
   min-height: 300px;
   text-align: center;
   width: 50%;
   margin: 0 auto;
   h2 {
      text-align: center;
      margin-top: 10px;
   }

}
.edit {
   text-align: left;
   position: relative;

   h1 {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      .draft {
         font-size: 0.95rem;
         padding: 0.5rem 0.75rem;
         border-radius: 0.3rem;
         border: 1px solid $uva-grey-100;
         background: $uva-grey-200;
      }
   }

   .sections {
      display: flex;
      flex-direction: column;

      .fields {
         display: flex;
         flex-direction: column;
         gap: 15px;
         .field {
            display: flex;
            flex-direction: column;
            gap: 10px;
         }
      }
      .note {
         font-style: italic;
         color: $uva-grey-A;
         margin-top: 0;
      }
   }

   .toolbar {
      background: $uva-grey-200;
      border-top: 1px solid $uva-grey;
      padding: 10px;
      margin-top: 50px;
      position: sticky;
      bottom: 0;
      display: flex;
      flex-flow: row wrap;
      gap: 10px;
      justify-content: space-between;
      align-items: center;
      .group {
         display: flex;
         flex-flow: row wrap;
         gap: 10px;
         justify-content: flex-start;
      }
   }
}
</style>