import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import AdminDashboard from '../views/AdminDashboard.vue'
import ETDPublicView from '../views/ETDPublicView.vue'
import ETDWorkForm from '../views/ETDWorkForm.vue'
import ETDDashboard from '../views/ETDDashboard.vue'
import Expired from '../views/Expired.vue'
import ForbiddenView from '../views/ForbiddenView.vue'
import NotFound from '../views/NotFound.vue'
import VueCookies from 'vue-cookies'
import { useUserStore } from '@/stores/user'

const router = createRouter({
   history: createWebHistory(import.meta.env.BASE_URL),
   routes: [
      {
         path: '/',
         name: 'home',
         component: HomeView
      },
      {
         path: '/admin',
         name: 'admin',
         component: AdminDashboard
      },
      {
         path: '/etd',
         name: 'libraetd',
         component: ETDDashboard
      },
      {
         path: '/etd/:id',
         alias: '/admin/etd/:id',
         name: 'edtworkform',
         component: ETDWorkForm
      },
      {
         path: '/public/etd/:id',
         name: 'etdpublic',
         component: ETDPublicView
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
   const noAuthRoutes = ["not_found", "forbidden", "expired", "home"]
   const isPublicPage = (to.name == "etdpublic" )
   const isAdminPage = (to.name == "admin" || to.path.includes("/admin"))

   // the /signedin endpoint called after authorization. it has no page itself; it just
   // processes the authorization response and redirects to the next page (or forbidden)
   if (to.path == '/signedin') {
      const jwtStr = VueCookies.get("libra3_jwt")
      userStore.setJWT(jwtStr)
      if ( userStore.isSignedIn  ) {
         console.log(`GRANTED [${jwtStr}]`)
         let priorURL = localStorage.getItem('prior_libra3_url')
         localStorage.removeItem("prior_libra3_url")
         if ( priorURL && priorURL != "/signedin" && priorURL != "/") {
            console.log("RESTORE "+priorURL)
            return {path: priorURL}
         }
         return {name: "home"}
      }
      return {name: "forbidden"}
   }

   // for all other routes, pull the existing jwt from storage from storage and set in the user store.
   // depending upon the page resuested, this token may or may not be used.
   const jwtStr = localStorage.getItem('libra3_jwt')
   userStore.setJWT(jwtStr)

   if ( noAuthRoutes.includes(to.name)) {
      console.log("NOT A PROTECTED PAGE")
   } else {
      if (userStore.isSignedIn == false) {
         if ( isPublicPage == false ) {
            console.log("AUTHENTICATE")
            localStorage.setItem("prior_libra3_url", to.fullPath)
            window.location.href = "/authenticate"
            return false   // cancel the original navigation
         }
         console.log("UNAUTHENTICATED REQUEST FOR PUBLIC VIEW PAGE")
      } else {
         if ( isAdminPage) {
            console.log(`REQUEST ADMIN PAGE WITH JWT`)
            if ( userStore.admin == false ) {
               console.log("REJECT NON-ADMIN REQUEST FOR ADMIN PAGES")
               return {name: "forbidden"}
            }
         } else {
            console.log(`REQUEST AUTHENTICATED PAGE WITH JWT`)
         }
      }

      // this page uses the auth token. be sure it is still valid before proceeding
      await userStore.validateAuth()
      if (userStore.isSignedIn == false) {
         console.log("JWT HAS EXPIRED")
         localStorage.setItem("prior_libra3_url", to.fullPath)
         return {name: "expired"}
      }
   }
})

export default router
