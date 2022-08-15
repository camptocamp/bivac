<script setup lang="ts">

defineProps<{
  object: object
}>()

</script>

<template>
    <div class="table">
        <div class="row" v-for="(value, name) in object">
            <div class="cell">{{name}}</div>
            <div class="cell">{{value}}</div>
        </div>
    </div>
</template>