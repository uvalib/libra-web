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
      role: "",
      orcid: {id: "", uri: ""},
      working: false,
      theses: [],
      requestInterceptor: null,
      responseInterceptor: null,
   }),
   getters: {
      isSignedIn: state => {
         return state.jwt != "" && state.computeID != ""
      },
      isAdmin: state => {
         return state.role == "admin"
      },
      isRegistrar: state => {
         return state.role == "registrar"
      },
      homePage: state => {
         if ( state.isAdmin ) return "/admin"
         if ( state.isRegistrar) return "/register"
         return "/"
      }
   },
   actions: {
      async validateAuth() {
         if ( this.isSignedIn ) {
            await axios.get(`/authcheck`).catch( err => {
               console.log("JWT VALIDATE FAILED: "+err)
               this.signOut()
            })
         } else {
            console.log("not sugned in, no auth to validate")
         }
      },
      getORCID() {
         if (this.computeID == "") return

         const orcidURL = `/api/users/orcid/${this.computeID}`
         axios.get(orcidURL).then(response => {
            console.log(response.data)
            this.orcid.id = response.data.orcid
            this.orcid.uri = response.data.uri
         }).catch( err => {
            if (err.response.status != 404) {
               console.log(err)
            }
         })
      },
      signOut() {
         localStorage.removeItem("libra3_jwt")
         localStorage.removeItem("libra3_impersonate")
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
         this.role = parsed.role
         console.log(`jwt is for user ${this.displayName} (${this.computeID}) with role ${this.role}`)

         // add interceptor to put bearer token in header
         if ( this.requestInterceptor ) {
            console.log("remove existing request intercptor")
            axios.interceptors.request.eject( this.requestInterceptor)
         }
         const system = useSystemStore()
         this.requestInterceptor = axios.interceptors.request.use(config => {
            config.headers['Authorization'] = 'Bearer ' + jwt
            return config
         }, error => {
            return Promise.reject(error)
         })

         // Catch 401 errors and redirect to an expired auth page
         if ( this.responseInterceptor ) {
            console.log("remove existing response intercptor")
            axios.interceptors.response.eject(this.responseInterceptor )
         }
        this.responseInterceptor = axios.interceptors.response.use(
            res => res,
            err => {
               console.log(`request ${err.config.url} failed with status ${err.response.status}`)
               console.log(err)
               if (err.config.url.match(/\/authenticate/)) {
                  this.router.push("/forbidden")
               } else {
                  if (err.response && err.response.status == 401) {
                     this.signOut()
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