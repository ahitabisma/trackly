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
            <NuxtLink to="/auth/login"
                class="inline-flex items-center gap-1.5 text-[13px] text-muted hover:text-ink transition-colors mb-5">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    stroke-linecap="round">
                    <path d="M19 12H5" />
                    <path d="m12 5-7 7 7 7" />
                </svg>
                Kembali ke login
            </NuxtLink>

            <h3 class="font-serif text-[24px] text-ink mb-1.5"
                style="letter-spacing:-.5px; font-family:'Instrument Serif',serif">
                Reset Password
            </h3>
            <p class="text-[13px] text-muted leading-relaxed mb-5">
                Masukkan email terdaftar. Kami akan kirimkan link untuk reset password kamu.
            </p>

            <div class="mb-5">
                <label class="field-label">Email</label>
                <input v-model="forgotEmail" type="email" placeholder="nama@email.com" class="neo-input"
                    :class="{ 'neo-input-err': errors.forgotEmail }" @keydown.enter="doForgot" />
                <p v-if="errors.forgotEmail" class="text-[11px] text-red-600 mt-1">{{ errors.forgotEmail }}</p>
            </div>

            <button class="neo-btn-primary" :disabled="loading" @click="doForgot">
                <span v-if="!loading">KIRIM LINK RESET</span>
                <span v-else class="flex items-center gap-2">
                    <SpinnerIcon /> Mengirim...
                </span>
            </button>
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
import { useRouter } from 'vue-router'
import { useAuthForm } from '~/composables/useAuth'
import { useAuthIcons } from '~/composables/useAuthIcons'

definePageMeta({ layout: false })

const router = useRouter()
const { SpinnerIcon } = useAuthIcons()
const { forgotEmail, errors, loading, toast, showToast } = useAuthForm()

const isValidEmail = (e: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e)

async function doForgot() {
    delete errors.forgotEmail
    if (!isValidEmail(forgotEmail.value)) {
        errors.forgotEmail = 'Masukkan email yang valid.'
        return
    }
    loading.value = true
    await new Promise(r => setTimeout(r, 1200))
    loading.value = false
    showToast(`Link reset dikirim ke ${forgotEmail.value}`, 'success')
    setTimeout(() => { router.push('/auth/login') }, 2000)
}
</script>

<style>
.field-label {
    display: block;
    font-family: 'Share Tech Mono', monospace;
    font-size: 10px;
    letter-spacing: .1em;
    text-transform: uppercase;
    margin-bottom: 0.375rem;
    color: #8a8178;
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
