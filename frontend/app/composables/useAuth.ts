import { ref, reactive, computed } from 'vue'

export const useAuthForm = () => {
    const loginForm = reactive({ email: '', password: '' })
    const registerForm = reactive({ name: '', email: '', password: '', confirm: '', terms: false })
    const forgotEmail = ref('')
    const errors = reactive<Record<string, string>>({})
    const loading = ref(false)

    const showLoginPass = ref(false)
    const showRegPass = ref(false)
    const showConfirmPass = ref(false)

    const toast = reactive({ show: false, message: '', type: '' })
    let toastTimer: ReturnType<typeof setTimeout>

    function showToast(message: string, type = '') {
        clearTimeout(toastTimer)
        toast.message = message
        toast.type = type
        toast.show = true
        toastTimer = setTimeout(() => { toast.show = false }, 3500)
    }

    function clearErr(key: string) {
        delete errors[key]
    }

    function clearAllErrors() {
        Object.keys(errors).forEach(k => delete errors[k])
    }

    function isValidEmail(e: string) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e)
    }

    const strengthScore = computed(() => {
        const v = registerForm.password
        if (!v) return 0
        let s = 0
        if (v.length >= 8) s++
        if (/[A-Z]/.test(v)) s++
        if (/[0-9]/.test(v)) s++
        if (/[^A-Za-z0-9]/.test(v)) s++
        return s
    })

    const strengthLabel = computed(() => {
        return ['Sangat Lemah', 'Lemah', 'Cukup', 'Kuat', 'Sangat Kuat'][strengthScore.value] ?? ''
    })

    function strengthBarClass(bar: number) {
        const score = strengthScore.value
        if (bar > score) return 'bg-bdr'
        const cls = ['', 'bg-red-400', 'bg-amber-400', 'bg-blue', 'bg-green-500']
        return cls[score] ?? 'bg-bdr'
    }

    const USERS_KEY = 'ksei_users'
    function getUsers(): any[] {
        try { return JSON.parse(localStorage.getItem(USERS_KEY) || '[]') } catch { return [] }
    }
    function saveUsers(u: any[]) { localStorage.setItem(USERS_KEY, JSON.stringify(u)) }

    return {
        loginForm,
        registerForm,
        forgotEmail,
        errors,
        loading,
        showLoginPass,
        showRegPass,
        showConfirmPass,
        toast,
        showToast,
        clearErr,
        clearAllErrors,
        isValidEmail,
        strengthScore,
        strengthLabel,
        strengthBarClass,
        getUsers,
        saveUsers,
    }
}
