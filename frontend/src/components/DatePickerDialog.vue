<template>
   <Button @click="show" label="Change" severity="secondary" class="change"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Set End Date" style="width:fit-content" position="top">
      <div class="embargo-date">
         <div class="help">Use one of the quick helper buttons or pick a custom end date</div>
         <div class="datepick">
            <Calendar v-model="endDate" inline :minDate="new Date()" :maxDate="tenYears"/>
            <div class="helpers">
               <Button label="6 Months" severity="secondary" @click="setEmbargoEndDate(6,'month')"/>
               <Button label="1 Year" severity="secondary" @click="setEmbargoEndDate(1,'year')"/>
               <Button label="2 Years" severity="secondary" @click="setEmbargoEndDate(2,'year')"/>
               <Button label="5 Years" severity="secondary" @click="setEmbargoEndDate(5,'year')"/>
               <Button v-if="showTenYear" label="10 Years" severity="secondary" @click="setEmbargoEndDate(10,'year')"/>
               <Button v-if="props.admin && props.visibility=='embargo'" label="Forever" severity="secondary" @click="endDate = null"/>
            </div>
         </div>
         <div class="controls">
            <span v-if="endDate" ><b>End date</b>: {{ $formatDate(endDate) }}</span>
            <span v-else>No expiration date</span>
            <span>
               <Button severity="secondary" label="Cancel" @click="isOpen=false"/>
               <Button label="OK" @click="okClicked"/>
            </span>
         </div>
      </div>
   </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import Calendar from 'primevue/calendar'

const emit = defineEmits( ['picked'])
const props = defineProps({
   admin: {
      type: Boolean,
      default: false
   },
   program: {
      type: String,
      default: ""
   },
   degree: {
      type: String,
      default: ""
   },
   type: {
      type: String,
      required: true,
      validator(value) {
         return ['oa', 'etd'].includes(value)
      },
   },
   visibility: {
      type: String,
      required: true
   },
   endDate: {
      type: String,
      default: null
   },
})
const isOpen = ref(false)
const endDate = ref()

const showTenYear = computed( () => {
   if ( props.admin) return true
   return ( props.program.includes('Creative Writing') && props.degree == 'MFA (Master of Fine Arts)' )
})
const tenYears = computed( () => {
   let d = new Date()
   d.setFullYear( d.getFullYear() + 10)
   return d
})

const show = (() => {
   isOpen.value = true
   endDate.value = new Date(props.endDate)
})

const okClicked = (() => {
   emit("picked", endDate.value.toJSON() )
   isOpen.value = false
})

const setEmbargoEndDate = ((count, type) => {
   endDate.value = new Date()
   if (type=="month") {
      endDate.value.setMonth( endDate.value.getMonth() + count)
   } else {
      endDate.value.setFullYear( endDate.value.getFullYear() + count)
   }
})

</script>

<style lang="scss" scoped>
button.change {
   margin-bottom: 5px;
   font-size: 0.85em;
   padding: 4px 10px;
}
.help {
   text-align: left;
   margin-bottom: 15px;
}
.datepick {
   display: flex;
   flex-flow: row nowrap;
   .helpers {
      display: flex;
      flex-direction: column;
      margin-left: 15px;
      button {
         font-size: 0.85em;
         margin-bottom: 5px;
      }
   }
}
.controls {
   margin:15px 0 5px 0;
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   align-items: center;
   button {
      font-size: 0.9em;
      margin-left: 10px;
   }
}
:deep(span.p-disabled) {
   color: #ddd;
}
:deep(.p-datepicker-today) {
   span {
      background: white;
      border: 1px solid $uva-grey-100;
   }
}
</style>