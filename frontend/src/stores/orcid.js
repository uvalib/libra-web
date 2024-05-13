import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useOrcidStore = defineStore('orcid', {
   state: () => ({
      working: false,
      orcid: {},
   }),
   actions: {
      find(computeID) {
         this.working = true
         let url = `/api/users/orcid/${computeID}`
         axios.get(url).then(response => {
            this.orcid.id = response.data.orcid
            this.orcid.uri = response.data.uri
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
         return this.orcid
      }
    }
  })