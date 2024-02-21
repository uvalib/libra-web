<template>
   <h1>Add New Work</h1>
   <div class="form">
      <FormKit type="form" id="openurl-book" :actions="false" @submit="submitClicked">
         <FormKit type="select" label="Resource Type" v-model="data.resourceType"
            placeholder="Select a resource type"
            :options="system.oaResourceTypes" validation="required"
         />
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
   resourceType: ""
})

const submitClicked = ( () => {
   alert("ER")
})
</script>

<style lang="scss" scoped>
.form {
   width: 70%;
   margin: 50px auto;
   min-height: 600px;
}
</style>