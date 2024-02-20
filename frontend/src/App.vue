<template>
   <Toast position="top-center" />
   <ConfirmDialog position="top"/>

   <header id="libra-header">
      <div class="main-header">
         <div class="library-link">
            <a target="_blank" href="https://library.virginia.edu">
               <UvaLibraryLogo />
            </a>
         </div>
         <div class="site-link">
            <router-link to="/">Libra</router-link>
         </div>
      </div>
      <div class="user-header">
         <span v-if="user.isSignedIn">{{ user.displayName }}</span>
         <span v-else>Not signed in</span>
      </div>
   </header>

   <RouterView v-if="configuring==false" />

   <LibraryFooter />

   <Dialog v-model:visible="systemStore.showError" :modal="true" header="System Error" @hide="errorClosed()" class="error">
      {{systemStore.error}}
      <template #footer>
         <Button label="OK" autofocus class="p-button-secondary" @click="errorClosed()"/>
      </template>
   </Dialog>

</template>

<script setup>
import { onBeforeMount, ref, watch } from 'vue'
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import LibraryFooter from "@/components/LibraryFooter.vue"
import { RouterView } from 'vue-router'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import Dialog from 'primevue/dialog'
import Toast from 'primevue/toast'
import { useToast } from "primevue/usetoast"

const toast = useToast()
const systemStore = useSystemStore()
const user = useUserStore()
const configuring = ref(true)

watch(() => systemStore.toast.show, (newShow) => {
   if ( newShow == true) {
      if ( systemStore.toast.error) {
         toast.add({severity:'error', summary:  systemStore.toast.summary, detail:  systemStore.toast.message, life: 10000})
      } else {
         toast.add({severity:'success', summary:  systemStore.toast.summary, detail:  systemStore.toast.message, life: 5000})
      }
      systemStore.clearToastMessage()
   }
})


onBeforeMount( async () => {
   document.title = `Libra`
   await systemStore.getVersion()
   configuring.value = false
})

const errorClosed = (() => {
   systemStore.setError("")
   systemStore.showError = false
})

</script>

<style scoped>
header {
   background-color: var(--uvalib-brand-blue);
   color: white;
   text-align: left;
   position: relative;
   box-sizing: border-box;
   .main-header {
      display: flex;
      flex-direction: row;
      flex-wrap: nowrap;
      justify-content: space-between;
      align-content: stretch;
      align-items: center;
      padding: 1vw 20px 5px 10px;
      a {
         color: white !important;
      }
   }
   p.version {
      margin: 5px 0 0 0;
      font-size: 0.5em;
      text-align: right;
   }
   div.library-link {
      width: 250px;
      order: 0;
      flex: 0 1 auto;
      align-self: flex-start;
   }
   div.site-link {
      order: 0;
      font-size: 1.5em;
      a {
         color: white;
         text-decoration: none;
         &:hover {
            text-decoration: underline;
         }
      }
   }
   .user-header {
      background-color: var(--uvalib-blue-alt-darkest);
      color: white;
      text-align: right;
      padding: 10px;
   }
}
</style>
