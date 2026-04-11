<template>
    <main class="min-h-screen bg-cream text-ink font-sans font-bold antialiased overflow-x-hidden">


        <!-- ================================================
         HERO SECTION
    ================================================ -->
        <section class="min-h-[50vh] flex flex-col items-center justify-center text-center px-5 pt-20 pb-10"
            :style="{ background: 'linear-gradient(180deg,#cce8f5 0%,#f2ede4 50%)' }">

            <!-- Slogan + Update Badge -->
            <div class="w-full max-w-3xl mb-6 flex flex-col items-center gap-4">
                <div class="font-mono text-md uppercase tracking-[0.3em] text-bluedk">Pantau Kepemilikan Saham Jadi
                    Lebih Mudah</div>
            </div>

            <!-- Ticker berjalan -->
            <aside class="relative w-full max-w-5xl mb-8">
                <div class="w-full bg-white rounded-xl overflow-hidden">
                    <div class="tradingview-widget-container" style="width: 100%; height: 44px;">
                        <iframe scrolling="no" allowtransparency="true" frameborder="0"
                            src="https://www.tradingview-widget.com/embed-widget/ticker-tape/?locale=en#%7B%22symbols%22%3A%5B%7B%22proName%22%3A%22IDX%3ABUMI%22%2C%22title%22%3A%22BUMI%22%7D%2C%7B%22proName%22%3A%22IDX%3ADEWA%22%2C%22title%22%3A%22DEWA%22%7D%2C%7B%22proName%22%3A%22IDX%3ADCII%22%2C%22title%22%3A%22DCII%22%7D%2C%7B%22proName%22%3A%22IDX%3ABBCA%22%2C%22title%22%3A%22BBCA%22%7D%2C%7B%22proName%22%3A%22IDX%3AANTM%22%2C%22title%22%3A%22ANTM%22%7D%2C%7B%22proName%22%3A%22IDX%3AARCI%22%2C%22title%22%3A%22ARCI%22%7D%5D%2C%22showSymbolLogo%22%3Atrue%2C%22colorTheme%22%3A%22light%22%2C%22isTransparent%22%3Afalse%2C%22displayMode%22%3A%22large%22%2C%22width%22%3A%22100%25%22%2C%22height%22%3A44%7D"
                            title="ticker tape TradingView widget" lang="en"
                            style="user-select: none; box-sizing: border-box; display: block; height: 44px; width: 100%;"></iframe>
                    </div>
                </div>
            </aside>

            <!-- SEARCH BOX -->
            <div class="fu3 w-full max-w-xl relative">
                <div class="flex flex-col sm:flex-row gap-3 items-stretch">
                    <div class="relative flex-1">
                        <input id="hero-search" v-model="searchInput" class="search-input" type="text"
                            placeholder="Cari ticker... mis. BRMS, BBCA" @keydown.enter="doSearch"
                            @input="handleSearchInput" @focus="onSearchInput" maxlength="4" />
                        <svg class="absolute right-4 top-1/2 -translate-y-1/2 text-muted w-4 h-4 pointer-events-none"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8" />
                            <path d="m21 21-4.35-4.35" />
                        </svg>

                        <!-- Dropdown suggestions -->
                        <div v-if="showDropdown" ref="dropdownRef"
                            class="absolute top-full mt-2 left-0 right-0 bg-white border-2 border-ink rounded-xl max-h-64 min-h-fit overflow-y-auto overflow-x-hidden z-50">
                            <div v-if="suggestions.length">
                                <div v-for="s in suggestions" :key="s.id"
                                    class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bluebg border-b border-bdr last:border-0 transition-colors"
                                    @mousedown.prevent="selectCompany(s)">
                                    <span
                                        class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md min-w-[46px] text-center">
                                        {{ s.kode }}
                                    </span>
                                    <span class="text-sm text-ink2">{{ s.nama_perusahaan }}</span>
                                </div>
                            </div>
                            <div v-else class="px-4 py-4 text-center">
                                <p class="text-sm text-muted font-mono">Ticker tidak ditemukan</p>
                            </div>
                        </div>
                    </div>

                    <button class="neo-btn w-full sm:w-auto flex items-center justify-center" @click="doSearch">TELUSURI
                        →</button>
                </div>

                <!-- Quick load chips -->
                <div class="flex gap-2 mt-4 flex-wrap justify-center items-center">
                    <span class="text-xs text-muted font-mono">COBA:</span>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="() => quickLoad('BUMI')">BUMI</button>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="() => quickLoad('DEWA')">DEWA</button>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="() => quickLoad('IMPC')">IMPC</button>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="() => quickLoad('BBCA')">BBCA</button>
                </div>
            </div>

            <!-- ================================================
         TABLE SECTION
    ================================================ -->
            <div id="table-section" class="w-full max-w-7xl mx-auto px-4 sm:px-8 pt-8 pb-24">
                <!-- Update date badge -->
                <div class="flex items-center justify-center w-full mb-6">
                    <div class="flex items-center gap-2 bg-card border-2 border-ink rounded-xl px-4 py-2 w-fit text-center"
                        :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                        <span class="w-2 h-2 rounded-full bg-green-500 animate-pulse flex-shrink-0"></span>
                        <span class="font-mono text-xs text-muted uppercase tracking-wider">Terakhir diperbarui</span>
                        <span class="font-mono text-xs text-ink font-bold">{{ lastUpdated }}</span>
                    </div>
                </div>

                <div>
                    <h2 class="font-serif text-3xl text-ink mb-1" :style="{ letterSpacing: '-1px' }">
                        Data Pemegang Saham di atas 1%
                    </h2>
                    <p class="text-sm text-muted">
                        Tampilan terstruktur data kepemilikan saham berdasarkan persentase.
                    </p>
                </div>

                <!-- STATS CARDS -->
                <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 my-4">
                    <div v-for="(s, i) in stats" :key="i" class="bg-card border-2 border-ink rounded-xl p-4"
                        :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                        <div class="font-mono text-xs uppercase tracking-wider text-muted mb-1">{{ s.label }}</div>
                        <div class="font-serif text-2xl text-ink leading-none mb-1">{{ s.val }}</div>
                        <div class="text-xs text-muted">{{ s.sub }}</div>
                    </div>
                </div>

                <!-- Loading state -->
                <div v-if="shareholdingLoading" class="text-center py-14">
                    <div class="animate-spin inline-block w-8 h-8 border-4 border-ink border-t-blue rounded-full"></div>
                    <p class="text-sm text-muted mt-4">Memuat data kepemilikan saham...</p>
                </div>

                <!-- Error state -->
                <div v-else-if="shareholdingError" class="bg-red-50 border-2 border-red-300 rounded-2xl p-6">
                    <p class="text-sm text-red-700 font-mono">Galat: {{ shareholdingError }}</p>
                </div>

                <!-- Table -->
                <div v-else class="border-2 border-ink rounded-2xl overflow-hidden bg-white"
                    :style="{ boxShadow: '4px 4px 0 #1a1612' }">
                    <div class="overflow-x-auto">
                        <table class="data-table w-full min-w-[920px] text-sm border-collapse">
                            <thead>
                                <tr class="border-b-2 border-ink bg-card">
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Kode</th>
                                    <th
                                        class="text-left px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Nama Investor</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Tipe</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Asal</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        %</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Saham</th>
                                    <th
                                        class="text-center px-5 py-3 font-mono text-xs text-muted font-bold uppercase tracking-wider">
                                        Tanggal</th>
                                </tr>
                            </thead>
                            <tbody>
                                <template v-if="tableData.length > 0">
                                    <tr v-for="(row, i) in tableData" :key="i"
                                        class="border-b border-bdr transition-colors hover:bg-card">
                                        <td class="px-5 py-3">
                                            <span
                                                class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md">
                                                {{ row.company_kode }}
                                            </span>
                                        </td>
                                        <td class="px-5 py-3 text-sm text-ink max-w-xs text-left">{{ row.investor_name
                                            }}</td>
                                        <td class="px-5 py-3">
                                            <span
                                                class="text-xs font-bold px-2 py-0.5 rounded-full bg-purple-100 text-purple-700">
                                                {{ row.investor_type }}
                                            </span>
                                        </td>
                                        <td class="px-5 py-3">
                                            <span class="font-mono text-xs"
                                                :class="row.local_foreign === 'F' ? 'text-violet-700' : 'text-green-700'">{{
                                                    row.local_foreign === 'F' ? 'FOREIGN' : 'LOCAL' }}</span>
                                        </td>
                                        <td class="px-5 py-3 text-right font-mono font-bold text-sm">{{
                                            row.percentage.toFixed(2) }}%</td>
                                        <td class="px-5 py-3 text-right font-mono text-xs text-muted">
                                            {{ row.total_holding_shares ? row.total_holding_shares.toLocaleString() :
                                                '—' }}
                                        </td>
                                        <td class="px-5 py-3 text-center font-mono text-xs text-muted">
                                            {{ row.date }}
                                        </td>
                                    </tr>
                                </template>
                                <tr v-else>
                                    <td colspan="7" class="text-center py-14 text-muted text-sm font-mono">
                                        // cari perusahaan di atas untuk melihat kepemilikan saham
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                </div>
            </div>
        </section>

        <!-- ================================================
         GRAPH SECTION
    ================================================ -->
        <section id="graph-section" class="w-full px-4 sm:px-8 pb-20 pt-2 lg:max-w-7xl lg:mx-auto">

            <!-- Header -->
            <div class="flex items-start justify-between flex-wrap gap-4 mb-5">
                <div>
                    <div class="font-serif text-ink"
                        :style="{ fontSize: '42px', letterSpacing: '-2px', lineHeight: 1 }">{{
                            currentTicker }}</div>
                    <div class="text-sm text-muted mt-0.5">{{ currentName }}</div>
                    <div class="text-xs text-muted mt-1 font-mono">
                        {{ tableData.length }} pemegang saham ≥1% // Data KSEI
                    </div>
                </div>
                <!-- <div class="flex gap-3 flex-wrap items-center">
                    <button class="neo-btn-sm" @click="exportCSV">↓ EKSPOR CSV</button>
                    <button class="neo-btn-sm" @click="resetView">↺ RESET TAMPILAN</button>
                </div> -->
            </div>

            <!-- Graph + Legend -->
            <div class="flex flex-col lg:flex-row gap-4 items-start">

                <!-- LEGEND PANEL -->
                <div class="flex-shrink-0 w-full lg:w-44 bg-card rounded-2xl p-4 border-2 border-ink"
                    :style="{ boxShadow: '3px 3px 0 #1a1612' }">
                    <div class="font-mono text-xs uppercase tracking-widest text-muted mb-3">Tipe Investor</div>
                    <div class="flex flex-col gap-0.5">
                        <div v-for="t in TYPES" :key="t.key"
                            class="leg-item flex items-center gap-2 rounded-lg px-2 py-1.5 select-none hover:bg-cream2 transition-colors cursor-pointer"
                            :class="{ off: hidden.has(t.key) }" @click="toggleVisibility(t.key)">
                            <span class="leg-dot w-2.5 h-2.5 rounded-full border-2 flex-shrink-0"
                                :style="{ background: t.fill, borderColor: t.stroke }"></span>
                            <span class="leg-name text-xs text-ink2 flex-1 leading-tight">{{ t.label }}</span>
                            <span class="eye-on opacity-40 text-muted text-xs">👁</span>
                            <span class="eye-off opacity-40 text-muted text-xs hidden">🔒</span>
                        </div>
                    </div>

                    <div class="mt-4 pt-3 border-t border-bdr">
                        <div class="font-mono text-xs uppercase tracking-widest text-muted mb-2.5">Tipe Relasi</div>
                        <div class="flex items-center gap-2 text-xs text-ink2 mb-1.5">
                            <div class="w-6 flex-shrink-0" :style="{ borderTop: '2px solid #6ab0e8' }"></div>
                            Langsung
                        </div>
                        <div class="flex items-center gap-2 text-xs text-ink2">
                            <div class="w-6 flex-shrink-0" :style="{ borderTop: '2px dashed #8b70d8' }"></div>
                            Kepemilikan Silang
                        </div>
                    </div>
                </div>

                <!-- CANVAS -->
                <div class="flex-1 min-w-0 w-full">
                    <div class="relative border-2 border-ink rounded-2xl overflow-hidden"
                        :style="{ boxShadow: '4px 4px 0 #1a1612' }">
                        <div ref="canvasRef" id="graph-canvas" style="height: 560px;">
                            <svg ref="svgRef" id="net-svg" width="100%" height="100%"></svg>
                        </div>

                        <!-- Zoom controls -->
                        <div class="absolute top-3 right-3 flex flex-col gap-1 bg-card border-2 border-ink rounded-xl p-1"
                            :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                            <button
                                class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-base font-bold transition-colors"
                                @click="zoomIn">+</button>
                            <button
                                class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-base font-bold transition-colors"
                                @click="zoomOut">−</button>
                            <button
                                class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-xs font-bold transition-colors"
                                @click="resetView">1:1</button>
                        </div>

                        <!-- Empty state -->
                        <div v-if="!graphLoaded"
                            class="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                            <div class="text-sm text-muted">Silahkan pilih emiten untuk melihat jaringan</div>
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <!-- ================================================
         TOOLTIP
    ================================================ -->
        <Teleport to="body">
            <div v-if="tooltip.visible"
                class="fixed pointer-events-none z-50 bg-white border-2 border-ink rounded-xl p-4" :style="{
                    boxShadow: '4px 4px 0 #1a1612',
                    maxWidth: '240px',
                    left: `${tooltip.x}px`,
                    top: `${tooltip.y}px`,
                }">
                <div class="font-mono text-xs font-bold text-bluedk mb-1">{{ tooltip.tick }}</div>
                <div class="font-sans text-sm font-semibold text-ink mb-2 leading-snug">{{ tooltip.name }}</div>
                <!-- eslint-disable-next-line vue/no-v-html -->
                <div class="font-sans text-xs text-muted leading-relaxed" v-html="tooltip.rows" />
            </div>
        </Teleport>

    </main>
</template>

<script setup lang="ts">
definePageMeta({
    layout: 'main',
})

import { ref, reactive, onMounted, onUnmounted, nextTick, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as d3 from 'd3'
import { TYPES, TM } from '~/composables/useDummyData'
import { useCompanySearch, useCompanyShareholdings } from '~/composables/useCompanySearch'
import type { Shareholding } from '~/types/shareholding.type'

// ── Composables ─────────────────────────────────────────────────────────────
const route = useRoute()
const router = useRouter()
const { suggestions, searchCompanies } = useCompanySearch()
const { shareholdings, loading: shareholdingLoading, error: shareholdingError, loadCompanyShareholdings } = useCompanyShareholdings()

// ── Reactive state ──────────────────────────────────────────────────────────
const lastUpdated = ref('31 Maret 2026')
const currentTicker = ref('BBCA')
const currentName = ref('BANK CENTRAL ASIA Tbk')
const searchInput = ref('BBCA')
const showDropdown = ref(false)
const hidden = ref<Set<string>>(new Set())
const tableData = computed(() => shareholdings.value)
const stats = ref<any[]>([])
const graphLoaded = ref(false)

// ── Debounce for search ──────────────────────────────────────────────────────
let searchTimeout: NodeJS.Timeout | null = null
const debounceSearch = (callback: () => void, delay: number = 300) => {
    if (searchTimeout) clearTimeout(searchTimeout)
    searchTimeout = setTimeout(callback, delay)
}

const tooltip = reactive({
    visible: false,
    x: 0,
    y: 0,
    tick: '',
    name: '',
    rows: '',
})

// ── Auto-load graph when ticker changes (data already fetched via selectCompany) ──
watch(currentTicker, (newTicker) => {
    if (newTicker && newTicker.trim() && shareholdings.value.length > 0) {
        loadGraph(newTicker, currentName.value)
    }
})

// ── Template refs ────────────────────────────────────────────────────────────
const svgRef = ref<SVGSVGElement | null>(null)
const canvasRef = ref<HTMLDivElement | null>(null)
const dropdownRef = ref<HTMLDivElement | null>(null)

// Safe type metadata accessor to avoid "possibly undefined"
const FALLBACK_TYPE_META = { fill: '#e6e0d8', stroke: '#8b8175', label: 'Other' }
function getTypeMeta(type?: string) {
    return (type && (TM as Record<string, any>)[type]) || TM.OTHER || FALLBACK_TYPE_META
}

const INVESTOR_TYPE_MAP: Record<string, string> = {
    CP: 'CORPORATE',
    IB: 'INVESTMENT_BANK',
    SC: 'SECURITIES',
    ID: 'INDIVIDUAL',
}

function normalizeInvestorType(type?: string) {
    if (!type) return 'OTHER'
    const mapped = INVESTOR_TYPE_MAP[type] || type
    return (TM as Record<string, any>)[mapped] ? mapped : 'OTHER'
}

function buildGraphDataFromShareholdings(records: Shareholding[], ticker: string) {
    return {
        ticker,
        name: records[0]?.company_name || ticker,
        edges: records.map(r => ({
            target: r.investor_name,
            pct: r.percentage,
            type: normalizeInvestorType(r.investor_type),
            origin: r.local_foreign === 'F' ? 'FOREIGN' : 'LOCAL',
            shares: r.total_holding_shares,
        })),
        cross: [] as any[],
    }
}

// D3 zoom instance stored for zoom controls
let zoomBehavior: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null
let svgSelection: d3.Selection<SVGSVGElement, unknown, null, undefined> | null = null
let nodeSelection: any = null
let linkSelection: any = null
let labelSelection: any = null

function applyLegendVisibility() {
    if (!nodeSelection || !linkSelection || !labelSelection) return
    const hiddenSet = hidden.value
    nodeSelection
        .attr('opacity', (d: any) => hiddenSet.has(d.type) ? 0.08 : 1)
        .attr('pointer-events', (d: any) => hiddenSet.has(d.type) ? 'none' : 'auto')
    linkSelection
        .attr('opacity', (d: any) => hiddenSet.has(d.type) ? 0.08 : 1)
    labelSelection
        .attr('opacity', (d: any) => hiddenSet.has(d.type) ? 0 : 1)
}

// ── Search ───────────────────────────────────────────────────────────────────
function handleSearchInput(event: Event) {
    const input = event.target as HTMLInputElement
    searchInput.value = input.value.toUpperCase()
    onSearchInput()
}

function onSearchInput() {
    const q = searchInput.value.trim()
    if (!q) {
        showDropdown.value = false
        return
    }
    debounceSearch(() => {
        searchCompanies(q)
        showDropdown.value = true
    }, 300)
}

async function selectCompany(company: any) {
    await loadCompanyShareholdings(company)

    // Navigate with ticker parameter
    router.push({ path: '/', query: { ticker: company.kode } })

    // Now update ticker to trigger graph render via watch
    currentTicker.value = company.kode
    currentName.value = company.nama_perusahaan
    searchInput.value = company.kode
    showDropdown.value = false
}

function doSearch() {
    if (suggestions.value.length > 0) {
        selectCompany(suggestions.value[0])
    }
}

function quickLoad(t: string) {
    // Find company from suggestions or use dummy data for now
    const found = suggestions.value.find(s => s.kode === t)
    if (found) {
        selectCompany(found)
    } else {
        // Fallback to dummy data
        selectCompany({ kode: t, nama_perusahaan: t })
    }
}

// ── Load data + draw graph ───────────────────────────────────────────────────
function loadGraph(t: string, n: string) {
    if (shareholdings.value.length === 0) {
        stats.value = []
        graphLoaded.value = false
        if (svgRef.value) {
            d3.select(svgRef.value).selectAll('*').remove()
        }
        return
    }

    const graphData = buildGraphDataFromShareholdings(shareholdings.value, t)
    currentTicker.value = t
    currentName.value = n
    hidden.value = new Set()
    graphLoaded.value = false

    // Build stats from shareholdings or dummy data
    const foreign = shareholdings.value.filter(s => s.local_foreign === 'F')
    const total = shareholdings.value.reduce((sum, s) => sum + s.percentage, 0)
    stats.value = [
        { label: 'Asing', val: foreign.length, sub: `dari ${shareholdings.value.length} total` },
        { label: 'Pemegang Terbesar', val: shareholdings.value[0]?.percentage.toFixed(2) + '%', sub: shareholdings.value[0]?.investor_name + '…' },
        { label: 'Total Tercatat', val: total.toFixed(1) + '%', sub: 'dari float tercatat' },
        { label: 'Jumlah Data', val: shareholdings.value.length, sub: 'data kepemilikan' },
    ]

    // Draw after DOM update
    nextTick(() => {
        drawGraph(t, graphData)
        graphLoaded.value = true
    })
}

// ── D3 graph ─────────────────────────────────────────────────────────────────
function drawGraph(t: string, data: any) {
    if (!svgRef.value || !canvasRef.value) return

    d3.select(svgRef.value).selectAll('*').remove()

    const W = canvasRef.value.clientWidth
    const H = canvasRef.value.clientHeight
    const svg = d3.select(svgRef.value)

    // Setup zoom
    zoomBehavior = d3.zoom<SVGSVGElement, unknown>()
        .scaleExtent([0.15, 5])
        .on('zoom', (event) => {
            gMain.attr('transform', event.transform)
        })
    svg.call(zoomBehavior)
    svgSelection = svg

    const defs = svg.append('defs')

        // Arrow markers
        ;[
            { id: 'arr-d', c: '#6ab0e8' },
            { id: 'arr-c', c: '#8b70d8' },
        ].forEach(m => {
            defs.append('marker')
                .attr('id', m.id)
                .attr('viewBox', '0 0 10 10')
                .attr('refX', 9).attr('refY', 5)
                .attr('markerWidth', 5).attr('markerHeight', 5)
                .attr('orient', 'auto')
                .append('path')
                .attr('d', 'M1 1L9 5L1 9')
                .attr('fill', 'none')
                .attr('stroke', m.c)
                .attr('stroke-width', 1.6)
                .attr('stroke-linecap', 'round')
        })

    const gMain = svg.append('g')
    const lG = gMain.append('g')
    const labG = gMain.append('g')
    const nG = gMain.append('g')

    // Build node map
    const nm: Record<string, any> = {}
    function addNode(id: string, type: string) {
        if (!nm[id]) nm[id] = { id, label: id, type }
    }
    addNode(t, 'stock')
    data.edges.forEach((e: any) => addNode(e.target, e.type))
    data.cross.forEach((c: any) => {
        addNode(c.investor, c.type)
        addNode(c.target, 'stock')
    })

    const nodes = Object.values(nm)

    const links = [
        ...data.edges.map((e: any) => ({
            source: t, target: e.target, pct: e.pct, shares: e.shares, cross: false, type: e.type,
        })),
        ...data.cross.map((c: any) => ({
            source: c.investor, target: c.target, pct: c.pct, cross: true, type: c.type,
        })),
    ]

    function nodeRadius(n: any) {
        if (n.id === t) return 40
        if (n.type === 'stock') return 22
        const l = data.edges.find((e: any) => e.target === n.id)
        return l ? Math.max(15, Math.min(36, 12 + l.pct * 0.95)) : 16
    }

    // Links
    const linkS = lG.selectAll('line').data(links).join('line')
    linkS
        .attr('stroke', (d: any) => d.cross ? '#8b70d8' : '#6ab0e8')
        .attr('stroke-opacity', (d: any) => d.cross ? 0.28 : 0.33)
        .attr('stroke-width', (d: any) => Math.max(1.2, d.pct * 0.2))
        .attr('stroke-dasharray', (d: any) => d.cross ? '6 3' : null)
        .attr('marker-end', (d: any) => `url(#${d.cross ? 'arr-c' : 'arr-d'})`)

    // Edge labels
    const labelS = labG.selectAll('text').data(links.filter((d: any) => d.pct >= 2)).join('text')
    labelS
        .attr('font-size', 9)
        .attr('fill', '#b0a898')
        .attr('text-anchor', 'middle')
        .attr('font-family', 'Share Tech Mono, monospace')
        .attr('pointer-events', 'none')

    // Node groups
    const nodeS = nG.selectAll<SVGGElement, any>('g').data(nodes).join('g')
        .attr('cursor', 'pointer')
        .call(
            d3.drag<SVGGElement, any>()
                .on('start', (ev: any, d: any) => {
                    if (!ev.active) sim.alphaTarget(0.3).restart()
                    d.fx = d.x; d.fy = d.y
                })
                .on('drag', (ev: any, d: any) => { d.fx = ev.x; d.fy = ev.y })
                .on('end', (ev: any, d: any) => {
                    if (!ev.active) sim.alphaTarget(0)
                    d.fx = null; d.fy = null
                }),
        )

    nodeSelection = nodeS
    linkSelection = linkS
    labelSelection = labelS
    applyLegendVisibility()

    // Glow ring for centre node
    nodeS.filter((d: any) => d.id === t).append('circle')
        .attr('r', 58)
        .attr('fill', 'none')
        .attr('stroke', '#6ab0e8')
        .attr('stroke-width', 1.5)
        .attr('stroke-opacity', 0.16)
        .attr('stroke-dasharray', '5 6')
        .attr('pointer-events', 'none')

    // Node circles
    nodeS.append('circle')
        .attr('r', (d: any) => nodeRadius(d))
        .attr('fill', (d: any) => getTypeMeta(d.type).fill)
        .attr('stroke', (d: any) => getTypeMeta(d.type).stroke)
        .attr('stroke-width', (d: any) => d.id === t ? 2.5 : 1.8)

    // Node labels
    nodeS.append('text')
        .attr('text-anchor', 'middle')
        .attr('dominant-baseline', 'central')
        .attr('font-family', 'Share Tech Mono, monospace')
        .attr('font-size', (d: any) => d.id === t ? 12 : d.type === 'stock' ? 10 : 8.5)
        .attr('font-weight', '700')
        .attr('fill', (d: any) => getTypeMeta(d.type).stroke)
        .attr('pointer-events', 'none')
        .text((d: any) => {
            const r = nodeRadius(d) * 1.85
            return d.label.length * 6 < r ? d.label : d.label.split(' ')[0].slice(0, 8)
        })

    // Tooltip interactions
    nodeS
        .on('mouseover', (ev: any, d: any) => {
            const e = data.edges.find((x: any) => x.target === d.id)
            let rows = ''
            if (e) {
                rows += `<div>Kepemilikan: <b style="color:#1a1612">${e.pct}%</b></div>`
                if (e.shares) rows += `<div>Saham: ${e.shares.toLocaleString()}</div>`
                rows += `<div>Asal: ${e.origin}</div>`
            }
            rows += `<div>Tipe: ${getTypeMeta(d.type).label}</div>`

            tooltip.visible = true
            tooltip.x = ev.clientX + 16
            tooltip.y = ev.clientY - 10
            tooltip.tick = d.id === t ? `★ ${t}` : getTypeMeta(d.type).label.toUpperCase()
            tooltip.name = d.label
            tooltip.rows = rows

            d3.select(ev.currentTarget).select('circle:not([r="58"])').attr('stroke-width', 3.5)
        })
        .on('mousemove', (ev: any) => {
            let tx = ev.clientX + 16, ty = ev.clientY - 10
            if (tx + 250 > window.innerWidth) tx = ev.clientX - 260
            if (ty + 170 > window.innerHeight) ty = ev.clientY - 180
            tooltip.x = tx
            tooltip.y = ty
        })
        .on('mouseout', (ev: any, d: any) => {
            tooltip.visible = false
            d3.select(ev.currentTarget).select('circle:not([r="58"])').attr('stroke-width', d.id === t ? 2.5 : 1.8)
        })

    // Force simulation
    const sim = d3.forceSimulation(nodes)
        .force('link',
            d3.forceLink(links)
                .id((d: any) => d.id)
                .distance((d: any) => d.cross ? 200 : Math.max(110, 165 - d.pct * 1.2))
                .strength(0.42),
        )
        .force('charge', d3.forceManyBody().strength(-360))
        .force('center', d3.forceCenter(W / 2, H / 2))
        .force('collision', d3.forceCollide().radius((d: any) => nodeRadius(d) + 16))
        .on('tick', () => {
            function ex(d: any, dim: string, tgt: boolean) {
                const s = d.source, tt = d.target
                const dx = tt.x - s.x, dy = tt.y - s.y
                const dist = Math.sqrt(dx * dx + dy * dy) || 1
                if (!tgt) return dim === 'x' ? s.x + (dx / dist) * nodeRadius(s) : s.y + (dy / dist) * nodeRadius(s)
                else return dim === 'x' ? tt.x - (dx / dist) * (nodeRadius(tt) + 9) : tt.y - (dy / dist) * (nodeRadius(tt) + 9)
            }
            linkS
                .attr('x1', (d: any) => ex(d, 'x', false))
                .attr('y1', (d: any) => ex(d, 'y', false))
                .attr('x2', (d: any) => ex(d, 'x', true))
                .attr('y2', (d: any) => ex(d, 'y', true))
            nodeS.attr('transform', (d: any) => `translate(${d.x || 0},${d.y || 0})`)
            labelS
                .attr('x', (d: any) => ((d.source as any).x + (d.target as any).x) / 2)
                .attr('y', (d: any) => ((d.source as any).y + (d.target as any).y) / 2 - 6)
                .text((d: any) => d.pct >= 2 ? d.pct.toFixed(1) + '%' : '')
        })
}

// ── Legend toggle ─────────────────────────────────────────────────────────────
function toggleVisibility(type: string) {
    const s = new Set(hidden.value)
    s.has(type) ? s.delete(type) : s.add(type)
    hidden.value = s
    applyLegendVisibility()
}

// ── Zoom controls ─────────────────────────────────────────────────────────────
function zoomIn() {
    if (svgSelection && zoomBehavior) {
        svgSelection.transition().duration(250).call(zoomBehavior.scaleBy, 1.3)
    }
}
function zoomOut() {
    if (svgSelection && zoomBehavior) {
        svgSelection.transition().duration(250).call(zoomBehavior.scaleBy, 1 / 1.3)
    }
}
function resetView() {
    if (svgSelection && zoomBehavior) {
        svgSelection.transition().duration(350).call(zoomBehavior.transform, d3.zoomIdentity)
    }
}

// ── Export CSV ────────────────────────────────────────────────────────────────
function exportCSV() {
    const rows = [['Kode', 'Nama Perusahaan', 'Investor', 'Tipe', 'Asal', 'Persentase', 'Saham', 'Tanggal']]
    tableData.value.forEach(e =>
        rows.push([
            currentTicker.value,
            e.company_name,
            e.investor_name,
            e.investor_type,
            e.local_foreign,
            e.percentage.toString(),
            e.total_holding_shares.toString(),
            e.date || ''
        ]),
    )
    const csv = rows.map(r => r.map(c => `"${c}"`).join(',')).join('\n')
    const a = document.createElement('a')
    a.href = 'data:text/csv;charset=utf-8,' + encodeURIComponent(csv)
    a.download = `shareholding_${currentTicker.value}_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
}

// ── Close dropdown on outside click ──────────────────────────────────────────
function handleOutsideClick(e: MouseEvent) {
    const target = e.target as HTMLElement
    if (!target.closest('#hero-search') && !target.closest('[data-dropdown]')) {
        showDropdown.value = false
    }
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────
onMounted(() => {
    document.addEventListener('click', handleOutsideClick)
    // Check if ticker passed from high-concentration-list page
    const tickerFromQuery = route.query.ticker as string | undefined
    if (tickerFromQuery) {
        // Load ticker from query param
        currentTicker.value = tickerFromQuery
        searchInput.value = tickerFromQuery
        loadCompanyShareholdings({ kode: tickerFromQuery, nama_perusahaan: tickerFromQuery, id: '', created_at: '', updated_at: '' } as any)
            .then(() => {
                loadGraph(tickerFromQuery, tickerFromQuery)
            })
    } else {
        // Auto-load default ticker (BBCA) on mount
        loadCompanyShareholdings({ kode: 'BBCA', nama_perusahaan: 'BANK CENTRAL ASIA Tbk', id: '', created_at: '', updated_at: '' } as any)
            .then(() => {
                loadGraph('BBCA', 'BANK CENTRAL ASIA Tbk')
            })
    }
})

onUnmounted(() => {
    document.removeEventListener('click', handleOutsideClick)
    if (searchTimeout) clearTimeout(searchTimeout)
})
</script>