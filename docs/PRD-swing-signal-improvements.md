# PRD: Peningkatan Analisis Teknikal & Trading Plan untuk Swing 1–4 Minggu

## Latar Belakang

Sistem saat ini terdiri dari dua bagian:

**Backend Go (`trackly-backend`, package `analisis`)**

- Endpoint `/analisis` memicu `AnalisisService.RunAnalisis`.
- Data diambil dari Google Sheets via `appsscript.Client`:
  - sheet `master` → daftar ticker (`SearchTickers`)
  - sheet `google_finance` → snapshot harga real-time (`GetTicker`)
  - sheet `chart` → data OHLCV historis, diisi formula `GOOGLEFINANCE()` setelah `selected_ticker`/`date_start`/`date_end` di-set, lalu di-poll (`pollInterval` / `pollMaxRetries`)
- OHLCV dikirim ke script Python via subprocess (`runPythonIndicator`), hasilnya (JSON: `indicators`, `signal`, `trading_plan` + chart PNG base64) di-map ke struct Go (`Indicators`, `SignalResult`, `TradingPlan`, dst).

**Modul Python** (`generate_chart.py`, `indicators.py`, `signals.py`, `trading_plan.py`, `chart_renderer.py`)

- Menghitung indikator teknikal, mengagregasi jadi sinyal, menyusun trading plan, dan merender chart candlestick.

Review kode sebelumnya (lihat README project) menemukan beberapa gap yang relevan khusus untuk tujuan **swing trading 1–4 minggu**. PRD ini merangkum perubahan yang perlu dikerjakan, di sisi Python maupun sisi integrasi Go.

## Tujuan

1. Chart & sinyal lebih representatif untuk keputusan swing 1–4 minggu (bukan day trading).
2. Trading plan lebih actionable & realistis untuk eksekusi saham Indonesia (BEI).
3. Integrasi Go ↔ Python lebih robust, mudah dikonfigurasi, dan mudah di-observe.

## Non-Goals

- Tidak mengganti sumber data (tetap Google Sheets + `GOOGLEFINANCE`).
- Tidak membangun backtesting engine di rilis ini.
- Tidak mengubah UI/consumer frontend — hanya kontrak data (JSON) dan konten chart image.

---

## Ruang Lingkup Perubahan

### A. Python — Chart & Indikator

**A1. Overlay tren & volatilitas di panel utama.**
Tambahkan SMA20/50/200, EMA20/50, dan Bollinger Bands sebagai `addplot(panel=0)` di `chart_renderer.py` (data sudah tersedia di `ind['_series']`, cuma belum pernah digambar).

- _AC_: chart PNG menampilkan garis SMA/EMA dan band BB di atas candlestick.

**A2. Gambar support/resistance & Fibonacci retracement.**
`indicators.py` sudah menghitung `support`, `resistance`, `fib_23_6`..`fib_61_8` — gambar sebagai horizontal line + label di panel utama.

- _AC_: garis-garis tsb muncul di chart dengan label harga.

**A3. Swing point berbasis index, bukan tebak-tebakan harga.**
Refactor `_swing_points()` agar mengembalikan pasangan `(index/date, harga)`, bukan cuma list harga. `chart_renderer.py` pakai index ini langsung untuk plotting, hapus heuristik pencocokan toleransi 0.5%.

- _AC_: marker swing tidak pernah salah tempat walau ada harga duplikat/mirip.

### B. Python — Logika Sinyal & Trading Plan

**B1. Bedakan bobot sinyal berdasar relevansi horizon.**
Indikator tren (SMA cross, ADX, MACD) diberi bobot lebih besar dari oscillator jangka pendek (RSI, Stochastic).

- _AC_: bobot per indikator didefinisikan di satu tempat (dict/config), gampang diubah.

**B2. Filter tren jangka panjang.**
Default: bias `buy` hanya "penuh" kalau close > SMA200 (kalau SMA200 tersedia); begitu juga arah sebaliknya untuk `sell`. Kalau bertentangan, turunkan confidence — jangan langsung diblokir total.

- _AC_: field baru `trend_filter_passed: bool` di output `signal`.

**B3. Confidence level.**
Tambah `confidence` (`high`/`medium`/`low`) di output signal berdasarkan jumlah indikator yang berkontribusi + konsistensi arah.

- _AC_: field `confidence` muncul di JSON `signal`.

**B4. Revisi ATR multiplier stop-loss.**
Dari 1.5x fixed → parameter configurable, default 2–2.5x (lebih sesuai holding mingguan, lebih tahan noise harian).

**B5. Time-stop.**
Tambah `time_stop_days` di trading plan (default ± 20 hari bursa ≈ 4 minggu), disebutkan juga di `invalidation_note`.

- _AC_: field baru `time_stop_days` di output `trading_plan`.

**B6. Position size dalam lot (opsional, butuh input tambahan).**
Tambah `suggested_lots` (asumsi 1 lot = 100 lembar) di trading plan, dihitung kalau CLI diberi argumen baru `--equity` (modal akun). Kalau tidak diberikan, field ini `null`.

- _AC_: field baru `suggested_lots` (nullable) di trading plan.

**B7. (Opsional/low-priority) Awareness batas auto-reject (ARA/ARB) BEI.**
Beri warning di `invalidation_note` kalau target price melewati batas ARA harian wajar dari harga saat ini. Butuh tabel aturan ARA terbaru BEI — kalau data tidak tersedia/tidak yakin, tandai TODO, jangan hardcode angka yang tidak diverifikasi.

### C. Python — Robustness

**C1.** `_safe_series()`: log exception (level debug), jangan silent-swallow total.

**C2. Validasi input di `generate_chart.py`.**
Cek kolom wajib (`date, open, high, low, close, volume`), urutan tanggal ascending, panjang data minimum (mis. ≥ 30 bar). Kalau gagal, kembalikan `{"error": "..."}` yang jelas — jangan biarkan exception mentah bocor ke stdout.

**C3.** Pin versi dependency di `requirements.txt`.

### D. Go — Integrasi & Robustness (`AnalisisService`)

**D1.** `pythonCmd` (`"python"` hardcoded di `runPythonIndicator`) dijadikan field konfigurasi (`pythonBin string`) di struct `AnalisisService`, di-set dari config/env — supaya bisa `python3` di Linux tanpa ubah kode.

**D2.** Propagasi `ctx` ke pemanggilan `s.client.GetSheet` / `SearchTickers` / `GetTicker`. Saat ini `ctx` diterima tapi tidak dipakai di method-method tsb (cuma dipakai di loop polling `RunAnalisis`). Kalau `appsscript.Client` belum mendukung context, tandai TODO di sana juga.

**D3.** Ganti `os.TempDir()` + nama file manual (`chart_%s_%d`) dengan `os.MkdirTemp` per-request, supaya file lebih terisolasi dan tidak predictable.

**D4.** Timeout subprocess (`120*time.Second`, hardcoded) sebaiknya juga configurable dari constructor `NewAnalisisService`, konsisten dengan `pollInterval`/`pollMaxRetries` yang sudah configurable.

**D5.** Jangan swallow error di `snapshot, _ := s.GetTicker(ctx, req.Ticker)` — minimal log warning kalau gagal (perilaku "tetap lanjut tanpa snapshot" boleh dipertahankan, tapi harus ada jejak log).

**D6. Perluas mapping struct untuk field baru dari B2/B3/B5/B6.**
Tambahkan field berikut ke struct terkait (lokasi definisi struct: file types dalam package `analisis`, di luar cuplikan yang di-review):

- `SignalResult`: `Confidence string`, `TrendFilterPassed bool`
- `TradingPlan`: `TimeStopDays *int`, `SuggestedLots *int`
  Update `mapSignal` dan `mapTradingPlan` untuk mem-parsing field-field JSON baru ini (`confidence`, `trend_filter_passed`, `time_stop_days`, `suggested_lots`).

**D7.** Rapikan `gofmt` pada var block (`masterCacheTTL`, dst) yang whitespace-nya tidak konsisten.

---

## Kontrak Data Baru (ringkas)

```json
// signal
{
  "overall": "bullish",
  "score": 0.45,
  "confidence": "medium",
  "trend_filter_passed": true,
  "breakdown": [ ... ]
}

// trading_plan
{
  "bias": "buy",
  "entry_price": 9150,
  "stop_loss": 8830,
  "targets": [ ... ],
  "suggested_position_size_pct": 4.59,
  "suggested_lots": 12,
  "time_stop_days": 20,
  "invalidation_note": "...",
  "disclaimer": "..."
}
```

## Urutan Implementasi yang Disarankan

1. **C2** (validasi input) — fondasi, cegah perubahan lain jadi susah didebug.
2. **A1–A3** (chart) — dampak visual paling langsung dirasakan pengguna.
3. **B1–B5** (signal & plan logic) — inti perbaikan kualitas sinyal.
4. **D1–D7** (Go robustness) — bisa dikerjakan paralel, tidak bergantung ke perubahan Python.
5. **B6, B7** — opsional/prioritas rendah, butuh input tambahan (equity) atau data eksternal (aturan ARA/ARB).

## Risiko & Catatan

- Field baru bersifat **additive** — backward-compatible selama consumer Go di-update bersamaan (lihat D6). Kalau ada consumer JSON lain di luar `trackly-backend`, pastikan mereka tolerant terhadap field tak dikenal.
- **B7** butuh data aturan ARA/ARB BEI yang bisa berubah dari waktu ke waktu — jangan hardcode tanpa sumber yang jelas dan tervalidasi.
- Perubahan bobot sinyal (**B1**) bisa mengubah histori sinyal yang sudah pernah dihasilkan sebelumnya — kalau ada tempat yang menyimpan histori sinyal untuk perbandingan, catat bahwa nilai sebelum/sesudah perubahan ini tidak apple-to-apple.

## Definition of Done

- Semua acceptance criteria di atas terpenuhi, kecuali **B7** (opsional).
- Ada unit test dasar untuk: swing point index-based (A3), trend filter (B2), perhitungan confidence (B3), validasi input (C2), dan mapping struct baru di Go (D6).
- README project diperbarui: field baru didokumentasikan, item roadmap yang sudah selesai dicoret.
