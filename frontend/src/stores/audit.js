import { defineStore } from 'pinia'
import axios from 'axios'

export const useAuditStore = defineStore('audit', {
   state: () => ({
      working: false,
      error: false,
      audits: [],
   }),

   actions: {
      getAudits(id) {
         this.error = false
         this.working = true
         axios.get(`/api/audits/${id}`).then(response => {
            this.audits = response.data
            this.working = false
         }).catch(err => {
            console.log("GET AUDITS FAILED:")
            console.log(err)
            this.error = true
            this.working = false
         })
      },
   },
})