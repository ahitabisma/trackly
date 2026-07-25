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

            <!-- Name -->
            <div class="mb-4">
                <label class="field-label">Nama Lengkap</label>
                <input v-model="registerForm.name" type="text" placeholder="John Doe" autocomplete="name"
                    class="neo-input" :class="{ 'neo-input-err': errors.regName }" @input="clearErr('regName')" />
                <p v-if="errors.regName" class="text-[11px] text-red-600 mt-1">{{ errors.regName }}</p>
            </div>

            <!-- Email -->
            <div class="mb-4">
                <label class="field-label">Email</label>
                <input v-model="registerForm.email" type="email" placeholder="nama@email.com" autocomplete="email"
                    class="neo-input" :class="{ 'neo-input-err': errors.regEmail }" @input="clearErr('regEmail')" />
                <p v-if="errors.regEmail" class="text-[11px] text-red-600 mt-1">{{ errors.regEmail }}</p>
            </div>

            <!-- Password -->
            <div class="mb-4">
                <label class="field-label">Password</label>
                <div class="relative">
                    <input v-model="registerForm.password" :type="showRegPass ? 'text' : 'password'"
                        placeholder="Min. 8 karakter" autocomplete="new-password" class="neo-input pr-12"
                        :class="{ 'neo-input-err': errors.regPassword }" @input="onPassInput" />
                    <button
                        class="absolute right-3.5 top-1/2 -translate-y-1/2 text-muted hover:text-ink transition-colors p-1"
                        type="button" @click="showRegPass = !showRegPass">
                        <EyeIcon v-if="!showRegPass" />
                        <EyeOffIcon v-else />
                    </button>
                </div>
                <p v-if="errors.regPassword" class="text-[11px] text-red-600 mt-1">{{ errors.regPassword }}</p>

                <!-- Strength meter -->
                <div v-if="registerForm.password.length > 0" class="mt-2">
                    <div class="flex gap-1 mb-1">
                        <div v-for="i in 4" :key="i" class="flex-1 h-[3px] rounded-full transition-all duration-300"
                            :class="strengthBarClass(i)" />
                    </div>
                    <span class="font-mono text-[10px] tracking-[.06em] uppercase text-muted">
                        {{ strengthLabel }}
                    </span>
                </div>
            </div>

            <!-- Confirm Password -->
            <div class="mb-5">
                <label class="field-label">Konfirmasi Password</label>
                <div class="relative">
                    <input v-model="registerForm.confirm" :type="showConfirmPass ? 'text' : 'password'"
                        placeholder="Ulangi password" autocomplete="new-password" class="neo-input pr-12"
                        :class="{ 'neo-input-err': errors.regConfirm }" @input="clearErr('regConfirm')" />
                    <button
                        class="absolute right-3.5 top-1/2 -translate-y-1/2 text-muted hover:text-ink transition-colors p-1"
                        type="button" @click="showConfirmPass = !showConfirmPass">
                        <EyeIcon v-if="!showConfirmPass" />
                        <EyeOffIcon v-else />
                    </button>
                </div>
                <p v-if="errors.regConfirm" class="text-[11px] text-red-600 mt-1">{{ errors.regConfirm }}</p>
            </div>

            <!-- Terms -->
            <div class="flex items-start gap-2.5 mb-5">
                <input v-model="registerForm.terms" type="checkbox" id="terms"
                    class="neo-checkbox mt-0.5 flex-shrink-0" />
                <label for="terms" class="text-[12px] text-muted leading-relaxed cursor-pointer">
                    Saya setuju dengan
                    <a href="#" class="text-bluedk hover:underline">Syarat &amp; Ketentuan</a> dan
                    <a href="#" class="text-bluedk hover:underline">Kebijakan Privasi</a> KSEIGraph.
                </label>
            </div>

            <!-- Submit -->
            <button class="neo-btn-primary mb-3.5" :disabled="loading" @click="handleRegister">
                <span v-if="!loading">BUAT AKUN</span>
                <span v-else class="flex items-center gap-2">
                    <SpinnerIcon /> Memproses...
                </span>
            </button>

            <!-- Divider -->
            <div class="flex items-center gap-3 my-5">
                <div class="flex-1 h-px bg-bdr" />
                <span class="font-mono text-[10px] tracking-[.1em] uppercase text-muted">atau daftar dengan</span>
                <div class="flex-1 h-px bg-bdr" />
            </div>

            <!-- Google -->
            <button class="neo-btn-google mb-5" @click="doGoogleAuth">
                <GoogleIcon />
                Daftar dengan Google
            </button>

            <p class="text-center text-[13px] text-muted">
                Sudah punya akun?
                <NuxtLink to="/auth/login" class="text-bluedk font-medium hover:underline ml-1">
                    Masuk di sini
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
import { useRouter } from 'vue-router'
import { useAuthForm } from '~/composables/useAuth'
import { useAuthIcons } from '~/composables/useAuthIcons'

definePageMeta({ layout: false })

const router = useRouter()
const { EyeIcon, EyeOffIcon, GoogleIcon, SpinnerIcon } = useAuthIcons()
const { registerForm, errors, loading, showRegPass, showConfirmPass, toast, showToast, clearErr, strengthLabel, strengthBarClass, doRegister } = useAuthForm()

function onPassInput() {
    clearErr('regPassword')
}

async function handleRegister() {
    const ok = await doRegister()
    if (ok) setTimeout(() => router.push('/analisis'), 1500)
}

async function doGoogleAuth() {
    showToast('Fitur Google Sign In belum tersedia', 'error')
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
