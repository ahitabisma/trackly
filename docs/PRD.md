# PRD: Fitur Analisis Teknikal & Trading Plan (Swing 1-4 Minggu)

## 1. Latar Belakang

Aplikasi memiliki frontend (Nuxt/Vue) + backend (Go) yang terhubung ke Google Sheets (via Apps Script API) sebagai sumber data Google Finance, dengan 4 sheet utama:

- `master` — daftar seluruh emiten (Kode, Nama Perusahaan, Tanggal Pencatatan, Saham, Papan Pencatatan)
- `config` — key-value control cell: `selected_ticker`, `date_start`, `date_end`
- `google_finance` — snapshot harga & fundamental per ticker
- `chart` — data historis OHLCV mengikuti ticker & rentang tanggal aktif di `config`

Dibutuhkan menu **Analisis** di mana user bisa mencari ticker, memilih rentang tanggal, lalu melihat grafik historis lengkap dengan indikator teknikal, sinyal (bullish/bearish/netral), serta **trading plan** (entry, stop loss, take profit) untuk **swing trading 1–4 minggu** di saham Indonesia (BEI).

> Output analisis & trading plan bersifat sinyal otomatis berbasis data historis, **bukan rekomendasi investasi**. Aplikasi wajib menampilkan disclaimer di UI.

## 2. Tujuan

- User bisa mencari ticker cepat (autocomplete dari `master`)
- User bisa melihat snapshot fundamental (price, 52w high/low, market cap, PE, EPS)
- User bisa melihat chart multi-panel: candlestick + overlay indikator + panel volume/RSI/MACD/Stochastic
- Sistem menghasilkan **skor/sinyal teknikal** dari confluence beberapa indikator dengan bobot berbeda
- Sistem menghasilkan **trading plan** otomatis: entry zone, stop loss, multi-level take profit, risk/reward, position sizing, time-stop
- Proses end-to-end target < 10 detik

## 3. Non-Goals

- Tidak membuat sistem multi-user/session terisolasi di versi awal
- Tidak membangun data warehouse/cache historis sendiri di luar Google Sheets
- Tidak mencakup fitur watchlist/portfolio/eksekusi order otomatis
- Bukan pengganti nasihat keuangan profesional
- Tidak mengganti sumber data (tetap Google Sheets + GOOGLEFINANCE)
- Tidak membangun backtesting engine di rilis ini

## 4. User Flow

1. User membuka menu **Analisis**
2. `onMounted` → frontend fetch semua ticker dari `GET /api/tickers`, simpan di memori
3. User mengetik nama/kode ticker → filter lokal via `computed` (instant)
4. User memilih ticker → memilih rentang tanggal (default: 1 tahun terakhir)
5. User klik "Analisis" → loading state
6. Backend:
   - Update `config.selected_ticker`, `date_start`, `date_end` via Apps Script
   - Poll sheet `chart` setiap 500ms (max 20× = 10s) sampai data tersedia
   - Ambil data chart terbaru
   - Jalankan Python subprocess: hitung indikator, sinyal, trading plan, render PNG
7. Frontend menampilkan: snapshot fundamental → chart multi-panel → ringkasan sinyal (badge + skor + breakdown) → trading plan (entry/SL/TP/RR) → disclaimer → tabel OHLCV

## 5. Functional Requirements

### 5.1 Backend (Go)

| Endpoint | Method | Deskripsi |
|----------|--------|-----------|
| `/api/tickers` | GET | Seluruh daftar ticker dari sheet `master` (cache TTL 5 menit) |
| `/api/analisis` | POST | Body `{ticker, date_start, date_end}` → orkestrasi update config, poll chart, jalankan Python, return hasil lengkap |

**Response `/api/analisis`:**

```json
{
  "snapshot": { "kode": "ANTM", "company_name": "...", "price": 3000, "high52": 3250, "low52": 2770, "volume": 61092600, "marketcap": 0, "pe": null, "eps": 0, "currency": "IDR" },
  "chart_image": "base64 PNG",
  "indicators": { ... },
  "signal": { "overall": "bullish", "score": 0.45, "confidence": "medium", "trend_filter_passed": true, "breakdown": [...] },
  "trading_plan": { "bias": "buy", "entry_price": 9150, "stop_loss": 8830, "targets": [...], "suggested_position_size_pct": 4.59, "suggested_lots": 12, "time_stop_days": 20, "invalidation_note": "...", "disclaimer": "..." },
  "error": ""
}
```

**Konfigurasi service:**

| Field | Default | Keterangan |
|-------|---------|------------|
| `pythonBin` | auto-detect (python3/py/python) | Binary Python yang dipakai |
| `pythonScriptPath` | `./scripts/generate_chart.py` | Path absolut ke script Python |
| `pollIntervalMs` | 500 | Interval polling sheet chart (ms) |
| `pollMaxRetries` | 20 | Maksimal retry polling |
| `subprocessTimeout` | 120s | Timeout eksekusi Python |

### 5.2 Python — Indikator, Sinyal, Trading Plan & Chart

**Input:** OHLCV dari sheet `chart` via temp JSON file. Kolom wajib: `date, open, high, low, close, volume`. Minimal 30 bar data, urutan ascending.

**CLI:** `python generate_chart.py --input <temp.json> --ticker <KODE> --out <chart.png> [--equity <modal>]`

**Stdout:** JSON `{ indicators: {...}, signal: {...}, trading_plan: {...} }`

#### Indikator

| Grup | Indikator |
|------|-----------|
| Trend | SMA20/50/200, EMA20/50, ADX + DI+/DI- |
| Momentum | RSI(14), MACD (12,26,9) + signal + histogram, Stochastic %K/%D |
| Volatilitas | Bollinger Bands (20,2), ATR(14) |
| Volume | OBV, Volume MA20, volume spike flag |
| Price Levels | Support/Resistance (swing high/low), Fibonacci retracement (23.6/38.2/50/61.8/78.6) |

#### Sinyal (Confluence-based)

- Tiap indikator memberi vote: bullish (+1) / bearish (-1) / netral (0)
- **Bobot berbeda**: indikator tren (SMA cross, ADX, MACD) lebih besar dari oscillator jangka pendek (RSI, Stochastic)
- Filter tren jangka panjang: `trend_filter_passed = bool(close > SMA200)` jika SMA200 tersedia
- Skor total dinormalisasi → label `bullish`/`bearish`/`netral`
- `confidence`: `high`/`medium`/`low` berdasarkan jumlah indikator berkontribusi + konsistensi arah

#### Trading Plan

- **Bias**: `buy`/`sell`/`hold` dari signal overall
- **Entry price**: area support terdekat (buy) / resistance terdekat (sell) atau pullback ke SMA20/EMA20
- **Stop loss**: di bawah support kunci / swing low, buffer ATR × 2.0–2.5 (configurable)
- **Take profit**: multi-level (1.5R, 2.5R, 3.5R) berdasarkan resistance berikutnya
- **Time stop**: default 20 hari bursa (≈ 4 minggu)
- **Position sizing**: `suggested_position_size_pct` (1-2% risk), `suggested_lots` (nullable, butuh `--equity`)
- **Invalidation note**: kondisi batal
- **Disclaimer**: teks permanen

#### Chart Render (multi-panel matplotlib PNG)

1. Panel utama: candlestick + SMA20/50/200 + EMA20/50 + Bollinger Bands + support/resistance + Fibonacci level + entry/SL/TP lines
2. Volume bar + Volume MA20 + volume spike highlight
3. RSI + garis 30/70
4. MACD + signal + histogram
5. Stochastic + garis 20/80

### 5.3 Frontend

- **Search box** autocomplete — fetch semua ticker sekali (`onMounted`), filter lokal via `computed`
- **Date range picker** (`<input type="date">`, default 1 tahun)
- **Tombol "Analisis"** + loading state
- **Area hasil** (berurutan):
  1. Snapshot card — price, 52w high/low, market cap, PE, EPS
  2. Chart multi-panel (PNG dari backend)
  3. Signal badge — overall + score + confidence + trend_filter_passed + expandable breakdown
  4. Indikator grid — SMA20/50/200, EMA20, RSI, MACD, ADX, ATR, BB Width, volume spike, S/R, Fibonacci
  5. Trading plan card — bias, entry, SL, multi-level TP table (dengan RR ratio), position size, time stop, invalidation note
  6. Disclaimer — permanen, selalu terlihat
  7. OHLCV table — collapsible (`<details>`)
- **Nilai `#N/A` / null** → tampil sebagai "N/A"

## 6. Perbaikan & Peningkatan (dari review kode)

### A. Python — Chart & Indikator

| Item | Deskripsi |
|------|-----------|
| A1 | Overlay SMA/EMA/BB di panel utama chart |
| A2 | Gambar support/resistance & Fibonacci sebagai horizontal line + label |
| A3 | `_swing_points()` return `(index, harga)` bukan list harga saja — marker swing tidak pernah salah tempat |

### B. Python — Logika Sinyal & Trading Plan

| Item | Deskripsi |
|------|-----------|
| B1 | Bobot sinyal berdasar relevansi horizon (tren > oscillator) |
| B2 | Filter tren jangka panjang: `trend_filter_passed` dari SMA200 |
| B3 | `confidence: high/medium/low` |
| B4 | ATR multiplier configurable, default 2.0–2.5× |
| B5 | `time_stop_days` default 20 |
| B6 | `suggested_lots` (nullable, butuh `--equity`) |
| B7 | (Opsional) ARA/ARB BEI awareness warning |

### C. Python — Robustness

| Item | Deskripsi |
|------|-----------|
| C1 | `_safe_series()` log exception, jangan silent-swallow |
| C2 | Validasi input: kolom wajib, urutan ascending, min 30 bar → error JSON jelas |
| C3 | Pin versi dependency di `requirements.txt` |

### D. Go — Integrasi & Robustness

| Item | Deskripsi |
|------|-----------|
| D1 | `pythonBin` configurable dari env, auto-detect python3/py/python |
| D2 | Propagasi `ctx` ke `GetSheet` / `SearchTickers` / `GetTicker` |
| D3 | `os.MkdirTemp` per-request ganti `os.TempDir()` manual |
| D4 | Timeout subprocess configurable dari constructor |
| D5 | Jangan swallow error `GetTicker` — minimal log warning |
| D6 | Field baru Go struct: `Confidence`, `TrendFilterPassed`, `TimeStopDays`, `SuggestedLots` |
| D7 | `gofmt` konsisten |

## 7. Keputusan Implementasi

1. **Timing refresh GOOGLEFINANCE** — Backend **poll** sheet `chart` setiap 500ms, max 20× (10s timeout). Jika timeout, return error.
2. **Handling #N/A** — Ditampilkan sebagai "N/A" di UI. Nilai numerik `0` dari `#N/A` dibedakan via `*float64`.
3. **Concurrency** — `config.selected_ticker` adalah shared single cell. Tooling internal **single-user**, aman diabaikan. Upgrade path: per-session sheet atau queue.
4. **Python output** — stdout JSON + PNG file via `--out`. Go baca PNG sebagai base64.
5. **Python discovery** — Auto-detect: `python3` → `py` → `python`, diverifikasi dengan `--version`.
6. **Error propagation** — Python error output (field `error` di JSON stdout) diteruskan ke response API.

## 8. Metrik Keberhasilan

- User bisa selesaikan satu siklus pencarian → chart, indikator, sinyal, trading plan tampil tanpa error
- Waktu end-to-end rata-rata < 10 detik
- Autocomplete terasa instant (filter lokal, tanpa network)
- Tidak ada crash saat data mengandung `#N/A`, data historis pendek, atau ticker tidak lengkap
- Trading plan konsisten matematis (RR ratio, SL/TP wajar relatif terhadap ATR & support/resistance)
- Chart PNG menampilkan seluruh overlay dan anotasi dengan benar
