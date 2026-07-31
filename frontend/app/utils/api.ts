import { $fetch } from 'ofetch'
import { useRuntimeConfig } from '#imports'

const SESSION_KEY = 'trackly_session'

function getToken(): string | null {
    try {
        const raw = localStorage.getItem(SESSION_KEY)
        if (!raw) return null
        const session = JSON.parse(raw)
        return session?.token || null
    } catch { return null }
}

function clearSession() {
    localStorage.removeItem(SESSION_KEY)
    if (process.client) {
        window.location.href = '/auth/login'
    }
}

function createApi() {
    const config = useRuntimeConfig()
    return $fetch.create({
        baseURL: config.public.apiUrl,
        onRequest({ options }) {
            const token = getToken()
            if (token) {
                options.headers = { ...options.headers, Authorization: `Bearer ${token}` }
            }
        },
        onResponseError({ response }) {
            if (response.status === 401) {
                clearSession()
            }
        },
    })
}

let _api: ReturnType<typeof $fetch.create> | undefined

export const api = new Proxy({} as any, {
    get(_, prop) {
        if (!_api) _api = createApi()
        return Reflect.get(_api, prop, _api)
    },
    apply(_, __, args) {
        if (!_api) _api = createApi()
        return Reflect.apply(_api as any, null, args)
    },
})
