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
            <Fieldset legend="Files">
               <div class="file" v-for="file in etdRepo.work.files">
                  <div>{{ file.name }}</div>
                  <div><label>Uploaded:</label>{{ $formatDate(file.createdAt) }}</div>
               </div>
            </Fieldset>
         </div>
         <div class="details">
            <div class="title" role="heading">{{ etdRepo.work.title }}</div>
            <Fieldset legend="Author:">{{  authorDisplay(etdRepo.work.author) }}</Fieldset>
            <Fieldset legend="Advisors:">
               <div v-for="advisor in  etdRepo.work.advisors" class="author">
                  <p>{{ advisorDisplay(advisor) }}</p>
                  <p>{{ advisor.institution }}</p>
               </div>
            </Fieldset>
            <Fieldset legend="Abstract:">{{  etdRepo.work.abstract }}</Fieldset>
            <Fieldset legend="Degree:">{{  etdRepo.work.degree }}</Fieldset>
            <Fieldset v-if="etdRepo.hasKeywords" legend="Keywords:">
               {{ etdRepo.work.keywords.join(", ") }}
            </Fieldset>
            <Fieldset v-if="etdRepo.hasSponsors" legend="Sponsoring Agency:">
               <div v-for="s in etdRepo.work.sponsors">{{ s }}</div>
            </Fieldset>
            <Fieldset v-if="etdRepo.hasRelatedURLs" legend="Related Links:">
               <ul>
                  <li v-for="url in etdRepo.work.relatedURLs"><a :href="url" target="_blank">{{ url }}</a></li>
               </ul>
            </Fieldset>
            <Fieldset v-if="etdRepo.work.notes" legend="Notes:">{{  etdRepo.work.notes }}</Fieldset>
            <Fieldset v-if="etdRepo.work.language" legend="Labguage:">{{  etdRepo.work.language }}</Fieldset>
            <Fieldset legend="Rights:">
               <a :href="etdRepo.work.licenseURL" target="_blank">{{ etdRepo.work.license }}</a>
            </Fieldset>
            <Fieldset legend="Persistent Link:">
               <a v-if="etdRepo.persistentLink" target="_blank" :href="etdRepo.persistentLink">{{ etdRepo.persistentLink }}</a>
               <span v-else>Persistent link will appear here after submission.</span>
            </Fieldset>
         </div>
         <div class="draft" v-if="etdRepo.isDraft">
            <h2 class="proof">Submission Proof</h2>
            <p>
               Before proceeding, we encourage you to review the information in this page.
               If you experience problems with your submission, please <a href="mailto:libra@virginia.edu">contact</a> us.
            </p>
            <div class="buttons">
               <Button severity="secondary" label="Edit" @click="editThesis()"/>
               <Button severity="primary" label="Submit Thesis" icon="pi pi-check" @click="submitThesis()"/>
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
const editThesis = (() => {
   router.push(`/etd/${route.params.id}`)

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
         border: 1px solid var(--uvalib-grey-light);
         box-shadow: 0 0 2px #b9b9b9;
         .author-header {
            .type {
               font-size: 0.85em;
               font-weight: normal;
               padding: 4px 10px;
            }
         }
      }
      div.files {
         width: 250px;
         margin: 320px 0 0 15px;
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
         border-radius: 3px;
         margin: 10px;
         .author-header {
            .type {
               font-size: 0.75em;
               font-weight: bold;
               padding: 2px 8px;
            }
         }
         ul {
            margin: 0;
            padding: 0 0 0 15px;
         }
      }
      div.files {
         width: 100%;
         margin-top: 10px;
         border-top: 1px solid var(--uvalib-grey-light);
         padding: 20px;
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
}
div.public-work {
   position: relative;
   min-height: 300px;

   div.published {
      background: var(--uvalib-yellow-light);
      padding: 20px;
      border: 1px solid var(--uvalib-yellow-dark);
      border-radius: 4px;
      text-align: left;
   }

   div.draft {
      background: var(--uvalib-yellow-light);
      padding: 0 20px 20px 20px;
      border: 1px solid var(--uvalib-yellow-dark);
      border-radius: 4px;
      h2.proof {
         padding: 0;
         margin: 20px 0 10px 0 !important;
         font-weight: normal !important;
      }
      p {
         text-align: left;
      }
      .buttons {
         button {
            margin-left: 5px;
            font-size: 0.9em;
         }
      }
   }

   div.error {
      border-radius: 5px;
      background-color: white;
      border: 5px solid var(--uvalib-red-dark);
      padding: 25px;
      p {
         text-align: left;
      }
      h2 {
         margin: 0 0 15px 0 !important;
         padding: 0;
      }
   }

   fieldset.p-fieldset {
      border: none;
      padding: 0;
      :deep(legend.p-fieldset-legend) {
         font-weight: bold;
         padding: 0;
      }
      :deep(div.p-fieldset-content) {
         padding: 5px;
      }
   }

   div.files {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      .file {
         margin-left: 10px;
         div {
            margin-bottom: 5px;
         }
         label {
            font-weight: bold;
            margin-right: 5px;
         }
      }
   }

   div.details {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      border-radius: 3px;

      .advisor-fieldset {
         :deep(legend.p-fieldset-legend) {
            width: 100%;
         }
         .author-header {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
            .legend {
               font-weight: bold;
            }
            .type {
               border-radius: 5px;
               background-color: var(--uvalib-grey-dark);
               color: white;
            }
         }
      }

      .title {
         color: var(--uvalib-text);
         font-size: 25px;
         font-weight: normal;
         margin-bottom: 20px;
      }

      .author {
         margin-bottom: 5px;
         font-size: 14px;
         p {
            margin: 0;
            padding: 0;
         }
      }
   }
}
</style>