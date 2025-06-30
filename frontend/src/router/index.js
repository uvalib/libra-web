import { createRouter, createWebHistory } from 'vue-router'
import AdminDashboard from '../views/AdminDashboard.vue'
import PublicView from '../views/PublicView.vue'
import EditWorkForm from '../views/EditWorkForm.vue'
import UserDashboard from '../views/UserDashboard.vue'
import RegistrationForm from '../views/RegistrationForm.vue'
import SignedOut from '../views/SignedOut.vue'
import Expired from '../views/Expired.vue'
import ForbiddenView from '../views/ForbiddenView.vue'
import NotFound from '../views/NotFound.vue'
import VueCookies from 'vue-cookies'
import { useUserStore } from '@/stores/user'
import { useAdminStore } from '@/stores/admin'
import { useToast } from "primevue/usetoast"

const router = createRouter({
   history: createWebHistory(import.meta.env.BASE_URL),
   routes: [
      {
         path: '/',
         alias: '/dashboard',
         name: 'dashboard',
         component: UserDashboard
      },
      {
         path: '/admin',
         name: 'admin',
         component: AdminDashboard,
         meta: { requiresAdmin: true }
      },
      {
         path: '/register',
         name: 'register',
         component: RegistrationForm,
         meta: { requiresRegistrar: true }
      },
      {
         path: '/etd/:id',
         name: 'edtworkform',
         component: EditWorkForm,
      },
       {
         path: '/admin/etd/:id',
         name: 'adminworkform',
         component: EditWorkForm,
         meta: { requiresAdmin: true }
      },
      {
         path: '/public/etd/:id',
         name: 'etdpublic',
         component: PublicView
      },
      {
         path: '/signedout',
         name: "signedout",
         component: SignedOut
      },
      {
         path: '/expired',
         name: "expired",
         component: Expired
      },
      {
         path: '/forbidden',
         name: "forbidden",
         component: ForbiddenView
      },
      {
         path: '/:pathMatch(.*)*',
         name: "not_found",
         component: NotFound
      }
   ],
   scrollBehavior(_to, _from, _savedPosition) {
      return new Promise(resolve => {
         setTimeout( () => {
            let bar = document.getElementsByClassName("user-header")[0]
            if ( bar ) {
               bar.classList.remove("sticky")
            }
            resolve({left: 0, top: 0})
         }, 100)
      })
   }
})

router.beforeEach( async (to) => {
   console.log("BEFORE ROUTE "+to.path)
   const userStore = useUserStore()
   const noAuthRoutes = ["not_found", "forbidden", "expired", "etdpublic", "signedout"]

   // close any lobgering toast messages
   useToast().removeAllGroups()

   // the /signedin endpoint called after authorization. it has no page itself; it just
   // processes the authorization response and redirects to the next page (or forbidden)
   if (to.path == '/signedin') {
      const jwtStr = VueCookies.get("libra3_jwt")
      userStore.setJWT(jwtStr)
      if ( userStore.isSignedIn  ) {
         console.log(`GRANTED [${jwtStr}]`)
         return userStore.homePage
      }
      return {name: "forbidden"}
   }

   // for all other routes, pull the existing jwt from storage from storage and set in the user store.
   // depending upon the page requested, this token may or may not be used.
   const jwtStr = localStorage.getItem('libra3_jwt')
   userStore.setJWT(jwtStr)

   // see if there is existing impersonate stored locally (in case browser was refreshed)
   const impersonateStr = localStorage.getItem("libra3_impersonate")
   const impersonateData = JSON.parse(impersonateStr)
   if (impersonateData) {
      console.log("cached impersonate data found")
      if (impersonateData.userID == userStore.computeID) {
         useAdminStore().impersonate = impersonateData
      } else {
         console.log("impersonate data fis mismatched; removing it")
         localStorage.removeItem("libra3_impersonate")
      }
   }

   // public view requires no auth, but will use it if present
   if ( noAuthRoutes.includes(to.name)) {
      console.log("NOT A PROTECTED PAGE")
   } else {
      // force authentication for all other pages
      if ( userStore.isSignedIn == false) {
         console.log("AUTHENTICATE")
         window.location.href = "/authenticate"
         return false   // cancel the original navigation
      }

      if ( to.meta.requiresAdmin ) {
         console.log(`REQUEST ADMIN PAGE WITH JWT`)
         if ( userStore.isAdmin == false ) {
            console.log("REJECT NON-ADMIN REQUEST FOR ADMIN PAGES")
            return {name: "forbidden"}
         }
      } else if ( to.meta.requiresRegistrar ) {
         console.log(`REQUEST REGISTER PAGE WITH JWT`)
         if ( userStore.isRegistrar == false && userStore.isAdmin == false) {
            console.log("REJECT NON-REGISTRAR REQUEST FOR REGISTER PAGE")
            return {name: "forbidden"}
         }
      }

      // this page uses the auth token. be sure it is still valid before proceeding
      await userStore.validateAuth()
      if (userStore.isSignedIn == false) {
         console.log("JWT HAS EXPIRED")
         return {name: "expired"}
      }
   }
})

export default router
