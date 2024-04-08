import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import { FilterMatchMode } from 'primevue/api'

export const useAdminStore = defineStore('admin', {
   state: () => ({
      working: false,
      scope: "etd",
      filters: {global: { value: null, matchMode: FilterMatchMode.CONTAINS }},
      hits: [],
      scopes: [
         {label: "All Works", value: "all"},
         {label: "SIS Works", value: "etd"},
         {label: "Optional Works", value: "oa"}
      ]
   }),
   actions: {
      search() {
         this.working = true
         let url = `/api/works/search?type=${this.scope}`
         console.log(url)
         axios.get(url).then(response => {
            this.hits = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
   }
})