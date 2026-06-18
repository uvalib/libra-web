import { defineStore } from 'pinia'
import axios from 'axios'
import { useSystemStore } from './system'

export const useRegistrarStore = defineStore('registrar', {
   state: () => ({
      working: false,
      total: 0,
      offset: 0,
      limit: 25,
      sortField: "submitted_at",
      sortOrder: "desc",
      filters: {},
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
         this.total = 0
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
      optDepositStatusSearch() {
         this.deposits = []
         this.total = 0
         this.working = true
         let url = `/api/registrar/optional?offset=${this.offset}&limit=${this.limit}&sort=${this.sortField}&order=${this.sortOrder}`
         let filters = [] 
         if ( this.filters.registrar ) {
            filters.push(`registrar=${encodeURIComponent(this.filters.registrar)}`)
         }
         if ( this.filters.submitted_at ) {
            filters.push(`submitted_at=${encodeURIComponent(this.filters.submitted_at)}`)
         }
         if ( this.filters.program ) {
            filters.push(`program=${encodeURIComponent(this.filters.program)}`)
         }
         if ( this.filters.degree ) {
            filters.push(`degree=${encodeURIComponent(this.filters.degree)}`)
         }
         if ( filters.length > 0 ) {
            url += `&${filters.join("&")}`
         }
         axios.get(url).then(response => {
            this.deposits = response.data.hits
            this.total = response.data.total
            this.working = false
         }).catch(err => {
            console.error(err)
            if (err.response && err.response.status == 404) {
               this.depositSearchMessage = `No matching optional deposits found`
            } else {
               this.depositSearchMessage = err
            }
            this.working = false
         })
      },
   }
})