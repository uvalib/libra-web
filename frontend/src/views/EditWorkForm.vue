<template>
   <div class="edit">
      <h1>
         <span>LibraETD Work</span>
         <span v-if="adminEdit==false && etdRepo.isDraft" class="draft">DRAFT</span>
      </h1>
      <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
      <Form v-else v-slot="$form" :initialValues="etdRepo.work" :resolver="resolver" class="sections" ref="etdForm" @submit="submitTest">
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

            <label for="file" class="libra-form-label">Upload Files</label>
            <FileUpload name="file" :url="`/api/upload/${etdRepo.work.id}`"
               @upload="fileUploaded($event)" @before-send="uploadRequested($event)"
               @removeUploadedFile="fileRemoved($event)"
               :multiple="true" :withCredentials="true" :auto="true"
               :showUploadButton="false" :showCancelButton="false">
               <template #empty>
                  <p>Click Choose or drag and drop files to upload. Uploaded files will be attached to the work upon submission.</p>
               </template>
            </FileUpload>
         </Panel>

         <Panel header="License" toggleable>
            <div class="fields">
               <div class="note">
                  Libra lets you choose an open license when you post your work, and will prominently display the
                  license you choose as part of the record for your work. See
                  <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
                  for option details.
               </div>
               <div class="field" >
                  <LabeledInput label="Rights" name="licenseID" :required="true" v-model="etdRepo.licenseID" type="select" :options="system.userLicenses" />
                  <Message v-if="$form.licenseID?.invalid" severity="error" size="small" variant="simple">{{ $form.licenseID.error.message }}</Message>
               </div>
               <div class="agree">
                  <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
                  <label for="agree-cb">
                     I have read and agree to the
                     <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank">Libra Deposit License</a>,
                     including discussing my deposit access options with my faculty advisor.
                  </label>
               </div>
            </div>
         </Panel>
      </Form>
      <div class="toolbar">
         <Button label="Discard changes" severity="danger" @click="router.push('/')" style="margin-right:auto"/>
         <Button label="Save" @click="etdForm.submit()"/>
         <Button label="Save and exit" />
         <Button label="Preview" severity="success" />
      </div>
   </div>
</template>

<script setup>
import { ref, onBeforeMount, computed, nextTick } from 'vue'
import AuditsPanel from '@/components/AuditsPanel.vue'
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

import { Form } from '@primevue/forms'
import { yupResolver } from '@primevue/forms/resolvers/yup'
import * as yup from 'yup'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Fieldset from 'primevue/fieldset'
import LabeledInput from '@/components/LabeledInput.vue'
import RepeatField from '@/components/RepeatField.vue'
import Checkbox from 'primevue/checkbox'

const confirm = useConfirm()
const router = useRouter()
const route = useRoute()
const system = useSystemStore()
const user = useUserStore()
const etdRepo = useETDStore()
const etdForm = ref(null)
const agree = ref(false)

const submitTest = ( (valid ) => {
   console.log("SUBMIT , LICENSE ["+etdRepo.licenseID+"]")
   let license = system.licenseDetail(etdRepo.licenseID)
   if (license) {
      etdRepo.work.license = license.label
      etdRepo.work.licenseURL = license.url
      console.log("SET LIC URL "+license.url)
   }
   console.log(valid)
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
   .sections, h1 {
      margin-left: 5%;
      margin-right: 5%;
   }
}
@media only screen and (max-width: 768px) {
   .sections, h1 {
      margin-left: 15px;
      margin-right: 15px;
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

      .fields {
         display: flex;
         flex-direction: column;
         gap: 15px;
         .field {
            display: flex;
            flex-direction: column;
            gap: 5px;
         }
         .agree {
            display: flex;
            flex-flow: row nowrap;
            justify-content: center;
            align-items: center;
            gap: 15px;
            padding: 20px 0 0 0;
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