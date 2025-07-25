import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import { useETDStore } from './etd'
import { useUserStore } from './user'
import VueCookies from 'vue-cookies'

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      hits: [],
      total: 0,
      offset: 0,
      limit: 20,
      query: "",
      searchCompleted: false,
      sortField: "",
      sortOrder: "",
      statusFilter: "any",
      sourceFilter: "any",
      impersonate: {
         adminJWT: "",
         userID: "",
         adminID: "",
         active: false
      }
   }),
   getters: {
      isImpersonating: state => {
         return state.impersonate.active
      },
      originalAdminID: state => {
         return state.impersonate.adminID
      }
   },
   actions: {
      resetSearch() {
         this.$reset()
      },
      async addRegistrations( program, degree, students ) {
         return axios.post(`/api/register`, {program: program, degree: degree, students: students}).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },

      search() {
         this.working = true
         let url = `/api/admin/search?q=${this.query}&offset=${this.offset}&limit=${this.limit}`
         if ( this.sortField != "" ) {
            url += `&sort=${this.sortField}&order=${this.sortOrder}`
         }
         if ( this.statusFilter != "any") {
            url += `&draft=${this.statusFilter == 'draft'}`
         }
         if ( this.sourceFilter != "any") {
            url += `&source=${this.sourceFilter}`
         }
         axios.get(url).then(response => {
            this.hits = response.data.hits
            this.total = response.data.total
            this.working = false
            this.searchCompleted = true
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
            this.searchCompleted = false
         })
      },

      becomeUser(tgtID) {
         let user = useUserStore()
         if ( user.isAdmin == false ) return

        this.impersonate.adminJWT = user.jwt
        this.impersonate.adminID = user.computeID
        this.impersonate.userID = tgtID
        this.impersonate.active = true

         axios.post(`/api/admin/impersonate/${tgtID}`).then(() => {
            const jwtStr = VueCookies.get("libra3_impersonate_jwt")
            user.setJWT(jwtStr)
            const strData = JSON.stringify(this.impersonate)
            localStorage.setItem("libra3_impersonate", strData)
            this.router.push("/")
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  `Unable to impersonate user '${tgtID}'` )
             this.endImpersonate()
         })
      },
      endImpersonate() {
         let user = useUserStore()
         user.setJWT(this.impersonate.adminJWT)
         this.impersonate.adminJWT = ""
         this.impersonate.adminID = ""
         this.impersonate.userID = ""
         this.impersonate.active = false
         localStorage.removeItem("libra3_impersonate")
         this.router.push( user.homePage )
      },

      unpublish(id) {
         axios.delete(`/api/admin/works/${id}/publish`).then(() => {
            let hit = this.hits.find( h=> h.id == id)
            if ( hit ) delete hit.publishedAt
            useETDStore().publishedAt = null

         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },

      delete(id) {
         this.working = true
         axios.delete(`/api/admin/works/${id}`).then( ()=> {
            let idx = this.hits.findIndex( h=> h.id == id)
            if (idx > -1) {
               this.hits.splice(idx,1)
            }
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})