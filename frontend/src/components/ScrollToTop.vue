<template>
   <transition name="message-transition"
         enter-active-class="animated faster fadeIn"
         leave-active-class="animated faster fadeOut">
      <div v-if="showScrollTop" role="button"
         @click.stop="backToTop" tab-index="0" aria-label="scroll to top of page"
         class="scroll-to-top"
      >
         <i class="pi pi-angle-up"></i>
      </div>
   </transition>
</template>

<script setup>
import { onMounted, ref,  onUnmounted} from 'vue'

const showScrollTop = ref(false)

function backToTop() {
   var scrollStep = -window.scrollY / (500 / 10),
   scrollInterval = setInterval(()=> {
      if ( window.scrollY != 0 ) {
         window.scrollBy( 0, scrollStep )
      } else {
         clearInterval(scrollInterval)
      }
   },10)
}
function scrollChecker() {
   if (window.window.scrollY > 150) {
      showScrollTop.value = true
   } else {
      showScrollTop.value = false
   }
}

onMounted( async () => {
   window.addEventListener("scroll", scrollChecker)
})
onUnmounted( async () => {
   window.removeEventListener("scroll", scrollChecker)
})
</script>

<style scoped>
.scroll-to-top {
   display: flex;
   flex-basis: auto;
   flex-direction: column;
   position: fixed;
   background-color: white;
   color: var(--uvalib-brand-orange);
   font-size: 2.9em;
   font-weight: 100;
   border: 2px solid var(--uvalib-brand-orange);
   border-radius: 50%;
   cursor: pointer;
   align-items: center;
   bottom: 5px;
   right: 5px;
   box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 2px 0px;
   box-sizing: border-box;
   justify-content: center;
   min-width: 0px;
   padding: 0px 0px 4px 0px;
   width: 56px;
   height: 56px;
   z-index: 1000;
   i {
      font-size: 0.95em;
      font-weight: 100;
   }
}
.scroll-to-top.mobile {
   position: fixed;
   background-color: white;
   color: var(--uvalib-brand-orange);
   border: 3px solid var(--uvalib-brand-orange);
   font-size: 2em;
   font-weight: 100;
   padding: 0px 12px;
   right: 5px;
   bottom: 45px;
   cursor: pointer;
}
.scroll-to-top:hover, .scroll-to-top.mobile:hover {
  color: white;
  background-color: var(--uvalib-brand-orange);
}
</style>
