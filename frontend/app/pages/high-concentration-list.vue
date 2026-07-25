<template>
    <main class="min-h-screen bg-cream text-ink font-sans font-bold antialiased overflow-x-hidden">

        <!-- ================================================
             HERO SECTION
        ================================================ -->
        <section class="min-h-[32vh] flex flex-col items-center justify-center text-center px-5 pt-20 pb-10"
            :style="{ background: 'linear-gradient(180deg,#cce8f5 0%,#f2ede4 50%)' }">
            <section class="w-full max-w-5xl mx-auto px-4 sm:px-8 pt-8 pb-24">
                <!-- Header row: stats + update date -->
                <div class="flex flex-col items-center justify-center md:flex-row md:items-end md:justify-between gap-4 mb-5">
                    <div>
                        <h2 class="font-serif text-2xl text-ink mb-1" :style="{ letterSpacing: '-1px' }">
                            High Concentration List
                        </h2>
                    </div>

                    <!-- Update date badge -->
                    <div class="flex items-center gap-2 bg-card border-2 border-ink rounded-xl px-4 py-2 self-start sm:self-auto"
                        :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                        <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse flex-shrink-0"></span>
                        <span class="font-mono text-xs text-muted uppercase tracking-wider">Terakhir diperbarui</span>
                        <span class="font-mono text-xs text-ink font-bold">{{ lastUpdated }}</span>
                    </div>
                </div>

                <!-- STATS CARDS -->
                <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 mb-6">
                    <div v-for="(s, i) in stats" :key="i" class="bg-card border-2 border-ink rounded-xl p-4"
                        :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                        <div class="font-mono text-xs uppercase tracking-wider text-muted mb-1">{{ s.label }}</div>
                        <div class="font-serif text-2xl text-ink leading-none mb-1">{{ s.val }}</div>
                        <div class="text-xs text-muted font-sans font-normal">{{ s.sub }}</div>
                    </div>
                </div>

                <!-- TABLE -->
                <div class="border-2 border-ink rounded-2xl overflow-hidden bg-white"
                    :style="{ boxShadow: '4px 4px 0 #1a1612' }">
                    <div class="overflow-x-auto">
                        <table class="w-full min-w-[600px] text-sm border-collapse">
                            <thead>
                                <tr class="border-b-2 border-ink bg-card">
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider w-12">
                                        No</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider w-24">
                                        Kode</th>
                                    <th
                                        class="text-left px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Nama Perusahaan</th>
                                    <th
                                        class="text-right px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider w-36">
                                        Konsentrasi</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="(row, i) in data" :key="row.code"
                                    class="border-b border-bdr transition-colors hover:bg-card group cursor-pointer"
                                    @click="goToTicker(row.code)">

                                    <!-- No -->
                                    <td class="px-5 py-4 text-center font-mono text-xs text-muted">
                                        {{ i + 1 }}
                                    </td>

                                    <!-- Kode -->
                                    <td class="px-5 py-4 text-center">
                                        <span
                                            class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md">
                                            {{ row.code }}
                                        </span>
                                    </td>

                                    <!-- Nama -->
                                    <td class="px-5 py-4 text-left">
                                        <span class="text-sm text-ink group-hover:text-bluedk transition-colors">
                                            {{ row.name }}
                                        </span>
                                    </td>

                                    <!-- Konsentrasi -->
                                    <td class="px-5 py-4 text-right">
                                        <div class="flex items-center justify-end gap-3">
                                            <!-- Progress bar -->
                                            <div
                                                class="w-20 h-1.5 bg-bluebg rounded-full overflow-hidden hidden sm:block">
                                                <div class="h-full rounded-full transition-all" :style="{
                                                    width: row.concentration + '%',
                                                    background: concentrationColor(row.concentration)
                                                }">
                                                </div>
                                            </div>
                                            <span class="font-mono font-bold text-sm"
                                                :style="{ color: concentrationColor(row.concentration) }">
                                                {{ row.concentration.toFixed(2) }}%
                                            </span>
                                        </div>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <!-- Footer note -->
                    <div
                        class="px-5 py-3 border-t border-bdr bg-card flex items-center justify-between flex-wrap gap-2">
                        <span class="font-mono text-xs text-muted">
                            // Data bersumber dari laporan KSEI
                        </span>
                        <span class="font-mono text-xs text-muted">
                            Klik baris untuk melihat detail kepemilikan
                        </span>
                    </div>
                </div>

                <!-- Info box -->
                <!-- <div class="mt-6 border-2 border-ink rounded-xl p-4 bg-card flex gap-3 items-start"
                :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                <span class="text-base mt-0.5 flex-shrink-0">ℹ</span>
                <div>
                    <div class="font-mono text-xs font-bold text-ink uppercase tracking-wider mb-1">Tentang Data Ini
                    </div>
                    <p class="text-xs text-muted font-sans font-normal leading-relaxed">
                        Konsentrasi kepemilikan dihitung dari total persentase saham yang dipegang oleh pemegang saham
                        dengan kepemilikan ≥1% (data KSEI). Nilai mendekati 100% mengindikasikan saham yang sangat
                        terkonsentrasi pada segelintir pemegang saham — umumnya founder atau entitas pengendali.
                    </p>
                </div>
            </div> -->
            </section>
        </section>

        <!-- ================================================
             TABLE SECTION
        ================================================ -->
        <!-- <section class="w-full max-w-5xl mx-auto px-4 sm:px-8 pt-8 pb-24"> -->



    </main>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

definePageMeta({
    layout: 'main',
})

const router = useRouter()

// ── Data ─────────────────────────────────────────────────────────────────────
const lastUpdated = '31 Mar 2026'

const data = ref([
    { code: 'ROCK', name: 'PT Rockfields Properti Indonesia Tbk', concentration: 99.85 },
    { code: 'IFSH', name: 'PT Ifishdeco Tbk', concentration: 99.77 },
    { code: 'SOTS', name: 'PT Satria Mega Kencana Tbk', concentration: 98.35 },
    { code: 'AGII', name: 'PT Samator Indo Gas Tbk', concentration: 97.75 },
    { code: 'BREN', name: 'PT Barito Renewables Energy Tbk', concentration: 97.31 },
    { code: 'MGLV', name: 'PT Panca Anugrah Wisesa Tbk', concentration: 95.94 },
    { code: 'DSSA', name: 'PT Dian Swastatika Sentosa Tbk', concentration: 95.76 },
    { code: 'LUCY', name: 'PT Lima Dua Lima Tiga Tbk', concentration: 95.47 },
    { code: 'RLCO', name: 'PT Abadi Lestari Indonesia Tbk', concentration: 95.35 },
])

// ── Stats ─────────────────────────────────────────────────────────────────────
const stats = computed(() => [
    {
        label: 'Rata-rata Konsentrasi',
        val: (data.value.reduce((s, r) => s + r.concentration, 0) / data.value.length).toFixed(2) + '%',
        sub: 'dari 9 emiten terdaftar',
    },
    {
        label: 'Tertinggi',
        val: data.value[0]?.concentration.toFixed(2) + '%',
        sub: data.value[0]?.code + ' · ' + data.value[0]?.name.split(' ').slice(0, 3).join(' '),
    },
    {
        label: 'Terendah dalam daftar',
        val: data.value?.[data.value?.length - 1]?.concentration.toFixed(2) + '%',
        sub: data.value?.[data.value?.length - 1]?.code,
    },
])

// ── Helpers ───────────────────────────────────────────────────────────────────
function concentrationColor(pct: number): string {
    if (pct >= 99) return '#c0392b'
    if (pct >= 97) return '#e67e22'
    if (pct >= 95) return '#d4a017'
    return '#2980b9'
}

function goToTicker(code: string) {
    router.push({ path: '/', query: { ticker: code } })
}
</script>