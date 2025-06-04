<template>
   <div class="edit">
      <h2>
         <span>LibraETD Work</span>
         <span v-if="etdRepo.isDraft" class="draft">DRAFT</span>
      </h2>
      <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
      <Form v-else v-slot="$form" :initialValues="etdRepo" :resolver="resolver" class="sections" ref="etdForm" @submit="saveChanges" :validateOnBlur="true" :validateOnMount="true">
         <div class="help">View <a target="_blank" href="https://www.library.virginia.edu/libra/etds/etds-checklist">ETD Submission Checklist</a> for help.</div>
         <Panel header="Metadata" toggleable>
            <ProgramPanel :admin="user.admin" @changed="programChanged = true"/>
            <div class="fields">
               <div class="field" >
                  <LabeledInput label="Title" name="work.title" :required="true" v-model="etdRepo.work.title"/>
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

               <Fieldset>
                  <template #legend>
                     <span>Advisors</span><span class="required"><span class="star">*</span>(required)</span>
                  </template>
                  <div v-for="(item, index) in etdRepo.work.advisors" class="advisor">
                     <div v-if="index==0" class="note">Lookup a UVA Computing ID to automatically fill the remaining fields for this advisor.</div>
                     <div class="id-field">
                        <div class="control-group">
                           <InputText type="text" v-model="item.computeID" :name="`work.advisors[${index}].computeID`" placeholder="Computing ID"/>
                           <Button class="check" icon="pi pi-search" label="Lookup Advisor"  severity="secondary" @click="checkAdvisorID(index)"/>
                        </div>
                        <Button v-if="index > 0" icon="pi pi-trash" severity="danger" aria-label="remove advisor" @click="removeAdvisor(index)"/>
                     </div>
                     <Message v-if="etdRepo.work.advisors[index].msg" severity="error" size="small" variant="simple">{{ etdRepo.work.advisors[index].msg }}</Message>
                     <div class="two-col">
                        <div class="field" >
                           <LabeledInput label="First Name" :name="`work.advisors[${index}].firstName`" :required="true" v-model="item.firstName"/>
                           <Message v-if="$form.work?.advisors?.[index]?.firstName?.invalid" severity="error" size="small" variant="simple">{{ $form.work.advisors[index].firstName.error.message }}</Message>
                        </div>
                        <div class="field" >
                           <LabeledInput label="Last Name" :name="`work.advisors[${index}].lastName`" :required="true" v-model="item.lastName"/>
                           <Message v-if="$form.work?.advisors?.[index]?.lastName?.invalid" severity="error" size="small" variant="simple">{{ $form.work.advisors[index].lastName.error.message }}</Message>
                        </div>
                     </div>
                     <div class="two-col">
                        <div class="field" >
                           <LabeledInput label="Department" :name="`work.advisors[${index}].department`" v-model="item.department"/>
                        </div>
                        <div class="field" >
                           <LabeledInput label="Institution" :name="`work.advisors[${index}].institution`" v-model="item.institution"/>
                        </div>
                     </div>
                  </div>
                  <div class="acts">
                     <Button label="Add Advisor" size="small" @click="addAdvisor"/>
                  </div>
               </Fieldset>

               <div class="field" >
                  <LabeledInput label="Abstract" name="work.abstract" :required="true" v-model="etdRepo.work.abstract" type="textarea" />
                  <Message v-if="$form.work?.abstract?.invalid" severity="error" size="small" variant="simple">{{ $form.work.abstract.error.message }}</Message>
               </div>

               <RepeatField label="Keywords" name="keyword" @change="listChanged=true" help="Add one keyword or keyword phrase per line" v-model="etdRepo.work.keywords" />
               <LabeledInput label="Language" name="work.language" v-model="etdRepo.work.language" type="select" :options="system.languages" />
               <RepeatField label="Related Links" @change="listChanged=true" name="related"
                  help="A link to a website or other specific content (audio, video, PDF document) related to the work"
                  v-model="etdRepo.work.relatedURLs"
               />
               <RepeatField label="Sponsoring Agencies" name="agency" @change="listChanged=true" v-model="etdRepo.work.sponsors"/>
               <LabeledInput label="Notes" name="work.notes" v-model="etdRepo.work.notes" type="textarea" />
               <LabeledInput v-if="user.admin" label="Admin Notes" name="work.adminNotes" v-model="etdRepo.work.adminNotes" type="textarea" />
               <div class="field" >
                  <LabeledInput label="Rights" name="licenseID" :required="true" v-model="etdRepo.licenseID" type="select" :options="system.userLicenses" />
                  <Message v-if="$form.licenseID?.invalid" severity="error" size="small" variant="simple">{{ $form.licenseID.error.message }}</Message>
                  <div class="note">
                     Libra lets you choose an open license when you post your work, and will prominently display the
                     license you choose as part of the record for your work. See
                     <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
                     for option details.
                  </div>
               </div>
            </div>
            <template #icons>
               <i v-if="metadataComplete" class="complete pi pi-check-circle"></i>
               <i v-else class="incomplete pi pi-exclamation-circle"></i>
            </template>
         </Panel>

         <Panel header="Files" toggleable>
            <FilesPanel />
            <template #icons>
               <i v-if="etdRepo.hasFiles" class="complete pi pi-check-circle"></i>
               <i v-else class="incomplete pi pi-exclamation-circle"></i>
            </template>
         </Panel>

         <Panel header="Access and Visibility" toggleable>
            <div class="license">
               <div class="note">
                  For more information, see the
                  <a href="https://uvapolicy.virginia.edu/policy/PROV-014" target="_blank">Provost Policy on Access Levels for Libra ETD deposits</a>.
               </div>

               <div v-if="etdRepo.visibility == 'embargo' && user.admin == false" class="embargo">
                  <!-- ETD can only be embargoed by an admin. When this happens, lock out the visibility for the user with a message -->
                  <div>This work is under embargo.</div>
                  <div>Files will NOT be available to anyone until {{ $formatDate(etdRepo.embargoReleaseDate) }}.</div>
               </div>
               <div v-else v-for="v in visibilityOpts" :key="v.value" class="visibility-opt">
                  <RadioButton v-model="etdRepo.visibility" name="visibility" :inputId="v.value"  :value="v.value" size="large" @update:model-value="visibilityUpdated"/>
                  <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
               </div>
               <Message v-if="$form.visibility?.invalid" severity="error" size="small" variant="simple">{{ $form.visibility.error.message }}</Message>
               <div v-if="etdRepo.visibility == 'uva' || (user.admin && etdRepo.visibility == 'embargo')" class="visibility-info">
                  <div v-if="etdRepo.visibility == 'uva'">Files available to UVA only until:</div>
                  <div v-else>Files unavailable to anyone until:</div>
                  <div class="embargo-date">
                     <span v-if="etdRepo.embargoReleaseDate">{{ $formatDate(etdRepo.embargoReleaseDate) }}</span>
                     <span v-else>Never</span>
                     <DatePickerDialog :endDate="etdRepo.embargoReleaseDate" :admin="user.admin"
                        :visibility="etdRepo.visibility" @picked="endDatePicked"
                        :degree="etdRepo.work.degree" :program="etdRepo.work.program" />
                  </div>
                  <div>After that, files will be be available worldwide.</div>
               </div>
               <div v-else class="visibility-info">
                  All files will be available worldwide.
               </div>
            </div>

            <template #icons>
               <i v-if="etdRepo.visibility != ''" class="complete pi pi-check-circle"></i>
               <i v-else class="incomplete pi pi-exclamation-circle"></i>
            </template>
         </Panel>
      </Form>
      <div class="toolbar">
         <span class="group">
            <template v-if="user.admin">
               <Button :disabled="!etdRepo.publishedAt" label="Unpublish" severity="danger" @click="unpublishClicked" />
               <Button :disabled="etdRepo.publishedAt" label="Delete" severity="danger" @click="deleteClicked" />
            </template>
            <Button label="Exit" severity="secondary" @click="exitClicked" />
         </span>
         <span class="unsaved" v-if="needsSave">UNSAVED EDITS</span>
         <span class="group">
            <Button label="Save" @click="saveClicked('edit')" :loading="etdRepo.saving" :disabled="needsSave==false"/>
            <Button label="Preview" severity="success" @click="previewClicked" :disabled="needsSave || metadataComplete==false || etdRepo.hasFiles==false"/>
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
import WaitSpinner from "@/components/WaitSpinner.vue"
import axios from 'axios'
import { useRouter, useRoute, onBeforeRouteLeave } from 'vue-router'

import { Form } from '@primevue/forms'
import ProgramPanel from '@/components/ProgramPanel.vue'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Fieldset from 'primevue/fieldset'
import LabeledInput from '@/components/LabeledInput.vue'
import RepeatField from '@/components/RepeatField.vue'
import DatePickerDialog from "@/components/DatePickerDialog.vue"
import FilesPanel from '@/components/FilesPanel.vue'
import RadioButton from 'primevue/radiobutton'
import { useConfirm } from "primevue/useconfirm"

const confirm = useConfirm()
const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const etdRepo = useETDStore()
const admin = useAdminStore()

const etdForm = ref(null)
const postSave = ref("edit")
const listChanged = ref(false)
const programChanged = ref(false)
const metadataComplete = ref(false)

onBeforeMount( async () => {
   document.title = "LibraETD"
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   await etdRepo.getWork( route.params.id )
})

onBeforeRouteLeave(() => {
   if (needsSave.value) {
      const exit = window.confirm('You have unsaved changes that will be lost if you return to the dashboard. Are you sure?')
      if (!exit) return false
   }
})

const needsSave  = computed( () => {
   if ( !etdForm.value) return false
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

   values.work.advisors.forEach( (a,idx) => {
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

   if ( values.work.abstract == "" ) {
      metadataComplete.value = false
      errors.work.abstract = [{ message: 'Abstract is required' }]
   }

   let licID = parseInt(values.licenseID)
   if ( licID == 0 ) {
      metadataComplete.value = false
      errors.licenseID = [{ message: 'Rights are required' }]
   }

   if ( !values.visibility || values.visibility == "") {
      errors.visibility = [{ message: 'Visibility is required' }]
   }

   return { values,errors }
}

const saveClicked = ((postSaveAct) => {
   postSave.value = postSaveAct
   etdForm.value.submit()
})

const previewClicked = (() => {
   router.push({ name: 'etdpublic', params: { id: etdRepo.work.computeID } })
})

const exitClicked = (() => {
   let exitRoute = "/"
   if ( user.admin ) {
      exitRoute = "/admin"
   }
   if ( isDirty(etdForm.value.states) ) {
      confirm.require({
         message: "Any unsaved changes will be lost if you exit. Are you sure?",
         header: 'Confirm Exit',
         icon: 'pi pi-question-circle',
         rejectClass: 'p-button-secondary',
         accept: (  ) => {
            clearEdits()
            router.push(exitRoute)
         },
      })
   } else {
      router.push(exitRoute)
   }
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
         router.push("/admin")
      },
   })
})

const isDirty = ((data) => {
   let dirty = ( etdRepo.pendingFileAdd.length > 0 || etdRepo.pendingFileDel.length > 0 || listChanged.value || programChanged.value)
   if (dirty ) return true

   Object.keys(data).some((key) => {
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

const clearEdits = (() => {
   etdForm.value.reset()
   listChanged.value = false
   programChanged.value = false
})

const saveChanges = ( async (data) => {
   if ( isDirty( data.states ) ) {
      console.log("data has been edited; saving")
      let license = system.licenseDetail(etdRepo.licenseID)
      if (license) {
         etdRepo.work.license = license.label
         etdRepo.work.licenseURL = license.url
      }

      await etdRepo.update( )
      if ( system.showError == false ) {
         system.toastMessage("Saved", "All changes have been saved")
         clearEdits()
      } else {
         return
      }
   }
})

const visibilityOpts = computed( () => {
   if (user.admin) {
      return system.visibility
   }
   return system.userVisibility
})

const addAdvisor = ( () => {
   etdRepo.work.advisors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})

const removeAdvisor = ((idx)=> {
   etdRepo.work.advisors.splice(idx,1)
   etdForm.value.validate()
})

const checkAdvisorID = ((idx) => {
   let cID = etdRepo.work.advisors[idx].computeID
   etdRepo.work.advisors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let department = ""
      if ( r.data.department && r.data.department.length > 0 ) {
         department = r.data.department[0]
      }
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: department, institution: "University of Virginia"}
      etdRepo.work.advisors.splice(idx,1, auth)
      // set firs/last name in the form state data so validation works
      etdForm.value.setFieldValue(`work.advisors[${idx}].firstName`, r.data.first_name)
      etdForm.value.setFieldValue(`work.advisors[${idx}].lastName`, r.data.last_name)
   }).catch( () => {
      etdRepo.work.advisors[idx].msg = cID+" is not a valid computing ID"
   })
})

const visibilityUpdated = (() => {
   if (etdRepo.visibility == "embargo" || etdRepo.visibility == "uva") {
      etdRepo.embargoReleaseVisibility = "open"
      let endDate = new Date()
      endDate.setMonth( endDate.getMonth() + 6)
      etdRepo.embargoReleaseDate = endDate.toJSON()
   }
})

const endDatePicked = ( (newDate) => {
   etdRepo.embargoReleaseDate = newDate
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .sections, h2 {
      margin-left: 5%;
      margin-right: 5%;
   }
   .visibility-panel {
      min-width: 375px;
   }
}
@media only screen and (max-width: 768px) {
   .sections, h2 {
      margin-left: 15px;
      margin-right: 15px;
   }
}

.unsaved {
   background: $uva-red;
   padding: 0.5rem 1rem;
   border-radius: 50px;
   color: white;
   border: 1px solid $uva-red-B;
   font-weight: bold;
}
.edit {
   text-align: left;
   position: relative;

   h2 {
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

   .complete {
      font-size: 1.25rem;
      color: $uva-green-A;
   }
   .incomplete {
      font-size: 1.25rem;
      color: $uva-red-A;
   }

   .sections {
      display: flex;
      flex-direction: column;
      gap: 25px;

      .advisor {
         border: 1px solid $uva-grey-100;
         padding: 10px;
         display: flex;
         flex-direction: column;
         gap: 10px;
         border-radius: 0.3rem;
      }

      .license {
         display: flex;
         flex-direction: column;
         gap: 10px;

         .visibility-opt {
            display: flex;
            flex-flow: row nowrap;
            gap: 15px;
            align-items: center;
            .visibility {
               width: 200px;
            }
         }
         .visibility-info {
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            gap: 10px;
            margin-top: 15px;
            .embargo-date {
               span {
                  margin-right: 20px;
               }
            }
         }
      }

      .fields {
         display: flex;
         flex-direction: column;
         gap: 15px;
         .field {
            display: flex;
            flex-direction: column;
            gap: 5px;
         }
         .id-field {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
         }
      }
      .control-group {
         display: flex;
         flex-flow: row nowrap;
         gap: 5px;
      }
      .note {
         font-style: italic;
         color: $uva-grey-A;
         margin-top: 0;
      }
      .acts {
         text-align: right;
         margin-top: 10px;
      }
   }

   .toolbar {
      background: $uva-grey-100;
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

   .two-col {
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      align-items: flex-start;
      gap: 25px;
      .field {
         flex-grow: 1;
      }
   }
}
</style>