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
                'Backup Running': vol.BackingUp ? 'Yes' : 'No',
                'Last Status': vol.LastBackupStatus,
                'Last Backup': vol.LastBackupDate,
                'Backup Directory': vol.BackupDir ? vol.BackupDir : '/',
                Mountpoint: vol.Mountpoint,
                'Read Only': vol.ReadOnly ? 'Yes' : 'No',
                'Hostname': vol.Hostname
            }

            if (typeof vol.Logs.backup !== 'undefined') {
                logRows.value.backup = vol.Logs.backup.substr(4).split("\n").filter((abc: string) => abc.length).map((abc: string) => { return JSON.parse(abc); })
                console.log(logRows.value.backup)
            } else {
                logRows.value.backup = undefined
            }


            return
        }
    }
    found.value = false
}
const autoreload = bivac.autoreload(load)
onUnmounted(() => { autoreload.cancel(); })

async function backup() {
    if (typeof route.params.id === 'string') {
        bivac.post('/backup/' + encodeURIComponent(route.params.id) + '?force=false')
    }
}

const menuItems = [
    {
        name: 'Refresh',
        handler: load
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

    <div class="table">
        <InfoTable :object="rows">

        </InfoTable>
        <InfoTable :object="logRows">

        </InfoTable>
    </div>
</template>

<style scoped>
</style>