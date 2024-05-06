import { defineStore } from 'pinia'
import axios from 'axios'

export const useAuditStore = defineStore('audit', {
   state: () => ({
    working: false,
    workID: "",
    error: "",
    audits: [],
   }),

   actions: {
     async getAudits(id, namespace) {
       this.$reset
       this.working = true
       return axios.get(`/api/audits/${namespace}/${id}`).then(response => {
         this.setAudits(response.data)
         this.working = false
       }).catch( err => {

         this.error = err
         this.working = false
       })
     },
     setAudits(data) {
       this.audits = data
     }
   },
})