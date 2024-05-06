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
      namespaces: [],
      degrees: [],
      programs: [],
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
      namespaceLabel: state => {
         return (ns) => {
            let nsv = state.namespaces.find( n => n.namespace == ns)
            if (nsv) {
               return nsv.label
            }
            return ns
         }
      },
      oaNamespace: state => {
         let tgtNS = state.namespaces.find( ns => ns.label == "LibraOpen")
         return tgtNS && tgtNS.namespace
      },
      etdNamespace: state => {
         let tgtNS = state.namespaces.find( ns => ns.label == "LibraETD")
         return tgtNS && tgtNS.namespace
      },
      visibilityLabel: state => {
         return (mode, key) => {
            if ( mode == "oa") {
               let oaV = state.visibility.filter( rt => rt.oa == true)
               if ( oaV ) {
                  let out = oaV.find(  v=>v.value == key)
                  if ( out ) return out.label
               }
            } else {
               let etdV = state.visibility.filter( rt => rt.etd == true)
               if (etdV) {
                  let out = etdV.find(  v=>v.value == key)
                  if ( out ) return out.label
               }
            }
            return key
         }
      },
      oaResourceTypes: state => {
         return state.resourceTypes.filter( rt => rt.oa == true)
      },
      etdResourceTypes: state => {
         return state.resourceTypes.filter( rt => rt.etd == true)
      },
      licenseDetail: state => {
         return (mode, id) => {
            if ( mode == "oa") {
               return state.licenses.filter( l => l.oa == true).find( oa => oa.value == id)
            } else {
               return state.licenses.filter( l => l.etd == true).find( oa => oa.value == id)
            }
         }
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
      sisDegrees: state => {
         return state.degrees.filter( d => d.type == "sis").map( d => d.degree)
      },
      optDegrees: state => {
         return state.degrees.filter( d => d.type == "optional").map( d => d.degree)
      },
      sisPrograms: state => {
         return state.programs.filter( d => d.type == "sis").map( d => d.program)
      },
      optPrograms: state => {
         return state.programs.filter( d => d.type == "optional").map( d => d.program)
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
            this.namespaces = response.data.namespaces
            this.degrees = response.data.degrees
            this.programs = response.data.programs
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
