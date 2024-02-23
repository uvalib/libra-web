import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useRepositoryStore = defineStore('repository', {
   state: () => ({
      depositToken: "",
   }),
   actions: {
      async getDepositToken() {
         this.depositToken = ""
         return axios.get("/api/token").then(response => {
            this.depositToken = response.data
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },
      cancel() {
         axios.post(`/api/cancel/${this.depositToken}`)
         this.depositToken = ""
      }
   }
})