import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useETDStore = defineStore('etd', {
   state: () => ({
      working: false,
      error: "",
      depositToken: "",
      work: {},
      visibility: "",
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
   },
   actions: {
      async getWork(id) {
         this.error = ""
         this.working = true
         this.pendingFileAdd = []
         this.pendingFileDel = []
         return axios.get(`/api/works/etd/${id}`).then(response => {
            this.visibility = response.data.visibility
            delete response.data.visibility
            this.work = response.data
            if ( this.work.keywords.length == 0) this.work.keywords.push("")
            if ( this.work.relatedURLs.length == 0) this.work.relatedURLs.push("")
            if ( this.work.sponsors.length == 0) this.work.sponsors.push("")
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
      initSubmission(compID, firstName, lastName, program) {
         this.work.title = "",
         this.work.author = {computeID: compID, firstName: firstName, lastName: lastName, program: program, institution: "University of Virginia"},
         this.work.advisors = [{computeID: "", firstName: "", lastName: "", department: "", institution: "University of Virginia", msg: ""}]
         this.work.abstract = ""
         this.work.license = ""
         this.work.language = ""
         this.work.keywords = []
         this.work.relatedURLs = []
         this.work.sponsors = []
         this.work.notes = ""
         this.work.degree = "MA (Master of Arts)"
         this.work.dateCreated = new Date()
         this.work.files = []
         this.visibility = ""
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
      cancel() {
         axios.post(`/api/cancel/${this.depositToken}`)
         this.depositToken = ""
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
            work: this.work, addFiles: this.pendingFileAdd, depositor: depositorComputeID, visibility: this.visibility
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
            work: this.work, addFiles: this.pendingFileAdd, delFiles: this.pendingFileDel, visibility: this.visibility
         }
         let url = `/api/works/etd/${this.work.id}`
         console.log(url)
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