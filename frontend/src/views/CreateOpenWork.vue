<template>
   <h1>Add New Work</h1>
   <div class="form">
      <FormKit type="form" :actions="false" @submit="submitClicked">
         <FormKit type="select" label="Resource Type" v-model="data.resourceType"
            placeholder="Select a resource type"
            :options="system.oaResourceTypes" validation="required"
         />
         <FormKit label="Title" type="text" v-model="data.title" validation="required"/>
         <FormKit label="Abstract" type="textarea" v-model="data.abstract" rows="10" validation="required"/>

         <FormKit type="select" label="Rights" v-model="data.rights"
            placeholder="Select rights"
            :options="system.oaLicenses" validation="required"
         />
         <p class="note">
            Libra lets you choose an open license when you post your work, and will prominently display the
            license you choose as part of the record for your work. See
            <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
            for option details.
         </p>

         <FormKit v-model="data.languages" type="list" dynamic #default="{ items }">
            <div v-for="(item, index) in items" :key="item" class="input-row">
               <div class="input-wrap">
                  <FormKit type="select" :label="inputLabel('Language', index)" :index="index"
                     placeholder="Select a language" :options="system.languages"
                  />
               </div>
               <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove language"
                  :disabled="data.languages.length == 1" @click="removeLanguage(index)"/>
            </div>
         </FormKit>
         <p class="note controls">
            <span>The language of the work's content.</span>
            <Button label="Add Language" @click="addLanguage"/>
         </p>

         <FormKit v-model="data.keywords" type="list" dynamic #default="{ items }">
            <div v-for="(item, index) in items" :key="item" class="input-row">
               <div class="input-wrap">
                  <FormKit :label="inputLabel('Keyword', index)" type="text" :index="index" />
               </div>
               <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove keyword"
                  :disabled="data.keywords.length == 1" @click="removeKeyword(index)"/>
            </div>
         </FormKit>
         <p class="note controls">
            <span>Add one keyword or keyword phrase per line.</span>
            <Button label="Add Keyword" @click="addKeyword"/>
         </p>

         <FormKit label="Publisher" type="text" v-model="data.publisher" validation="required"/>
         <p class="note">
            Libra lets you choose an open license when you post your work, and will prominently display the
            license you choose as part of the record for your work. See
            <a href="https://creativecommons.org/share-your-work" target="_blank">Choose a Creative Commons License</a>
            for option details.
         </p>
         <FormKit label="Source citation" type="text" v-model="data.citaion"/>
         <p class="note">The bibliographic citation of the work that reflects where it was originally published.</p>
         <FormKit label="Published date" type="text" v-model="data.pubDate"/>

         <FormKit v-model="data.relatedURLs" type="list" dynamic #default="{ items }">
            <div v-for="(item, index) in items" :key="item" class="input-row">
               <div class="input-wrap">
                  <FormKit :label="inputLabel('Related URL', index)" type="text" :index="index" />
               </div>
               <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove url"
                  :disabled="data.relatedURLs.length == 1" @click="removeURL(index)"/>
            </div>
         </FormKit>
         <p class="note controls">
            <span>Links to another version, another location with the file, website or other specific content (audio, video, PDF document) related to the work.</span>
            <Button label="Add URL" @click="addURL"/>
         </p>

         <FormKit v-model="data.sponsors" type="list" dynamic #default="{ items }">
            <div v-for="(item, index) in items" :key="item" class="input-row">
               <div class="input-wrap">
                  <FormKit :label="inputLabel('Sponsoring Agency', index)" type="text" :index="index" />
               </div>
               <Button class="remove" icon="pi pi-trash" severity="danger" aria-label="remove agency"
                  :disabled="data.sponsors.length == 1" @click="removeAgency(index)"/>
            </div>
         </FormKit>
         <p class="note controls">
            <Button label="Add Agency" @click="addAgency"/>
         </p>

         <FormKit label="Notes" type="textarea" v-model="data.notes" rows="10"/>

      </FormKit>
   </div>
</template>

<script setup>
import { ref } from 'vue'
import { useSystemStore } from "@/stores/system"

const system = useSystemStore()

const data = ref({
   resourceType: null,
   title: "",
   abstract: "",
   rights: null,
   languages: [""],
   keywords: [""],
   publisher: "University of Virginia",
   citation: "",
   pubDate: "",
   relatedURLs: [""],
   sponsors: [""],
   notes: ""
})

const inputLabel = ( (lbl, idx) => {
   console.log(lbl+" idx "+idx)
   if (idx==0) return lbl
   return null
})
const removeKeyword = ((idx)=> {
   data.value.keywords.splice(idx,1)
})
const addKeyword = ( () => {
   data.value.keywords.push("")
})
const removeLanguage = ((idx)=> {
   data.value.languages.splice(idx,1)
})
const addLanguage = ( () => {
   data.value.languages.push("")
})
const removeURL = ((idx)=> {
   data.value.relatedURLs.splice(idx,1)
})
const addURL = ( () => {
   data.value.relatedURLs.push("")
})
const removeAgency = ((idx)=> {
   data.value.sponsors.splice(idx,1)
})
const addAgency = ( () => {
   data.value.sponsors.push("")
})

const submitClicked = ( () => {
   alert("ER")
})
</script>

<style lang="scss" scoped>
.form {
   width: 50%;
   margin: 50px auto;
   min-height: 600px;
   text-align: left;
   .input-row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      .remove {
         padding: 6.25px 15px;
         margin-bottom: 0.3em;
         border: 0;
         margin-left: 5px;
      }
      .input-wrap {
         flex-grow: 1;
      }
   }
   .note {
      font-size: 0.85em;
      font-style: italic;
      color: var(--uvalib-grey);
      margin-top: 0;
      padding-top: 5px;
   }
   .note.controls {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
      button {
         font-size: 0.9em;
         padding: 4px 10px;
         white-space: nowrap;
         margin-left: auto;
      }
   }
}
</style>