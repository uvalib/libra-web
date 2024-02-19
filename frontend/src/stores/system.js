import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
   state: () => ({
      working: false,
		version: "unknown",
      error: "",
      showError: false,
      toast: {
         error: false,
         summary: "",
         message: "",
         show: false
      }
   }),
   getters: {

   },
   actions: {
      async getVersion() {
         this.working = true
         return axios.get("/version").then(response => {
            this.version = `v${response.data.version}-build-${response.data.build}`
         }).catch( err => {
            this.setError(  err )
         })
      },
      setError( e ) {
         this.error = e
         if (e.response && e.response.data) {
            this.error = e.response.data
         }
         this.showError = true
         this.working = false
      },
      toastMessage( summary, message ) {
         this.toast.summary = summary
         this.toast.message = message
         this.toast.show = true
         this.toast.error = false
      },
      toastError( summary, message ) {
         this.toast.summary = summary
         this.toast.message = message
         this.toast.show = true
         this.toast.error = true
      },
      clearToastMessage() {
         this.toast.summary = ""
         this.toast.message = ""
         this.toast.show = false
         this.toast.error = false
      },
   }
})
