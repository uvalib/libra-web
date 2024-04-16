<template>
   <div class="work-bkg"></div>
   <WaitSpinner v-if="oaRepo.working" :overlay="true" message="<div>Please wait...</div><p>Loading Work</p>" />
   <div v-else class="public-work">
      <div v-if="oaRepo.error" class="error">
         <h2>System Error</h2>
         <p>Sorry, a system error has occurred!</p>
         <p>{{ oaRepo.error }}</p>
      </div>
      <template v-else>
         <div class="files">
            <Fieldset legend="Files">
               <span class="file-embargo owner" v-if="oaRepo.visibility == 'embargo' && oaRepo.work.files">
                  <p>The files listed below will NOT be available until the embargo date has passed.</p>
                  <p>
                     Embargo end date:<br/>
                     {{ $formatDate(oaRepo.embargoReleaseDate) }}
                  </p>
               </span>
               <span  v-else-if="oaRepo.visibility == 'embargo'" class="file-embargo">
                  This item is embargoed and will not available until {{ $formatDate(oaRepo.embargoReleaseDate) }}.
               </span>
               <span  v-else-if="oaRepo.visibility == 'uva'" class="file-embargo">
                  This item is restricted to UVA.
               </span>
               <div class="file" v-for="file in oaRepo.work.files">
                  <div class="name">{{ file.name }}</div>
                  <div class="upload"><label>Uploaded:</label>{{ $formatDate(file.createdAt) }}</div>
                  <Button icon="pi pi-cloud-download" label="Download" severity="secondary" @click="downloadFileClicked(file.name)"/>
               </div>
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
               <a :href="oaRepo.work.licenseURL" target="_blank">{{ oaRepo.work.license }}</a>
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
            <Fieldset v-if="oaRepo.persistentLink" legend="Persistent Link:">
               <a target="_blank" :href="oaRepo.persistentLink">{{ oaRepo.persistentLink }}</a>
            </Fieldset>
         </div>
         <div class="draft" v-if="oaRepo.isDraft">
            <h2 class="proof">Publication Proof</h2>
            <p>
               Before proceeding, we encourage you to review the information in this page.
               If you experience problems with your publication, please <a href="mailto:libra@virginia.edu">contact</a> us.
            </p>
            <div class="buttons">
               <Button severity="secondary" label="Cancel" @click="cancelPreview()"/>
               <Button severity="secondary" label="Edit" @click="editWork()"/>
               <Button severity="primary" label="Publish" @click="publishWork()"/>
            </div>
         </div>
         <div class="published" v-if="justPublished">
            Thank you for publishing your work.
         </div>
      </template>
   </div>
</template>

<script setup>
import { onBeforeMount, ref } from 'vue'
import { useOAStore } from "@/stores/oa"
import { useSystemStore } from "@/stores/system"
import { useRoute, useRouter } from 'vue-router'
import Fieldset from 'primevue/fieldset'
import WaitSpinner from "@/components/WaitSpinner.vue"
import { useConfirm } from "primevue/useconfirm"

const oaRepo = useOAStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()
const system = useSystemStore()
const justPublished = ref(false)

const authorDisplay = ((a) => {
   return `${a.lastName}, ${a.firstName}, ${a.department}`
})
onBeforeMount( async () => {
   document.title = "LibraOpen"
   await oaRepo.getWork( route.params.id )
})
const downloadFileClicked = ( (name) => {
   oaRepo.downloadFile(name)
})
const editWork = (() => {
   router.push(`/oa/${route.params.id}`)
})
const cancelPreview = ( () => {
   router.back()
})
const publishWork = ( () => {
   confirm.require({
      message: `Publish this work? Depending upon your visibility settings, this work will become visible worldwide or to UVA only. Are you sure?`,
      header: 'Confirm Publication',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: async () => {
         await oaRepo.publish()
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
            padding: 6px 12px;
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
      .file-embargo {
         display: block;
         margin: 10px 0 10px 10px;
         font-style: italic;
         font-size: 1.15em;
      }
      .file-embargo.owner {
         margin: 10px 0 20px 0;
         background: var(--uvalib-yellow-light);
         padding: 15px 15px 5px 15px;
         font-size: 1em;
         border-radius: 5px;
         border: 1px solid var(--uvalib-yellow);
         p {
            margin:0 0 10px 0;
         }
      }
      .file {
         margin: 10px 0 0 10px;
         border: 1px solid var(--uvalib-grey-light);
         border-radius: 4px;
         padding: 10px;
         .name {
            text-align: left;
            margin-bottom: 10px;
            font-size: 1.1em;
         }
         .upload {
            font-size: 0.9em;
            margin-bottom: 5px;
         }
         label {
            font-weight: bold;
            margin-right: 5px;
         }
         button {
            margin: 15px auto 0 auto;
            font-size: 0.9em;
            padding: 4px 10px;
            display: block;
         }
      }
   }

   div.details {
      font-family: 'Open Sans', sans-serif;
      background: white;
      text-align: left;
      border-radius: 3px;
      min-height: 600px;

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