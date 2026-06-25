import { defineStore } from 'pinia'
import axios from 'axios'
import dayjs from 'dayjs'
import { useSystemStore } from './system'
import { useETDStore } from './etd'
import { useUserStore } from './user'
import { useCookies } from "vue3-cookies"

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      hits: [],
      total: 0,
      offset: 0,
      limit: 50,
      query: "",
      searchCompleted: false,
      sortField: "",
      sortOrder: "",
      statusFilter: "any",
      sourceFilter: "any",
      publishedFilter: { from: null, to: null},
      createdFilter:  { from: null, to: null},
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
      },
      createdFilterSet: state => {
         return state.createdFilter.to != null && state.createdFilter.from != null
      },
      isCreatedFilterValid: state => {
         return (
            state.createdFilter.to == null && state.createdFilter.from == null ||
            state.createdFilter.to != null && state.createdFilter.from != null
         )
      },
      publishedFilterSet: state => {
         return state.publishedFilter.to != null && state.publishedFilter.from != null
      },
      isPublishedFilterValid: state => {
         return (
            state.publishedFilter.to == null && state.publishedFilter.from == null ||
            state.publishedFilter.to != null && state.publishedFilter.from != null
         )
      },
   },
   actions: {
      resetSearch() {
         this.$reset()
         this.getRecentActivity()
      },
      getRecentActivity() {
         this.working = true
         let url = `/api/admin/search?q=${this.query}&offset=${this.offset}&limit=${this.limit}&recent=1`
         axios.get(url).then(response => {
            this.hits = response.data.hits
            this.total = response.data.total
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      clearPublishedFiter() {
         this.publishedFilter = { from: null, to: null}   
      },
      clearCreatedFiter() {
         this.createdFilter = { from: null, to: null}   
      },
      search() {
         const system = useSystemStore()
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
         if ( this.publishedFilterSet ) {
            const from = dayjs(this.publishedFilter.from).format("YYYY-MM-DD")
            const to = dayjs(this.publishedFilter.to).format("YYYY-MM-DD")
            url += `&published=${from} to ${to}`
         }
         if ( this.createdFilterSet ) {
            const from = dayjs(this.createdFilter.from).format("YYYY-MM-DD")
            const to = dayjs(this.createdFilter.to).format("YYYY-MM-DD")
            url += `&created=${from} to ${to}`
         }
         axios.get(url).then(response => {
            this.hits = response.data.hits
            this.total = response.data.total
            if (this.total > system.maxSearchHits ) {
               this.total = system.maxSearchHits
            }
            this.working = false
            this.searchCompleted = true
         }).catch( err => {
            console.error(err)
            system.setError(  err )
            this.working = false
            this.searchCompleted = false
         })
      },

      exportCSV() {
         // TODO just call search with &export={total}
         const system = useSystemStore()
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
         if ( this.publishedFilterSet ) {
            url += `&published=${this.publishedFilter.from} to ${this.publishedFilter.to}`
         }
         if ( this.createdFilterSet ) {
            url += `&created=${this.createdFilter.from} to ${this.createdFilter.to}`
         }
         url += `&export=${this.total}`
         axios.get(url).then(response => {
            const fileURL = window.URL.createObjectURL(new Blob([response.data], { type: 'application/vnd.ms-excel' }))
            const fileLink = document.createElement('a')
            fileLink.href =  fileURL
            fileLink.setAttribute('download', `libraetd-export.csv`)
            document.body.appendChild(fileLink)
            fileLink.click()
            window.URL.revokeObjectURL(fileURL)
            this.working = false
         }).catch((error) => {
            console.log(error)
            if (error.message) {
               useSystemStore().setError(error.message)
            } else {
               useSystemStore().setError(error)
            }
            this.working = false
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
            const { cookies } = useCookies()
            const jwtStr = cookies.get("libra3_impersonate_jwt")
            user.setJWT(jwtStr)
            const strData = JSON.stringify(this.impersonate)
            localStorage.setItem("libra3_impersonate", strData)
            this.router.push("/")
         }).catch( err => {
            console.error("Unable to impersonate user: "+err)
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
      },
      updateMimeTypes( types ) {
         axios.post(`/api/admin/mimetypes`, types).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })   
      }
   }
})