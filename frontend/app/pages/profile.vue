<template>
    <main class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24">
        <div class="max-w-xl mx-auto px-5 sm:px-8">
            <div class="fu mb-8">
                <h1 class="font-serif text-4xl text-ink" style="letter-spacing:-1px">Profil</h1>
                <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">Informasi akun kamu</p>
            </div>

            <div class="fu1 bg-card border-2 border-ink rounded-2xl p-6 sm:p-8" style="box-shadow:4px 4px 0 #1a1612">
                <div class="flex items-center gap-4 mb-6 pb-6 border-b border-bdr">
                    <div
                        class="w-16 h-16 rounded-full bg-bluedk text-cream flex items-center justify-center font-serif text-3xl font-bold">
                        {{ (user?.name?.charAt(0) || '?').toUpperCase() }}
                    </div>
                    <div>
                        <div class="font-serif text-2xl text-ink">{{ user?.name }}</div>
                        <div class="font-mono text-xs text-muted">{{ user?.email }}</div>
                    </div>
                </div>

                <div class="space-y-4">
                    <div class="flex items-center justify-between py-2">
                        <span class="font-mono text-xs text-muted uppercase tracking-wider">Nama</span>
                        <span class="font-sans text-sm font-bold text-ink">{{ user?.name || '—' }}</span>
                    </div>
                    <div class="flex items-center justify-between py-2 border-t border-bdr">
                        <span class="font-mono text-xs text-muted uppercase tracking-wider">Email</span>
                        <span class="font-sans text-sm text-ink">{{ user?.email || '—' }}</span>
                    </div>
                    <div v-if="user?.role === 'admin'" class="flex items-center justify-between py-2 border-t border-bdr">
                        <span class="font-mono text-xs text-muted uppercase tracking-wider">Role</span>
                        <span class="font-mono text-xs font-bold px-2 py-0.5 rounded-md"
                            :class="user?.role === 'admin' ? 'bg-amber-100 text-amber-800' : 'bg-bluebg text-bluedk'">
                            {{ user?.role === 'admin' ? 'Admin' : 'User' }}
                        </span>
                    </div>
                </div>

                <div class="mt-8 pt-6 border-t border-bdr">
                    <button class="neo-btn-sm w-full text-red-600 flex items-center justify-center gap-2" @click="doLogout">
                        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline points="16 17 21 12 16 7" /><line x1="21" y1="12" x2="9" y2="12" />
                        </svg>
                        Logout
                    </button>
                </div>
            </div>
        </div>
    </main>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'main', middleware: 'auth' })

import { useAuthSession } from '~/composables/useAuth'
import { useRouter } from 'vue-router'

const router = useRouter()
const { user, clear: logout } = useAuthSession()

const doLogout = () => {
    logout()
    router.push('/')
}
</script>
