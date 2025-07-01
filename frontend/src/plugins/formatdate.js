import dayjs from 'dayjs'

export default {
   install: (app) => {
      app.config.globalProperties.$formatDate = (dateStr) => {
         if (dateStr) {
            if ( dateStr.includes("T00:00:00Z") ) {
               return dateStr.split("T")[0]
            }
            let d = dayjs(dateStr)
            return d.format("YYYY-MM-DD")
         }
         return ""
      }
   }
}