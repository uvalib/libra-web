<template>
   <div class="expired">
      <h1>Session Expired</h1>
      <div class="message">
         <p>Your session has expired and you have been signed out.</p>
      </div>
      <Button @click="signinClick">Sign in with NetBadge</Button>
   </div>
</template>

<script setup>
function signinClick() {
   window.location.href = "/authenticate"
}
</script>

<style scoped lang="scss">
.expired {
   padding: 0;
   min-height: 600px;
   .message {
      margin: 10px 0 30px 0;
      font-size: 1.15em;
   }
}
</style>
