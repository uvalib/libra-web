import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useRegistrarStore = defineStore('registrar', {
   state: () => ({
      working: false,
      deposits: [],
      depositSearchMessage: "",
   }),
   getters: {
   },
   actions: {
      async addRegistrations( program, degree, students ) {
         return axios.post(`/api/registrar/register`, {program: program, degree: degree, students: students}).catch( err => {
            const system = useSystemStore()
            system.setError(  err )
         })
      },

      sisDepositStatusSearch(searchType, query) {
         this.working = true
         this.deposits = []
         this.depositSearchMessage = ""
         let url = `/api/registrar/sis?q=${query}&type=${searchType}`
         axios.get(url).then(response => {
            this.deposits = response.data
            this.working = false
         }).catch(err => {
            console.error(err)
            if (err.response && err.response.status == 404) {
               this.depositSearchMessage = `No items found matching ${query}`
            } else {
               this.depositSearchMessage = err
            }
            this.working = false
         })
      },
   }
})