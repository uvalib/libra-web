<template>
   <label>
      <div class="text">
         <span>{{ props.label }}</span>
         <span v-if="required && !props.readonly" class="required"><span class="star">*</span>(required)</span>
      </div>
      <InputText v-if="props.type=='text'" v-model="model" :name="props.name" fluid :readonly="props.readonly"/>
      <Textarea v-if="props.type=='textarea'" v-model="model" :name="props.name" fluid rows="10" :readonly="props.readonly"/>
      <select v-if="props.type=='select'" v-model="model" >
         <option v-for="o in options" :value="o.value">{{ o.label }}</option>
      </select>
   </label>
</template>

<script setup>
const model = defineModel()
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'

const props = defineProps({
   label: {
      type: String,
      required: true
   },
   name: {
      type: String,
      required: true
   },
   type: {
      type: String,
      default: "text"
   },
   options: {
      type: Array,
      default: []
   },
   required: {
      type: Boolean,
      default: false
   },
   readonly: {
      type: Boolean,
      default: false
   }
})
</script>

<style lang="scss" scoped>
label {
   display: flex;
   flex-direction: column;
   gap: 5px;
}
</style>