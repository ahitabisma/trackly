<script setup lang="ts">
const currentYear = computed(() => new Date().getFullYear());
</script>

<template>
    <footer class="bg-cream" style="border-top: 1.5px solid rgba(26,22,18,.1)">
        <div class="max-w-7xl mx-auto px-8 py-8 flex flex-wrap items-center justify-between gap-4">
            <div>
                <NuxtLink to="/" class="font-serif text-lg text-ink">
                    Trackly
                </NuxtLink>
                <p class="text-xs text-muted mt-1">
                    Data sourced from KSEI. For research &amp; educational purposes
                    only.
                </p>
            </div>

            <div class="flex gap-6 text-sm text-muted">
                <NuxtLink to="/about" class="hover:text-ink transition-colors">
                    About
                </NuxtLink>
                <NuxtLink to="/docs" class="hover:text-ink transition-colors">
                    API Docs
                </NuxtLink>
                <a href="https://github.com" target="_blank" class="hover:text-ink transition-colors">
                    GitHub
                </a>
            </div>

            <div class="font-mono text-xs text-muted">
                © {{ currentYear }} Trackly
            </div>
        </div>
    </footer>
</template>