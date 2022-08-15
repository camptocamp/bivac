<script setup lang="ts">

import bivac from '../bivac'
import { ref, computed, onMounted, onUnmounted } from 'vue';
import MenuBar from '../components/MenuBar.vue'
import { useRoute } from 'vue-router'
import InfoTable from '../components/InfoTable.vue'
const route = useRoute()



const found = ref(true)

async function load() {
    let all = await bivac.get('/volumes')
    for (const vol of all) {
        if (vol.ID == route.params.id) {
            found.value = true

            rows.value = {
                ID: vol.ID,
                Name: vol.Name,
                'Last Status': vol.LastBackupStatus,
                'Last Backup': vol.LastBackupDate,
                'Backup Directory': vol.BackupDir ? vol.BackupDir : '/',
                Mountpoint: vol.Mountpoint,
                'Read Only': vol.ReadOnly ? 'Yes' : 'No',
                'Hostname': vol.Hostname
            }

            return
        }
    }
    found.value = false
}
const autoreload = bivac.autoreload(load)
onUnmounted(() => {autoreload.cancel();})

const menuItems = [
    {
        name: 'Refresh',
        handler: load
    }
]

const rows = ref({})
</script>

<template>

    <MenuBar :items="menuItems"></MenuBar>

    <div class="table">
        <InfoTable :object="rows">

        </InfoTable>
    </div>
</template>

<style scoped>
</style>