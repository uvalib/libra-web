<template>
   <div class="signedout">
      <h1>Signed Out</h1>
      <div class="message">
         <p>You have successfuly signed out of LibraETD.</p>
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
.signedout {
   padding: 0;
   min-height: 600px;
   .message {
      margin: 10px 0 30px 0;
      font-size: 1.15em;
   }
}
</style>
