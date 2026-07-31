import { $fetch } from 'ofetch'

const SESSION_KEY = 'trackly_session'

// Vite replaces NUXT_PUBLIC_* vars at build/dev time
const BASE_URL = import.meta.env.NUXT_PUBLIC_API_URL || 'http://localhost:8080'

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

export const api = $fetch.create({
    baseURL: BASE_URL,
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
