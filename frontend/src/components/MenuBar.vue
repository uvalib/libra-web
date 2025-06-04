<template>
   <Menubar :model="libraMenu"
      :pt="{
            item: {
               'aria-level': null
            }
         }"
      >
      <template #start v-if="user.admin == false && user.registrar == false">
         <a class="menu-link" href="mailto:libra@virginia.edu" target="_blank">
            <i class="pi pi-question-circle"></i>
            <span>Libra Support</span>
         </a>
      </template>
      <template #item="{ item, props, hasSubmenu }">
         <router-link v-if="item.route" v-slot="{ href, navigate }" :to="item.route" custom>
            <a :href="href" v-bind="props.action" @click="navigate">
               <span :class="item.icon" />
               <span>{{ item.label }}</span>
            </a>
         </router-link>
         <a v-else :href="item.url" :target="item.target" v-bind="props.action">
            <span :class="item.icon" v-if="item.icon" />
            <img :src="item.image" :alt="item.alt" v-if="item.image" style="width:25px;"/>
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
import { useRouter} from "vue-router"

const router = useRouter()
const user = useUserStore()

onMounted(()=>{
   user.getORCID()
})

const libraMenu = computed( () => {
   let menu = []
   if ( user.admin ) {
      menu.push({label: "Dashboard", route: "/admin", icon: "pi pi-home"})
   } else if ( user.registrar ) {
      menu.push({label: "Dashboard", route: "/register", icon: "pi pi-home"})
   } else {
      menu.push({label: "Dashboard", route: "/", icon: "pi pi-home"})
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
      // let orcidMenu = []
      // if ( user.orcid.id != "") {
      //     orcidMenu.push( {label: "Manage", url: "https://orciddev.lib.virginia.edu", target: "_blank", image:"./orcid_id.svg", alt:"ORCID logo"} )
      //    orcidMenu.push( {label: user.orcid.id, url:  user.orcid.uri, target: "_blank", image:"./orcid_id.svg", alt:"ORCID logo"} )
      // } else {
      //    orcidMenu.push( {label: "Register or connect", url: "https://orciddev.lib.virginia.edu", target: "_blank", image:"./orcid_id.svg", alt:"ORCID logo"} )
      // }

      let userMenu = { label: `${user.firstName} ${user.lastName}`, icon: "pi pi-user", items: [] }
      if ( user.orcid.id == "") {
          userMenu.items.push( {label: "Register or connect ORCID ID", url: "https://orciddev.lib.virginia.edu", target: "_blank", image:"./orcid_id.svg", alt:"ORCID logo"} )
      } else {
         userMenu.items.push( {label: "Manage ORCID ID", url: "https://orciddev.lib.virginia.edu", target: "_blank", icon: "pi pi-external-link"} )
         userMenu.items.push( {label: user.orcid.id, url:  user.orcid.uri, target: "_blank", image:"./orcid_id.svg", alt:"ORCID logo"} )
      }

      userMenu.items.push( { label: "Sign out", icon: "pi pi-sign-out", command: ()=>signOut()} )

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
a {
   display: flex;
   flex-flow: row nowrap;
   align-items: center;
   gap: 10px;
}
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

