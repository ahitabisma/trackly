<script setup lang="ts">
import Footer from '~/components/common/Footer.vue';
import Navbar from '~/components/common/Navbar.vue';
</script>

<template>
    <div>
        <Navbar />
        <slot />
        <Footer />
    </div>
</template>