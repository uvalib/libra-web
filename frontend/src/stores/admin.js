import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import { useETDStore } from './etd'

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      hits: [],
   }),
   actions: {
      clearAll() {
         this.hits = []
      },
      async addRegistrations( program, degree, students ) {
         return axios.post(`/api/admin/register`, {program: program, degree: degree, students: students}).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },

      search( computeID ) {
         this.working = true
         let url = `/api/admin/search?cid=${computeID}`
         axios.get(url).then(response => {
            this.hits = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
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