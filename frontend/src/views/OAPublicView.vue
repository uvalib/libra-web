<template>
   <div class="work-bkg"></div>
   <WaitSpinner v-if="oaRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
   <div v-else class="public-work">
      <div class="files">
         <Fieldset legend="Files">

         </Fieldset>
      </div>
      <div class="details">
         <div class="title" role="heading">{{ oaRepo.work.title }}</div>
         <Fieldset class="author-fieldset">
            <template #legend>
               <div class="author-header">
                  <span class="legend">Authors:</span>
                  <span class="type">{{ oaRepo.work.resourceType }}</span>
               </div>
            </template>
            <div v-for="author in  oaRepo.work.authors" class="author">
               <p>{{ authorDisplay(author) }}</p>
               <p>{{ author.institution }}</p>
            </div>
         </Fieldset>
         <Fieldset legend="Abstract:">{{  oaRepo.work.abstract }}</Fieldset>
         <Fieldset v-if="oaRepo.hasKeywords" legend="Keywords:">
            {{ oaRepo.work.keywords.join(", ") }}
         </Fieldset>
         <Fieldset legend="Rights:">
            <a :href="system.licenseDetail('oa', oaRepo.work.license).url" target="_blank">
               {{ system.licenseDetail("oa", oaRepo.work.license).label }}
            </a>
         </Fieldset>
         <Fieldset v-if="oaRepo.hasContributors" legend="Contributors:">
            <div v-for="contributor in  oaRepo.work.contributors" class="author">
               <p>{{ authorDisplay(contributor) }}</p>
               <p>{{ contributor.institution }}</p>
            </div>
         </Fieldset>
         <Fieldset v-if="oaRepo.hasLanguages" legend="Languages:">
            {{ oaRepo.work.languages.join(", ") }}
         </Fieldset>
         <Fieldset v-if="oaRepo.work.citation" legend="Source Citation::">
            {{ oaRepo.work.citation }}
         </Fieldset>
         <Fieldset legend="Publisher:">{{  oaRepo.work.publisher }}</Fieldset>
         <Fieldset v-if="oaRepo.work.pubDate" legend="Published Date:">{{  oaRepo.work.pubDate }}</Fieldset>
         <Fieldset v-if="oaRepo.hasRelatedURLs" legend="Related Links:">
            <ul>
               <li v-for="url in oaRepo.work.relatedURLs"><a :href="url" target="_blank">{{ url }}</a></li>
            </ul>
         </Fieldset>
         <Fieldset v-if="oaRepo.hasSponsors" legend="Sponsoring Agency:">
            <div v-for="s in oaRepo.work.sponsors">{{ s }}</div>
         </Fieldset>
         <Fieldset v-if="oaRepo.work.notes" legend="Notes:">{{  oaRepo.work.notes }}</Fieldset>
         <!--
         <%= display_doi_link( @work ) %> -->
      </div>
   </div>
</template>

<script setup>
import { ref, onBeforeMount, computed } from 'vue'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { useOAStore } from "@/stores/oa"
import { useRoute } from 'vue-router'
import Fieldset from 'primevue/fieldset'
import WaitSpinner from "@/components/WaitSpinner.vue"

const system = useSystemStore()
const user = useUserStore()
const oaRepo = useOAStore()
const route = useRoute()

const authorDisplay = ((a) => {
   return `${a.lastName}, ${a.firstName}, ${a.department}`
})
onBeforeMount( async () => {
   console.log("BEFORE MOUNT")
   document.title = "LibraOpen"
   await oaRepo.getWork( route.params.id )
})
</script>

<style lang="scss" scoped>
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
   display: flex;
   flex-flow: row nowrap;
   justify-content: center;
   align-items: flex-start;
   position: relative;
   min-height: 300px;

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
      width: 250px;
      margin-top: 320px;
   }

   div.details {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      border: 1px solid var(--uvalib-grey-light);
      box-shadow: 0 0 2px #b9b9b9;
      max-width: 640px;
      padding: 30px;
      border-radius: 3px;
      margin: 20px;

      .author-fieldset {
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
               font-weight: normal;
               font-size: 0.85em;
               border-radius: 5px;
               background-color: var(--uvalib-grey-dark);
               color: white;
               padding: 4px 10px;
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