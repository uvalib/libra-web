<template>
   <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Thesis</p>" />
   <div v-else-if="etdRepo.error" class="error">
      <h1>System Error</h1>
      <div>Sorry, a system error has occurred!</div>
      <div>{{ etdRepo.error }}</div>
   </div>
   <template v-else>
      <div class="work-bkg"></div>
      <div class="public-work">
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
               <div class="upload">Uploaded on {{ $formatDate(file.createdAt) }}</div>
               <Button label="Download" icon="pi pi-cloud-download" severity="secondary" size="small" @click="etdRepo.downloadFile(file.name)" />
            </div>
         </div>

         <div class="details">
            <div class="draft" v-if="etdRepo.isDraft">
               <div class="proof">Submission Proof</div>
               <div>
                  Before proceeding, we encourage you to review the information in this page.
                  If you experience problems with your submission, please <a href="mailto:libra@virginia.edu">contact</a> us.
               </div>
               <div class="agree">
                  <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
                  <label for="agree-cb">
                     I have read and agree to the
                     <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank" aria-describedby="new-window">Libra Deposit License</a>,
                     including discussing my deposit access options with my faculty advisor.
                  </label>
               </div>
               <div class="buttons">
                  <Button severity="secondary" label="Exit" size="small" @click="cancelPreview"/>
                  <Button severity="secondary" label="Edit" size="small" @click="editThesis"/>
                  <Button severity="primary" label="Submit Thesis" size="small"  @click="submitThesis" :disabled="!agree"/>
               </div>
            </div>
            <div class="published" v-if="justPublished">
               Thank you for submitting your thesis. Be sure to take note of and use
               the Persistent Link when you refer to this work.
            </div>

            <div class="metadata">
               <h1>{{ etdRepo.work.title }}</h1>
               <section>
                  <h2>Author</h2>
                  <div class="content">{{  authorDisplay(etdRepo.work.author) }}</div>
               </section>
               <section>
                  <h2>Advisors</h2>
                  <ul class="advisors">
                     <li v-for="advisor in  etdRepo.work.advisors" class="advisor">
                        <span>{{ advisor.lastName }}, {{ advisor.firstName }}</span>
                        <span v-if="advisor.department">,&nbsp;{{ advisor.department }}</span>
                        <span v-if="advisor.institution">,&nbsp;{{ advisor.institution }}</span>
                     </li>
                  </ul>
               </section>
               <section>
                  <h2>Abstract</h2>
                  <div style="white-space: pre-wrap;">{{ etdRepo.work.abstract }}</div>
               </section>
               <section>
                  <h2>Degree</h2>
                  <div>{{ etdRepo.work.degree }}</div>
               </section>
               <section v-if="etdRepo.hasKeywords">
                  <h2>Keywords</h2>
                  <div>{{ etdRepo.work.keywords.join("; ") }}</div>
               </section>
               <section v-if="etdRepo.hasSponsors">
                  <h2>Sponsors</h2>
                  <ul>
                     <li v-for="s in etdRepo.work.sponsors">{{ s }}</li>
                  </ul>
               </section>
               <section v-if="etdRepo.hasRelatedURLs">
                  <h2>Related Links</h2>
                  <ul>
                     <li v-for="url in etdRepo.work.relatedURLs"><a :href="url" target="_blank" aria-describedby="new-window">{{ url }}</a></li>
                  </ul>
               </section>
               <section v-if="etdRepo.work.notes">
                  <h2>Notes</h2>
                  <div style="white-space: pre-wrap;">{{ etdRepo.work.notes }}</div>
               </section>
               <section v-if="etdRepo.work.language">
                  <h2>Language</h2>
                  <div>{{ etdRepo.work.language }}</div>
               </section>
               <section>
                  <h2>Rights</h2>
                  <a v-if="etdRepo.work.licenseURL" :href="etdRepo.work.licenseURL" target="_blank" aria-describedby="new-window">{{ etdRepo.work.license }}</a>
                  <div v-else>{{ etdRepo.work.license }}</div>
               </section>
               <section v-if="!etdRepo.isDraft">
                  <h2>Issued Date:</h2>
                  <div>{{ $formatDate(etdRepo.publishedAt) }}</div>
               </section>
               <section v-if="!etdRepo.isDraft">
                  <h2>Persistent Link</h2>
                  <a v-if="etdRepo.persistentLink" target="_blank" aria-describedby="new-window" :href="etdRepo.persistentLink">{{ etdRepo.persistentLink }}</a>
                  <span v-else>Persistent link will appear here after submission.</span>
               </section>
               <section v-if="!etdRepo.isDraft">
                  <h2>Suggested Citation</h2>
                  <span id="citation">{{ etdRepo.suggestedCitation }}</span>
                  <Button v-if="canClipboard" severity="secondary" size="small" label="Copy citation" @click="copyCitation"/>
               </section>
            </div>
         </div>
      </div>
   </template>
</template>

<script setup>
import { onBeforeMount, ref } from 'vue'
import { useETDStore } from "@/stores/etd"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useRoute, useRouter } from 'vue-router'
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useConfirm } from "primevue/useconfirm"
import Checkbox from 'primevue/checkbox'
import { useClipboard, usePermission } from '@vueuse/core'
import { useSeoMeta } from '@unhead/vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc';
dayjs.extend(utc)

const etdRepo = useETDStore()
const system = useSystemStore()
const user = useUserStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()

const canClipboard = usePermission('clipboard-write')
const { copy, copied } = useClipboard()

const justPublished = ref(false)
const agree = ref(false)

onBeforeMount( async () => {
   await etdRepo.getWork( route.params.id )
})
useSeoMeta({
   citation_title: ()=> etdRepo.work.title,
   citation_author: ()=> etdRepo.work.author ? `${etdRepo.work.author.lastName}, ${etdRepo.work.author.firstName}` : null,
   citation_publication_date: ()=> etdRepo.publishedAt ? dayjs(etdRepo.publishedAt).utc().format("YYYY-MM-DD") : null,
   citation_dissertation_institution: ()=> etdRepo.work.author.institution,
   citation_pdf_url: null, // TODO
   citation_publisher: 'University of Virginia',
   citation_doi: ()=> etdRepo.work.doi,
   citation_keywords: ()=> etdRepo.work.keywords ? etdRepo.work.keywords.filter((value) => value != null && value != "").join('; ') : null,
   citation_dissertation_institution: ()=> etdRepo.work.author ? [etdRepo.work.author.institution, etdRepo.work.program].filter((value) => value != null && value != "").join(", ") : null,
   citation_dissertation_name: ()=> etdRepo.work.degree,
})

const authorDisplay = ((a) => {
   return `${a.lastName}, ${a.firstName}, ${etdRepo.work.program}, ${a.institution}`
})

const copyCitation = (() => {
   copy( etdRepo.suggestedCitation )
   if (copied) {
      system.toastMessage("Copied", "Citation has been copied to the clipboard.")
   }
})

const editThesis = (() => {
   if (user.isAdmin) {
      router.push(`/admin/etd/${route.params.id}`)
   } else {
      router.push(`/etd/${route.params.id}`)
   }
})

const cancelPreview = ( () => {
   router.push(user.homePage)
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
      gap: 1rem;
      padding: 2.5rem 2.5rem 0 1rem;
   }

   .details {
      flex-basis: 70%;
   }

   .files {
      margin-top: 280px;
      flex-basis: 30%;
   }
}

@media only screen and (max-width: 768px) {
   div.work-bkg {
      display: none;
   }
   div.public-work {
      display: flex;
      flex-direction: column-reverse;
   }
}
div.error {
   background-color: white;
   padding: 25px;
   min-height: 300px;
   text-align: center;
   width: 50%;
   margin: 0 auto;

}

.note {
   text-align: center;
   font-style: italic;
   font-style: 0.9em;
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
    h1 {
      font-size: 1.5em;
      font-weight: 400;
      padding: 0;
      margin: 0;
   }
   section {
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      gap: 10px;
   }
   .details {
      border: 1px solid $uva-grey-100;
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      margin-bottom: 50px;

      div.draft {
         background: $uva-yellow-100;
         padding: 20px;
         border: 2px solid $uva-yellow-A;
         display: flex;
         flex-direction: column;
         gap: 10px;
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
         .agree {
            display: flex;
            flex-flow: row nowrap;
            justify-content: center;
            align-items: flex-start;
            gap: 10px;
            margin: 10px 0;
         }
      }

      h3 {
         font-size: 1.5em;
         font-weight: normal;
         padding:0;
         margin:0;
      }

      .metadata {
         padding: 30px;
         display: flex;
         flex-direction: column;
         gap: 1.5rem;
         ul {
            display: flex;
            flex-direction: column;
            gap: 5px;
            margin: 0;
            list-style: none;
            padding: 0 0 0 0px;
         }
         h2 {
            text-align: left;
            font-size: 1.15em;
            padding:0;
            margin:0 0 5px 0;
         }
      }
   }

   div.published {
      background: $uva-yellow-100;
      padding: 20px;
      border: 1px solid $uva-yellow-A;
      text-align: left;
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
}
</style>
