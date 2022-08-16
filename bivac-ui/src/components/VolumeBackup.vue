<script setup lang="ts">
import bivac from '../bivac'
import Table from './Table.vue'
import Row from './Row.vue'
import Cell from './Cell.vue'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

const props = defineProps<{
    id: string
}>()

const vol = computed(() => { return bivac.volumes.value[props.id] })

const forceBackup = ref(false)

const localLock = ref(false)

function backup() {
    localLock.value = true
    bivac.backup(props.id, forceBackup.value)
}

</script>

<template>

    <div class="wrapper">
        <div>
            <input type="checkbox" v-model="forceBackup"> Force Backup
        </div>
        <div>
            <button @click="backup()" :disabled="vol.BackingUp && !forceBackup">Backup</button>
        </div>
    </div>

</template>

<style scoped>
.wrapper {
    width: 100%;
    border: 1px solid var(--color-border);
}
</style>