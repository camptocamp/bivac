<script setup lang="ts">

defineProps<{
  items: MenuItem[]
}>()

type MenuItem = {
    name: string,
    handler: Function
}

function executeHandler(item: MenuItem) {
    item.handler()
}


</script>

<template>
    <div class="menubar">
        <div class="menuItem" v-for="item of items">
            <button @click="() => executeHandler(item)">{{item.name}}</button>
        </div>
    </div>
</template>

<style scoped>

.menubar {
    box-sizing: border-box;
    width: 100%;
    color: var(--color-accent);
    margin-bottom: 20px;
    border-bottom: 1px solid var(--color-border);
    border-top: 1px solid var(--color-border);
}

.menuItem {
    display: inline-block;
    padding: 10px;
}

</style>