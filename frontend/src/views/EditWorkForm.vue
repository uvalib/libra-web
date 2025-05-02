<template>
   <div class="edit">
      <h1>
         <span>LibraETD Work</span>
         <span v-if="adminEdit==false && etdRepo.isDraft" class="draft">DRAFT</span>
      </h1>
      <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
      <Form v-else v-slot="$form" :initialValues="etdRepo.work" :resolver="resolver" class="sections" ref="etdForm" @submit="saveChanges">
         <Panel header="Metadata" toggleable>
            <div class="two-col">
               <table>
                  <tbody>
                     <tr>
                        <td class="label">Institution:</td><td>{{ etdRepo.work.author.institution  }}</td>
                     </tr>
                     <tr>
                        <td class="label">Program:</td><td>{{ etdRepo.work.program  }}</td>
                     </tr>
                     <tr>
                        <td class="label">Degree:</td><td>{{ etdRepo.work.degree }}</td>
                     </tr>
                  </tbody>
               </table>
               <table>
                  <tbody>
                     <tr>
                        <td class="label">Date Created:</td><td>{{ $formatDate(etdRepo.createdAt) }}</td>
                     </tr>
                     <tr v-if="etdRepo.isDraft==false">
                        <td class="label">Date Published:</td><td>{{ $formatDate(etdRepo.publishedAt) }}</td>
                     </tr>
                     <tr>
                        <td></td><td><AuditsPanel :workID="etdRepo.work.id"/></td>
                     </tr>
                  </tbody>
               </table>
            </div>
            <div class="fields">
               <div class="field" >
                  <LabeledInput label="Title" name="title" :required="true" v-model="etdRepo.work.title"/>
                  <Message v-if="$form.title?.invalid" severity="error" size="small" variant="simple">{{ $form.title.error.message }}</Message>
               </div>
               <Fieldset legend="Author">
                  <div class="two-col">
                     <div class="field" >
                        <LabeledInput label="First Name" name="author.firstName" :required="true" v-model="etdRepo.work.author.firstName"/>
                        <Message v-if="$form.author?.firstName?.invalid" severity="error" size="small" variant="simple">{{ $form.author.firstName.error.message }}</Message>
                     </div>
                     <div class="field" >
                        <LabeledInput label="Last Name" name="author.lastName" :required="true" v-model="etdRepo.work.author.lastName"/>
                        <Message v-if="$form.author?.lastName?.invalid" severity="error" size="small" variant="simple">{{ $form.author.lastName.error.message }}</Message>
                     </div>
                  </div>
               </Fieldset>
               <Fieldset>
                  <template #legend>
                     <span>Advisors</span><span class="libra-required"><span class="star">*</span>(one or more required)</span>
                  </template>
                  <div class="note">Lookup a UVA Computing ID to automatically fill the remaining fields for this advisor.</div>
                  <div v-for="(item, index) in etdRepo.work.advisors" class="advisor">
                     <div class="id-field">
                        <div class="control-group">
                           <InputText type="text" v-model="item.computeID" :name="`advisors[${index}].computeID`" placeholder="Computing ID"/>
                           <Button class="check" icon="pi pi-search" severity="secondary" @click="checkAdvisorID(index)"/>
                        </div>
                        <Button v-if="index > 0" icon="pi pi-trash" severity="danger" aria-label="remove advisor" @click="removeAdvisor(index)"/>
                     </div>
                     <Message v-if="etdRepo.work.advisors[index].msg" severity="error" size="small" variant="simple">{{ etdRepo.work.advisors[index].msg }}</Message>
                     <div class="two-col">
                        <div class="field" >
                           <LabeledInput label="First Name" :name="`advisors[${index}].firstName`" :required="true" v-model="item.firstName"/>
                           <Message v-if="$form.advisors?.[index]?.firstName?.invalid" severity="error" size="small" variant="simple">{{ $form.advisors[index].firstName.error.message }}</Message>
                        </div>
                        <div class="field" >
                           <LabeledInput label="Last Name" :name="`advisors[${index}].lastName`" :required="true" v-model="item.lastName"/>
                           <Message v-if="$form.advisors?.[index]?.lastName?.invalid" severity="error" size="small" variant="simple">{{ $form.advisors[index].lastName.error.message }}</Message>
                        </div>
                     </div>
                     <div class="two-col">
                        <div class="field" >
                           <LabeledInput label="Department" :name="`advisors[${index}].department`" v-model="item.department"/>
                        </div>
                        <div class="field" >
                           <LabeledInput label="Institution" :name="`advisors[${index}].institution`" v-model="item.institution"/>
                        </div>
                     </div>
                  </div>
                  <div class="acts">
                     <Button label="Add Advisor" size="small" @click="addAdvisor"/>
                  </div>
               </Fieldset>

               <div class="field" >
                  <LabeledInput label="Abstract" name="abstract" :required="true" v-model="etdRepo.work.abstract" type="textarea" />
                  <Message v-if="$form.abstract?.invalid" severity="error" size="small" variant="simple">{{ $form.abstract.error.message }}</Message>
               </div>
               <RepeatField label="Keywords" help="Add one keyword or keyword phrase per line" v-model="etdRepo.work.keywords"/>
               <LabeledInput label="Language" name="notes" v-model="etdRepo.work.language" type="select" :options="system.languages" />
               <RepeatField label="Related Links" help="A link to a website or other specific content (audio, video, PDF document) related to the work" v-model="etdRepo.work.relatedURLs"/>
               <RepeatField label="Sponsoring Agencies" v-model="etdRepo.work.sponsors"/>
               <LabeledInput label="Notes" name="language" v-model="etdRepo.work.notes" type="textarea" />
            </div>
         </Panel>

         <Panel header="Files" toggleable>
            <FilesPanel />
            <template #icons>
               <i v-if="etdRepo.hasFiles" class="complete pi pi-check-circle"></i>
               <i v-else class="incomplete pi pi-exclamation-circle"></i>
            </template>
         </Panel>

         <Panel header="License" toggleable>
            <div class="license-content">
               <Fieldset legend="Visibility" class="visibility-panel">
                  <div v-if="etdRepo.visibility == 'embargo' && adminEdit == false" class="embargo">
                     <!-- ETD can only be embargoed by an admin. When this happens, lock out the visibility for the user with a message -->
                     <div>This work is under embargo.</div>
                     <div>Files will NOT be available to anyone until {{ $formatDate(etdRepo.embargoReleaseDate) }}.</div>
                  </div>
                  <div v-else v-for="v in visibilityOpts" :key="v.value" class="visibility-opt">
                     <RadioButton v-model="etdRepo.visibility" :inputId="v.value"  :value="v.value" @update:model-value="visibilityUpdated"/>
                     <label :for="v.value" class="visibility" :class="v.value">{{ v.label }}</label>
                  </div>
                  <div v-if="etdRepo.visibility == 'uva' || (adminEdit && etdRepo.visibility == 'embargo')" class="limited">
                     <div v-if="etdRepo.visibility == 'uva'">Files available to UVA only until:</div>
                     <div v-else>Files unavailable to anyone until:</div>
                     <div class="embargo-date">
                        <span v-if="etdRepo.embargoReleaseDate">{{ $formatDate(etdRepo.embargoReleaseDate) }}</span>
                        <span v-else>Never</span>
                        <DatePickerDialog :endDate="etdRepo.embargoReleaseDate" :admin="adminEdit"
                           :visibility="etdRepo.visibility" @picked="endDatePicked"
                           :degree="etdRepo.work.degree" :program="etdRepo.work.program" />
                     </div>
                     <div>After that, files will be be available worldwide.</div>
                  </div>
               </Fieldset>
               <div class="license" >
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
            <div class="agree" v-if="adminEdit == false">
               <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
               <label for="agree-cb">
                  I have read and agree to the
                  <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank">Libra Deposit License</a>,
                  including discussing my deposit access options with my faculty advisor.
               </label>
            </div>
            <template #icons>
               <i v-if="etdRepo.hasLicense" class="complete pi pi-check-circle"></i>
               <i v-else class="incomplete pi pi-exclamation-circle"></i>
            </template>
         </Panel>
      </Form>
      <div class="toolbar">
         <Button label="Discard changes" severity="danger" @click="router.push('/')" style="margin-right:auto"/>
         <Button label="Save" @click="saveClicked('edit')"/>
         <Button label="Save and exit" @click="saveClicked('exit')"/>
         <Button label="Preview" severity="success" @click="saveClicked('preview')"/>
      </div>
   </div>
</template>

<script setup>
import { ref, onBeforeMount, computed } from 'vue'
import AuditsPanel from '@/components/AuditsPanel.vue'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useETDStore } from "@/stores/etd"
import Panel from 'primevue/panel'
import WaitSpinner from "@/components/WaitSpinner.vue"
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'

import { Form } from '@primevue/forms'
import { yupResolver } from '@primevue/forms/resolvers/yup'
import * as yup from 'yup'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Fieldset from 'primevue/fieldset'
import LabeledInput from '@/components/LabeledInput.vue'
import RepeatField from '@/components/RepeatField.vue'
import Checkbox from 'primevue/checkbox'
import DatePickerDialog from "@/components/DatePickerDialog.vue"
import RadioButton from 'primevue/radiobutton'
import FilesPanel from '@/components/FilesPanel.vue'

const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const etdRepo = useETDStore()
const etdForm = ref(null)
const agree = ref(false)
const postSave = ref("edit")

const saveClicked = ((postSaveAct) => {
   postSave.value = postSaveAct
   etdForm.value.submit()
   // etdForm.value.validate() // manual calls
   // console.log(etdForm.value.states) // grab state data
})

const saveChanges = ( async (valid ) => {
   let license = system.licenseDetail(etdRepo.licenseID)
   if (license) {
      etdRepo.work.license = license.label
      etdRepo.work.licenseURL = license.url
   }
   console.log(valid)
   await etdRepo.update( )
   if ( system.showError == false ) {
      system.toastMessage("Saved", "All changes have been saved")
      if ( postSave.value == "exit") {
         router.push("/")
      } else if ( postSave.value == "preview") {
         router.push({ name: 'etdpublic', params: { id: etdRepo.work.computeID } })
      }
   }
})

const resolver = ref(
    yupResolver(
        yup.object().shape({
            title: yup.string().required('Title is required'),
            author:  yup.object({
               firstName: yup.string().required('First name is required'),
               lastName: yup.string().required('Last name is required')
            }),
            advisors: yup.array(
               yup.object({
                  firstName: yup.string().required('Advisor first name is required'),
                  lastName: yup.string().required('Advisor last name is required')
               })
            ),
            abstract: yup.string().required('Abstract is required'),
            licenseID: yup.string().required('Rights are required'),
        })
    )
)

const visibilityOpts = computed( () => {
   if (adminEdit.value) {
      return system.visibility
   }
   return system.userVisibility
})

const adminEdit = computed( () => {
   return route.path.includes("/admin")
})

onBeforeMount( async () => {
   document.title = "LibraETD"
   if ( user.isSignedIn == false) {
      router.push("/forbidden")
      return
   }
   await etdRepo.getWork( route.params.id )
})
const addAdvisor = ( () => {
   etdRepo.work.advisors.push({computeID: "", firstName: "", lastName: "", department: "", institution: ""})
})
const removeAdvisor = ((idx)=> {
   etdRepo.work.advisors.splice(idx,1)
})
const checkAdvisorID = ((idx) => {
   let cID = etdRepo.work.advisors[idx].computeID
   etdRepo.work.advisors[idx].msg = ""
   axios.get(`/api/users/lookup/${cID}`).then(r => {
      let auth = {computeID: r.data.cid, firstName: r.data.first_name, lastName: r.data.last_name, department: r.data.department[0], institution: "University of Virginia"}
      etdRepo.work.advisors.splice(idx,1, auth)
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
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .sections, h1 {
      margin-left: 5%;
      margin-right: 5%;
   }
   .visibility-panel {
      min-width: 375px;
   }
   .license {
      max-width: 50%;
   }
}
@media only screen and (max-width: 768px) {
   .sections, h1 {
      margin-left: 15px;
      margin-right: 15px;
   }
   .visibility-panel, .license {
     flex-grow: 1;
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
      table {
         td.label {
            font-weight: bold;
            text-align: right;
            padding-right: 10px;
         }
         margin-bottom: 25px;
      }

      .advisor {
         border: 1px solid $uva-grey-100;
         padding: 10px;
         display: flex;
         flex-direction: column;
         gap: 10px;
         border-radius: 0.3rem;
      }

      .license-content {
         display: flex;
         flex-flow: row wrap;
         gap: 25px;
         .license {
            display: flex;
            flex-direction: column;
            gap: 10px;
         }
         .visibility-panel {
            .p-fieldset-content {
               display: flex;
               flex-direction: column;
               gap: 15px;
               .visibility-opt {
                  display: flex;
                  flex-flow: row nowrap;
                  gap: 15px;
                  align-items: center;
                  .visibility {
                     flex-grow: 1;
                  }
               }
               .limited {
                  display: flex;
                  flex-direction: column;
                  align-items: center;
                  gap: 5px;
                  margin-top: 15px;
                  .embargo-date {
                     display: flex;
                     flex-flow: row nowrap;
                     align-items: baseline;
                     gap: 20px;
                  }
               }
            }
         }
      }

      .agree {
         display: flex;
         flex-flow: row nowrap;
         justify-content: center;
         align-items: center;
         gap: 15px;
         padding: 20px 0 0 0;
         margin-top: 15px;
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
         color: $uva-grey;
         margin-top: 0;
      }
      .acts {
         text-align: right;
         margin-top: 10px;
      }
   }

   .toolbar {
      background: $uva-grey-200;
      border-top: 2px solid $uva-grey-100;
      padding: 15px;
      margin-top: 50px;
      position: sticky;
      bottom: 0;
      display: flex;
      flex-flow: row wrap;
      gap: 5px;
      justify-content: flex-end;
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