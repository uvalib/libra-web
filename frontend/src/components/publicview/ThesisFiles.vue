<template>
   <div class="files">
      <h2 class="title">Files</h2>
      <template  v-if="etdRepo.visibility == 'embargo' || etdRepo.visibility == 'restricted' ">
         <span class="file-embargo">
         This item is restricted to abstract view only until {{ $formatDate(etdRepo.embargoReleaseDate) }}.
         </span>
         <span  v-if="etdRepo.isDraft" class="file-embargo author">
            The files listed below will NOT be available to anyone until the embargo date has passed.
         </span>
      </template>
      <span  v-if="etdRepo.visibility == 'uva'" class="file-embargo">
         This item is restricted to UVA until {{ $formatDate(etdRepo.embargoReleaseDate) }}.
      </span>

      <div  v-if="etdRepo.isDraft || (etdRepo.visibility != 'embargo' && etdRepo.visibility != 'restricted')" class="file" v-for="file in etdRepo.work.files">
         <div class="name">{{ file.name }}</div>
         <div class="file-stat">Uploaded on {{ $formatDate(file.createdAt) }}</div>
         <div class="file-stat">Downloads: {{ file.downloads }}</div>
         <Button label="Download" icon="pi pi-cloud-download" severity="secondary"
            :ariaLabel="`download file ${file.name}`"
            @click="etdRepo.downloadFile(file.name, 'view')" :loading="etdRepo.downloading==file.name" />
      </div>
   </div>
</template>

<script setup>
import { useETDStore } from "@/stores/etd"

const etdRepo = useETDStore()
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .files {
      margin-top: 280px;
      flex-basis: 30%;
   }
}

@media only screen and (max-width: 768px) {
}
div.files {
   font-family: 'Open Sans', sans-serif;
   background: white;
   text-align: left;
   display: flex;
   flex-direction: column;
   gap: 10px;
   padding: 20px;
   border: 1px solid $uva-grey-100;
   .title {
      font-weight: bold;
   }

   .file-stat {
      font-size: 0.9em;
   }

   .file-embargo {
      padding: 10px;
      font-style: normal;
      background: $uva-yellow-100;
      border: 1px solid $uva-yellow-A;
      border-radius: 4px;
   }

   .file {
      display: flex;
      flex-direction: column;
      gap: 10px;
      padding-top: 15px;
      border-top: 1px solid $uva-grey-100;
   }
}

</style>
