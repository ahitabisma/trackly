<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';

const route = useRoute();
const menuOpen = ref(false);
const isMobile = ref(false);

const getLinkClassName = (href: string) => {
    let classes = "neo-btn-sm";

    if (route.path === href) {
        classes += " active";
    }

    return classes;
};

const toggleMenu = () => {
    menuOpen.value = !menuOpen.value;
};

const closeMenu = () => {
    menuOpen.value = false;
};

const checkScreenSize = () => {
    isMobile.value = window.innerWidth < 640;
};

onMounted(() => {
    checkScreenSize();
    window.addEventListener('resize', checkScreenSize);
});

onUnmounted(() => {
    window.removeEventListener('resize', checkScreenSize);
});
</script>

<template>
    <nav class="fixed top-0 left-0 right-0 z-50 bg-cream" style="border-bottom: 1.5px solid rgba(26,22,18,.1)">
        <div class="max-w-7xl mx-auto px-5 sm:px-8 h-14 flex items-center justify-between">
            <NuxtLink to="/" class="font-serif font-bold text-xl text-ink tracking-tight" @click="closeMenu">
                TRACKLY
            </NuxtLink>

            <div class="hidden sm:flex items-center gap-5">
                <NuxtLink to="/" :class="getLinkClassName('/')">
                    1% Share Ownership
                </NuxtLink>

                <NuxtLink to="/high-concentration-list" :class="getLinkClassName('/high-concentration-list')">
                    High Concentration List
                </NuxtLink>

                <NuxtLink to="/analisis" :class="getLinkClassName('/analisis')">
                    Analisis
                </NuxtLink>
            </div>

            <button v-if="isMobile" class="neo-btn-sm" type="button" :aria-expanded="menuOpen" aria-label="Toggle menu"
                @click="toggleMenu">
                {{ menuOpen ? 'TUTUP' : 'MENU' }}
            </button>
        </div>

        <div v-if="isMobile && menuOpen" class="border-t border-ink/10 bg-cream">
            <div class="max-w-7xl mx-auto px-5 py-3 flex flex-col gap-2">
                <NuxtLink to="/" :class="getLinkClassName('/')" @click="closeMenu">
                    1% Share Ownership
                </NuxtLink>
                <NuxtLink to="/high-concentration-list" :class="getLinkClassName('/high-concentration-list')"
                    @click="closeMenu">
                    High Concentration List
                </NuxtLink>
                <NuxtLink to="/analisis" :class="getLinkClassName('/analisis')" @click="closeMenu">
                    Analisis
                </NuxtLink>
            </div>
        </div>
    </nav>
</template>