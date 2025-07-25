import dayjs from 'dayjs'

export default {
   install: (app) => {
      app.config.globalProperties.$formatDateTime = (dateStr) => {
         if (dateStr) {
            let d = dayjs(dateStr)
            if ( dateStr.includes("T00:00:00Z") ) {
               return dateStr.split("T")[0]
            } else {
               return d.format("YYYY-MM-DD hh:mm A")
            }
         }
         return ""
      }
   }
}