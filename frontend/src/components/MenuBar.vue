<template>
   <Menubar :model="libraMenu" aria-label="main">
      <template #start v-if="user.isAdmin == false && user.isRegistrar == false">
         <a class="menu-link" href="mailto:libra@virginia.edu" target="_blank" aria-describedby="new-window">
            <i class="pi pi-question-circle" aria-hidden="true"></i>
            <span>Libra Support</span>
         </a>
      </template>
       <template #end v-if="user.isAdmin">
         <span class='admin'><i class="pi pi-star"></i>Administrator</span>
       </template>
   </Menubar>
</template>

<script setup>
import Menubar from 'primevue/menubar'
import { computed, onMounted } from 'vue'
import { useUserStore } from "@/stores/user"
import { useSystemStore } from "@/stores/system"
import { useRouter} from "vue-router"

const router = useRouter()
const user = useUserStore()
const system = useSystemStore()

onMounted(()=>{
   user.getORCID()
})

const libraMenu = computed( () => {
   let menu = []
   if ( user.isAdmin ) {
      menu.push({label: "Dashboard", icon: "pi pi-home", command: ()=>router.push("/admin")})
      menu.push({label: "Deposit Registration", icon: "pi pi-user-plus", command: ()=>router.push("/register")})
      menu.push({label: "Deposit Status", icon: "pi pi-calendar-clock", command: ()=>router.push("/status")})
   } else if ( user.isRegistrar ) {
      menu.push({label: "Dashboard", icon: "pi pi-home", command: ()=>router.push("/register")})
      menu.push({label: "Deposit Status", icon: "pi pi-calendar-clock", command: ()=>router.push("/status")})
   } else {
      menu.push({label: "Dashboard", icon: "pi pi-home", command: ()=>router.push("/")})
   }

   if ( user.isAdmin || user.isRegistrar) {
      let userMenu =
      {
         label: `${user.firstName} ${user.lastName}`, icon: "pi pi-user", items: [
            {label: "Sign out", icon: "pi pi-sign-out",  command: ()=>signOut()}
         ]
      }
      menu.push(userMenu)
   } else {
      let userMenu = { label: `${user.firstName} ${user.lastName}`, icon: "pi pi-user", items: [] }
      if ( user.orcid.id == "") {
          userMenu.items.push( {label: "Register or connect ORCID ID", url: system.orcidURL, target: "_blank"} )
      } else {
         userMenu.items.push( {label: "Manage ORCID ID", url: system.orcidURL, target: "_blank", icon: "pi pi-external-link"} )
         userMenu.items.push( {label: user.orcid.id, url:  user.orcid.uri, target: "_blank"} )
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
.admin {
   padding: 0.5rem 0.75rem;
   background: $uva-grey-A;
   color: white;
   display: flex;
   flex-flow: row nowrap;
   gap: 8px;
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

