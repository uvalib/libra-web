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
            <router-link v-if="isLibraETD" to="/etd"><img src="@/assets/LibraETD.svg"/></router-link>
            <router-link v-else to="/"><img src="@/assets/Libra.svg"/></router-link>
            <div class="sub">Online Archive of University of Virginia Scholarship</div>
         </div>
      </div>
      <div class="user-header" v-if="user.isSignedIn">
         <MenuBar />
      </div>
   </header>

   <main  >
      <RouterView  v-if="configuring == false"/>
      <h1 v-else style="min-height:">Authenticating...</h1>
   </main>

   <LibraryFooter />
   <ScrollTop />

   <Dialog v-model:visible="systemStore.showError" :modal="true" header="System Error" @hide="errorClosed()" class="error">
      {{systemStore.error}}
      <template #footer>
         <Button label="OK" autofocus class="p-button-secondary" @click="errorClosed()"/>
      </template>
   </Dialog>

</template>

<script setup>
import { onBeforeMount, ref, watch, computed } from 'vue'
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import LibraryFooter from "@/components/LibraryFooter.vue"
import MenuBar from "@/components/MenuBar.vue"
import { RouterView } from 'vue-router'
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import Dialog from 'primevue/dialog'
import Toast from 'primevue/toast'
import ScrollTop from 'primevue/scrolltop'
import { useToast } from "primevue/usetoast"
import { useRoute } from 'vue-router'

const toast = useToast()
const systemStore = useSystemStore()
const user = useUserStore()
const configuring = ref(true)
const route = useRoute()

const isLibraETD = computed( () => {
   return route.fullPath.indexOf("/etd") > -1
})

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
   await systemStore.getConfig()
   configuring.value = false
})

const errorClosed = (() => {
   systemStore.setError("")
   systemStore.showError = false
})

</script>

<style lang="scss">
html,
body {
   margin: 0;
   padding: 0;
   font-family: "franklin-gothic-urw", arial, sans-serif;
   -webkit-font-smoothing: antialiased;
   -moz-osx-font-smoothing: grayscale;
   color: $uva-grey-A;
   background: $uva-blue-alt-B;
}

#app {
   font-family: "franklin-gothic-urw", arial, sans-serif;
   -webkit-font-smoothing: antialiased;
   -moz-osx-font-smoothing: grayscale;
   text-align: center;
   color: $uva-text-color-base;
   margin: 0;
   padding: 0;
   background: #fafafa;
   outline: 0;
   border: 0;

   h1 {
      color: $uva-brand-blue;
      margin: 2rem 2.5rem;
      font-weight: 700;
      position: relative;
      text-align: center;
   }
   h2 {
      color: $uva-brand-blue;
      text-align: center;
      font-size: 1.3em;
      margin: 1rem 1.5rem;
   }
}

a {
   color: $uva-brand-blue-100;
   font-weight: 500;
   text-decoration: none;
   &:hover {
      text-decoration: underline;
      color: $uva-brand-blue-200;
   }
}

a:focus, input:focus, select:focus, textarea:focus, button.pool:focus, .pre-footer a:focus  {
   outline: 2px dotted $uva-brand-blue-100;
   outline-offset: 3px;
}
a:focus {
   border-radius: 0.3rem;
}
footer, div.header, nav.menu {
   a:focus {
      outline: 2px dotted $uva-grey-200;
      outline-offset: 3px;
   }
}

header {
   background-color: $uva-brand-blue;
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
      text-align: right;
      .sub {
         display: block;
         font-size: 0.9em;
      }
   }
}
</style>
