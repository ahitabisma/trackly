"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import * as d3 from "d3";

// DUMMY DATA
const DB: Record<string, any> = {
  BRMS: {
    ticker: "BRMS",
    name: "BUMI RESOURCES MINERALS Tbk",
    edges: [
      {
        target: "EMIRATES TARIAN GLOBAL VENTURES SPC",
        pct: 25.1,
        type: "OTHER",
        origin: "FOREIGN",
        shares: 35592738434,
      },
      {
        target: "GLAS TRUST (SINGAPORE) LTD.",
        pct: 7.61,
        type: "INVESTMENT_BANK",
        origin: "FOREIGN",
        shares: 10784607811,
      },
      {
        target: "SUGIMAN HALIM",
        pct: 7.45,
        type: "INDIVIDUAL",
        origin: "LOCAL",
        shares: 10568888888,
      },
      {
        target: "CGS INTL SEKURITAS INDONESIA",
        pct: 5.25,
        type: "SECURITIES",
        origin: "LOCAL",
        shares: 7439770171,
      },
      {
        target: "PETROMINE ENERGY TRADING",
        pct: 4.28,
        type: "CORPORATE",
        origin: "LOCAL",
        shares: 6067891376,
      },
      {
        target: "PT TRIMEGAH SEKURITAS INDONESIA TBK",
        pct: 3.91,
        type: "SECURITIES",
        origin: "LOCAL",
        shares: 5538637200,
      },
      {
        target: "CGS INTL SECURITIES SINGAPORE PTE",
        pct: 3.77,
        type: "INVESTMENT_BANK",
        origin: "FOREIGN",
        shares: 5351412625,
      },
      {
        target: "PT MAYBANK SEKURITAS INDONESIA",
        pct: 3.24,
        type: "SECURITIES",
        origin: "LOCAL",
        shares: 4589589548,
      },
      {
        target: "BUMI RESOURCES TBK",
        pct: 3.08,
        type: "CORPORATE",
        origin: "LOCAL",
        shares: 4365383689,
      },
      {
        target: "VANECK GOLD MINERS ETF",
        pct: 2.87,
        type: "MUTUAL_FUND",
        origin: "FOREIGN",
        shares: 4075687300,
      },
      {
        target: "WEXLER CAPITAL PTE. LTD",
        pct: 2.19,
        type: "CORPORATE",
        origin: "FOREIGN",
        shares: 3100000032,
      },
    ],
    cross: [
      {
        investor: "GLAS TRUST (SINGAPORE) LTD.",
        type: "INVESTMENT_BANK",
        target: "TRIO",
        pct: 25.53,
      },
      {
        investor: "GLAS TRUST (SINGAPORE) LTD.",
        type: "INVESTMENT_BANK",
        target: "BUMI",
        pct: 2.08,
      },
      {
        investor: "PETROMINE ENERGY TRADING",
        type: "CORPORATE",
        target: "DEWA",
        pct: 4.46,
      },
      {
        investor: "PETROMINE ENERGY TRADING",
        type: "CORPORATE",
        target: "VKTR",
        pct: 2.83,
      },
      {
        investor: "PETROMINE ENERGY TRADING",
        type: "CORPORATE",
        target: "ELTY",
        pct: 2.61,
      },
    ],
  },
};

const EMITEN = [
  { ticker: "BRMS", name: "BUMI RESOURCES MINERALS Tbk" },
  { ticker: "BBCA", name: "BANK CENTRAL ASIA Tbk" },
  { ticker: "TLKM", name: "TELEKOMUNIKASI INDONESIA Tbk" },
  { ticker: "BMRI", name: "BANK MANDIRI Tbk" },
  { ticker: "ASII", name: "ASTRA INTERNATIONAL Tbk" },
  { ticker: "GOTO", name: "GOTO GOJEK TOKOPEDIA Tbk" },
  { ticker: "BUMI", name: "BUMI RESOURCES Tbk" },
  { ticker: "DEWA", name: "DARMA HENWA Tbk" },
  { ticker: "TRIO", name: "TRIKOMSEL OKE Tbk" },
];

const TYPES = [
  {
    key: "stock",
    label: "Stock (Emiten)",
    fill: "#daedf9",
    stroke: "#2d7ab5",
  },
  { key: "CORPORATE", label: "Corporate", fill: "#e8e2f8", stroke: "#6040c0" },
  {
    key: "INVESTMENT_BANK",
    label: "Investment Bank",
    fill: "#ede0f8",
    stroke: "#8b50d8",
  },
  {
    key: "INDIVIDUAL",
    label: "Individual",
    fill: "#fce8cc",
    stroke: "#c07830",
  },
  {
    key: "MUTUAL_FUND",
    label: "Mutual Fund",
    fill: "#d4f0dc",
    stroke: "#3a8a58",
  },
  {
    key: "SECURITIES",
    label: "Securities",
    fill: "#fce4e4",
    stroke: "#b04040",
  },
  { key: "OTHER", label: "Other", fill: "#f0ece8", stroke: "#807870" },
];

const TM: Record<string, (typeof TYPES)[0]> = {};
TYPES.forEach((t) => {
  TM[t.key] = t;
});

export default function Home() {
  const [ticker, setTicker] = useState("BRMS");
  const [name, setName] = useState("BUMI RESOURCES MINERALS Tbk");
  const [searchInput, setSearchInput] = useState("");
  const [showDropdown, setShowDropdown] = useState(false);
  const [suggestions, setSuggestions] = useState<typeof EMITEN>([]);
  const [hidden, setHidden] = useState<Set<string>>(new Set());
  const [tableData, setTableData] = useState<any[]>([]);
  const [stats, setStats] = useState<any[]>([]);
  const [graphLoaded, setGraphLoaded] = useState(false);
  const [tooltip, setTooltip] = useState({
    visible: false,
    x: 0,
    y: 0,
    tick: "",
    name: "",
    rows: "",
  });
  const [zoom, setZoom] = useState({ scale: 1, x: 0, y: 0 });
  const svgRef = useRef<SVGSVGElement>(null);
  const canvasRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Handle search input
  useEffect(() => {
    const q = searchInput.trim().toUpperCase();
    if (!q) {
      setShowDropdown(false);
      return;
    }
    const ms = EMITEN.filter(
      (e) => e.ticker.includes(q) || e.name.toUpperCase().includes(q),
    );
    setSuggestions(ms.slice(0, 7));
    setShowDropdown(ms.length > 0);
  }, [searchInput]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node)
      ) {
        const target = e.target as HTMLElement;
        if (!target.closest("#hero-search")) {
          setShowDropdown(false);
        }
      }
    };
    document.addEventListener("click", handleClick);
    return () => document.removeEventListener("click", handleClick);
  }, []);

  const selectEmiten = (t: string, n: string) => {
    setTicker(t);
    setName(n);
    setSearchInput(t);
    setShowDropdown(false);
    loadGraph(t, n);
  };

  const doSearch = () => {
    const q = searchInput.trim().toUpperCase();
    const f = EMITEN.find((x) => x.ticker === q);
    selectEmiten(
      f ? f.ticker : "BRMS",
      f ? f.name : "BUMI RESOURCES MINERALS Tbk",
    );
  };

  const quickLoad = (t: string) => {
    const f = EMITEN.find((x) => x.ticker === t);
    if (f) {
      selectEmiten(f.ticker, f.name);
    }
  };

  const loadGraph = (t: string, n: string) => {
    const data = DB[t] || DB["BRMS"];
    setTicker(t);
    setName(n);
    setHidden(new Set());
    setGraphLoaded(false);

    // Build stats
    const foreign = data.edges.filter((e: any) => e.origin === "FOREIGN");
    const total = data.edges.reduce((s: number, e: any) => s + e.pct, 0);
    const items = [
      {
        label: "Shareholders",
        val: data.edges.length,
        sub: "investors ≥1%",
      },
      {
        label: "Foreign",
        val: foreign.length,
        sub: `of ${data.edges.length} total`,
      },
      {
        label: "Largest Holder",
        val: data.edges[0].pct + "%",
        sub: data.edges[0].target.split(" ")[0] + "…",
      },
      {
        label: "Total Recorded",
        val: total.toFixed(1) + "%",
        sub: "of float recorded",
      },
      {
        label: "Cross-Holdings",
        val: data.cross.length,
        sub: "connections found",
      },
    ];
    setStats(items);

    // Build table data
    setTableData(data.edges);

    // Draw graph
    setTimeout(() => {
      drawGraph(t, data);
      setGraphLoaded(true);
    }, 0);
  };

  const drawGraph = (t: string, data: any) => {
    if (!svgRef.current || !canvasRef.current) return;

    d3.select(svgRef.current).selectAll("*").remove();

    const W = canvasRef.current.clientWidth;
    const H = canvasRef.current.clientHeight;

    const svg = d3.select(svgRef.current);

    const defs = svg.append("defs");

    // Markers
    [
      { id: "arr-d", c: "#6ab0e8" },
      { id: "arr-c", c: "#8b70d8" },
    ].forEach((m) => {
      defs
        .append("marker")
        .attr("id", m.id)
        .attr("viewBox", "0 0 10 10")
        .attr("refX", 9)
        .attr("refY", 5)
        .attr("markerWidth", 5)
        .attr("markerHeight", 5)
        .attr("orient", "auto")
        .append("path")
        .attr("d", "M1 1L9 5L1 9")
        .attr("fill", "none")
        .attr("stroke", m.c)
        .attr("stroke-width", 1.6)
        .attr("stroke-linecap", "round");
    });

    const gMain = svg.append("g");

    // Build nodes
    const nm: Record<string, any> = {};
    function addNode(id: string, label: string, type: string) {
      if (!nm[id]) nm[id] = { id, label, type };
    }
    addNode(t, t, "stock");
    data.edges.forEach((e: any) => addNode(e.target, e.target, e.type));
    data.cross.forEach((c: any) => {
      addNode(c.investor, c.investor, c.type);
      addNode(c.target, c.target, "stock");
    });

    const nodes = Object.values(nm);

    const links = [
      ...data.edges.map((e: any) => ({
        source: t,
        target: e.target,
        pct: e.pct,
        shares: e.shares,
        cross: false,
      })),
      ...data.cross.map((c: any) => ({
        source: c.investor,
        target: c.target,
        pct: c.pct,
        cross: true,
      })),
    ];

    function nodeRadius(n: any) {
      if (n.id === t) return 40;
      if (n.type === "stock") return 22;
      const l = data.edges.find((e: any) => e.target === n.id);
      return l ? Math.max(15, Math.min(36, 12 + l.pct * 0.95)) : 16;
    }

    const lG = gMain.append("g");
    const labG = gMain.append("g");
    const nG = gMain.append("g");

    const linkS = lG
      .selectAll("line")
      .data(links)
      .join("line")
      .attr("stroke", (d: any) => (d.cross ? "#8b70d8" : "#6ab0e8"))
      .attr("stroke-opacity", (d: any) => (d.cross ? 0.28 : 0.33))
      .attr("stroke-width", (d: any) => Math.max(1.2, d.pct * 0.2))
      .attr("stroke-dasharray", (d: any) => (d.cross ? "6 3" : null))
      .attr("marker-end", (d: any) => `url(#${d.cross ? "arr-c" : "arr-d"})`);

    const labelS = labG
      .selectAll("text")
      .data(links.filter((d: any) => d.pct >= 2))
      .join("text")
      .attr("font-size", 9)
      .attr("fill", "#b0a898")
      .attr("text-anchor", "middle")
      .attr("font-family", "Share Tech Mono, monospace")
      .attr("pointer-events", "none");

    const nodeS = nG
      .selectAll<SVGGElement, any>("g")
      .data(nodes)
      .join("g")
      .attr("cursor", "pointer")
      .call(
        d3
          .drag<SVGGElement, any>()
          .on("start", (ev: any, d: any) => {
            if (!ev.active) sim.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on("drag", (ev: any, d: any) => {
            d.fx = ev.x;
            d.fy = ev.y;
          })
          .on("end", (ev: any, d: any) => {
            if (!ev.active) sim.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }),
      );

    // Glow ring for center
    nodeS
      .filter((d: any) => d.id === t)
      .append("circle")
      .attr("r", 58)
      .attr("fill", "none")
      .attr("stroke", "#6ab0e8")
      .attr("stroke-width", 1.5)
      .attr("stroke-opacity", 0.16)
      .attr("stroke-dasharray", "5 6")
      .attr("pointer-events", "none");

    nodeS
      .append("circle")
      .attr("r", (d: any) => nodeRadius(d))
      .attr("fill", (d: any) => (TM[d.type] || TM.OTHER).fill)
      .attr("stroke", (d: any) => (TM[d.type] || TM.OTHER).stroke)
      .attr("stroke-width", (d: any) => (d.id === t ? 2.5 : 1.8));

    nodeS
      .append("text")
      .attr("text-anchor", "middle")
      .attr("dominant-baseline", "central")
      .attr("font-family", "Share Tech Mono, monospace")
      .attr("font-size", (d: any) =>
        d.id === t ? 12 : d.type === "stock" ? 10 : 8.5,
      )
      .attr("font-weight", "700")
      .attr("fill", (d: any) => (TM[d.type] || TM.OTHER).stroke)
      .attr("pointer-events", "none")
      .text((d: any) => {
        const r = nodeRadius(d) * 1.85;
        return d.label.length * 6 < r
          ? d.label
          : d.label.split(" ")[0].slice(0, 8);
      });

    // Tooltip interactions
    nodeS
      .on("mouseover", (ev: any, d: any) => {
        const e = data.edges.find((x: any) => x.target === d.id);
        let rows = "";
        if (e) {
          rows += `<div>Ownership: <b style="color:#1a1612">${e.pct}%</b></div>`;
          if (e.shares)
            rows += `<div>Shares: ${e.shares.toLocaleString()}</div>`;
          rows += `<div>Origin: ${e.origin}</div>`;
        }
        rows += `<div>Type: ${(TM[d.type] || TM.OTHER).label}</div>`;

        setTooltip({
          visible: true,
          x: ev.clientX + 16,
          y: ev.clientY - 10,
          tick:
            d.id === t
              ? `★ ${t}`
              : (TM[d.type] || TM.OTHER).label.toUpperCase(),
          name: d.label,
          rows,
        });

        d3.select(ev.target).attr("stroke-width", 3.5);
      })
      .on("mousemove", (ev: any) => {
        let tx = ev.clientX + 16,
          ty = ev.clientY - 10;
        if (tx + 250 > window.innerWidth) tx = ev.clientX - 260;
        if (ty + 170 > window.innerHeight) ty = ev.clientY - 180;
        setTooltip((prev) => ({ ...prev, x: tx, y: ty }));
      })
      .on("mouseout", (ev: any, d: any) => {
        setTooltip((prev) => ({ ...prev, visible: false }));
        d3.select(ev.target).attr("stroke-width", d.id === t ? 2.5 : 1.8);
      });

    const sim = d3
      .forceSimulation(nodes)
      .force(
        "link",
        d3
          .forceLink(links)
          .id((d: any) => d.id)
          .distance((d: any) =>
            d.cross ? 200 : Math.max(110, 165 - d.pct * 1.2),
          )
          .strength(0.42),
      )
      .force("charge", d3.forceManyBody().strength(-360))
      .force("center", d3.forceCenter(W / 2, H / 2))
      .force(
        "collision",
        d3.forceCollide().radius((d: any) => nodeRadius(d) + 16),
      )
      .on("tick", () => {
        function ex(d: any, dim: string, tgt: boolean) {
          const s = d.source,
            t = d.target;
          const dx = t.x - s.x,
            dy = t.y - s.y;
          const dist = Math.sqrt(dx * dx + dy * dy) || 1;
          if (!tgt)
            return dim === "x"
              ? s.x + (dx / dist) * nodeRadius(s)
              : s.y + (dy / dist) * nodeRadius(s);
          else
            return dim === "x"
              ? t.x - (dx / dist) * (nodeRadius(t) + 9)
              : t.y - (dy / dist) * (nodeRadius(t) + 9);
        }
        linkS
          .attr("x1", (d: any) => ex(d, "x", false))
          .attr("y1", (d: any) => ex(d, "y", false))
          .attr("x2", (d: any) => ex(d, "x", true))
          .attr("y2", (d: any) => ex(d, "y", true));
        nodeS.attr(
          "transform",
          (d: any) => `translate(${d.x || 0},${d.y || 0})`,
        );
        labelS
          .attr("x", (d: any) => (d.source.x + d.target.x) / 2)
          .attr("y", (d: any) => (d.source.y + d.target.y) / 2 - 6)
          .text((d: any) => (d.pct >= 2 ? d.pct.toFixed(1) + "%" : ""));
      });
  };

  const toggleVisibility = (type: string) => {
    const newHidden = new Set(hidden);
    if (newHidden.has(type)) {
      newHidden.delete(type);
    } else {
      newHidden.add(type);
    }
    setHidden(newHidden);
  };

  const exportCSV = () => {
    const rows = [
      ["Ticker", "Investor", "Type", "Origin", "Percentage", "Shares"],
    ];
    tableData.forEach((e) =>
      rows.push([ticker, e.target, e.type, e.origin, e.pct, e.shares || ""]),
    );
    const csv = rows.map((r) => r.join(",")).join("\n");
    const a = document.createElement("a");
    a.href = "data:text/csv;charset=utf-8," + encodeURIComponent(csv);
    a.download = `ksei_${ticker}.csv`;
    a.click();
  };

  const resetView = () => {
    setZoom({ scale: 1, x: 0, y: 0 });
  };

  // Initial load
  useEffect(() => {
    setTimeout(() => loadGraph("BRMS", "BUMI RESOURCES MINERALS Tbk"), 200);
  }, []);

  return (
    <main className="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden">
      {/* HERO */}
      <section
        className="min-h-screen flex flex-col items-center justify-center text-center px-5 pt-20 pb-16"
        style={{
          background: "linear-gradient(180deg,#cce8f5 0%,#f2ede4 50%)",
        }}
      >
        <div
          className="fu flex items-center gap-2 mb-5 px-4 py-1.5 rounded-full text-xs font-mono uppercase tracking-widest text-bluedk"
          style={{
            background: "rgba(106,176,232,.12)",
            border: "1.5px solid rgba(106,176,232,.32)",
          }}
        >
          <span className="pulse-dot w-2 h-2 rounded-full bg-blue inline-block"></span>
          KSEI Real-time Data
        </div>

        <h1
          className="fu1 font-serif text-ink leading-none mb-6"
          style={{
            fontSize: "clamp(40px,7.5vw,92px)",
            letterSpacing: "-3px",
            maxWidth: "820px",
            lineHeight: 0.97,
          }}
        >
          Visualizing the IDX
          <br />
          <em>Shareholder Network.</em>
        </h1>

        <p
          className="fu2 font-light text-ink2 mb-10"
          style={{ fontSize: "17px", maxWidth: "530px", lineHeight: 1.85 }}
        >
          Explore complex <span className="hl hl-b">cross-ownership</span> and{" "}
          <span className="hl hl-o">institutional holdings</span> across{" "}
          <span className="hl hl-g">IDX-listed companies</span> in a single{" "}
          <span className="hl hl-p">interactive view</span>.
        </p>

        {/* SEARCH */}
        <div className="fu3 w-full max-w-xl relative">
          <div className="flex gap-3 items-stretch">
            <div className="relative flex-1">
              <input
                id="hero-search"
                className="search-input"
                type="text"
                placeholder="Search ticker... e.g. BRMS, BBCA"
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && doSearch()}
              />
              <svg
                className="absolute right-4 top-1/2 -translate-y-1/2 text-muted w-4 h-4 pointer-events-none"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.35-4.35" />
              </svg>
              <div
                ref={dropdownRef}
                className={`${
                  showDropdown ? "" : "hidden"
                } absolute top-full mt-2 left-0 right-0 bg-white border-2 border-ink rounded-xl overflow-hidden z-50`}
              >
                {suggestions.map((s) => (
                  <div
                    key={s.ticker}
                    onClick={() => selectEmiten(s.ticker, s.name)}
                    className="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bluebg border-b border-bdr last:border-0 transition-colors"
                  >
                    <span className="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md min-w-[46px] text-center">
                      {s.ticker}
                    </span>
                    <span className="text-sm text-ink2">{s.name}</span>
                  </div>
                ))}
              </div>
            </div>
            <button className="neo-btn" onClick={doSearch}>
              EXPLORE →
            </button>
          </div>

          {/* QUICK LOAD */}
          <div className="flex gap-2 mt-4 flex-wrap justify-center items-center">
            <span className="text-xs text-muted font-mono">TRY:</span>
            <button
              onClick={() => quickLoad("BRMS")}
              className="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
            >
              BRMS
            </button>
            <button
              onClick={() => quickLoad("GLAS")}
              className="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
            >
              GLAS NETWORK
            </button>
            <button
              onClick={() => quickLoad("PETROMINE")}
              className="text-xs px-3 py-1 rounded-full border border-bdr2 text-ink2 hover:border-ink hover:text-ink transition-all font-mono"
            >
              PETROMINE
            </button>
          </div>
        </div>
      </section>

      {/* GRAPH SECTION */}
      <section
        id="graph-section"
        className="max-w-7xl mx-auto px-4 sm:px-8 pb-20 pt-2"
      >
        {/* Header */}
        <div className="flex items-start justify-between flex-wrap gap-4 mb-5">
          <div>
            <div
              id="g-ticker"
              className="font-serif text-ink"
              style={{
                fontSize: "42px",
                letterSpacing: "-2px",
                lineHeight: 1,
              }}
            >
              {ticker}
            </div>
            <div id="g-name" className="text-sm text-muted mt-0.5">
              {name}
            </div>
            <div id="g-meta" className="text-xs text-muted mt-1 font-mono">
              {tableData.length} shareholders ≥1% // KSEI Data
            </div>
          </div>
          <div className="flex gap-3 flex-wrap items-center">
            <button className="neo-btn-sm" onClick={exportCSV}>
              ↓ EXPORT CSV
            </button>
            <button className="neo-btn-sm" onClick={resetView}>
              ↺ RESET VIEW
            </button>
          </div>
        </div>

        {/* Graph + Legend */}
        <div className="flex gap-4 items-start">
          {/* LEGEND */}
          <div
            id="legend-panel"
            className="flex-shrink-0 w-44 bg-card rounded-2xl p-4 border-2 border-ink"
            style={{ boxShadow: "3px 3px 0 #1a1612" }}
          >
            <div className="font-mono text-xs uppercase tracking-widest text-muted mb-3">
              Investor Type
            </div>
            <div className="flex flex-col gap-0.5">
              {TYPES.map((t) => (
                <div
                  key={t.key}
                  className={`leg-item flex items-center gap-2 rounded-lg px-2 py-1.5 select-none hover:bg-cream2 transition-colors cursor-pointer ${
                    hidden.has(t.key) ? "off" : ""
                  }`}
                  onClick={() => toggleVisibility(t.key)}
                >
                  <span
                    className="leg-dot w-2.5 h-2.5 rounded-full border-2 flex-shrink-0"
                    style={{
                      background: t.fill,
                      borderColor: t.stroke,
                    }}
                  ></span>
                  <span className="leg-name text-xs text-ink2 flex-1 leading-tight">
                    {t.label}
                  </span>
                  <span className="eye-on opacity-40 text-muted text-xs">
                    👁
                  </span>
                  <span className="eye-off opacity-40 text-muted text-xs hidden">
                    🔒
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-4 pt-3 border-t border-bdr">
              <div className="font-mono text-xs uppercase tracking-widest text-muted mb-2.5">
                Edge Type
              </div>
              <div className="flex items-center gap-2 text-xs text-ink2 mb-1.5">
                <div
                  className="w-6 flex-shrink-0"
                  style={{ borderTop: "2px solid #6ab0e8" }}
                ></div>
                Direct
              </div>
              <div className="flex items-center gap-2 text-xs text-ink2">
                <div
                  className="w-6 flex-shrink-0"
                  style={{ borderTop: "2px dashed #8b70d8" }}
                ></div>
                Cross-holding
              </div>
            </div>
          </div>

          {/* CANVAS */}
          <div className="flex-1 min-w-0">
            <div
              className="relative border-2 border-ink rounded-2xl overflow-hidden"
              style={{ boxShadow: "4px 4px 0 #1a1612" }}
            >
              <div
                id="graph-canvas"
                ref={canvasRef}
                style={{ height: "560px" }}
              >
                <svg ref={svgRef} id="net-svg"></svg>
              </div>

              {/* Zoom controls */}
              <div
                className="absolute top-3 right-3 flex flex-col gap-1 bg-card border-2 border-ink rounded-xl p-1"
                style={{ boxShadow: "2px 2px 0 #1a1612" }}
              >
                <button
                  className="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-base font-bold transition-colors"
                  onClick={() =>
                    setZoom((z) => ({
                      ...z,
                      scale: Math.min(5, z.scale * 1.3),
                    }))
                  }
                >
                  +
                </button>
                <button
                  className="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-base font-bold transition-colors"
                  onClick={() =>
                    setZoom((z) => ({
                      ...z,
                      scale: Math.max(0.15, z.scale / 1.3),
                    }))
                  }
                >
                  −
                </button>
                <button
                  className="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-cream2 font-mono text-xs font-bold transition-colors"
                  onClick={resetView}
                >
                  1:1
                </button>
              </div>

              {/* Empty state */}
              {!graphLoaded && (
                <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                  <div className="text-5xl mb-4">📊</div>
                  <div className="font-serif text-2xl text-ink mb-2">
                    Search an emiten above
                  </div>
                  <div className="text-sm text-muted">
                    or click a quick-load chip to explore the network
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* STATS */}
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mt-4">
          {stats.map((s, i) => (
            <div
              key={i}
              className="bg-card border-2 border-ink rounded-xl p-4"
              style={{ boxShadow: "2px 2px 0 #1a1612" }}
            >
              <div className="font-mono text-xs uppercase tracking-wider text-muted mb-1">
                {s.label}
              </div>
              <div className="font-serif text-2xl text-ink leading-none mb-1">
                {s.val}
              </div>
              <div className="text-xs text-muted">{s.sub}</div>
            </div>
          ))}
        </div>
      </section>

      {/* TABLE SECTION */}
      <section
        id="table-section"
        className="max-w-7xl mx-auto px-4 sm:px-8 pb-24"
      >
        <div className="mb-6">
          <h2
            className="font-serif text-3xl text-ink mb-1"
            style={{ letterSpacing: "-1px" }}
          >
            Shareholder Data Table
          </h2>
          <p className="text-sm text-muted">
            Structured view of all ownership records corresponding to the graph
            above.
          </p>
        </div>

        <div
          className="border-2 border-ink rounded-2xl overflow-hidden bg-white"
          style={{ boxShadow: "4px 4px 0 #1a1612" }}
        >
          <div className="overflow-x-auto">
            <table className="data-table w-full text-sm border-collapse">
              <thead>
                <tr className="border-b-2 border-ink bg-card">
                  <th className="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                    Ticker
                  </th>
                  <th className="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                    Investor Name
                  </th>
                  <th className="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden sm:table-cell">
                    Type
                  </th>
                  <th className="text-left px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden sm:table-cell">
                    Origin
                  </th>
                  <th className="text-right px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider">
                    %
                  </th>
                  <th className="text-right px-5 py-3 font-mono text-xs text-muted font-normal uppercase tracking-wider hidden md:table-cell">
                    Shares
                  </th>
                </tr>
              </thead>
              <tbody>
                {tableData.length > 0 ? (
                  tableData.map((row, i) => (
                    <tr
                      key={i}
                      className="border-b border-bdr transition-colors hover:bg-card"
                    >
                      <td className="px-5 py-3">
                        <span className="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md">
                          {ticker}
                        </span>
                      </td>
                      <td className="px-5 py-3 text-sm text-ink max-w-xs">
                        {row.target}
                      </td>
                      <td className="px-5 py-3 hidden sm:table-cell">
                        <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-purple-100 text-purple-700">
                          {row.type}
                        </span>
                      </td>
                      <td className="px-5 py-3 hidden sm:table-cell">
                        <span
                          className={`font-mono text-xs ${row.origin === "FOREIGN" ? "text-violet-700" : "text-green-700"}`}
                        >
                          {row.origin}
                        </span>
                      </td>
                      <td className="px-5 py-3 text-right font-mono font-bold text-sm">
                        {row.pct}%
                      </td>
                      <td className="px-5 py-3 text-right font-mono text-xs text-muted hidden md:table-cell">
                        {row.shares ? row.shares.toLocaleString() : "—"}
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td
                      colSpan={6}
                      className="text-center py-14 text-muted text-sm font-mono"
                    >
                      // search an emiten to populate
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* TOOLTIP */}
      {tooltip.visible && (
        <div
          id="gtt"
          className="fixed pointer-events-none z-50 bg-white border-2 border-ink rounded-xl p-4"
          style={{
            boxShadow: "4px 4px 0 #1a1612",
            maxWidth: "240px",
            left: `${tooltip.x}px`,
            top: `${tooltip.y}px`,
          }}
        >
          <div className="font-mono text-xs font-bold text-bluedk mb-1">
            {tooltip.tick}
          </div>
          <div className="font-sans text-sm font-semibold text-ink mb-2 leading-snug">
            {tooltip.name}
          </div>
          <div
            className="font-sans text-xs text-muted leading-relaxed"
            dangerouslySetInnerHTML={{ __html: tooltip.rows }}
          />
        </div>
      )}
    </main>
  );
}
