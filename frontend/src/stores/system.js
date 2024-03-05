import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
   state: () => ({
      working: false,
		version: "unknown",
      languages: [],
      licenses: [],
      visibility: [],
      resourceTypes: [],
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
      visibilityLabel: state => {
         return (mode, key) => {
            if ( mode == "oa") {
               let oaV = state.visibility.filter( rt => rt.oa == true)
               let out = oaV.find(  v=>v.value == key)
               return out.label
            }
            let etdV = state.visibility.filter( rt => rt.etd == true)
            let out = etdV.find(  v=>v.value == key)
            return out.label
         }
      },
      oaResourceTypes: state => {
         return state.resourceTypes.filter( rt => rt.oa == true)
      },
      etdResourceTypes: state => {
         return state.resourceTypes.filter( rt => rt.etd == true)
      },
      oaLicenses: state => {
         return state.licenses.filter( rt => rt.oa == true)
      },
      etdLicenses: state => {
         return state.licenses.filter( rt => rt.etd == true)
      },
      oaVisibility: state => {
         return state.visibility.filter( rt => rt.oa == true)
      },
      etdVisibility: state => {
         return state.visibility.filter( rt => rt.etd == true)
      },
   },
   actions: {
      async getConfig() {
         this.working = true
         return axios.get("/config").then(response => {
            this.version = response.data.version
            this.languages = response.data.languages
            this.licenses = response.data.licenses
            this.resourceTypes = response.data.resourceTypes
            this.visibility = response.data.visibility
            console.log("CONFIGURE SUCCESS")
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
