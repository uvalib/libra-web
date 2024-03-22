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
         <Fieldset legend="Authors:">
            <div v-for="author in  oaRepo.work.authors" class="author">
               <p>{{ authorDisplay(author) }}</p>
               <p>{{ author.institution }}</p>
            </div>
         </Fieldset>
         <Fieldset legend="Abstract:">{{  oaRepo.work.abstract }}</Fieldset>
         <Fieldset legend="Rights:">
            <a :href="system.licenseDetail('oa', oaRepo.work.license).url" target="_blank">
               {{ system.licenseDetail("oa", oaRepo.work.license).label }}
            </a>
         </Fieldset>
         <!-- <% display_resource_type @work %>
         <% display_authors( Author.sort(@work.authors) ) %>
         <%= display_abstract(@work.abstract) %>
         <%= display_keywords( @work ) %>
         <%= display_rights(@work.rights_ids) %>
         <% display_contributors( Contributor.sort(@work.contributors )) %>
         <%= display_language( @work.language ) %>
         <%= display_source_citation( @work.source_citation ) %>
         <%= display_generic( 'Publisher', @work.publisher) %>
         <%= display_generic_date( 'Published Date', date_formatter( @work.published_date ) ) %>
         <%= display_related_links( @work.related_url ) %>
         <%= display_sponsoring_agency( @work.sponsoring_agency) %>
         <%= display_notes( @work.notes ) %>
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