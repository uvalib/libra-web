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
            <td class="label">Date Published:</td><td>{{ $formatDate(etdRepo.publishedAt) }}</td>
         </tr>
         <tr>
            <td colspan="2"><AuditsPanel :workID="etdRepo.work.id"/></td>
         </tr>
      </tbody>
   </table>
</div>
</template>

<script setup>
import { computed } from 'vue'
import AuditsPanel from '@/components/AuditsPanel.vue'
import { useSystemStore } from "@/stores/system"
import { useETDStore } from "@/stores/etd"
import Select from 'primevue/select'

const system = useSystemStore()
const etdRepo = useETDStore()

const emit = defineEmits( ['changed'])

const props = defineProps({
   admin: {
      type: Boolean,
      default: false
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
      margin-bottom: 25px;
   }
}
</style>

