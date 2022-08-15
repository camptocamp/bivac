<script setup lang="ts">
import MainMenuItem from "./MainMenuItem.vue"
import BivacLogo from "./icons/bivac-logo.vue"

type menuEntry = {
    name: string,
    to: string
}

const menu: menuEntry[] = [
    { name: "info", to: "/" },
    { name: "volumes", to: "/volumes" },
]

</script>

<template>
    <div class="mainMenuContainer">
        <div class="logo">
            <RouterLink class="logoLink" to="/">
                <BivacLogo></BivacLogo>
            </RouterLink>
        </div>
        <div class="navContainer">
            <div class="innerNav">
                <nav>
                    <template v-for="item in menu">
                        <MainMenuItem :name="item.name" :to="item.to" />
                    </template>
                </nav>
            </div>
        </div>
    </div>
</template>

<style scoped>
.mainMenuContainer {
    box-sizing: border-box;
    height: 100%;
    padding:  30px 0;
    border-left: 1px solid var(--color-border);
    border-right: 1px solid var(--color-border);
}

.logoLink {
    display: block;
    padding: 20px;
}

.logoLink:hover {
    background-color:  var(--color-background-active);
}

.innerNav {
    width: 100%;
    text-align: center;
}

.navContainer {
    width: 100%;
}

@media (min-width: 1024px) {}
</style>