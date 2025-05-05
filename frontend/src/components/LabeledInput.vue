<template>
   <label>
      <div class="text">
         <span>{{ props.label }}</span>
         <span v-if="required" class="required"><span class="star">*</span>(required)</span>
      </div>
      <InputText v-if="props.type=='text'" v-model="model" :name="props.name" fluid/>
      <Textarea v-if="props.type=='textarea'" v-model="model" :name="props.name" fluid rows="10" />
      <Select v-if="props.type=='select'" :options="props.options" optionLabel="label" optionValue="value"
         v-model="model" :name="props.name" :placeholder="`Select ${props.label}`" />
   </label>
</template>

<script setup>
const model = defineModel()
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'

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
})
</script>

<style lang="scss" scoped>
label {
   display: flex;
   flex-direction: column;
   gap: 5px;
}
</style>