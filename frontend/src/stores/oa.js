import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useOAStore = defineStore('oa', {
   state: () => ({
      working: false,
      depositToken: "",
      work: {},
      pendingFileAdd: [],
      pendingFileDel: []
   }),
   actions: {
      initSubmission(compID, firstName, lastName, department) {
         this.work.resourceType = "Book"
         this.work.title = ""
         this.work.authors = [{
            computeID: compID, firstName: firstName, lastName: lastName,
            department: department, institution: "University of Virginia", msg: ""}
         ]
         this.work.abstract = ""
         this.work.license = ""
         this.work.languages = [""]
         this.work.keywords = [""]
         this.work.contributors = [{computeID: "", firstName: "", lastName: "", department: "", institution: "", msg: ""}]
         this.work.publisher = "University of Virginia"
         this.work.citation = ""
         this.work.pubDate = ""
         this.work.relatedURLs = []
         this.work.sponsors = []
         this.work.notes = ""
         this.work.files = []
         this.work.visibility = ""
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
            // This file has already been submitted. remove it from the files
            // list and added to a pending delete list. When the update is submitted
            // the files on the pending delete list will be removed from the store.
            // this allows file deletions to be canceled,
            let idx = this.work.files.findIndex( f => f == file)
            if ( idx > -1) {
               this.work.files.splice(idx, 1)
               this.pendingFileDel.push(file)
            }
         }

      } ,
      async deposit( ) {
         this.working = true
         let payload = {work: this.work, addFiles: this.pendingFileAdd}
         return axios.post(`/api/submit/oa/${this.depositToken}`, payload).then(response => {
            this.work = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async update( ) {
         const system = useSystemStore()
         system.setError(  "not yet implemented" )
         this.working = false
      },
      async getWork(id) {
         this.working = true
         return axios.get(`/api/works/oa/${id}`).then(response => {
            this.work = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      async deleteWork( id ) {
         this.working = true
         return axios.delete(`/api/works/oa/${id}`).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})