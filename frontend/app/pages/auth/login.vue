<template>
    <div class="min-h-screen bg-cream flex flex-col items-center justify-center px-5 py-12 relative overflow-x-hidden">
        <!-- Grid texture background -->
        <div class="absolute inset-0 pointer-events-none" style="
      background-image:
        linear-gradient(rgba(26,22,18,.03) 1px, transparent 1px),
        linear-gradient(90deg, rgba(26,22,18,.03) 1px, transparent 1px);
      background-size: 32px 32px;
    " />

        <!-- Radial glow -->
        <div class="absolute inset-0 pointer-events-none" style="
      background: radial-gradient(ellipse at 60% 0%, rgba(106,176,232,.13) 0%, transparent 55%),
                  radial-gradient(ellipse at 10% 100%, rgba(106,176,232,.08) 0%, transparent 50%);
    " />

        <!-- Logo -->
        <div class="fu relative z-10 mb-8 text-center">
            <NuxtLink to="/" class="inline-block group">
                <div class="font-serif text-ink text-4xl tracking-tight leading-none transition-opacity group-hover:opacity-70"
                    style="letter-spacing:-1.5px; font-family:'Instrument Serif',serif">
                    Trackly
                </div>
                <div class="font-mono text-muted text-[10px] tracking-[.15em] uppercase mt-1">
                    IDX Shareholder Network
                </div>
            </NuxtLink>
        </div>

        <!-- Card -->
        <div class="fu1 relative z-10 w-full max-w-[420px]">

            <!-- Email -->
            <div class="mb-4">
                <div
                    class="flex items-center justify-between font-mono text-[10px] tracking-[.1em] uppercase text-muted mb-1.5">
                    <span>Email</span>
                </div>
                <input v-model="loginForm.email" type="email" placeholder="nama@email.com" autocomplete="email"
                    class="neo-input" :class="{ 'neo-input-err': errors.loginEmail }" @input="clearErr('loginEmail')"
                    @keydown.enter="doLogin" />
                <p v-if="errors.loginEmail" class="text-[11px] text-red-600 mt-1">{{ errors.loginEmail }}</p>
            </div>

            <!-- Password -->
            <div class="mb-5">
                <div
                    class="flex items-center justify-between font-mono text-[10px] tracking-[.1em] uppercase text-muted mb-1.5">
                    <span>Password</span>
                    <NuxtLink to="/auth/forgot"
                        class="text-bluedk hover:text-ink transition-colors normal-case font-sans text-[11px]">
                        Lupa password?
                    </NuxtLink>
                </div>
                <div class="relative">
                    <input v-model="loginForm.password" :type="showLoginPass ? 'text' : 'password'"
                        placeholder="Masukkan password" autocomplete="current-password" class="neo-input pr-12"
                        :class="{ 'neo-input-err': errors.loginPassword }" @input="clearErr('loginPassword')"
                        @keydown.enter="doLogin" />
                    <button
                        class="absolute right-3.5 top-1/2 -translate-y-1/2 text-muted hover:text-ink transition-colors p-1"
                        type="button" @click="showLoginPass = !showLoginPass">
                        <EyeIcon v-if="!showLoginPass" />
                        <EyeOffIcon v-else />
                    </button>
                </div>
                <p v-if="errors.loginPassword" class="text-[11px] text-red-600 mt-1">{{ errors.loginPassword }}</p>
            </div>

            <!-- Submit -->
            <button class="neo-btn-primary mb-3.5" :disabled="loading" @click="doLogin">
                <span v-if="!loading">MASUK</span>
                <span v-else class="flex items-center gap-2">
                    <SpinnerIcon /> Memproses...
                </span>
            </button>

            <!-- Divider -->
            <div class="flex items-center gap-3 my-5">
                <div class="flex-1 h-px bg-bdr" />
                <span class="font-mono text-[10px] tracking-[.1em] uppercase text-muted">atau lanjut dengan</span>
                <div class="flex-1 h-px bg-bdr" />
            </div>

            <!-- Google -->
            <button class="neo-btn-google mb-5" @click="doGoogleAuth">
                <GoogleIcon />
                Masuk dengan Google
            </button>

            <p class="text-center text-[13px] text-muted">
                Belum punya akun?
                <NuxtLink to="/auth/register" class="text-bluedk font-medium hover:underline ml-1">
                    Daftar sekarang
                </NuxtLink>
            </p>
        </div>

        <!-- Toast -->
        <Transition name="toast">
            <div v-if="toast.show"
                class="fixed bottom-7 left-1/2 -translate-x-1/2 px-5 py-3 rounded-xl font-mono text-[11px] tracking-[.06em] z-50 whitespace-nowrap"
                :class="toast.type === 'error' ? 'bg-red-700 text-white' : toast.type === 'success' ? 'bg-green-700 text-white' : 'bg-ink text-cream'"
                style="box-shadow: 4px 4px 0 rgba(0,0,0,.3)">
                {{ toast.message }}
            </div>
        </Transition>
    </div>
</template>

<script setup lang="ts">
import { nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthForm } from '~/composables/useAuth'
import { useAuthIcons } from '~/composables/useAuthIcons'

definePageMeta({ layout: false })

const router = useRouter()
const { EyeIcon, EyeOffIcon, GoogleIcon, SpinnerIcon } = useAuthIcons()
const { loginForm, errors, loading, showLoginPass, toast, showToast, clearErr, getUsers, saveUsers } = useAuthForm()

function triggerSuccess(title: string, msg: string) {
    setTimeout(() => router.push('/'), 3000)
    showToast(title, 'success')
    localStorage.setItem('ksei_session', msg)
}

async function doLogin() {
    errors.loginEmail = ''
    errors.loginPassword = ''
    let valid = true
    const isValidEmail = (e: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e)

    if (!isValidEmail(loginForm.email)) { errors.loginEmail = 'Email tidak valid.'; valid = false }
    if (loginForm.password.length < 6) { errors.loginPassword = 'Password minimal 6 karakter.'; valid = false }
    if (!valid) return

    loading.value = true
    await new Promise(r => setTimeout(r, 1200))
    loading.value = false

    const users = getUsers()
    const user = users.find((u: any) => u.email === loginForm.email && u.password === btoa(loginForm.password))

    if (user) {
        localStorage.setItem('ksei_session', JSON.stringify({ name: user.name, email: user.email }))
        showToast(`Selamat datang kembali, ${user.name}. Mengalihkan...`, 'success')
        setTimeout(() => router.push('/'), 3000)
    } else {
        errors.loginEmail = 'Email atau password salah.'
        errors.loginPassword = 'Cek kembali email & password kamu.'
        showToast('Login gagal — email atau password salah', 'error')
    }
}

async function doGoogleAuth() {
    showToast('Membuka Google Sign In...', '')
    loading.value = true
    await new Promise(r => setTimeout(r, 1600))
    loading.value = false

    const mockName = 'Demo User'
    const mockEmail = 'demo@gmail.com'
    const users = getUsers()
    if (!users.find((u: any) => u.email === mockEmail)) {
        users.push({ name: mockName, email: mockEmail, password: null, provider: 'google', createdAt: new Date().toISOString() })
        saveUsers(users)
    }
    localStorage.setItem('ksei_session', JSON.stringify({ name: mockName, email: mockEmail }))
    showToast(`Selamat datang, ${mockName}. Mengalihkan...`, 'success')
    setTimeout(() => router.push('/'), 3000)
}
</script>

<style>
.tab-shadow {
    box-shadow: 3px 3px 0 #1a1612;
}

@keyframes fadeUp {
    from {
        opacity: 0;
        transform: translateY(18px);
    }

    to {
        opacity: 1;
        transform: translateY(0);
    }
}

.fu {
    animation: fadeUp .6s ease both;
}

.fu1 {
    animation: fadeUp .6s .1s ease both;
}

@keyframes spin {
    to {
        transform: rotate(360deg);
    }
}

.toast-enter-active,
.toast-leave-active {
    transition: all .35s cubic-bezier(.175, .885, .32, 1.275);
}

.toast-enter-from,
.toast-leave-to {
    opacity: 0;
    transform: translateX(-50%) translateY(20px);
}
</style>
