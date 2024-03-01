import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useRepositoryStore = defineStore('repository', {
   state: () => ({
      working: false,
      depositToken: "",
      oaWork: null,
      etdWork: null
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
            this.oaWork = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async depositETD( jsonPayload ) {
         this.working = true
         return axios.post(`/api/etd/${this.depositToken}`, jsonPayload).then(response => {
            this.etdWork = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})