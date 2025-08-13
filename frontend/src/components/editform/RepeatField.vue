<template>
    <Fieldset :legend="props.title">
      <div v-if="props.help" class="note">{{ props.help }}</div>
      <div class="control-group">
         <label :for="props.name">{{ props.label }}</label>
         <InputText type="text" :id="props.name" :name="props.name" v-model="newValue" @keyup.enter="addValue" fluid />
         <Button label="Add" severity="secondary" @click="addValue" />
      </div>
      <div class="list">
         <div class="note" v-if="model.length == 0">None</div>
          <div v-for="(k,idx) in model" class="value">
            <span @click="removeValue(idx)">{{ k }}</span>
            <Button icon="pi pi-times" severity="danger" rounded small :aria-label="`remove ${k}`" @click="removeValue(idx)"/>
         </div>
      </div>
    </Fieldset>
</template>

<script setup>
import { ref } from 'vue'
import InputText from 'primevue/inputtext'
import Fieldset from 'primevue/fieldset'

const model = defineModel()

const emit = defineEmits(['change'])

const props = defineProps({
   title: {
      type: String,
      required: true,
   },
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
   document.getElementById(props.name).focus()
   emit('change')
})

const addValue = (() => {
   if (newValue.value.length > 0 ) {
      model.value.push(newValue.value)
      newValue.value = ""
      emit('change')
   }
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
   align-items: center;
   gap: 5px;
}
.note {
   font-style: italic;
   color: $uva-grey-A;
   margin: 0 0 5px 0;
}
label {
   display: inline-block;
   margin-right: 5px;
}
.value {
   display: flex;
   flex-flow: row nowrap;
   gap: 8px;
   align-items: center;
   border: 1px solid $uva-grey-100;
   border-radius: 50px;
   padding: 6px 6px 6px 12px;
   background-color: #fafafa;
   cursor: default;
   &:hover {
      background-color: #f5f5f5;
   }

   button {
      width: 24px;
      height: 24px;
   }
}
</style>