<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { defineComponent } from 'vue'

const props = defineProps<{
    name: string
    to: string
}>()

</script>

<template>
    <div class="linkContainer">
        <RouterLink class="mainMenuItem" :to="to">
            <div>{{ name }}</div>
        </RouterLink>
    </div>
</template>

<style scoped>

.mainMenuItem {
    display:block;
    padding: 20px 30px;
    color: var(--color-accent);
}

.mainMenuItem:hover {
    background-color: var(--color-background-active);
}

@media (min-width: 1024px) {
}
</style>