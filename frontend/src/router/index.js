import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import OADashboard from '../views/OADashboard.vue'
import OAWorkForm from '../views/OAWorkForm.vue'
import ETDWorkForm from '../views/ETDWorkForm.vue'
import ETDDashboard from '../views/ETDDashboard.vue'
import Granted from '../views/Granted.vue'
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
         path: '/etd',
         name: 'edt',
         component: ETDDashboard
      },
      {
         // NOTE: this route is temporary as users can't create new ETD works; they are
         // autocreated by a back end process and users just edit them
         path: '/etd/new',
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
         path: '/granted',
         name: 'granted',
         component: Granted
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
            bar.classList.remove("sticky")
            resolve({left: 0, top: 0})
         }, 100)
      })
   }
})

router.beforeEach((to, _from, next) => {
   console.log("BEFORE ROUTE "+to.path)
   const userStore = useUserStore()
   if (to.path == '/granted') {
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
   } else if (to.name !== 'not_found' && to.name !== 'forbidden' && to.name !== "expired") {
      localStorage.setItem("prior_libra3_url", to.fullPath)
      let jwtStr = localStorage.getItem('libra3_jwt')
      console.log(`GOT JWT [${jwtStr}]`)
      if (jwtStr != null && jwtStr != "" && jwtStr != "null") {
         userStore.setJWT(jwtStr)
         next()
      } else {
         console.log("AUTHENTICATE")
         window.location.href = "/authenticate"
      }
   } else {
      next()
   }
})

export default router
