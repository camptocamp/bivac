var apiKey: string

export default {
    async setApiKey(key: string) {
        apiKey = key
    },

    api(path: string): Promise<Response> {
        return fetch(path, {
            headers: {
                Authorization: "Bearer " + apiKey
            }
        })
    },

    async get(path: string) {
        return await this.api(path).then(res => res.json())
    },

    autoreload(handler: Function) {
        const container = {
            timer: null,
            async reload() {
                await handler()
                this.timer = setTimeout(() => {this.reload();}, 5000)
            },
            cancel() {
                clearTimeout(this.timer)
            }
        }
        container.reload()
        return container
    }
}