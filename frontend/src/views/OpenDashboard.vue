<template>
   <h1>Dashboard</h1>
   <div class="dashboard">
      <Button label="Create new work" @click="createWorkClicked"/>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const createWorkClicked = (() => {
   router.push("/oa/new")
})
</script>

<style lang="scss" scoped>
.dashboard {
   width: 70%;
   margin: 50px auto;
   min-height: 600px;
}
</style>