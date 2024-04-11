import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import { FilterMatchMode } from 'primevue/api'

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      scope: "etd",
      search: {
         computeID: ""
      },
      filters: {global: { value: null, matchMode: FilterMatchMode.CONTAINS }},
      hits: [],
   }),
   actions: {
      search() {
         this.working = true
         let url = `/api/admin/search?cid=${this.search.computeID}`
         axios.get(url).then(response => {
            this.hits = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      resetSearch() {
         this.$reset
      },
      async delete(type, id) {
         this.working = true
         return axios.delete(`/api/admin/${type}/${id}`).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})