<script setup lang="ts">
import { computed } from 'vue';

const currentYear = computed(() => new Date().getFullYear());
</script>

<template>
    <footer class="bg-cream" style="border-top: 1.5px solid rgba(26,22,18,.1)">
        <div class="max-w-7xl mx-auto px-5 sm:px-8 py-8">
            <div class="flex flex-col items-center justify-center  md:flex-row md:items-center md:justify-between gap-4">
                <div>
                    <NuxtLink to="/" class="font-serif text-lg text-ink">
                        Trackly
                    </NuxtLink>
                </div>

                <div class="font-mono text-xs text-muted">
                    © {{ currentYear }} Trackly
                </div>
            </div>
        </div>
    </footer>
</template>