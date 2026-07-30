import { $fetch } from 'ofetch'

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

export const api = $fetch.create({
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
