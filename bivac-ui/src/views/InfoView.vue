<script setup lang="ts">

import bivac from '../bivac'
import { ref, onUnmounted } from 'vue';
import MenuBar from '../components/MenuBar.vue'

const info = ref({})

async function load() {
    info.value = await bivac.info()
}
const autoreload = bivac.autoreload(load)
onUnmounted(() => {autoreload.cancel();})

const menuItems = [
    {
        name: 'Refresh',
        handler: load
    }
]
</script>

<template>

    <MenuBar :items="menuItems"></MenuBar>

    <div class="table">
        <div class="row" v-for="(value, name) in info">
            <div class="cell">{{name}}</div>
            <div class="cell">{{value}}</div>
        </div>
    </div>
</template>

<style scoped>


</style>