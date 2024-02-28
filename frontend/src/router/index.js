import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import OpenDashboard from '../views/OpenDashboard.vue'
import CreateOpenWork from '../views/CreateOpenWork.vue'
import ETDDashboard from '../views/ETDDashboard.vue'
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
         path: '/oa',
         name: 'open',
         component: OpenDashboard
      },
      {
         path: '/oa/new',
         name: 'newopen',
         component: CreateOpenWork
      },
      {
         path: '/expired',
         name: "expired",
         component: () => import('../views/Expired.vue')
      },
      {
         path: '/forbidden',
         name: "forbidden",
         component: () => import('../views/ForbiddenView.vue')
      },
      {
         path: '/:pathMatch(.*)*',
         name: "not_found",
         component: () => import('../views/NotFound.vue')
      }
   ],
   scrollBehavior(_to, _from, _savedPosition) {
      return new Promise(resolve => {
         setTimeout( () => {
            resolve({left: 0, top: 0})
         }, 100)
      })
   }
})

router.beforeEach((to, _from, next) => {
   const userStore = useUserStore()
   if (to.path === '/granted') {
      let jwtStr = VueCookies.get("libra3_jwt")
      userStore.setJWT(jwtStr)
      let priorURL = localStorage.getItem('prior_libra3_url')
      localStorage.removeItem("prior_libra3_url")
      if ( priorURL && priorURL != "/granted" && priorURL != "/") {
         console.log("RESTORE "+priorURL)
         next(priorURL)
      } else {
         next("/")
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
