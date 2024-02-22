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
         <FormKit label="Notes" type="textarea" v-model="data.notes" rows="10"/>

      </FormKit>
   </div>
</template>

<script setup>
import { ref } from 'vue'
import { useSystemStore } from "@/stores/system"

const system = useSystemStore()
// import { useRouter } from 'vue-router'

// const router = useRouter()

// const createWorkClicked = (() => {
//    router.push("/oa/new")
// })

const data = ref({
   resourceType: null,
   title: "",
   abstract: "",
   rights: null,
   publisher: "University of Virginia",
   citation: "",
   pubDate: "",
   notes: ""
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
   .note {
      font-size: 0.85em;
      font-style: italic;
      color: var(--uvalib-grey);
      margin-top: 0;
      padding-top: 5px;
   }
}
</style>