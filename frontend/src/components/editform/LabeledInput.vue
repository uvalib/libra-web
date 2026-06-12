<template>
   <label>
      <div class="text">
         <span class="label">{{ props.label }}</span>
         <span v-if="required && !props.readonly" class="required"><span class="star">*</span>(required)</span>
      </div>
      <InputText v-if="props.type=='text'" v-model="model" :name="props.name" fluid :readonly="props.readonly" @value-change="emit('change')"/>
      <Textarea v-if="props.type=='textarea'" v-model="model" :name="props.name" fluid rows="10" :readonly="props.readonly" @value-change="emit('change')"/>
      <Select v-if="props.type=='select'" :options="props.options" optionLabel="label" optionValue="value"
         v-model="model" :name="props.name" :placeholder="`Select ${props.label}`" :readonly="props.readonly" @value-change="emit('change')"/>
   </label>
</template>

<script setup>
import InputText from 'primevue/inputtext'
import Textarea from 'primevue/textarea'
import Select from 'primevue/select'

const model = defineModel()

const emit = defineEmits(['change'])

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
   .label {
      font-weight: bold;
   }
}
</style>