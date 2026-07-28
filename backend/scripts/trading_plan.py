import math

RISK_PER_TRADE_PCT = 0.01
ATR_MULT_SL = 2.5              # fallback kalau support nggak tersedia/nggak valid
STRUCTURE_BUFFER_ATR_MULT = 0.5  # buffer di luar level support, biar nggak kena stop noise
MIN_STOP_ATR_MULT = 1.0        # jarak minimum stop dari entry
R_MULTIPLIERS = [1.0, 2.0, 3.0]  # RR target: 1:1, 1:2, 1:3
TIME_STOP_DAYS = 20

# CATATAN: trading plan ini LONG-ONLY (asumsi akun retail IDX standar, tanpa
# fasilitas short selling/margin trading). Sinyal 'bearish' TIDAK menghasilkan
# rencana short — cuma dikasih status 'avoid' (hindari entry, tunggu sinyal
# berubah). Kalau suatu saat akunmu punya fasilitas short, logic short bisa
# ditambah lagi sebagai jalur terpisah.


def _f(v):
    if v is None:
        return None
    return round(float(v), 2)


def compute(ind, signal):
    close = ind.get('_close')
    last_close = close.iloc[-1] if close is not None else None
    if last_close is None:
        return _empty('no price data')

    atr = ind.get('atr')
    support = ind.get('support')
    overall = signal.get('overall', 'neutral')
    entry = float(last_close)

    if overall == 'bearish':
        return _avoid(entry, signal)

    # bullish atau neutral -> long-only plan
    stop_loss, stop_basis = _compute_stop_loss(entry, atr, support)

    if stop_loss is None or stop_loss == entry:
        return _empty('Unable to determine entry/stop due to insufficient data')

    risk_per_share = abs(entry - stop_loss)
    if risk_per_share <= 0 or entry <= 0:
        return _empty('Invalid risk calculation')

    suggested_size = round(RISK_PER_TRADE_PCT / (risk_per_share / entry) * 100, 2)

    # Target diturunkan LANGSUNG dari risk_per_share, jadi RR pasti presisi 1:1, 1:2, 1:3
    tp_list = []
    for i, rr in enumerate(R_MULTIPLIERS):
        tp = _f(entry + risk_per_share * rr)
        tp_list.append({'level': i + 1, 'price': tp, 'rr_ratio': rr})

    entry_zone = _f(entry * 0.99)
    invalidation = _invalidation_note(stop_loss, signal, stop_basis)

    return {
        'bias': 'buy',
        'entry_zone': entry_zone,
        'entry_price': _f(entry),
        'stop_loss': _f(stop_loss),
        'stop_loss_basis': stop_basis,
        'targets': tp_list,
        'suggested_position_size_pct': suggested_size,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': invalidation,
        'disclaimer': _disclaimer(),
    }


def _compute_stop_loss(entry, atr, support):
    """
    Prioritas: pakai level support + buffer kecil dari ATR, supaya stop
    loss nempel ke level teknikal yang relevan (bukan jarak ATR murni yang
    bisa nggak nyambung sama struktur harga). Kalau support nggak ada atau
    posisinya salah arah (di atas entry), fallback ke ATR multiple.

    Ada pengaman jarak minimum (MIN_STOP_ATR_MULT x ATR) supaya kalau
    support KETERLALUAN deket entry, stop tetap dilebarkan dikit biar
    nggak gampang kena stop cuma gara-gara noise harian.
    """
    buffer = atr * STRUCTURE_BUFFER_ATR_MULT if atr else 0
    min_distance = atr * MIN_STOP_ATR_MULT if atr else None

    candidate = None
    if support is not None and support < entry:
        candidate = support - buffer

    if candidate is not None:
        distance = abs(entry - candidate)
        if min_distance and distance < min_distance:
            candidate = entry - min_distance
        return _f(candidate), 'structure'  # support + ATR buffer

    if atr:
        return _f(entry - atr * ATR_MULT_SL), 'atr_fallback'  # support nggak valid, balik ke ATR murni

    return None, None


def _avoid(entry, signal):
    """Sinyal bearish, tapi akun long-only -> nggak ada trading plan sama
    sekali, cuma rekomendasi hindari entry / tunggu."""
    score = signal.get('score')
    confidence = signal.get('confidence')
    reason = (
        f'Sinyal bearish (score {score}, confidence {confidence}). Akun ini asumsinya long-only '
        f'(tidak ada fasilitas short selling), jadi TIDAK ada entry buy yang disarankan sekarang. '
        f'Tunggu sinyal berubah jadi bullish/neutral. Kalau kamu sudah hold saham ini, pertimbangkan '
        f'review/keluar posisi sesuai rencana manajemen risiko kamu sendiri (di luar cakupan trading plan ini).'
    )
    return {
        'bias': 'avoid',
        'entry_zone': None,
        'entry_price': _f(entry),
        'stop_loss': None,
        'stop_loss_basis': None,
        'targets': [],
        'suggested_position_size_pct': 0,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': reason,
        'disclaimer': _disclaimer(),
    }


def _empty(reason):
    return {
        'bias': 'hold', 'entry_zone': None, 'entry_price': None,
        'stop_loss': None, 'stop_loss_basis': None, 'targets': [],
        'suggested_position_size_pct': 0,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': reason,
        'disclaimer': _disclaimer(),
    }


def _invalidation_note(stop, signal, stop_basis):
    parts = []
    if signal.get('overall') == 'neutral':
        parts.append('Overall signal is neutral — plan has lower confidence')
    basis_txt = 'support level + ATR buffer' if stop_basis == 'structure' else 'ATR-based (no valid support nearby)'
    parts.append(f'Stop loss at {stop} ({basis_txt})')
    parts.append('Invalidated if opposite signal confluence >= 3 indicators')
    parts.append(f'Time stop: {TIME_STOP_DAYS} trading days (~4 weeks)')
    return '. '.join(parts)


def _disclaimer():
    return (
        'Analisis dan trading plan ini bersifat informatif dan tidak mengandung ajakan untuk membeli atau menjual '
        'saham. Keputusan investasi sepenuhnya berada di tangan investor. Kinerja masa lalu tidak menjamin hasil '
        'di masa depan. Selalu lakukan due diligence sebelum bertransaksi.'
    )