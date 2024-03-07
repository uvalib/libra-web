import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useETDStore = defineStore('etd', {
   state: () => ({
      working: false,
      depositToken: "",
      work: {},
   }),
   actions: {
      initNewSubmission(compID, firstName, lastName, program) {
         this.work.title = "ETDTitle",
         this.work.author = {computeID: compID, firstName: firstName, lastName: lastName, program: program, institution: "University of Virginia"},
         this.work.advisors = [{computeID: "", firstName: "", lastName: "", department: "", institution: "University of Virginia", msg: ""}]
         this.work.abstract = "ABS"
         this.work.license = "1"
         this.work.language = "English"
         this.work.keywords = ["key1"]
         this.work.relatedURLs = ["fake_url"]
         this.work.sponsors = ["sponsor"]
         this.work.notes = "note text"
         this.work.degree = "MA (Master of Arts)"
         this.work.dateCreated = new Date()
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
         return axios.post(`/api/submit/etd/${this.depositToken}`, this.work).then(response => {
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
         return axios.delete(`/api/works/etd/${id}`).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      }
   }
})