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