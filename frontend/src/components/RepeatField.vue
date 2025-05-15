<template>
    <Fieldset :legend="props.label">
      <div class="list">
         <div class="note" v-if="model.length == 0">None</div>
         <Chip v-for="(k,idx) in model" removable @remove="removeValue(idx)" :label="k" />
      </div>
      <div class="control-group">
         <InputText type="text" :name="name" v-model="newValue" fluid />
         <Button label="Add" severity="secondary" @click="addValue"/>
      </div>
      <div v-if="props.help" class="note">{{ props.help }}</div>
    </Fieldset>
</template>

<script setup>
import { ref } from 'vue'
import InputText from 'primevue/inputtext'
import Fieldset from 'primevue/fieldset'
import Chip from 'primevue/chip'

const model = defineModel()

const emit = defineEmits(['change'])

const props = defineProps({
   label: {
      type: String,
      required: true,
   },
   help: {
      type: String,
      default: ""
   },
   name: {
      type: String,
      required: true
   }
})
const newValue = ref("")

const removeValue = ((idx) => {
   model.value.splice(idx,1)
   emit('change')
})

const addValue = (() => {
   model.value.push(newValue.value)
   newValue.value = ""
   emit('change')
})

</script>

<style lang="scss" scoped>
 .list {
   border: 1px solid $uva-grey-200;
   border-radius: 0.3rem;
   padding: 0.5rem 0.75rem;
   display: flex;
   flex-flow: row wrap;
   gap: 5px;
}
.control-group {
   display: flex;
   flex-flow: row nowrap;
   gap: 5px;
}
.note {
   font-style: italic;
   color: $uva-grey;
   margin-top: 0;
}
</style>