<script setup lang="ts">
const route = useRoute();

const getLinkClassName = (href: string) => {
    let classes = "neo-btn-sm";

    if (route.path === href) {
        classes += " active";
    }

    return classes;
};
</script>

<template>
    <nav class="fixed top-0 left-0 right-0 z-50 bg-cream" style="border-bottom: 1.5px solid rgba(26,22,18,.1)">
        <div class="max-w-7xl mx-auto px-5 sm:px-8 h-14 flex items-center justify-between">
            <NuxtLink to="/" class="font-serif text-xl text-ink tracking-tight">
                TRACKLY
            </NuxtLink>

            <div class="flex items-center gap-5">
                <NuxtLink to="/" :class="getLinkClassName('/')">
                    1% Share Ownership
                </NuxtLink>

                <NuxtLink to="/" :class="getLinkClassName('/')">
                    Shareholder Concentration List
                </NuxtLink>
            </div>
        </div>
    </nav>
</template>