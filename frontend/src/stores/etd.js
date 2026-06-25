import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'
import dayjs from 'dayjs'

export const useETDStore = defineStore('etd', {
   state: () => ({
      working: false,
      fileChange: false,
      saving: false,
      downloading: "",
      error: "",
      work: {},
      isDraft: true,
      visibility: "",
      embargoReleaseDate: null,
      embargoReleaseVisibility: "",
      licenseID: 0,
      persistentLink: "",
      source: "",
      sourceID: "",
      depositor: "",
      createdAt: null,
      modifiedAt: null,
      publishedAt: null,
   }),
   getters: {
      hasKeywords: state => {
         return ( state.work && state.work.keywords.length > 0 )
      },
      hasLicense: state => {
         return (  parseInt(state.licenseID,10) != 0)
      },
      hasRelatedURLs: state => {
         return ( state.work && state.work.relatedURLs.length > 0)
      },
      hasSponsors: state => {
         return ( state.work && state.work.sponsors.length > 0)
      },
      hasFiles: state => {
         return ( state.work.files.length > 0 )
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
      async getWork(id, usage) {
         this.$reset()
         this.working = true
         return axios.get(`/api/works/${id}?for=${usage}`).then(response => {
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

      async downloadFile(fileName) {
         this.downloading = name
         return axios.get(`/api/works/${this.work.id}/files/${fileName}`).then((response) => {
            const element = document.createElement('a')
            element.setAttribute('href', response.data)
            element.setAttribute('download', name)
            element.style.display = 'none'
            document.body.appendChild(element)
            element.click()
            document.body.removeChild(element)
            this.downloading = ""
         }).catch((error) => {
            this.downloading = ""
            const system = useSystemStore()
            system.setError( error)
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
         this.work.files.forEach( f => f.url = "")

         // lookup licence ID based on URL
         this.licenseID = 0
         if ( this.work.licenseURL || this.work.license) {
            const system = useSystemStore()
            let lic = system.licenses.find( l => l.url == this.work.licenseURL )
            if (lic) {
               this.licenseID = lic.value
            }
         }
      },

      addFile( file ) {
         this.work.files.push( file )
      },

      removeFile( file ) {
         console.log("delete file "+file)
         this.fileChange = true
         axios.delete(`/api/works/${this.work.id}/files/${file}`).then(() => {
            let idx = this.work.files.findIndex( f => f.name == file)
            if ( idx > -1) {
               this.work.files.splice(idx, 1)
            }
         }).catch((error) => {
            const system = useSystemStore()
            system.setError( error)
         }).finally( () => {
             this.fileChange = false
         })
      },

      async replaceFile(tgtFileName, newFile) {
         this.fileChange = true
         const system = useSystemStore()

         // see if extension has changed...
         // If so, first rename the original file to the new extension
         let replaceTarget = tgtFileName
         const origNameParts = tgtFileName.split(".")
         const origExt = origNameParts.pop().toLowerCase() 
         const newExt = newFile.name.split(".").pop().toLowerCase() 
         if (origExt != newExt ) {
            replaceTarget = `${origNameParts[0]}.${newExt}`
            console.log(`rename ${tgtFileName} to have new extenstion ${replaceTarget}`)
            await this.renameFile(tgtFileName, replaceTarget)
            console.log("rename done, now do the replace")
         }

          // override the name of newFile with thename of the existing file
         let formData = new FormData()
         formData.append('file', newFile, replaceTarget)

         await axios.post(`/api/admin/works/${this.work.id}/files/${replaceTarget}/replace`, formData, {
            headers: {
               'Content-Type': 'multipart/form-data',
            }
         }).then(() => {
            system.toastMessage("Replace Success", `'${tgtFileName}' has been replaced with '${newFile.name}'`)
         }).catch((error) => {
            system.toastError("Replace Failed", error)
         }).finally( () => {
            this.fileChange = false
         })
      },

      async renameFile(origName, newName) {
         this.fileChange = true
         let payload = {orignalName: origName, newName: newName}
         axios.put(`/api/works/${this.work.id}/files/rename`, payload).then(() => {
            let tgtFile = this.work.files.find( f => f.name == origName )
            tgtFile.name = newName
         }).catch((error) => {
            const system = useSystemStore()
            system.setError( error)
         }).finally( () => {
             this.fileChange = false
         })
      },

      async update( ) {
         this.saving = true
         let payload = {work: this.work, visibility: this.visibility}
         if ( this.visibility == "embargo" || this.visibility == "uva") {
            payload.embargoReleaseDate = this.embargoReleaseDate
            payload.embargoReleaseVisibility = this.embargoReleaseVisibility
         }
         let url = `/api/works/${this.work.id}`
         return axios.put(url, payload).then(response => {
             this.modifiedAt = response.data.modifiedAt
             this.persistentLink = response.data.persistentLink
            this.saving = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.saving = false
         })
      },

      async updatePublishedDate( newDateStr ) {
         this.saving = true
         let d = dayjs(newDateStr)
         let payload = { newDate: d.format("YYYY-MM-DDTHH:mm:ss[Z]")}
         return axios.put(`/api/admin/works/${this.work.id}/published`, payload).then((response)=> {
            this.publishedAt = response.data
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
            this.publishedAt = dayjs().format("YYYY-MM-DDTHH:mm:ss[Z]")
            this.working = false
         }).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
   }
})