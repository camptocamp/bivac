<script setup lang="ts">
import { RouterView } from 'vue-router'
import BivacLogo from './components/icons/bivac-logo.vue'
import AuthWrapper from './components/AuthWrapper.vue';
import MainMenu from './components/MainMenu.vue';
</script>

<template>

  <AuthWrapper>
    <MainMenu></MainMenu>

    <div class="viewWrapper">
    <RouterView></RouterView>
    </div>

  </AuthWrapper>
  
</template>

<style scoped>

.viewWrapper {
  padding: 20px;
}

</style>
