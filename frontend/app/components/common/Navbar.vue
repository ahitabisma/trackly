<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthSession } from '~/composables/useAuth'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const isMobile = ref(false)
const profileOpen = ref(false)
const { isLoggedIn, user, clear: logout } = useAuthSession()

const getLinkClassName = (href: string) => {
    let classes = "neo-btn-sm"
    if (route.path === href) classes += " active"
    return classes
}

const doLogout = () => {
    logout()
    router.push('/')
    profileOpen.value = false
    menuOpen.value = false
}

const toggleMenu = () => { menuOpen.value = !menuOpen.value }
const closeMenu = () => { menuOpen.value = false }
const checkScreenSize = () => { isMobile.value = window.innerWidth < 640 }

const handleClickOutside = (e: MouseEvent) => {
    const target = e.target as HTMLElement
    if (!target.closest('[data-profile-dropdown]')) {
        profileOpen.value = false
    }
}

onMounted(() => {
    checkScreenSize()
    window.addEventListener('resize', checkScreenSize)
    document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
    window.removeEventListener('resize', checkScreenSize)
    document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
    <nav class="fixed top-0 left-0 right-0 z-50 bg-cream" style="border-bottom: 1.5px solid rgba(26,22,18,.1)">
        <div class="max-w-7xl mx-auto px-5 sm:px-8 h-14 flex items-center justify-between">
            <NuxtLink to="/" class="font-serif font-bold text-xl text-ink tracking-tight" @click="closeMenu">
                TRACKLY
            </NuxtLink>

            <div class="hidden sm:flex items-center gap-5">
                <NuxtLink v-if="isLoggedIn" to="/analisis" :class="getLinkClassName('/analisis')">
                    Analisis
                </NuxtLink>

                <template v-if="isLoggedIn">
                    <div data-profile-dropdown class="relative">
                        <button class="neo-btn-sm flex items-center gap-2" @click.stop="profileOpen = !profileOpen">
                            <span
                                class="w-6 h-6 rounded-full bg-bluedk text-cream flex items-center justify-center font-mono text-xs font-bold">
                                {{ (user?.name?.charAt(0) || '?').toUpperCase() }}
                            </span>
                            <span class="font-mono text-xs font-bold text-ink">{{ user?.name }}</span>
                            <svg class="w-3 h-3 text-muted transition-transform" :class="{ 'rotate-180': profileOpen }"
                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                        </button>
                        <div v-if="profileOpen"
                            class="absolute right-0 top-full mt-2 bg-white border-2 border-ink rounded-xl min-w-[200px] z-50"
                            style="box-shadow:4px 4px 0 #1a1612">
                            <div class="px-4 py-3 border-b border-bdr">
                                <div class="font-mono text-sm font-bold text-ink">{{ user?.name }}</div>
                                <div class="font-mono text-[10px] text-muted mt-0.5">{{ user?.email }}</div>
                            </div>
                            <NuxtLink to="/profile"
                                class="flex items-center gap-2 px-4 py-2.5 text-sm hover:bg-cream2 font-sans transition-colors"
                                @click="profileOpen = false">
                                <svg class="w-4 h-4 text-muted" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
                                </svg>
                                Profil
                            </NuxtLink>
                            <button
                                class="w-full flex items-center gap-2 px-4 py-2.5 text-sm hover:bg-red-50 font-sans text-red-600 transition-colors"
                                @click="doLogout">
                                <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
                                </svg>
                                Logout
                            </button>
                        </div>
                    </div>
                </template>
                <template v-else>
                    <NuxtLink to="/auth/login" class="neo-btn-sm">Login</NuxtLink>
                </template>
            </div>

            <button v-if="isMobile" class="neo-btn-sm" type="button" :aria-expanded="menuOpen" aria-label="Toggle menu"
                @click="toggleMenu">
                {{ menuOpen ? 'TUTUP' : 'MENU' }}
            </button>
        </div>

        <div v-if="isMobile && menuOpen" class="border-t border-ink/10 bg-cream">
            <div class="max-w-7xl mx-auto px-5 py-3 flex flex-col gap-2">
                <NuxtLink v-if="isLoggedIn" to="/analisis" :class="getLinkClassName('/analisis')" @click="closeMenu">
                    Analisis
                </NuxtLink>

                <template v-if="isLoggedIn">
                    <div class="px-2 py-2 border-b border-bdr mb-1">
                        <div class="font-mono text-sm font-bold text-ink">{{ user?.name }}</div>
                        <div class="font-mono text-[10px] text-muted">{{ user?.email }}</div>
                    </div>
                    <NuxtLink to="/profile" class="neo-btn-sm text-left" @click="closeMenu">Profil</NuxtLink>
                    <button class="neo-btn-sm text-left text-red-600" @click="doLogout">Logout</button>
                </template>
                <template v-else>
                    <NuxtLink to="/auth/login" class="neo-btn-sm" @click="closeMenu">Login</NuxtLink>
                </template>
            </div>
        </div>
    </nav>
</template>
