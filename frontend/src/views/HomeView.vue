<template>
   <h1>Libra</h1>
   <div class="info">
      <p>
         Libra is the University of Virginia's open access institutional repository, the online archive of scholarship
         created by the University community. It provides a central, stable location for the scholarly output of the
         University community (journal articles, datasets, theses/dissertations, etc.). Anyone can search, view and
         download content.
      </p>
      <p>
         For questions or feedback, contact us: <a href="mailto:libra@virginia.edu">libra@virginia.edu</a>
      </p>
   </div>
   <h2>Submit your work</h2>
   <div class="submissions">
      <Card>
         <template #title>LibraOpen: Open content</template>
         <template #content>
            <p>
               Submit your scholarly articles, books, presentations, and other creative works. Please review our
               <a target="_blank" href="https://library.virginia.edu/libra/open/oc-checklist">LibraOpen Checklist</a> or
               contact
               <a href="mailto:libra@virginia.edu">libra@virginia.edu</a> for deposit help.
            </p>
            <p>For any UVA-affilliated individual</p>
            <Button label="Submit OA Content" @click="oaClicked" />
         </template>
      </Card>
      <Card>
         <template #title>LibraETD: Theses and dissertations</template>
         <template #content>
            <p>
               Deposit your approved UVA electronic thesis or dissertation.
               Before uploading your work, please review our
               <a target="_blank" href="https://library.virginia.edu/libra/etds/etds-checklist">
                  ETD Submission Checklist
               </a>.
            </p>
            <p>For students submitting approved theses or dissertations</p>
            <Button label="Submit ETD Content" @click="etdClicked" />
         </template>
      </Card>
   </div>
</template>

<script setup>
import Card from 'primevue/card'
import { useRouter } from 'vue-router'

const router = useRouter()

const oaClicked = (() => {
   router.push(`/oa`)
})
const etdClicked = (() => {
   router.push(`/etd`)
})
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .info {
      width: 75%;
      margin: 10px auto 50px auto;
      font-size: 1.15em;
   }
   .submissions {
      width: 75%;
      div {
         max-width: 60%;
      }
   }
}
@media only screen and (max-width: 768px) {
   .info {
      width: 95%;
      font-size: 1em;
   }
   .submissions {
      width: 95%;
      div {
         max-width: 90%;
      }
   }
}
.info {
   margin: 10px auto 30px auto;
}
.submissions {
   margin: 30px auto;
   display: flex;
   flex-flow: column;
   justify-content: flex-start;
   align-items: center;
   gap: 1rem;
}
</style>

