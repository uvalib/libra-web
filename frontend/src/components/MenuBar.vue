<template>
   <Menubar :model="libraMenu">
      <template #item="{ label, item, props }">
         <router-link v-if="item.route" :to="item.route">
            {{ label }}
         </router-link>
         <a v-else :href="item.url" :target="item.target" v-bind="props.action">
            <span v-bind="props.icon" />
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
import { useRouter, useRoute } from "vue-router"

const route = useRoute()
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
}
</style>

