<template>
   <div class="work-bkg"></div>
   <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Thesis</p>" />
   <div v-else class="public-work">
      <div v-if="etdRepo.error" class="error">
         <h2>System Error</h2>
         <p>Sorry, a system error has occurred!</p>
         <p>{{ etdRepo.error }}</p>
      </div>
      <template v-else>

         <div class="files">
            <div class="title">Files</div>
            <template  v-if="etdRepo.visibility == 'embargo'">
               <span class="file-embargo">
               This item is restricted to abstract view only until {{ $formatDate(etdRepo.embargoReleaseDate) }}.
               </span>
               <span  v-if="etdRepo.isDraft" class="file-embargo author">
                  The files listed below will NOT be available to anyone until the embargo date has passed.
               </span>
            </template>
            <span  v-if="etdRepo.visibility == 'uva'" class="file-embargo">
               This item is restricted to UVA until {{ $formatDate(etdRepo.embargoReleaseDate) }}.
            </span>

            <div  v-if="etdRepo.isDraft || etdRepo.visibility != 'embargo'" class="file" v-for="file in etdRepo.work.files">
               <div class="name">{{ file.name }}</div>
               <div class="upload"><label>Uploaded:</label>{{ $formatDate(file.createdAt) }}</div>
               <Button icon="pi pi-cloud-download" label="Download" severity="secondary" size="small" @click="downloadFileClicked(file.name)"/>
            </div>
         </div>

         <div class="details">
            <div class="title" role="heading">{{ etdRepo.work.title }}</div>
            <Fieldset legend="Author">{{  authorDisplay(etdRepo.work.author) }}</Fieldset>
            <Fieldset legend="Advisors">
               <div v-for="advisor in  etdRepo.work.advisors" class="author">
                  <p>{{ advisorDisplay(advisor) }}</p>
                  <p>{{ advisor.institution }}</p>
               </div>
            </Fieldset>
            <Fieldset legend="Abstract">{{  etdRepo.work.abstract }}</Fieldset>
            <Fieldset legend="Degree">{{  etdRepo.work.degree }}</Fieldset>
            <Fieldset v-if="etdRepo.hasKeywords" legend="Keywords">
               {{ etdRepo.work.keywords.join(", ") }}
            </Fieldset>
            <Fieldset v-if="etdRepo.hasSponsors" legend="Sponsoring Agency">
               <div v-for="s in etdRepo.work.sponsors">{{ s }}</div>
            </Fieldset>
            <Fieldset v-if="etdRepo.hasRelatedURLs" legend="Related Links">
               <ul>
                  <li v-for="url in etdRepo.work.relatedURLs"><a :href="url" target="_blank">{{ url }}</a></li>
               </ul>
            </Fieldset>
            <Fieldset v-if="etdRepo.work.notes" legend="Notes">{{  etdRepo.work.notes }}</Fieldset>
            <Fieldset v-if="etdRepo.work.language" legend="Labguage">{{  etdRepo.work.language }}</Fieldset>
            <Fieldset legend="Rights">
               <a :href="etdRepo.work.licenseURL" target="_blank">{{ etdRepo.work.license }}</a>
            </Fieldset>
            <Fieldset legend="Persistent Link">
               <a v-if="etdRepo.persistentLink" target="_blank" :href="etdRepo.persistentLink">{{ etdRepo.persistentLink }}</a>
               <span v-else>Persistent link will appear here after submission.</span>
            </Fieldset>
         </div>
         <div class="draft" v-if="etdRepo.isDraft">
            <div class="proof">Submission Proof</div>
            <div>
               Before proceeding, we encourage you to review the information in this page.
               If you experience problems with your submission, please <a href="mailto:libra@virginia.edu">contact</a> us.
            </div>
            <div class="buttons">
               <Button severity="secondary" label="Cancel" size="small" @click="cancelPreview"/>
               <Button severity="secondary" label="Edit" size="small" @click="editThesis"/>
               <Button severity="primary" label="Submit Thesis" size="small"  @click="submitThesis"/>
            </div>
         </div>
         <div class="published" v-if="justPublished">
            Thank you for submitting your thesis. Be sure to take note of and use
            the Persistent Link when you refer to this work.
         </div>
      </template>
   </div>
</template>

<script setup>
import { onBeforeMount, ref } from 'vue'
import { useETDStore } from "@/stores/etd"
import { useSystemStore } from "@/stores/system"
import { useRoute, useRouter } from 'vue-router'
import Fieldset from 'primevue/fieldset'
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useConfirm } from "primevue/useconfirm"

const etdRepo = useETDStore()
const system = useSystemStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()

const justPublished = ref(false)

const authorDisplay = ((a) => {
   return `${a.lastName}, ${a.firstName}, ${a.program}, ${a.institution}`
})
const advisorDisplay = ((a) => {
   return `${a.lastName}, ${a.firstName}, ${a.department}`
})
onBeforeMount( async () => {
   document.title = "LibraETD"
   await etdRepo.getWork( route.params.id )
})
const downloadFileClicked = ( (name) => {
   etdRepo.downloadFile(name)
})
const editThesis = (() => {
   router.push(`/etd/${route.params.id}`)
})
const cancelPreview = ( () => {
   router.push("/etd")
})
const submitThesis = ( () => {
   confirm.require({
      message: `This is your final step and you cannot change the document afterwards. Are you sure?`,
      header: 'Confirm Submission',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: async () => {
         await etdRepo.publish()
         if (system.error == "") {
            justPublished.value = true
         }
      },
   })
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   div.public-work {
      display: flex;
      flex-flow: row nowrap;
      justify-content: center;
      align-items: flex-start;
      div.error {
         max-width: 60%;
         margin: 50px auto 0 auto;
         box-shadow: 0 0 15px 5px black;
      }
      div.draft, div.published {
         margin: 20px 20px 0 0;
         max-width: 300px;
      }
      div.details {
         max-width: 60%;
         min-width: 525px;
         padding: 30px;
         margin: 20px;
         border: 1px solid $uva-grey-100;
      }
      div.files {
         width: 250px;
         margin: 320px 0 0 0;
      }
   }
}
@media only screen and (max-width: 768px) {
   div.work-bkg {
      display: none;
   }
   div.public-work {
      display: flex;
      flex-direction: column-reverse;
      div.draft, div.published  {
         margin: 20px auto 0 auto;
         width: 90%;
      }
      div.error {
         max-width: 100%;
         margin: 5px;
      }
      div.details {
         max-width: none;
         padding: 20px;
         border-radius: 0;
         margin: 0;
         ul {
            margin: 0;
            padding: 0 0 0 15px;
         }
      }
      div.files {
         width: 100%;
         margin: 0;
         border-top: 1px solid $uva-grey-100;
         padding: 0;
      }
   }
}

div.work-bkg {
   background-image: url('@/assets/header.jpg');
   background-position: center center;
   background-repeat: no-repeat;
   height: 300px;
   background-size: cover;
   position: absolute;
   left: 0;
   right: 0;
   border-top: 1px solid $uva-grey-A;
   border-bottom: 1px solid $uva-grey-A;
}

div.public-work {
   position: relative;
   min-height: 300px;

   div.published {
      background: $uva-yellow-100;
      padding: 20px;
      border: 1px solid $uva-yellow-A;
      border-radius: 4px;
      text-align: left;
   }

   div.draft {
      background: $uva-yellow-100;
      padding: 20px;
      border: 1px solid $uva-yellow-A;
      border-radius: 0.3rem;
      display: flex;
      flex-direction: column;
      gap: 15px;
      text-align: left;
      .proof {
         font-size: 1.2em;
         font-weight: bold;
         text-align: center;
      }
      .buttons {
         display: flex;
         flex-flow: row wrap;
         justify-content: center;
         gap: 5px;
      }
   }

   div.error {
      border-radius: 5px;
      background-color: white;
      border: 5px solid $uva-red-A;
      padding: 25px;
      p {
         text-align: left;
      }
      h2 {
         margin: 0 0 15px 0 !important;
         padding: 0;
      }
   }

   div.files {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      display: flex;
      flex-direction: column;
      gap: 15px;
      padding: 20px;
      border: 1px solid $uva-grey-100;
      border-radius: 0.3rem;
      .title {
         font-weight: bold;
      }

      .upload {
         display: flex;
         flex-flow: row nowrap;
         gap: 10px;
      }

      .file-embargo {
         padding: 10px;
         font-style: normal;
         background: $uva-yellow-100;
         border: 1px solid $uva-yellow-A;
         border-radius: 4px;
      }
      .file {
         display: flex;
         flex-direction: column;
         gap: 10px;
         padding-top: 15px;
         border-top: 1px solid $uva-grey-100;
      }
   }

   div.details {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      display: flex;
      flex-direction: column;
      gap: 10px;
      border-radius: 0.3rem;

      .title {
         font-size: 25px;
         font-weight: normal;
      }
   }
}
</style>