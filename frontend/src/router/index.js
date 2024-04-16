import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import AdminDashboard from '../views/AdminDashboard.vue'
import OADashboard from '../views/OADashboard.vue'
import OAWorkForm from '../views/OAWorkForm.vue'
import OAPublicView from '../views/OAPublicView.vue'
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
         path: '/admin/etd/:id',
         name: 'edtworkform',
         component: ETDWorkForm
      },
      {
         path: '/admin/oa/:id',
         name: 'oaworkform',
         component: OAWorkForm
      },
      {
         path: '/etd',
         name: 'edt',
         component: ETDDashboard
      },
      {
         path: '/etd/new',
         name: 'edtworkform',
         component: ETDWorkForm
      },
      {
         path: '/etd/:id',
         name: 'edtworkform',
         component: ETDWorkForm
      },
      {
         path: '/oa',
         name: 'open',
         component: OADashboard
      },
      {
         path: '/oa/new',
         name: 'openworkform',
         component: OAWorkForm
      },
      {
         path: '/oa/:id',
         name: 'openworkform',
         component: OAWorkForm
      },
      {
         path: '/public/oa/:id',
         name: 'oapublic',
         component: OAPublicView
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
   scrollBehavior(to, _from, _savedPosition) {
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

router.beforeEach((to, _from, next) => {
   console.log("BEFORE ROUTE "+to.path+": "+to.name)
   const userStore = useUserStore()

   // the /signedin endpoint called after authorization. it has no page itself; it just
   // processes the authorization response and redirects to the next page
   if (to.path == '/signedin') {
      let jwtStr = VueCookies.get("libra3_jwt")
      console.log(`GRANTED [${jwtStr}]`)
      if (jwtStr != null && jwtStr != "" && jwtStr != "null") {
         userStore.setJWT(jwtStr)
         let priorURL = localStorage.getItem('prior_libra3_url')
         localStorage.removeItem("prior_libra3_url")
         if ( priorURL && priorURL != "/granted" && priorURL != "/") {
            console.log("RESTORE "+priorURL)
            next(priorURL)
         } else {
            next("/")
         }
      } else {
         next("/forbidden")
      }
      return
   }

   // request for home page or public work metadata. authorization is not required, but if it is present, it will be used
   if (to.name == "oapublic" || to.name == "etdpublic" ||  to.name == "home") {
      console.log("PUBLIC REQUEST; AUTH OPTIONAL")
      let jwtStr = localStorage.getItem('libra3_jwt')
      if (jwtStr != null && jwtStr != "" && jwtStr != "null") {
         console.log("    AUTH PRESENT")
         userStore.setJWT(jwtStr)
      } else {
         console.log("    NO AUTH PRESENT")
      }
      next()
      return
   }

   // some routes are accessible by all. do not check auth and just proceed to the route
   let publicRoutes = ["not_found", "forbidden", "expired"]
   if ( publicRoutes.includes(to.name)) {
      console.log("NOT A PROTECTED ROUTE")
      next()
      return
   }

   // all other routes require jwt authorization
   console.log("AUTH REQUIRED")
   localStorage.setItem("prior_libra3_url", to.fullPath)
   let jwtStr = localStorage.getItem('libra3_jwt')
   console.log(`GOT JWT [${jwtStr}]`)
   if (jwtStr != null && jwtStr != "" && jwtStr != "null") {
      userStore.setJWT(jwtStr)
      if (to.name == "admin" ) {
         if ( userStore.admin == false ) {
            console.log("   REJECT NON-ADMIN REQUEST FOR ADMIN PAGES")
            next("/forbidden")
            return
         }
         console.log("    ADMIN REQUEST GRANTED")
      }
      next()
   } else {
      console.log("AUTHENTICATE")
      window.location.href = "/authenticate"
   }
})

export default router
