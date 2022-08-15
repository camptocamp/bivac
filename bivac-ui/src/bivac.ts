var apiKey: string

export default {
    setApiKey(key: string) {
        apiKey = key
    },

    api(path: string, method: string = 'GET', body: string | object = undefined): Promise<Response> {
        let options = {
            method: method.toUpperCase(),
            headers: {
                Authorization: "Bearer " + apiKey
            }
        }
        if (typeof body !== 'undefined') {
            options.body = body
        }
        return fetch(path, options)
    },

    async get(path: string) {
        return await this.api(path).then(res => res.json())
    },

    async post(path: string, body: string | object = null) {
        return await this.api(path).then(res => res.json())
    },

    autoreload(handler: Function) {
        const container = {
            timer: null,
            async reload() {
                try {
                    await handler()
                } catch (e) {
                    console.log(e)
                }
                this.timer = setTimeout(() => { this.reload(); }, 5000)
            },
            cancel() {
                clearTimeout(this.timer)
            }
        }
        container.reload()
        return container
    }
}