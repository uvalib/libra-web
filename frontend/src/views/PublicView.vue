<template>
   <WaitSpinner v-if="etdRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Thesis</p>" />
   <div v-else-if="etdRepo.error" class="error">
      <h1>System Error</h1>
      <div>Sorry, a system error has occurred!</div>
      <div>{{ etdRepo.error }}</div>
   </div>
   <template v-else>
      <div class="work-bkg"></div>
      <div class="public-work" aria-live="polite">
         <div class="details">
            <SubmitPanel />

            <div class="metadata">
               <h1>
                  <span v-html="etdRepo.work.title"></span>
                  <span class="view-cnt">{{ etdRepo.work.views }} views</span>
               </h1>
               <section>
                  <h2>Author</h2>
                  <div class="content author">
                     <span>{{ authorDisplay(etdRepo.work.author) }}</span>
                     <a v-if="etdRepo.work.author.orcid" class="orcid" :href="etdRepo.work.author.orcid" target="_blank" aria-describedby="new-window">
                        <img class="orcid-img" src="@/assets/orcid_id.svg" aria-hidden="true"/>
                        <span>{{ etdRepo.work.author.orcid.split("/").pop() }}</span>
                     </a>
                  </div>
               </section>
               <section v-if="etdRepo.work.advisors.length >0">
                  <h2>Advisors</h2>
                  <ul class="unstyled">
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
                  <ul class="unstyled">
                     <li v-for="s in etdRepo.work.sponsors">{{ s }}</li>
                  </ul>
               </section>
               <section v-if="etdRepo.hasRelatedURLs">
                  <h2>Related Links</h2>
                  <ul class="links">
                     <li v-for="url in etdRepo.work.relatedURLs">
                        <span v-html="extractLink(url)"/>
                     </li>
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
         <ThesisFiles />
      </div>
   </template>
</template>

<script setup>
import { onBeforeMount } from 'vue'
import { useETDStore } from "@/stores/etd"
import { useSystemStore } from "@/stores/system"
import { useRoute } from 'vue-router'
import WaitSpinner from "@/components/WaitSpinner.vue"
import ThesisFiles from "@/components/publicview/ThesisFiles.vue"
import SubmitPanel from "@/components/publicview/SubmitPanel.vue"
import { useClipboard, usePermission } from '@vueuse/core'
import { useSeoMeta } from '@unhead/vue'
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc';
dayjs.extend(utc)

const etdRepo = useETDStore()
const system = useSystemStore()
const route = useRoute()

const canClipboard = usePermission('clipboard-write')
const { copy, copied } = useClipboard()

onBeforeMount( async () => {
   await etdRepo.getWork( route.params.id, "view" )
})

useSeoMeta({
   title: () => etdRepo.work.title,
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

const extractLink = ( data) => {
   let regex  = /(http|ftp|https):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])/
	let final = []
   let tokens = data.split(" ")
	tokens.forEach(  token => {
      let trimmed = token.trim()
      if (regex.test( trimmed ) ) {
         let link = `<a href="${trimmed}" target="_blank" aria-describedby="new-window">${trimmed}</a>`
         final.push(link)
      } else {
         final.push(trimmed)
      }
   })
   return final.join(" ")
}
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   div.public-work {
      display: flex;
      flex-flow: row-reverse nowrap;
      justify-content: center;
      align-items: flex-start;
      gap: 1rem;
      padding: 2.5rem 2.5rem 0 1rem;
   }

   .details {
      flex-basis: 70%;
   }
}

@media only screen and (max-width: 768px) {
   div.work-bkg {
      display: none;
   }
   div.public-work {
      display: flex;
      flex-direction: column;
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
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-start;
      gap: 20px;
      .view-cnt {
         padding: 10px;
         font-size: 0.5em;
         text-align: center;
         display: inline-block;
         border: 1px solid $uva-grey-50;
         border-radius: 100px;
      }
   }

   .content.author {
      display: flex;
      flex-flow: row wrap;
      gap: 10px;
      a.orcid {
         display: inline-flex;
         flex-flow: row nowrap;
         gap: 5px;
         justify-content: flex-start;
         align-self: center;

         .orcid-img {
            width: 20px;
         }
      }
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
         ul.unstyled {
            display: flex;
            flex-direction: column;
            gap: 5px;
            margin: 0;
            list-style: none;
            padding: 0 0 0 0px;
         }
         ul.links {
            margin-top: 0;
            margin-bottom: 0;
            padding-left: 20px;
            li {
               margin-top: 5px;
            }
         }
         h2 {
            text-align: left;
            font-size: 1.15em;
            padding:0;
            margin:0 0 5px 0;
         }
      }
   }
}
</style>
