import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import { FilterMatchMode } from 'primevue/api'

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      hits: [],
   }),
   actions: {
      clearAll() {
         this.hits = []
      },
      async addRegistrations( department, degree, users ) {
         let students = []
         users.forEach( u => {
            u.department = department
            students.push(u)
         })
         return axios.post(`/api/admin/register`, {department: department, degree: degree, students: students}).catch( err => {
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

      async delete(type, id) {
         this.working = true
         return axios.delete(`/api/admin/${type}/${id}`).then( ()=> {
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