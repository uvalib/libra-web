<template>
   <Menubar :model="libraMenu">
      <template #start v-if="user.admin == false && user.registrar == false">
         <a class="menu-link" href="mailto:libra@virginia.edu" target="_blank">Libra Support</a>
      </template>
      <template #item="{ item, props, hasSubmenu }">
         <router-link v-if="item.route" v-slot="{ href, navigate }" :to="item.route" custom>
            <a :href="href" v-bind="props.action" @click="navigate">
               <span :class="item.icon" />
               <span>{{ item.label }}</span>
            </a>
         </router-link>
         <a v-else :href="item.url" :target="item.target" v-bind="props.action">
            <span :class="item.icon" />
            <span>{{ item.label }}</span>
            <span v-if="hasSubmenu" class="pi pi-fw pi-angle-down" />
         </a>
      </template>
   </Menubar>
</template>

<script setup>
import Menubar from 'primevue/menubar'
import { computed, onMounted } from 'vue'
import { useUserStore } from "@/stores/user"
import { useOrcidStore } from "@/stores/orcid"
import { useRouter} from "vue-router"
import ORCIDLogo from '@/assets/ORCID-iD_icon-vector.svg'

const router = useRouter()
const user = useUserStore()
const orcid = useOrcidStore()

onMounted(()=>{
   orcid.find(user.computeID)
})

const libraMenu = computed( () => {
   let menu = []
   if ( user.admin ) {
      menu.push({label: "Dashboard", route: "/admin"})
   } else if ( user.registrar ) {
      menu.push({label: "Dashboard", route: "/register"})
   } else {
      menu.push({label: "Dashboard", route: "/"})
   }

   if ( user.admin || user.registrar) {
      let userMenu =
      {
         label: `${user.firstName} ${user.lastName}`, items: [
            {label: "Sign out",  command: ()=>signOut()}
         ]
      }
      menu.push(userMenu)
   } else {
      let orcidMenu = [
         {label: "Manage", url: "https://orciddev.lib.virginia.edu", target: "_blank", icon:"pi pi-external-link"},
         { visible: (orcid.userURI != "") , label: orcid.userURI, img: ORCIDLogo, alt: "ORCID logo", target: "_blank", url: orcid.userURI, icon: "pi pi-external-link"}
      ]
      let userMenu = {
         label: `${user.firstName} ${user.lastName}`, items: [
            { label: "ORCID", items: orcidMenu },
            { label: "Sign out",  command: ()=>signOut()}
         ]
      }

      menu.push(userMenu)
   }
   return menu
})

const signOut = (() => {
   user.signOut()
   router.push("/signedout")
})

</script>

<style scoped lang="scss">
.menu-link {
   color: $uva-text-color-dark;
   border-radius: 0.4rem;
   padding: 0.5rem 0.75rem;
   &:hover {
      text-decoration: none;
      background: $uva-brand-blue-200;
      color: white;
   }
}
</style>

