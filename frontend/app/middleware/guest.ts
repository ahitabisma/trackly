export default defineNuxtRouteMiddleware((to, from) => {
    if (process.client) {
        const raw = localStorage.getItem('trackly_session')
        if (!raw) return
        try {
            const session = JSON.parse(raw)
            if (session.token && session.user) {
                return navigateTo('/')
            }
        } catch {}
    }
})
