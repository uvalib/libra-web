import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import dayjs from 'dayjs'

export const useETDStore = defineStore('etd', {
   state: () => ({
      working: false,
      saving: false,
      error: "",
      work: {},
      isDraft: true,
      visibility: "",
      embargoReleaseDate: null,
      embargoReleaseVisibility: "",
      licenseID: "0",
      persistentLink: "",
      source: "",
      sourceID: "",
      depositor: "",
      createdAt: null,
      modifiedAt: null,
      publishedAt: null,
      pendingFileAdd: [],
      pendingFileDel: [],
   }),
   getters: {
      hasKeywords: state => {
         return ( state.work.keywords.length > 0 )
      },
      hasLicense: state => {
         return (  parseInt(state.licenseID,10) != 0)
      },
      hasRelatedURLs: state => {
         return ( state.work.relatedURLs.length > 0)
      },
      hasSponsors: state => {
         return ( state.work.sponsors.length > 0)
      },
      hasFiles: state => {
         return (state.pendingFileAdd.length > 0 || state.work.files.length > 0)
      },
      hasAdvisor: state => {
         if ( state.work.advisors.length == 0) return false
         if ( state.work.advisors.length > 1) return true
         let a = state.work.advisors[0]
         return a.firstName != "" && a.lastName != ""
      },
      suggestedCitation: state => {
         //[Author LastName], [Author FirstName]. [Title]. [Author Institution], [program], [Degree], [Published Year], [DOI URI].
         let c = `${state.work.author.lastName}, ${state.work.author.firstName}. ${state.work.title}. ${state.work.author.institution}, `
         c += `${state.work.program}, ${state.work.degree}`
         if ( state.publishedAt) {
            c += `, ${state.publishedAt.split("T")[0]}`
         }
         if ( state.persistentLink) {
            c += `, ${state.persistentLink}`
         }
         c+="."
         return c
      }
   },
   actions: {
      async getWork(id) {
         this.$reset
         this.working = true
         return axios.get(`/api/works/${id}`).then(response => {
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

      setWorkDetails( data ) {
         this.isDraft = data.isDraft
         delete data.isDraft
         this.source = data.source
         delete data.source
         this.sourceID = data.sourceID
         delete data.sourceID
         this.visibility = data.visibility
         delete data.visibility
         this.depositor = data.depositor
         delete data.depositor
         this.persistentLink = data.persistentLink
         delete data.persistentLink
         this.createdAt = data.createdAt
         delete data.createdAt
         if ( data.modifiedAt ) {
            this.modifiedAt = data.modifiedAt
            delete data.modifiedAt
         }
         if ( data.publishedAt ) {
            this.publishedAt = data.publishedAt
            delete data.publishedAt
         }
         if ( data.embargo ) {
            this.embargoReleaseDate = data.embargo.releaseDate
            this.embargoReleaseVisibility  = data.embargo.releaseVisibility
            delete data.embargo
         }
         this.work = data
         if (this.work.advisors.length == 0) {
            this.work.advisors.push({firstName: "", lastName:"", department: "", institution: ""})
         }
         this.work.files.forEach( f => f.url = "")

         // lookup licence ID based on URL
         this.licenseID = "0"
         if ( this.work.licenseURL || this.work.license) {
            const system = useSystemStore()
            let lic = system.licenses.find( l => l.url == this.work.licenseURL )
            if (lic) {
               this.licenseID =""+lic.value
            }
         }
      },

      cancelEdit() {
         if ( this.pendingFileAdd.length > 0) {
            axios.post(`/api/cancel/${this.work.id}`)
         }
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
            axios.delete(`/api/${this.work.id}/${file}`)
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
         return axios.get(`/api/works/${this.work.id}/files/${name}`).then((response) => {
            this.work.files.forEach( f => {
               if (f.name == name ) {
                  f.url = response.data
                  let now = dayjs()
                  let exp = now.add(1, 'hour')
                  f.expire = exp.format("YYYY-MM-DD hh:mm A")
                  setTimeout( ()=>{
                     f.url=""
                     delete f.expire
                  }, 60*60*1000)
               }
            })
         }).catch((error) => {
            const system = useSystemStore()
            system.setError( error)
         })
      },

      async update( ) {
         this.saving = true
         let payload = {work: this.work, addFiles: this.pendingFileAdd, delFiles: this.pendingFileDel, visibility: this.visibility}
         if ( this.visibility == "embargo" || this.visibility == "uva") {
            payload.embargoReleaseDate = this.embargoReleaseDate
            payload.embargoReleaseVisibility = this.embargoReleaseVisibility
         }
         let url = `/api/works/${this.work.id}`
         return axios.put(url, payload).then(response => {
            this.setWorkDetails( response.data )
            this.pendingFileAdd = []
            this.pendingFileDel = []
            this.saving = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.saving = false
         })
      },

      async publish(  ) {
         this.working = true
         return axios.post(`/api/works/${this.work.id}/publish`).then(()=> {
            this.isDraft = false
            this.publishedAt = new Date()
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
   }
})