<template>
<div class="work-overview">
   <table>
      <tbody>
         <tr>
            <td class="label">Institution:</td><td>{{ etdRepo.work.author.institution  }}</td>
         </tr>
         <tr>
            <td class="label" id="admin-program">Program:</td>
            <td v-if="props.admin == false">{{ etdRepo.work.program  }}</td>
            <td v-else>
               <Select v-model="etdRepo.work.program" :options="programs" editable fluid
                  ariaLabelledby="admin-program" @update:modelValue="emit('changed')"/>
            </td>
         </tr>
         <tr>
            <td class="label" id="admin-degree">Degree:</td>
            <td v-if="props.admin == false">{{ etdRepo.work.degree }}</td>
            <td v-else>
               <Select v-model="etdRepo.work.degree" :options="degrees" fluid
                  ariaLabelledby="admin-degree" @update:modelValue="emit('changed')"/>
            </td>
         </tr>
      </tbody>
   </table>
   <table>
      <tbody>
         <tr>
            <td class="label">Date Created:</td><td>{{ $formatDate(etdRepo.createdAt) }}</td>
         </tr>
         <tr v-if="etdRepo.isDraft==false">
            <td class="label">Date Published:</td>
            <td class="pub-date">
               <template v-if="editDate">
                  <InputMask v-model="newDate" autofocus mask="9999-99-99" slotChar="yyyy-mm-dd" @keydown.enter="updateDate" />
                  <Button class="action" icon="pi pi-times" rounded severity="secondary" aria-label="cancel" size="small" @click="editDate=false" :disabled="etdRepo.saving"/>
                  <Button class="action" icon="pi pi-check" rounded severity="secondary" aria-label="rename" size="small" @click="updateDate" :loading="etdRepo.saving"/>
               </template>
               <template v-else>
                  <span>{{ $formatDate(etdRepo.publishedAt) }}</span>
                  <Button v-if="props.admin" label="Edit" size="small" severity="secondary" @click="editDateClicked"/>
               </template>
            </td>
         </tr>
         <tr>
            <td colspan="2"><AuditsPanel :workID="etdRepo.work.id"/></td>
         </tr>
      </tbody>
   </table>
</div>
</template>

<script setup>
import { computed, ref } from 'vue'
import AuditsPanel from '@/components/AuditsPanel.vue'
import { useSystemStore } from "@/stores/system"
import { useETDStore } from "@/stores/etd"
import Select from 'primevue/select'
import InputMask from 'primevue/inputmask'

const system = useSystemStore()
const etdRepo = useETDStore()

const editDate = ref(false)
const newDate = ref()

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
   table {
      td.label {
         font-weight: bold;
         text-align: right;
         padding-right: 10px;
      }
      td.pub-date {
         display: flex;
         flex-flow: row nowrap;
         gap: 5px;
         justify-content: flex-start;
         align-items: center;
      }
      margin-bottom: 25px;
   }
}
</style>

