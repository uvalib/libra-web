import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useETDStore = defineStore('etd', {
   state: () => ({
      working: false,
      error: "",
      depositToken: "",
      work: {},
      isDraft: true,
      visibility: "",
      embargoReleaseDate: "",
      embargoReleaseVisibility: "",
      licenseID: "",
      persistentLink: "",
      datePublished: null,
      pendingFileAdd: [],
      pendingFileDel: [],
   }),
   getters: {
      hasKeywords: state => {
         if ( state.work.keywords.length == 0 ) return false
         if ( state.work.keywords.length > 1) return true
         return state.work.keywords[0] != ""
      },
      hasRelatedURLs: state => {
         if ( state.work.relatedURLs.length > 1) return true
         return state.work.relatedURLs[0] != ""
      },
      hasSponsors: state => {
         if ( state.work.sponsors.length > 1) return true
         return state.work.sponsors[0] != ""
      },
      hasAdvisor: state => {
         if ( state.work.advisors.length == 0) return false
         if ( state.work.advisors.length > 1) return true
         let a = state.work.advisors[0]
         return a.firstName != "" && a.lastName != "" && a.computeID != ""
      },
   },
   actions: {
      async getWork(id) {
         this.$reset
         this.working = true
         return axios.get(`/api/works/etd/${id}`).then(response => {
            this.setWorkDetails(response.data)
            this.working = false
         }).catch( err => {
            if (err.response.status == 404) {
               this.router.push("/not_found")
            } else if (err.response.status == 403) {
               this.router.push("/forbidden")
            } else {
               this.error = err
            }
            this.working = false
         })
      },
      setWorkDetails( data) {
         this.isDraft = data.isDraft
         delete data.isDraft
         this.visibility = data.visibility
         delete data.visibility
         this.persistentLink = data.persistentLink
         delete data.persistentLink
         if ( data.datePublished ) {
            this.datePublished = data.datePublished
            delete data.datePublished
         }
         if ( data.embargo ) {
            this.embargoReleaseDate = data.embargo.releaseDate
            this.embargoReleaseVisibility  = data.embargo.releaseVisibility
            delete data.embargo
         }
         this.work = data
         if ( this.work.keywords.length == 0) this.work.keywords.push("")
         if ( this.work.relatedURLs.length == 0) this.work.relatedURLs.push("")
         if ( this.work.sponsors.length == 0) this.work.sponsors.push("")

         // lookup licence ID based on URL
         const system = useSystemStore()
         let lic = system.etdLicenses.find( l => l.url == this.work.licenseURL )
         if (lic) {
            this.licenseID = lic.value
         }
      },
      initSubmission(compID, firstName, lastName, program) {
         this.work.title = "",
         this.work.author = {computeID: compID, firstName: firstName, lastName: lastName, program: program, institution: "University of Virginia"},
         this.work.advisors = [{computeID: "", firstName: "", lastName: "", department: "", institution: "University of Virginia", msg: ""}]
         this.work.abstract = ""
         this.work.license = ""
         this.work.licenseURL = ""
         this.work.language = ""
         this.work.keywords = []
         this.work.relatedURLs = []
         this.work.sponsors = []
         this.work.notes = ""
         this.work.degree = "MA (Master of Arts)"
         this.work.createdAt = new Date()
         this.work.files = []

         this.licenseID = ""
         this.visibility = ""
         this.embargoReleaseDate = ""
         this.embargoReleaseVisibility = ""
         this.pendingFileAdd = []
         this.pendingFileDel = []
      },
      async getDepositToken() {
         this.depositToken = ""
         return axios.get("/api/token").then(response => {
            this.depositToken = response.data
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },
      cancelCreate() {
         axios.post(`/api/cancel/${this.depositToken}`)
         this.depositToken = ""
      },
      cancelEdit() {
         axios.post(`/api/cancel/${this.work.id}`)
      },
      addFile( file ) {
         this.pendingFileAdd.push( file )
      },
      removeFile( file) {
         let pendingIdx = this.pendingFileAdd.findIndex( f => f == file )
         if ( pendingIdx > -1) {
            // this file has not been attached to a work in easystore; just remove
            // if from the pending add list and delete the version that was uploaded to temp storage
            this.pendingFileAdd.splice(pendingIdx, 1)
            axios.delete(`/api/${this.depositToken}/${file}`)
         } else {
            console.log("delete previously added file "+file)
            // This file has already been submitted. remove it from the files
            // list. When the update is submitted the files will be replaced with those in the file list
            let idx = this.work.files.findIndex( f => f.name == file)
            if ( idx > -1) {
               this.work.files.splice(idx, 1)
               this.pendingFileDel.push(file)
            }
         }
      },
      async downloadFile( name ) {
         return axios.get(`/api/works/etd/${this.work.id}/files/${name}`,{responseType: "blob"}).then((response) => {
            let ct = response.headers["content-type"]
            const fileURL = window.URL.createObjectURL(new Blob([response.data], {type: ct}))
            const fileLink = document.createElement('a')

            fileLink.href = fileURL;
            fileLink.setAttribute('download', response.headers["content-disposition"].split("filename=")[1])
            document.body.appendChild(fileLink);

            fileLink.click();
            window.URL.revokeObjectURL(fileURL);

         }).catch((error) => {
            const system = useSystemStore()
            system.setError( error)
         })
      },
      async deposit( depositorComputeID ) {
         this.working = true
         let payload = {
            work: this.work, addFiles: this.pendingFileAdd, depositor: depositorComputeID, visibility: this.visibility,
            embargoReleaseDate: this.embargoReleaseDate, embargoReleaseVisibility: this.embargoReleaseVisibility
         }
         return axios.post(`/api/submit/etd/${this.depositToken}`, payload).then(response => {
            this.work = response.data
            this.working = false
            this.pendingFileAdd = []
            this.pendingFileDel = []
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async update( ) {
         this.working = true
         let payload = {
            work: this.work, addFiles: this.pendingFileAdd, delFiles: this.pendingFileDel, visibility: this.visibility,
            embargoReleaseDate: this.embargoReleaseDate, embargoReleaseVisibility: this.embargoReleaseVisibility
         }
         let url = `/api/works/etd/${this.work.id}`
         return axios.put(url, payload).then(response => {
            this.work = response.data
            this.working = false
            this.pendingFileAdd = []
            this.pendingFileDel = []
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async publish(  ) {
         this.working = true
         return axios.post(`/api/works/etd/${this.work.id}/publish`).then(()=> {
            this.isDraft = false
            this.datePublished = new Date()
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async deleteWork( id ) {
         this.working = true
         return axios.delete(`/api/works/etd/${id}`).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})