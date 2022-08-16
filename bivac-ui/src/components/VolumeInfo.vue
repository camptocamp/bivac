<script setup lang="ts">
import bivac from '../bivac'
import Table from './Table.vue'
import Row from './Row.vue'
import Cell from './Cell.vue'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

const props = defineProps<{
  id: string
}>()

const vol = computed(() => {return bivac.volumes.value[props.id]})

const age = ref('')

function loadAge() {

  const last = new Date(vol.value.LastBackupDate + "Z")
  const seconds = Math.floor((Date.now() - last.valueOf())/1000)
  const minutes = Math.floor(seconds/60)
  const hours = Math.floor(minutes/60)
  const days = Math.floor(hours/24)

  let suffix = 'second'
  let value = seconds

  if(days > 0) {
    value = days
    suffix = 'day'
  } else if (hours > 0) {
    value = hours
    suffix = 'hour'
  }  else if (minutes > 0) {
    value = minutes
    suffix = 'minute'
  }

  if(value > 1) {
    suffix += 's'
  }
  age.value = value + ' ' + suffix
}
const reload = bivac.autoreload(() => {loadAge();})
onUnmounted(() => {reload.cancel();})

</script>

<template>
  <Table>
    <Row>
      <Cell>ID</Cell><Cell>{{id}}</Cell>
    </Row>
    <Row>
      <Cell>Name</Cell><Cell>{{vol.Name}}</Cell>
    </Row>
    <Row>
      <Cell>Last Status</Cell>
      <Cell>
          <div class="status" :class="vol.LastBackupStatus === 'Success' ? 'success' : 'fail'">{{vol.LastBackupStatus}}</div>
      </Cell>
    </Row>
    <Row>
      <Cell>Last Backup</Cell><Cell>{{vol.LastBackupDate}} ({{age}} old)</Cell>
    </Row>
    <Row>
      <Cell>Backup Directory</Cell><Cell>{{vol.BackupDir ? vol.BackupDir : '/'}}</Cell>
    </Row>
    <Row>
      <Cell>Mountpoint</Cell><Cell>{{vol.Mountpoint}}</Cell>
    </Row>
  </Table>
</template>

<style scoped>

.success {
  background-color: var(--color-background-success);
  color: var(--color-text-success);
}

.fail {
  background-color: var(--color-background-fail);
  color: var(--color-text-fail);
}

.status {
  width: 100%;
  text-align: center;
  padding: 0.5em;
  font-size: 1.5em;
}

</style>