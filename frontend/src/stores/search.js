import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useSearchStore = defineStore('search', {
   state: () => ({
      working: false,
      hits: [],
   }),
   actions: {
      search(type, computeID, title="") {
         this.working = true
         let url = `/api/works/search?type=${type}`
         if (computeID != "") {
            url += `&cid=${computeID}`
         }
         if (title != "") {
            url += `&title=${title}`
         }
         axios.get(url).then(response => {
            this.hits = response.data
            console.log(this.hits)
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      removeDeletedWork( id ) {
         let idx = this.hits.findIndex( h => h.id == id)
         this.hits.splice(idx,1)
      }
   }
})