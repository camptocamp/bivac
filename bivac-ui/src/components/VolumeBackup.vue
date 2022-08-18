<script setup lang="ts">
import bivac from '../bivac'
import Table from './Table.vue'
import Row from './Row.vue'
import Cell from './Cell.vue'
import { computed, onMounted, onUnmounted, reactive, ref, type ComputedRef, type Ref } from 'vue';
import LoadingIcon from './icons/LoadingIcon.vue';

const props = defineProps<{
    id: string
}>()

const vol = computed(() => { return bivac.volumes.value[props.id] })

const forceBackup = ref(false)

const localLock = ref(false)

function backup() {
    localLock.value = true
    forceBackup.value = false
    bivac.backup(props.id, forceBackup.value)
}

const backupAvailable: ComputedRef<boolean> = computed(() => {
    if (localLock.value && vol.value.BackingUp) {
        localLock.value = false
    }
    const available = !(localLock.value || vol.value.BackingUp) || forceBackup.value
    return available
})

</script>

<template>

    <div class="wrapper">
        <div>
            <input type="checkbox" v-model="forceBackup"> Force Backup
        </div>
        <div>
            <button class="backupButton" @click="backup()" :disabled="!backupAvailable">
                <template v-if="backupAvailable">
                    <div>Backup</div>
                </template>
                <template v-else>
                    <LoadingIcon />
                    <div>Backup running...</div>
                </template>
            </button>
        </div>
    </div>

</template>

<style scoped>
.wrapper {
    width: 100%;
    padding: 20px;
    border: 1px solid var(--color-border);
}

.backupButton {
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 1.8em;
    padding: 0.5em;
    width: 400px;
}

.backupButton div {
    display: inline-block;
}
</style>