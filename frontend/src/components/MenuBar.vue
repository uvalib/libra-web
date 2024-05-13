<template>
   <Menubar :model="libraMenu">
      <template #item="{ label, item, props }">
         <router-link v-if="item.route" :to="item.route">
            {{ label }}
         </router-link>
         <a v-else-if="!item.if" :href="item.url" class="flex-inline" :target="item.target" v-bind="props.action">
            <span v-bind="props.icon" class=""/>
            <img v-if="item.img" :src="item.img" />
            <span v-bind="props.label">{{ label }}</span>
            <span v-if="item.items" class="pi pi-fw pi-angle-down" v-bind="props.submenuicon" />
         </a>
      </template>
   </Menubar>
</template>

<script setup>
import Menubar from 'primevue/menubar'
import { computed } from 'vue'
import { useUserStore } from "@/stores/user"
import { useRouter} from "vue-router"
import ORCIDLogo from '@/assets/ORCID-iD_icon-vector.svg';


const router = useRouter()
const user = useUserStore()

const libraMenu = computed( () => {
   let menu = [
      {label: "Home", route: "/"},
      {label: "LibraETD", route: "/etd"},
      {label: "LibraOpen", route: "/oa"},
   ]
   if ( user.admin ) {
      menu.push({label: "Admin", route: "/admin"})
   }
   let userMenu =
      {label: `${user.firstName} ${user.lastName}`, items: [
         {  label: "ORCID",
            items: [
               {if: typeof(user.orcid.uri) === 'undefined', label: user.orcid.uri, img: ORCIDLogo, alt: "ORCID logo", target: "_blank", url: user.orcid.uri, icon: "pi pi-external-link"},
               {label: "Manage Connection", url: "https://orciddev.lib.virginia.edu", target: "_blank", icon:"pi pi-external-link"},
            ]
         },
         {label: "Sign out",  command: ()=>signOut()}
      ]}
   menu.push(userMenu)
   return menu
})

const signOut = (() => {
   user.signOut()
   router.push("/")
})

</script>

<style scoped lang="scss">
div.p-menuitem-content {
   a {
      color: var(--uvalib-text) !important;
      padding: 0.75rem 1rem !important;
      display: block;
      border-radius: 0;
      white-space: nowrap;
      &:hover {
         text-decoration: none !important;
      }
   }
   img {
      height: 1em
   }
}
</style>

