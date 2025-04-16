import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
   define: {
      // enable hydration mismatch details in production build
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'true'
   },
   plugins: [
      vue(),
   ],
   resolve: {
      alias: {
         '@': fileURLToPath(new URL('./src', import.meta.url))
      }
   },
   server: { // this is used in dev mode only
      port: 8080,
      proxy: {
         '/api': {
            target: process.env.LIBRA_SRV,  //export LIBRA_SRV=http://localhost:8085
            changeOrigin: true
         },
         '/authenticate': {
            target: process.env.LIBRA_SRV,
            changeOrigin: true
         },
         '/authcheck': {
            target: process.env.LIBRA_SRV,
            changeOrigin: true
         },
         '/config': {
            target: process.env.LIBRA_SRV,
            changeOrigin: true
         },
         '/healthcheck': {
            target: process.env.LIBRA_SRV,
            changeOrigin: true
         },
         '/version': {
            target: process.env.LIBRA_SRV,
            changeOrigin: true
         },
      }
   },
   css: {
      preprocessorOptions : {
          scss: {
              api: "modern-compiler",
              additionalData: `@use "@/assets/theme/colors.scss" as *;`
          },
      }
   },
})
