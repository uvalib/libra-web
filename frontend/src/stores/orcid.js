import { defineStore } from 'pinia'
import axios from 'axios'
import { useUserStore } from './user'

export const useOrcidStore = defineStore('orcid', {
   state: () => ({
      working: false,
      orcids: [],
      userURI: "",
      refresh: 5
   }),
   actions: {
      find(computeID) {
         this.working = true

         // check if we have this ID already

         if (this.orcids[computeID] && (this.orcids[computeID].timestamp + (this.refresh*60) < Date.now()) ) {
            console.log(`Using cached ORCID for user ${computeID}` )
            return this.orcids[computeID]
         }

         let url = `/api/users/orcid/${computeID}`
         axios.get(url).then(response => {
            this.orcids[computeID] = response.data
            this.orcids[computeID].timestamp = Date.now()

            let user = useUserStore()

            if (user.computeID == computeID) {
               this.userURI = this.orcids[computeID].uri
            }
            console.log(`Found ORCID for user ${computeID}: ${this.userURI}` )

            this.working = false
         }).catch( err => {
            if (err.response.status != 404) {
               console.log(err)
            }
            this.working = false
            return {}
         })
         return this.orcids[computeID] || {}
      }
    }
  })