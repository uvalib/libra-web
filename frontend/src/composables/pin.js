import { onMounted, ref, watch, nextTick } from 'vue'
import { useWindowScroll, useElementBounding, useResizeObserver } from '@vueuse/core'

export function usePinnable( pinID, scrollID, footerID ) {
   const { y } = useWindowScroll()
   const main = ref( document.getElementById(scrollID) )
   // const footer = ref( document.getElementById(footerID) )
   const toolbar = ref( document.getElementById(pinID) )
   const toolbarBounds = ref(  useElementBounding( toolbar ) )
   const pinnedY = ref(-1)
   const noPin = ref(toolbarBounds.value.top > 300)

   useResizeObserver(main, () => {
      noPin.value = toolbarBounds.value.top > 300
      if ( pinnedY.value > -1) {
         unpin()
         nextTick( () => pin() )
      }
   })

   watch(y, (newY) => {
      if (noPin.value) return

      if ( pinnedY.value < 0) {
         if ( toolbarBounds.value.top <= 0 ) {
            pinnedY.value = y.value+toolbarBounds.value.top
            pin()
         }
      } else {
         if ( newY <=  pinnedY.value) {
            console.log("UNPIN")
            pinnedY.value = -1
            unpin()
         }
      }
   })

   const unpin = ( () => {
      toolbar.value.classList.remove("sticky")
      toolbar.value.style.width = `auto`
      main.value.style.top = `auto`
      // footer.value.style.top = `auto`
   })

   const pin = (() => {
      toolbar.value.classList.add("sticky")
      toolbar.value.style.width = `${toolbarBounds.value.width}px`
      // main.value.style.top = `${toolbarBounds.value.height}px`
      // footer.value.style.top = `${toolbarBounds.value.height}px`
   })

   return {}
}