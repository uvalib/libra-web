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
      <Form v-else v-slot="$form" :initialValues="etdRepo" :resolver="resolver" class="sections" ref="etdForm" :validateOnBlur="true" :validateOnMount="true">
         <Panel header="Metadata" toggleable pt:title:id="metadata-panel" pt:contentContainer:aria-labelledby="metadata-panel">
            <ProgramPanel :admin="user.isAdmin" @changed="programChanged = true"/>
            <div class="fields">
               <div class="field" >
                  <LabeledInput label="Title" name="work.title" :required="true"
                     v-model="etdRepo.work.title" :readonly="etdRepo.source == 'sis' && user.isAdmin == false"
                  />
                  <Message v-if="$form.work?.title?.invalid" severity="error" size="small" variant="simple">{{ $form.work.title.error.message }}</Message>
               </div>
               <Fieldset legend="Author">
                  <div class="two-col">
                     <div class="field" >
                        <LabeledInput label="First Name" name="work.author.firstName" :required="true" v-model="etdRepo.work.author.firstName"/>
                        <Message v-if="$form.work?.author?.firstName?.invalid" severity="error" size="small" variant="simple">{{ $form.work.author.firstName.error.message }}</Message>
                     </div>
                     <div class="field" >
                        <LabeledInput label="Last Name" name="work.author.lastName" :required="true" v-model="etdRepo.work.author.lastName"/>
                        <Message v-if="$form.work?.author?.lastName?.invalid" severity="error" size="small" variant="simple">{{ $form.work.author.lastName.error.message }}</Message>
                     </div>
                  </div>
               </Fieldset>

               <AdvisorsPanel :form="$form" @change="advisorsUpdated" />

               <div class="field" >
                  <LabeledInput label="Abstract" name="work.abstract" :required="true" v-model="etdRepo.work.abstract" type="textarea" />
                  <Message v-if="$form.work?.abstract?.invalid" severity="error" size="small" variant="simple">{{ $form.work.abstract.error.message }}</Message>
               </div>

               <RepeatField title="Keywords" label="Keyword" name="keyword" @change="listChanged=true" help="Add one keyword or keyword phrase per line." v-model="etdRepo.work.keywords" />
               <LabeledInput label="Language" name="work.language" v-model="etdRepo.work.language" type="select" :options="system.languages" />
               <RepeatField title="Related Links" label="Link" @change="listChanged=true" name="related"
                  help="A link to a website or other specific content (audio, video, PDF document) related to the work. Add one link per line."
                  v-model="etdRepo.work.relatedURLs"
               />
               <RepeatField title="Sponsoring Agencies" label="Agency" name="agency" @change="listChanged=true" v-model="etdRepo.work.sponsors" help="Add one agency per line."/>
               <LabeledInput label="Notes" name="work.notes" v-model="etdRepo.work.notes" type="textarea" />
               <LabeledInput v-if="user.isAdmin" label="Admin Notes" name="work.adminNotes" v-model="etdRepo.work.adminNotes" type="textarea" />
               <div class="field" >
                  <LabeledInput label="Rights" name="licenseID" :required="true" v-model="etdRepo.licenseID" type="select" :options="system.licenses" />
                  <Message v-if="$form.licenseID?.invalid" severity="error" size="small" variant="simple">{{ $form.licenseID.error.message }}</Message>
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
         <VisibilityPanel :form="$form" @embargo-changed="embargoChanged = true"/> 
         
      </Form>
      <div class="toolbar" v-if="!etdRepo.error && !etdRepo.working">
         <span class="group">
            <template v-if="user.isAdmin">
               <Button :disabled="!etdRepo.publishedAt" label="Unpublish" severity="danger" @click="unpublishClicked" />
               <Button :disabled="etdRepo.publishedAt" label="Delete" severity="danger" @click="deleteClicked" />
            </template>
            <Button label="Exit" severity="secondary" @click="exitClicked" />
         </span>
         <Message v-if="saved" severity="success" variant="simple" icon="pi pi-save" size="large" :life="3000" @life-end="saved=false">Changes saved</Message>
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
import { Form } from '@primevue/forms'
import Message from 'primevue/message'
import Fieldset from 'primevue/fieldset'
import { useConfirm } from "primevue/useconfirm"
import { useRouter, useRoute } from 'vue-router'
import { onKeyStroke } from '@vueuse/core'
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

const etdForm = ref(null)
const listChanged = ref(false)
const advisorsChanged = ref(false)
const programChanged = ref(false)
const embargoChanged = ref(false)
const metadataComplete = ref(false)
const lastKeyStrokeAt = ref( new Date().getTime())
const saved = ref(false)

useIntervalFn(() => {
   const nowMs = new Date().getTime()
   const delta = nowMs - lastKeyStrokeAt.value
   if ( delta > 3000 && needsSave.value ) {
      // if no recent activity and form is dirty, save it
      saveChanges()
   }
}, 2000)

onBeforeMount( async () => {
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   system.getMimeTypes()
   await etdRepo.getWork( route.params.id, "edit" )
   setInterval
})

onKeyStroke((e) => {
   // watch for any keystroke targeted at form inputs. reset the activity timer
   if ( e.target.classList.contains("p-textarea") || e.target.classList.contains("p-inputtext") ) {
      lastKeyStrokeAt.value = new Date().getTime()
   }
})

const previewDisabled = computed( () => {
   if (needsSave.value == true || metadataComplete.value ==false || etdRepo.hasFiles==false || etdRepo.visibility == "") {
      return true
   }
   return false
})

const needsSave  = computed( () => {
   if ( !etdForm.value) {
      return false
   }
   return isDirty(etdForm.value.states)
})

const resolver = ({ values }) => {
   const errors = {
      work: {
         title: [],
         author: { lastName: [], firstName: [] },
         advisors: [ {advisor: { lastName: [], firstName: [] } } ],
         abstract: [] },
      licenseID: [],
      visibility: [],
   }
   metadataComplete.value = true

   if ( values.work.title == "" ) {
      metadataComplete.value = false
      errors.work.title = [{ message: 'Title is required' }]
   }

   if (values.work.author.firstName == "") {
      metadataComplete.value = false
      errors.work.author.firstName = [{ message: 'Author first name is required' }]
   }
   if (values.work.author.lastName == "") {
      metadataComplete.value = false
      errors.work.author.lastName = [{ message: 'Author last name is required' }]
   }

   if ( user.isAdmin == false ) {
      if ( etdRepo.work.advisors.length == 0) {
         metadataComplete.value = false
      } else {
         etdRepo.work.advisors.forEach( (a,idx) => {
            errors.work.advisors.push ({ firstName: [], lastName: []})
            if ( a.firstName == "") {
               metadataComplete.value = false
               errors.work.advisors[ idx ].firstName = [{ message: 'Advisor first name is required' }]
            }
            if ( a.lastName == "") {
               metadataComplete.value = false
               errors.work.advisors[ idx ].lastName = [{ message: 'Advisor last name is required' }]
            }
         })
      }
   }

   if ( values.work.abstract == "" ) {
      metadataComplete.value = false
      errors.work.abstract = [{ message: 'Abstract is required' }]
   }

   let licID = parseInt(values.licenseID)
   if ( !values.licenseID || licID == 0 ) {
      metadataComplete.value = false
      errors.licenseID = [{ message: 'Rights are required' }]
   }

   if ( !values.visibility || values.visibility == "") {
      errors.visibility = [{ message: 'Visibility is required' }]
   }

   return { values, errors }
}

const previewClicked = (() => {
    router.push(`/preview/${etdRepo.work.id}`)
})

const exitClicked = (async () => {
   if ( needsSave.value ) {
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

const isDirty = ((data) => {
   // these are repeat value items that are not handled by the resolver
   let dirty = ( listChanged.value || programChanged.value || advisorsChanged.value || embargoChanged.value )
   if ( dirty ) return true

   Object.keys(data).some((key) => {
      // these are input fields on the form for entering repeat values, but are not actully part of the form. Ignore
      if ( key == "keyword" || key=="agency" || key=="related" || key.includes("computeID")) return false
      
      if (key == "dirty") {
         if (data[key]) {
            dirty = true
         }
         return dirty
      }
      if (data[key] && typeof data[key] === 'object') {
         dirty = isDirty(data[key])
         return dirty
      }
   })

   return dirty
})

const saveChanges = ( async () => {
   console.log("data has been edited; saving")
   let license = system.licenseDetail(etdRepo.licenseID)
   if (license) {
      etdRepo.work.license = license.label
      etdRepo.work.licenseURL = license.url
   }

   await etdRepo.update( )
   if ( system.showError == false ) {
      saved.value = true
      etdForm.value.reset()
      listChanged.value = false
      programChanged.value = false
      advisorsChanged.value = false
      embargoChanged.value = false
      etdForm.value.validate()
   } else {
      return
   }
})

const advisorsUpdated = ( () => {
   advisorsChanged.value = true
   etdForm.value.validate()
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