export default defineNuxtRouteMiddleware((to, from) => {
    if (process.client) {
        const raw = localStorage.getItem('trackly_session')
        if (!raw) {
            return navigateTo('/auth/login')
        }
        try {
            const session = JSON.parse(raw)
            if (!session.token || !session.user) {
                localStorage.removeItem('trackly_session')
                return navigateTo('/auth/login')
            }
        } catch {
            localStorage.removeItem('trackly_session')
            return navigateTo('/auth/login')
        }
    }
})
