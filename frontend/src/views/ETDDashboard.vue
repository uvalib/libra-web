<template>
   <h1>LibraETD Dashboard</h1>
   <div class="dashboard">
      <Button label="Edit Work" @click="editWorkClicked"/>
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()

const editWorkClicked = (() => {
   router.push("/etd/new")
})
</script>

<style lang="scss" scoped>
.dashboard {
   width: 70%;
   margin: 50px auto;
   min-height: 600px;
}
</style>