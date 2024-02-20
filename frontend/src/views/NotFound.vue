<template>
   <h1>Not Found</h1>
   <div class="notfound">
      <div>Sorry, the URL you requested could not be found!</div>
      <Button label="Return to Home Page"  @click="homeClicked()" />
   </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
function homeClicked() {
   router.push("/")
}
</script>

<style scoped lang="scss">
.notfound {
   min-height: 600px;
   font-size: 1.25em;
   margin: 0;
   padding: 5px 0 15px 0;
   div {
      margin-bottom: 30px;
   }

}
</style>
