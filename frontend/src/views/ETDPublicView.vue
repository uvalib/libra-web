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
               <a :href="system.licenseDetail('oa', etdRepo.work.license).url" target="_blank">
                  {{ system.licenseDetail("oa", etdRepo.work.license).label }}
               </a>
            </Fieldset>
            <Fieldset v-if="etdRepo.work.pubDate" legend="Publication Date:">{{  etdRepo.work.pubDate }}</Fieldset>
            <!--
            <%= display_doi_link( @work ) %> -->
         </div>
      </template>
   </div>
</template>

<script setup>
import { onBeforeMount } from 'vue'
import { useSystemStore } from "@/stores/system"
import { useETDStore } from "@/stores/etd"
import { useRoute } from 'vue-router'
import Fieldset from 'primevue/fieldset'
import WaitSpinner from "@/components/WaitSpinner.vue"

const system = useSystemStore()
const etdRepo = useETDStore()
const route = useRoute()

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
         margin-top: 320px;
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