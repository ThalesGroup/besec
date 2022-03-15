import * as api from './client'

// The client is a singleton here, rather than in the store, because it can't be serialised
export default class ApiClient {
    private static _instance: ApiClient
    private _client!: api.Client

    private constructor() {
        let apiHost
        if (process.env.NODE_ENV === 'production') {
            apiHost = '/' // in prod, API host is assumed to be the same as the host the frontend is on
        } else {
            if (!process.env.REACT_APP_API_HOST) {
                throw new Error('REACT_APP_API_HOST environment variable must be set when not a production build.')
            }
            apiHost = process.env.REACT_APP_API_HOST
        }
        const apiVersion = 'v1alpha1'

        this._client = new api.Client(apiHost + apiVersion, { fetch: fetchWithTimeout })
    }

    public static get client() {
        if (!ApiClient._instance) {
            ApiClient._instance = new ApiClient()
        }
        return ApiClient._instance._client
    }
}

// None of the API Client requests should take long to resolve - there are no heavy operations
// Timeout quickly so the user (or developer!) isn't left wondering
async function fetchWithTimeout(input: RequestInfo, init?: RequestInit) {
    const timeout = 8000

    const controller = new AbortController()
    const id = setTimeout(() => controller.abort(), timeout)
    const response = await fetch(input, {
        ...init,
        signal: controller.signal,
    })
    clearTimeout(id)
    return response
}
