import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useOAStore = defineStore('oa', {
   state: () => ({
      working: false,
      depositToken: "",
      work: {},
   }),
   actions: {
      initNewSubmission(compID, firstName, lastName, department) {
         this.work.resourceType = "Book"
         this.work.title = "title"
         this.work.authors = [{
            computeID: compID, firstName: firstName, lastName: lastName,
            department: department, institution: "University of Virginia", msg: ""}
         ]
         this.work.abstract = "ABS"
         this.work.license = "1"
         this.work.languages = ["English"]
         this.work.keywords = ["key1"]
         this.work.contributors = [{computeID: "", firstName: "", lastName: "", department: "", institution: "", msg: ""}]
         this.work.publisher = "University of Virginia"
         this.work.citation = "fake citation"
         this.work.pubDate = "1980"
         this.work.relatedURLs = ["fake_url"]
         this.work.sponsors = ["sponsor"]
         this.work.notes = "note text"
         this.work.files = []
         this.work.visibility = ""
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
         this.work.files.push( file )
      },
      removeFile( file) {
         axios.delete(`/api/${this.depositToken}/${file}`)
      } ,
      async deposit( ) {
         this.working = true
         return axios.post(`/api/submit/oa/${this.depositToken}`, this.work).then(response => {
            this.work = response.data
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      getWork(id) {

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