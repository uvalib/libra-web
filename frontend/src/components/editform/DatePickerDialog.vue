<template>
   <Button @click="show" label="Change" severity="secondary" size="small"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Set End Date" style="width:fit-content" position="top">
      <div class="embargo-date">
         <template v-if="props.admin">
            <div class="help">Set the end date by clicking a helper button, clicking a date on the calendar or enter it directly.</div>
            <div class="datepick">
               <div class="dates">
                  <DatePicker v-model="pickedEndDate" inline :minDate="new Date()" :maxDate="tenYears" @update:modelValue="datePicked"/>
                  <div class="setter">
                     <InputMask v-model="enteredDate" mask="9999-99-99" placeholder="YYYY-MM-DD" slotChar="YYYY-MM-DD" fluid @keydown="dateKeyPressed"/>
                     <Button label="Set" severity="secondary" @click="dateEntered($event)"/>
                  </div>
               </div>
               <div class="helpers">
                  <Button label="6 Months" severity="secondary" @click="setEmbargoEndDate(6,'month')"/>
                  <Button label="1 Year" severity="secondary" @click="setEmbargoEndDate(1,'year')"/>
                  <Button label="2 Years" severity="secondary" @click="setEmbargoEndDate(2,'year')"/>
                  <Button label="5 Years" severity="secondary" @click="setEmbargoEndDate(5,'year')"/>
                  <Button v-if="showTenYear" label="10 Years" severity="secondary" @click="setEmbargoEndDate(10,'year')"/>
                  <Button v-if="props.admin && props.visibility=='embargo'" label="Forever" severity="secondary" @click="pickedEndDate = null"/>
               </div>
            </div>
         </template>
         <div v-else class="helpers">
            <div>Select the Limited Access end date:</div>
            <Button label="6 Months" severity="secondary" @click="setEmbargoEndDate(6,'month')"/>
            <Button label="1 Year" severity="secondary" @click="setEmbargoEndDate(1,'year')"/>
            <Button label="2 Years" severity="secondary" @click="setEmbargoEndDate(2,'year')"/>
            <Button label="5 Years" severity="secondary" @click="setEmbargoEndDate(5,'year')"/>
            <Button v-if="showTenYear" label="10 Years" severity="secondary" @click="setEmbargoEndDate(10,'year')"/>
         </div>
         <div class="controls">
            <span v-if="pickedEndDate" ><b>End date</b>: {{ pickedDateStr }}</span>
            <span v-else>No expiration date</span>
            <div class="buttons">
               <Button severity="secondary" label="Cancel" @click="isOpen=false"/>
               <Button label="OK" @click="okClicked"/>
            </div>
         </div>
      </div>
   </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import DatePicker from 'primevue/datepicker'
import InputMask from 'primevue/inputmask'
import dayjs from 'dayjs'

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
const pickedEndDate = ref()
const enteredDate = ref("")

const pickedDateStr = computed(() => {
   return dayjs(pickedEndDate.value).format("YYYY-MM-DD")
})

const showTenYear = computed( () => {
   if ( props.admin) return true
   return ( props.program.includes('Creative Writing') && props.degree == 'MFA (Master of Fine Arts)' )
})
const tenYears = computed( () => {
   let d = new Date()
   d.setFullYear( d.getFullYear() + 10)
   return d
})

const datePicked = (() => {
   enteredDate.value = pickedDateStr.value
})

const show = (() => {
   isOpen.value = true
   let dateStr = props.endDate
   if ( dateStr.includes("T00:00:00Z") ) {
      dateStr = dateStr.split("T")[0]
   }
   pickedEndDate.value = dayjs(dateStr).toDate()
   enteredDate.value = dateStr
})

const okClicked = (() => {
   let dateStr = dayjs(pickedEndDate.value).format("YYYY-MM-DDTHH:mm:ss[Z]")
   emit("picked", dateStr)
   isOpen.value = false
})

const dateKeyPressed = ((event) => {
   if (event.keyCode == 13) {
      dateEntered()
   }
})

const dateEntered = (() => {
   pickedEndDate.value = dayjs(enteredDate.value).toDate()
})

const setEmbargoEndDate = ((count, type) => {
   pickedEndDate.value = new Date()
   if (type=="month") {
      pickedEndDate.value.setMonth( pickedEndDate.value.getMonth() + count)
   } else {
      pickedEndDate.value.setFullYear( pickedEndDate.value.getFullYear() + count)
   }
})

</script>

<style lang="scss" scoped>
.help {
   text-align: left;
   margin-bottom: 15px;
   white-space: break-spaces;
   max-width: 400px;
}
.datepick {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   gap: 10px;
   .dates {
      display: flex;
      flex-direction: column;
      gap: 10px;
      .setter {
         display: flex;
         flex-flow: row nowrap;
         gap: 5px;
      }
   }
}
.helpers {
   display: flex;
   flex-direction: column;
   gap: 5px;
}
.controls {
   display: flex;
   flex-direction: column;
   margin-top: 20px;
   padding-top: 20px;
   border-top: 1px solid $uva-grey-100;
   align-items: flex-start;
   gap: 10px;
   .buttons {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 5px;
      margin-top: 10px;
      width: 100%;
   }
}
</style>