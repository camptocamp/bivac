<script setup lang="ts">

import bivac from '../bivac'
import { ref, computed, onMounted, onUnmounted } from 'vue';
import MenuBar from '../components/MenuBar.vue'
import { useRoute } from 'vue-router'
import InfoTable from '../components/InfoTable.vue'
import VolumeInfo from '../components/VolumeInfo.vue'
import VolumeBackup from '../components/VolumeBackup.vue'
const route = useRoute()

const id: string = route.params.id as string

const menuItems = [
    {
        name: 'Refresh',
        handler: () => { bivac.loadVolumes(); }
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
    <div class="sections" v-else>
        <VolumeInfo :id="id"></VolumeInfo>
        <div class="spacer"></div>
        <VolumeBackup :id="id"></VolumeBackup>
    </div>
</template>

<style scoped>
.sections {
    width: 100%;
}

.spacer {
    width: 100%;
    height: 20px;
}
</style>