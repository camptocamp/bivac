<script setup lang="ts">

import bivac from '../bivac'
import { ref, onUnmounted } from 'vue';
import MenuBar from '../components/MenuBar.vue'

const volumes = bivac.volumes

async function load() {
    bivac.loadVolumes();
}



const cols = ref({
        ID: 'ID',
        Name: 'Name',
        LastBackupStatus: 'Last Status',
        LastBackupDate: 'Last Backup',
        Mountpoint: "Mountpoint",
})

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
        <div class="row header">
            <div class="cell" v-for="(header, key) in cols">{{header}}</div>
            <div class="cell">Backup Running</div>
        </div>
        <template v-for="vol in volumes">
            <RouterLink class="row" :to="'/volume/' + vol.ID">
            <div class="cell" v-for="(header, key) in cols">{{vol[key]}}</div>
            <div class="cell">{{vol.BackingUp}}</div>
            </RouterLink>
        </template>
    </div>
</template>

<style scoped>

.row:hover:not(.header) {
    background-color: var(--color-background-active);
}

</style>