<script setup lang="ts">

import { ref } from 'vue';
import bivac from '../bivac'

const authenticated = ref(false)

const apiKey = ref('')
var dirty: boolean = true;

async function authenticate() {
    if (apiKey.value.length > 0) {
        let code = await bivac.setApiKey(apiKey.value)
        ping()
    }
}

async function ping() {
    const status: number = await bivac.ping()

    switch (status) {
        case 401:
            authenticated.value = false
            if (!dirty) {
                forgetSessionKey()
            }
            break;

        case 200:
            authenticated.value = true
            saveSessionKey()
            break;

        //unknown error
        default:
            authenticated.value = false
            break;
    }
}

const sessionIndex = 'bivac-api-key';

function loadSessionKey() {
    let key = sessionStorage.getItem(sessionIndex);
    if (typeof key === 'string') {
        apiKey.value = key;
        dirty = false
    }
}

function saveSessionKey() {
    sessionStorage.setItem(sessionIndex, apiKey.value)
    dirty = false
}

function forgetSessionKey() {
    sessionStorage.removeItem(sessionIndex)
}

loadSessionKey()
authenticate()
ping()

</script>


<template>


    <slot v-if="authenticated"></slot>
    <template v-else>
        <div class="overlay">
            <div class="wrapper">
                <div class="dialog">
                    <div>
                        <input v-model="apiKey">
                    </div>
                    <div>
                        <button @click="authenticate">Authenticate</button>
                    </div>
                </div>
            </div>
        </div>

    </template>


</template>

<style scoped>
.overlay {
    width: 100vw;
    height: 100vh;
    position: fixed;
    top: 0;
    left: 0;
    display: table;
}

.wrapper {
    display: table-cell;
    vertical-align: middle;

}

.dialog {
    width: fit-content;
    padding: 50px;
    background-color: var(--color-dialog);
    margin-left: auto;
    margin-right: auto;
    text-align: center;
    box-shadow: 14px 28px 84px 6px rgba(97, 103, 109, 0.4);
}

.dialog>div {
    padding: 10px 0;
}

input {
    font-size: 1em;
    padding: 0.5em;
    width: 500px;
}

button {
    font-size: 1.5em;
    padding: 0.5em;
}
</style>