<template>
   <div class="draft" v-if="etdRepo.isDraft">
      <div class="proof">Submission Proof</div>
      <div>
         Before proceeding, we encourage you to review the information in this page.
         If you experience problems with your submission, please <a href="mailto:libra@virginia.edu">contact</a> us.
      </div>
      <div class="agree">
         <Checkbox inputId="agree-cb" v-model="agree" :binary="true" />
         <label for="agree-cb">
            I have read and agree to the
            <a href="https://www.library.virginia.edu/libra/etds/etd-license" target="_blank" aria-describedby="new-window">Libra Deposit License</a>,
            including discussing my deposit access options with my faculty advisor.
         </label>
      </div>
      <div class="buttons">
         <RouterLink :to="user.homePage">Dashboard</RouterLink>
         <RouterLink :to="editThesisLink()">Edit thesis</RouterLink>
         <Button severity="primary" label="Submit Thesis" size="small" class="submit" @click="submitThesis" :disabled="!agree"/>
      </div>
   </div>
   <div class="published" v-if="justPublished">
      Thank you for submitting your thesis. Be sure to take note of and use
      the Persistent Link when you refer to this work.
   </div>
</template>

<script setup>
import { ref } from 'vue'
import { useETDStore } from "@/stores/etd"
import { useUserStore } from "@/stores/user"
import Checkbox from 'primevue/checkbox'
import { useConfirm } from "primevue/useconfirm"
import { useRoute } from 'vue-router'

const etdRepo = useETDStore()
const user = useUserStore()
const confirm = useConfirm()
const route = useRoute()

const justPublished = ref(false)
const agree = ref(false)

const editThesisLink = (() => {
   if (user.isAdmin) {
      return `/admin/etd/${route.params.id}`
   }
   return `/etd/${route.params.id}`
})

const submitThesis = ( () => {
   confirm.require({
      message: `This is your final step and you cannot change the document afterwards. Are you sure?`,
      header: 'Confirm Submission',
      icon: 'pi pi-question-circle',
      rejectClass: 'p-button-secondary',
      accept: async () => {
         await etdRepo.publish()
         if (system.error == "") {
            justPublished.value = true
         }
      },
   })
})
</script>

<style lang="scss" scoped>
 div.draft {
   background: $uva-yellow-100;
   padding: 20px;
   border: 2px solid $uva-yellow-A;
   display: flex;
   flex-direction: column;
   gap: 10px;
   text-align: left;
   .proof {
      font-size: 1.2em;
      font-weight: bold;
      text-align: center;
   }
   .buttons {
      margin-top: 10px;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-items: center;
      gap: 1.15rem;
      .submit {
         margin-left: auto;
      }
   }
   .agree {
      display: flex;
      flex-flow: row nowrap;
      justify-content: center;
      align-items: flex-start;
      gap: 10px;
      margin: 10px 0;
   }
}

div.published {
   background: $uva-yellow-100;
   padding: 20px;
   border: 1px solid $uva-yellow-A;
   text-align: left;
}
</style>