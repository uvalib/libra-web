import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

function parseJwt(token) {
   var base64Url = token.split('.')[1]
   var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
   var jsonPayload = decodeURIComponent(atob(base64).split('').map(function (c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
   }).join(''))

   return JSON.parse(jsonPayload);
}

export const useUserStore = defineStore('user', {
   state: () => ({
      jwt: "",
      computeID: "",
      uvaID: "",
      displayName: "",
      firstName: "",
      initials: "",
      lastName: "",
      description: [],
      department: [],
      title: [],
      office: [],
      phone: [],
      affiliation: [],
      email: "",
      private: "",
      admin: false,
      registrar: false,
      working: false,
      theses: []
   }),
   getters: {
      isSignedIn: state => {
         return state.jwt != "" && state.computeID != ""
      },
   },
   actions: {
      async validateAuth() {
         if ( this.isSignedIn ) {
            await axios.get(`/authcheck`).catch( err => {
               console.log("JWT VALIDATE FAILED: "+err)
               localStorage.removeItem("libra3_jwt")
               this.$reset()
            })
         } else {
            console.log("not sugned in, no auth to validate")
         }
      },
      signOut() {
         localStorage.removeItem("libra3_jwt")
         this.$reset()
      },
      getTheses() {
         this.working = true
         let url = `/api/works/search?cid=${this.computeID}`
         axios.get(url).then(response => {
            this.theses = response.data.hits
            this.working = false
         }).catch( err => {
            console.error(err)
            const system = useSystemStore()
            system.setError(  err )
            this.working = false
         })
      },
      setJWT(jwt) {
         if (jwt == null || jwt == "" || jwt == "null") {
            return
         }
         if (jwt == this.jwt) return

         this.jwt = jwt
         localStorage.setItem("libra3_jwt", jwt)

         let parsed = parseJwt(jwt)
         this.computeID = parsed.cid
         this.uvaID = parsed.uva_id
         this.displayName = parsed.display_name
         this.firstName = parsed.first_name
         this.initials = parsed.initials
         this.lastName = parsed.last_name
         this.description = parsed.description
         this.department = parsed.department
         this.title = parsed.title
         this.office = parsed.office
         this.phone = parsed.phone
         this.affiliation = parsed.affiliation
         this.email = parsed.email
         this.private = parsed.private
         this.admin = parsed.admin
         this.registrar = parsed.registrar

         // add interceptor to put bearer token in header
         const system = useSystemStore()
         axios.interceptors.request.use(config => {
            config.headers['Authorization'] = 'Bearer ' + jwt
            return config
         }, error => {
            return Promise.reject(error)
         })

         // Catch 401 errors and redirect to an expired auth page
         axios.interceptors.response.use(
            res => res,
            err => {
               console.log("failed response for "+err.config.url)
               console.log(err)
               if (err.config.url.match(/\/authenticate/)) {
                  this.router.push("/forbidden")
               } else {
                  if (err.response && err.response.status == 401) {
                     localStorage.removeItem("libra3_jwt")
                     this.$reset
                     system.working = false
                     this.router.push("/expired")
                     return new Promise(() => { })
                  }
               }
               return Promise.reject(err)
            }
         )
      }
   }
})