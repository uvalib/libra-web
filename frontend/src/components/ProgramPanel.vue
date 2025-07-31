<template>
<div class="work-overview">
   <dl>
      <dt>Institution:</dt>
      <dd>{{ etdRepo.work.author.institution  }}</dd>

      <dt id="admin-program">Program:</dt>
      <dd v-if="props.admin == false">{{ etdRepo.work.program  }}</dd>
      <dd v-else>
         <Select v-model="etdRepo.work.program" :options="programs" editable fluid
            ariaLabelledby="admin-program" @update:modelValue="emit('changed')"/>
      </dd>

      <dt id="admin-degree">Degree:</dt>
      <dd v-if="props.admin == false">{{ etdRepo.work.degree }}</dd>
      <dd v-else>
         <Select v-model="etdRepo.work.degree" :options="degrees" fluid
            ariaLabelledby="admin-degree" @update:modelValue="emit('changed')"/>
      </dd>
   </dl>
   <div class="data-column">
      <dl>
         <dt>Date Created:</dt>
         <dd>{{ $formatDate(etdRepo.createdAt) }}</dd>

         <template v-if="etdRepo.isDraft==false">
            <dt class="label">Date Published:</dt>
            <dd v-if="editDate" class="pub-date">
               <InputMask v-model="newDate" autofocus mask="9999-99-99" slotChar="yyyy-mm-dd" @keydown.enter="updateDate" />
               <Button class="action" icon="pi pi-times" rounded severity="secondary" aria-label="cancel" size="small" @click="editDate=false" :disabled="etdRepo.saving"/>
               <Button class="action" icon="pi pi-check" rounded severity="secondary" aria-label="rename" size="small" @click="updateDate" :loading="etdRepo.saving"/>
            </dd>
            <dd v-else class="pub-date">
               <span>{{ $formatDate(etdRepo.publishedAt) }}</span>
               <Button v-if="props.admin" label="Edit" size="small" rounded icon="pi pi-pen-to-square" severity="secondary" @click="editDateClicked"/>
            </dd>
         </template>
      </dl>
      <div class="error" v-if="dateError">{{ dateError }}</div>
      <AuditsPanel :workID="etdRepo.work.id"/>
   </div>
</div>
</template>

<script setup>
import { computed, ref } from 'vue'
import AuditsPanel from '@/components/AuditsPanel.vue'
import { useSystemStore } from "@/stores/system"
import { useETDStore } from "@/stores/etd"
import Select from 'primevue/select'
import InputMask from 'primevue/inputmask'
import dayjs from 'dayjs'
import customParseFormat from 'dayjs/plugin/customParseFormat';
dayjs.extend(customParseFormat)

const system = useSystemStore()
const etdRepo = useETDStore()

const editDate = ref(false)
const newDate = ref()
const dateError = ref("")

const emit = defineEmits( ['changed'])

const props = defineProps({
   admin: {
      type: Boolean,
      default: false
   }
})

const editDateClicked = (() => {
   newDate.value = etdRepo.publishedAt
   editDate.value = true
})
const updateDate = ( async () => {
   dateError.value = ""
   var nd = new Date(newDate.value).getTime()
   var now = new Date().getTime()
   if ( nd > now ) {
      dateError.value = "Publication date cannot be in the future"
      return
   }

   if ( dayjs( newDate.value, 'YYYY-MM-DD', true).isValid() == false ) {
      dateError.value = `${newDate.value} is not a valid date`
      return
   }

   await etdRepo.updatePublishedDate( newDate.value )
   if ( system.showError == false ) {
      system.toastMessage("Updated", "The publication date has been updated")
      editDate.value = false
   }
})

const programs = computed( () =>{
   if (etdRepo.source == "sis") return system.sisPrograms
   return system.optPrograms
})

const degrees = computed( () =>{
   if (etdRepo.source == "sis") return system.sisDegrees
   return system.optDegrees
})
</script>

<style lang="scss" scoped>
.work-overview {
   display: flex;
   flex-flow: row wrap;
   justify-content: space-between;
   align-items: flex-start;
   gap: 25px;
   margin-bottom: 25px;
   .data-column {
      display: flex;
      flex-direction: column;
      gap: 15px;
   }
   dl {
      grid-template-columns: max-content auto;
      display: grid;
      grid-column-gap: 0.75rem;
      padding: 0;
      margin: 0;
      dt {
         font-weight: bold;
         text-align: right;
         padding: 0.3rem 0;
         white-space: nowrap;
      }
      dd {
         padding: 0.3rem 0;
         margin: 0;
      }
      dd.pub-date {
         display: flex;
         flex-flow: row nowrap;
         gap: 5px;
         justify-content: flex-start;
         align-items: center;
         padding: 0;
      }
   }
   .error {
      color: $uva-red-A;
      text-align: center;
   }
}
</style>

