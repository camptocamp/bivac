<template>
    <div class="iconWrapper">
        <div class="icon"></div>
    </div>
</template>

<style scoped>
@keyframes spinner {
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
}

.iconWrapper {
    display: inline-block;
    height: 40px;
    width: 40px;
}

.icon {
    display: inline-block;
    box-sizing: border-box;
    height: 100%;
    width: 100%;
    animation: 1.5s linear infinite spinner;
    animation-play-state: inherit;
    border: solid 5px var(--color-border);
    border-bottom-color: var(--color-accent);
    border-radius: 50%;
    transform-origin: 50% 50%;
    will-change: transform;
}
</style>