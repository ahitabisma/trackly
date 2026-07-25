import { ref, reactive, computed } from 'vue'
import { authService } from '~/services/auth.service'
import type { User } from '~/types/auth.type'

const SESSION_KEY = 'trackly_session'

export function useAuthSession() {
    const session = ref<{ token: string; user: User } | null>(null)

    const init = () => {
        try {
            const raw = localStorage.getItem(SESSION_KEY)
            if (raw) session.value = JSON.parse(raw)
        } catch { session.value = null }
    }

    const save = (token: string, user: User) => {
        const data = { token, user }
        localStorage.setItem(SESSION_KEY, JSON.stringify(data))
        session.value = data
    }

    const clear = () => {
        localStorage.removeItem(SESSION_KEY)
        session.value = null
    }

    const isLoggedIn = computed(() => !!session.value)
    const isAdmin = computed(() => session.value?.user?.role === 'admin')
    const token = computed(() => session.value?.token)
    const user = computed(() => session.value?.user)

    if (process.client) init()

    return { session, isLoggedIn, isAdmin, token, user, save, clear, init }
}

export function useAuthForm() {
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

    function clearErr(key: string) { delete errors[key] }

    function clearAllErrors() { Object.keys(errors).forEach(k => delete errors[k]) }

    function isValidEmail(e: string) { return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e) }

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

    async function doLogin(): Promise<boolean> {
        clearAllErrors()
        let valid = true
        if (!isValidEmail(loginForm.email)) { errors.loginEmail = 'Email tidak valid.'; valid = false }
        if (loginForm.password.length < 6) { errors.loginPassword = 'Password minimal 6 karakter.'; valid = false }
        if (!valid) return false

        loading.value = true
        try {
            const res = await authService.login({ email: loginForm.email, password: loginForm.password })
            const session = useAuthSession()
            session.save(res.access_token, res.user)
            showToast(`Selamat datang, ${res.user.name}!`, 'success')
            return true
        } catch (e: any) {
            const msg = e?.data?.message || 'Email atau password salah.'
            errors.loginEmail = msg
            errors.loginPassword = msg
            showToast('Login gagal — ' + msg, 'error')
            return false
        } finally {
            loading.value = false
        }
    }

    async function doRegister(): Promise<boolean> {
        clearAllErrors()
        let valid = true
        if (!registerForm.name) { errors.regName = 'Nama tidak boleh kosong.'; valid = false }
        if (!isValidEmail(registerForm.email)) { errors.regEmail = 'Masukkan email yang valid.'; valid = false }
        if (registerForm.password.length < 8) { errors.regPassword = 'Password minimal 8 karakter.'; valid = false }
        if (registerForm.password !== registerForm.confirm) { errors.regConfirm = 'Password tidak cocok.'; valid = false }
        if (!registerForm.terms) { showToast('Setujui syarat & ketentuan terlebih dahulu', 'error'); valid = false }
        if (!valid) return false

        loading.value = true
        try {
            await authService.register({
                name: registerForm.name,
                email: registerForm.email,
                password: registerForm.password,
            })
            const res = await authService.login({ email: registerForm.email, password: registerForm.password })
            const session = useAuthSession()
            session.save(res.access_token, res.user)
            showToast(`Halo ${registerForm.name}! Akun kamu sudah aktif.`, 'success')
            return true
        } catch (e: any) {
            const errData = e?.data
            if (errData?.errors && typeof errData.errors === 'object') {
                Object.entries(errData.errors).forEach(([k, v]) => {
                    errors['reg' + k.charAt(0).toUpperCase() + k.slice(1)] = v as string
                })
            } else {
                errors.regEmail = errData?.message || 'Pendaftaran gagal. Coba lagi.'
            }
            showToast(errors.regEmail || 'Pendaftaran gagal', 'error')
            return false
        } finally {
            loading.value = false
        }
    }

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
        doLogin,
        doRegister,
    }
}
