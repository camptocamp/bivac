<template>
    <div class="row">
        <slot></slot>
    </div>
</template>

<style scoped>
.row {
    display: table-row;
    border: 1px solid var(--color-border);
}
</style>