// this file doesn't directly form part of the app, but gets embedded in the nswag generated client code // ignore
import * as generated from './client'

// due to nswag issue #1859
// we've set  "importRequiredTypes": false in the nswag config
// and add the imports here instead using ES6 syntax
import moment from 'moment'

class ClientBase {
    // Get an up to date ID token - set by the caller
    // If null, no authorization header is included in requests
    getIdToken: (() => Promise<string>) | null = null

    protected async transformOptions(options: { headers?: any }) {
        // should be RequestInit, but it doesn't actually appear to be - it's just an object

        if (options.headers && this.getIdToken) {
            options.headers['Authorization'] = 'Bearer ' + (await this.getIdToken())
        }
        return Promise.resolve(options)
    }
}
