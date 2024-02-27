import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useRepositoryStore = defineStore('repository', {
   state: () => ({
      working: false,
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
      },
      removeFile( file) {
         axios.delete(`/api/${this.depositToken}/${file}`)
      } ,
      async depositOA( jsonPayload ) {
         this.working = true
         return axios.post(`/api/oa/${this.depositToken}`, jsonPayload).then(response => {
            this.depositToken = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})