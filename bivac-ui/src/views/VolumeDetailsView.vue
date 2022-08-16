<script setup lang="ts">

import bivac from '../bivac'
import { ref, computed, onMounted, onUnmounted } from 'vue';
import MenuBar from '../components/MenuBar.vue'
import { useRoute } from 'vue-router'
import InfoTable from '../components/InfoTable.vue'
import VolumeInfo from '../components/VolumeInfo.vue'
const route = useRoute()

const id: string = route.params.id


async function backup() {
    if (typeof route.params.id === 'string') {
        bivac.backup(route.params.id)
    }
}

const menuItems = [
    {
        name: 'Refresh',
        handler: () => { bivac.loadVolumes(); }
    },
    {
        name: 'Backup',
        handler: backup
    }
]

const rows = ref({})

const logRows = ref({})
</script>

<template>

    <MenuBar :items="menuItems"></MenuBar>

    <div v-if="typeof bivac.volumes.value[id] === 'undefined'">
        ERROR 404
    </div>
    <template v-else>
        <VolumeInfo :id="id"></VolumeInfo>
    </template>
</template>

<style scoped>
</style>