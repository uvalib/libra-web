<template>
   <div class="status">
      <h1>Optional Deposit Status</h1>
   </div>
</template>

<script setup>
</script>

<style lang="scss" scoped>
@media only screen and (min-width: 768px) {
   .status {
      margin: 0 50px 50px;
   }
}
@media only screen and (max-width: 768px) {
   .status {
      margin: 0 10px;
   }
}
.status {
   display: flex;
   flex-direction: column;
   gap: 20px;
   min-height: 600px;
   text-align: left;
}
</style>