<template>
    <div class="table">
        <slot></slot>
    </div>
</template>

<style scoped>

.table {
    display: table;
    border: 1px solid var(--color-border);
}
</style>