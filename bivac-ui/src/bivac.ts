import { ref, type Ref } from 'vue'

class bivac {
    private apiKey: string | null = null
    public volumes: Ref<{ [key: string]: { [key: string]: any } }> = ref({})

    constructor() {
        this.autoreload(() => this.loadVolumes())
    }

    public setApiKey(apiKey: string): void {
        this.apiKey = apiKey
        this.loadVolumes()
    }

    private api(path: string, method: string = 'GET', body: string | object | null = null): Promise<Response> {
        //console.log('requesting: ' + path)
        const options: { [key: string]: any } = {
            method: method.toUpperCase(),
            headers: {
                Authorization: "Bearer " + this.apiKey
            }
        }
        if (body !== null) {
            options.body = body
        }
        return fetch(path, options)
    }

    private async get(path: string) {
        return await this.api(path).then(res => res.json())
    }

    private async post(path: string, body: string | object | null = null) {
        return await this.api(path).then(res => res.json())
    }

    public autoreload(handler: Function) {
        const container = {
            timer: 0,
            async reload() {
                try {
                    await handler()
                } catch (e) {
                    console.log(e)
                }
                container.timer = setTimeout(() => { container.reload(); }, 5000)
            },
            cancel() {
                clearTimeout(container.timer)
            }
        }
        container.reload()
        return container
    }

    public async ping(): Promise<number> {
        const response = await this.api('/ping');
        if (response.status === 200) {
            const parsed = await response.json()
            if (typeof parsed === 'object' && typeof parsed.type === 'string' && parsed.type === 'pong') {
                return 200;
            }
            return 500;
        }
        return response.status
    }

    public async info() {
        return await this.get('/info').then(res => res.data)
    }

    public async loadVolumes() {
        if (this.apiKey !== null) {
            const volumesArray: { [key: string]: any }[] = await this.get('/volumes')
            volumesArray.sort((a, b) => {
                return a.ID.localeCompare(b.ID);
            })
            const volumes: { [key: string]: { [key: string]: any } } = {}
            for (const vol of volumesArray) {
                volumes[vol.ID] = vol;
            }
            this.volumes.value = volumes
        }
    }

    public async backup(id: string, force: boolean = false) {
        return await this.post('/backup/' + encodeURIComponent(id) + '?force=' + JSON.stringify(force))
    }

    public async rawRestic(volume: string, command: string[]) {
        this.post(
            '/restic/' + encodeURIComponent(volume),
            {
                cmd: command
            }
        )
    }

}

class restic {

}

export default new bivac()