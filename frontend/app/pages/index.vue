<template>
    <main class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden">

        <!-- ================================================
         HERO SECTION
    ================================================ -->
        <section class="min-h-screen flex flex-col items-center justify-center text-center px-5 pt-20 pb-16"
            :style="{ background: 'linear-gradient(180deg,#cce8f5 0%,#f2ede4 50%)' }">
            <!-- Badge -->
            <div class="fu flex items-center gap-2 mb-5 px-4 py-1.5 rounded-full text-xs font-mono uppercase tracking-widest text-bluedk"
                :style="{ background: 'rgba(106,176,232,.12)', border: '1.5px solid rgba(106,176,232,.32)' }">
                <span class="pulse-dot w-2 h-2 rounded-full bg-blue inline-block"></span>
                KSEI Real-time Data
            </div>

            <!-- Headline -->
            <h1 class="fu1 font-serif text-ink leading-none mb-6"
                :style="{ fontSize: 'clamp(40px,7.5vw,92px)', letterSpacing: '-3px', maxWidth: '820px', lineHeight: 0.97 }">
                Visualizing the IDX
                <br />
                <em>Shareholder Network.</em>
            </h1>

            <!-- Subtext -->
            <p class="fu2 font-light text-ink2 mb-10"
                :style="{ fontSize: '17px', maxWidth: '530px', lineHeight: 1.85 }">
                Explore complex <span class="hl hl-b">cross-ownership</span> and
                <span class="hl hl-o">institutional holdings</span> across
                <span class="hl hl-g">IDX-listed companies</span> in a single
                <span class="hl hl-p">interactive view</span>.
            </p>

            <!-- SEARCH BOX -->
            <div class="fu3 w-full max-w-xl relative">
                <div class="flex gap-3 items-stretch">
                    <div class="relative flex-1">
                        <input id="hero-search" v-model="searchInput" class="search-input" type="text"
                            placeholder="Search ticker... e.g. BRMS, BBCA" @keydown.enter="doSearch"
                            @input="onSearchInput" @focus="onSearchInput" />
                        <svg class="absolute right-4 top-1/2 -translate-y-1/2 text-muted w-4 h-4 pointer-events-none"
                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8" />
                            <path d="m21 21-4.35-4.35" />
                        </svg>

                        <!-- Dropdown suggestions -->
                        <div v-if="showDropdown && suggestions.length" ref="dropdownRef"
                            class="absolute top-full mt-2 left-0 right-0 bg-white border-2 border-ink rounded-xl overflow-hidden z-50">
                            <div v-for="s in suggestions" :key="s.ticker"
                                class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bluebg border-b border-bdr last:border-0 transition-colors"
                                @mousedown.prevent="selectEmiten(s.ticker, s.name)">
                                <span
                                    class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md min-w-[46px] text-center">
                                    {{ s.ticker }}
                                </span>
                                <span class="text-sm text-ink2">{{ s.name }}</span>
                            </div>
                        </div>
                    </div>

                    <button class="neo-btn" @click="doSearch">EXPLORE →</button>
                </div>

                <!-- Quick load chips -->
                <div class="flex gap-2 mt-4 flex-wrap justify-center items-center">
                    <span class="text-xs text-muted font-mono">TRY:</span>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="quickLoad('BRMS')">BRMS</button>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="quickLoad('GLAS')">GLAS NETWORK</button>
                    <button
                        class="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
                        @click="quickLoad('PETROMINE')">PETROMINE</button>
                </div>
            </div>
        </section>

        <!-- ================================================
         GRAPH SECTION
    ================================================ -->
        <section id="graph-section" class="max-w-7xl mx-auto px-4 sm:px-8 pb-20 pt-2">

            <!-- Header -->
            <div class="flex items-start justify-between flex-wrap gap-4 mb-5">
                <div>
                    <div class="font-serif text-ink"
                        :style="{ fontSize: '42px', letterSpacing: '-2px', lineHeight: 1 }">{{ currentTicker }}</div>
                    <div class="text-sm text-muted mt-0.5">{{ currentName }}</div>
                    <div class="text-xs text-muted mt-1 font-mono">
                        {{ tableData.length }} shareholders ≥1% // KSEI Data
                    </div>
                </div>
                <div class="flex gap-3 flex-wrap items-center">
                    <button class="neo-btn-sm" @click="exportCSV">↓ EXPORT CSV</button>
                    <button class="neo-btn-sm" @click="resetView">↺ RESET VIEW</button>
                </div>
            </div>

            <!-- Graph + Legend -->
            <div class="flex gap-4 items-start">

                <!-- LEGEND PANEL -->
                <div class="flex-shrink-0 w-44 bg-card rounded-2xl p-4 border-2 border-ink"
                    :style="{ boxShadow: '3px 3px 0 #1a1612' }">
                    <div class="font-mono text-xs uppercase tracking-widest text-muted mb-3">Investor Type</div>
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
                        <div class="font-mono text-xs uppercase tracking-widest text-muted mb-2.5">Edge Type</div>
                        <div class="flex items-center gap-2 text-xs text-ink2 mb-1.5">
                            <div class="w-6 flex-shrink-0" :style="{ borderTop: '2px solid #6ab0e8' }"></div>
                            Direct
                        </div>
                        <div class="flex items-center gap-2 text-xs text-ink2">
                            <div class="w-6 flex-shrink-0" :style="{ borderTop: '2px dashed #8b70d8' }"></div>
                            Cross-holding
                        </div>
                    </div>
                </div>

                <!-- CANVAS -->
                <div class="flex-1 min-w-0">
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
                            <div class="text-5xl mb-4">📊</div>
                            <div class="font-serif text-2xl text-ink mb-2">Search an emiten above</div>
                            <div class="text-sm text-muted">or click a quick-load chip to explore the network</div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- STATS CARDS -->
            <div class="grid grid-cols-2 sm:grid-cols-5 gap-3 mt-4">
                <div v-for="(s, i) in stats" :key="i" class="bg-card border-2 border-ink rounded-xl p-4"
                    :style="{ boxShadow: '2px 2px 0 #1a1612' }">
                    <div class="font-mono text-xs uppercase tracking-wider text-muted mb-1">{{ s.label }}</div>
                    <div class="font-serif text-2xl text-ink leading-none mb-1">{{ s.val }}</div>
                    <div class="text-xs text-muted">{{ s.sub }}</div>
                </div>
            </div>
        </section>

        <!-- ================================================
         TABLE SECTION
    ================================================ -->
        <section id="table-section" class="max-w-7xl mx-auto px-4 sm:px-8 pb-24">
            <div class="mb-6">
                <h2 class="font-serif text-3xl text-ink mb-1" :style="{ letterSpacing: '-1px' }">
                    Shareholder Data Table
                </h2>
                <p class="text-sm text-muted">
                    Structured view of all ownership records corresponding to the graph above.
                </p>
            </div>

            <div class="border-2 border-ink rounded-2xl overflow-hidden bg-white"
                :style="{ boxShadow: '4px 4px 0 #1a1612' }">
                <div class="overflow-x-auto">
                    <table class="data-table w-full text-sm border-collapse">
                        <thead>
                            <tr class="border-b-2 border-ink bg-card">
                                <th
                                    class="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                                    Ticker</th>
                                <th
                                    class="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                                    Investor Name</th>
                                <th
                                    class="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden sm:table-cell">
                                    Type</th>
                                <th
                                    class="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden sm:table-cell">
                                    Origin</th>
                                <th
                                    class="text-right px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                                    %</th>
                                <th
                                    class="text-right px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden md:table-cell">
                                    Shares</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="tableData.length > 0">
                                <tr v-for="(row, i) in tableData" :key="i"
                                    class="border-b border-bdr transition-colors hover:bg-card">
                                    <td class="px-5 py-3">
                                        <span
                                            class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md">
                                            {{ currentTicker }}
                                        </span>
                                    </td>
                                    <td class="px-5 py-3 text-sm text-ink max-w-xs">{{ row.target }}</td>
                                    <td class="px-5 py-3 hidden sm:table-cell">
                                        <span
                                            class="text-xs font-medium px-2 py-0.5 rounded-full bg-purple-100 text-purple-700">
                                            {{ row.type }}
                                        </span>
                                    </td>
                                    <td class="px-5 py-3 hidden sm:table-cell">
                                        <span class="font-mono text-xs"
                                            :class="row.origin === 'FOREIGN' ? 'text-violet-700' : 'text-green-700'">{{
                                                row.origin }}</span>
                                    </td>
                                    <td class="px-5 py-3 text-right font-mono font-bold text-sm">{{ row.pct }}%</td>
                                    <td class="px-5 py-3 text-right font-mono text-xs text-muted hidden md:table-cell">
                                        {{ row.shares ? row.shares.toLocaleString() : '—' }}
                                    </td>
                                </tr>
                            </template>
                            <tr v-else>
                                <td colspan="6" class="text-center py-14 text-muted text-sm font-mono">
                                    // search an emiten to populate
                                </td>
                            </tr>
                        </tbody>
                    </table>
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

import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import * as d3 from 'd3'
import { DB, EMITEN, TYPES, TM } from '~/composables/useDummyData'

// ── Reactive state ──────────────────────────────────────────────────────────
const currentTicker = ref('BRMS')
const currentName = ref('BUMI RESOURCES MINERALS Tbk')
const searchInput = ref('')
const showDropdown = ref(false)
const suggestions = ref<typeof EMITEN>([])
const hidden = ref<Set<string>>(new Set())
const tableData = ref<any[]>([])
const stats = ref<any[]>([])
const graphLoaded = ref(false)

const tooltip = reactive({
    visible: false,
    x: 0,
    y: 0,
    tick: '',
    name: '',
    rows: '',
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

// D3 zoom instance stored for zoom controls
let zoomBehavior: d3.ZoomBehavior<SVGSVGElement, unknown> | null = null
let svgSelection: d3.Selection<SVGSVGElement, unknown, null, undefined> | null = null

// ── Search ───────────────────────────────────────────────────────────────────
function onSearchInput() {
    const q = searchInput.value.trim().toUpperCase()
    if (!q) { showDropdown.value = false; return }
    const ms = EMITEN.filter(e => e.ticker.includes(q) || e.name.toUpperCase().includes(q))
    suggestions.value = ms.slice(0, 7)
    showDropdown.value = ms.length > 0
}

function selectEmiten(t: string, n: string) {
    currentTicker.value = t
    currentName.value = n
    searchInput.value = t
    showDropdown.value = false
    loadGraph(t, n)
}

function doSearch() {
    const q = searchInput.value.trim().toUpperCase()
    const f = EMITEN.find(x => x.ticker === q)
    selectEmiten(f ? f.ticker : 'BRMS', f ? f.name : 'BUMI RESOURCES MINERALS Tbk')
}

function quickLoad(t: string) {
    const f = EMITEN.find(x => x.ticker === t)
    if (f) selectEmiten(f.ticker, f.name)
    else {
        // Fallback to BRMS for demo tickers not in EMITEN list
        selectEmiten('BRMS', 'BUMI RESOURCES MINERALS Tbk')
    }
}

// ── Load data + draw graph ───────────────────────────────────────────────────
function loadGraph(t: string, n: string) {
    const data = DB[t] || DB['BRMS']
    currentTicker.value = t
    currentName.value = n
    hidden.value = new Set()
    graphLoaded.value = false

    // Build stats
    const foreign = data.edges.filter((e: any) => e.origin === 'FOREIGN')
    const total = data.edges.reduce((s: number, e: any) => s + e.pct, 0)
    stats.value = [
        { label: 'Shareholders', val: data.edges.length, sub: 'investors ≥1%' },
        { label: 'Foreign', val: foreign.length, sub: `of ${data.edges.length} total` },
        { label: 'Largest Holder', val: data.edges[0].pct + '%', sub: data.edges[0].target.split(' ')[0] + '…' },
        { label: 'Total Recorded', val: total.toFixed(1) + '%', sub: 'of float recorded' },
        { label: 'Cross-Holdings', val: data.cross.length, sub: 'connections found' },
    ]

    // Build table
    tableData.value = data.edges

    // Draw after DOM update
    nextTick(() => {
        drawGraph(t, data)
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
            source: t, target: e.target, pct: e.pct, shares: e.shares, cross: false,
        })),
        ...data.cross.map((c: any) => ({
            source: c.investor, target: c.target, pct: c.pct, cross: true,
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
        .attr('stroke', (d: any) => d.cross ? '#8b70d8' : '#6ab0e8')
        .attr('stroke-opacity', (d: any) => d.cross ? 0.28 : 0.33)
        .attr('stroke-width', (d: any) => Math.max(1.2, d.pct * 0.2))
        .attr('stroke-dasharray', (d: any) => d.cross ? '6 3' : null)
        .attr('marker-end', (d: any) => `url(#${d.cross ? 'arr-c' : 'arr-d'})`)

    // Edge labels
    const labelS = labG.selectAll('text').data(links.filter((d: any) => d.pct >= 2)).join('text')
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
                rows += `<div>Ownership: <b style="color:#1a1612">${e.pct}%</b></div>`
                if (e.shares) rows += `<div>Shares: ${e.shares.toLocaleString()}</div>`
                rows += `<div>Origin: ${e.origin}</div>`
            }
            rows += `<div>Type: ${getTypeMeta(d.type).label}</div>`

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
    const rows = [['Ticker', 'Investor', 'Type', 'Origin', 'Percentage', 'Shares']]
    tableData.value.forEach(e =>
        rows.push([currentTicker.value, e.target, e.type, e.origin, e.pct, e.shares || '']),
    )
    const csv = rows.map(r => r.join(',')).join('\n')
    const a = document.createElement('a')
    a.href = 'data:text/csv;charset=utf-8,' + encodeURIComponent(csv)
    a.download = `ksei_${currentTicker.value}.csv`
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
    setTimeout(() => loadGraph('BRMS', 'BUMI RESOURCES MINERALS Tbk'), 200)
})

onUnmounted(() => {
    document.removeEventListener('click', handleOutsideClick)
})
</script>