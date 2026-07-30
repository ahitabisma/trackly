import math

RISK_PER_TRADE_PCT = 0.01
ATR_MULT_SL = 1.5              # turun dari 2.5 -- 2.5x ATR kegedean buat swing, apalagi saham IDX yang ATR-nya bisa lebar
STRUCTURE_BUFFER_ATR_MULT = 0.5
MIN_STOP_ATR_MULT = 1.0
MAX_STOP_LOSS_PCT = 0.08       # HARD CAP: stop loss max 8% dari entry, berapapun ATR-nya -- ini yang bikin SL 25% kemarin nggak akan kejadian lagi
R_MULTIPLIERS = [1.0, 2.0, 3.0]
TIME_STOP_DAYS = 20
EXTENDED_ATR_THRESHOLD = 1.5   # kalau harga > 1.5x ATR di atas level referensi -> dianggap "extended"
PULLBACK_ZONE_BUFFER_ATR = 0.3  # jarak zona pullback dari level referensi, biar nggak pas mepet level

# Entry SELALU direpresentasikan sebagai RANGE (zone_low - zone_high), lebar
# zona ikut ATR tapi dibatasi persentase biar nggak kekecilan/kegedean.
ENTRY_ZONE_ATR_FRACTION = 0.15  # lebar zona per sisi = 0.15x ATR
ENTRY_ZONE_MIN_PCT = 0.003      # minimal 0.3% dari harga per sisi (saham low-ATR)
ENTRY_ZONE_MAX_PCT = 0.015      # maksimal 1.5% dari harga per sisi (saham high-ATR)

# CATATAN: trading plan ini LONG-ONLY. Sinyal 'bearish' -> 'avoid'.


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
    sma20 = ind.get('sma20')
    overall = signal.get('overall', 'neutral')
    current_price = float(last_close)

    if overall == 'bearish':
        return _avoid(current_price, signal)

    # Tentuin ENTRY PLAN dulu -- bukan asal current_price. Kalau harga udah
    # extended jauh dari support/SMA20, sarankan pullback zone.
    planned_entry, zone_low, zone_high, entry_type, ref_used = _compute_entry_plan(
        current_price, atr, support, sma20
    )

    # Stop-loss dihitung dari PLANNED ENTRY (bukan current_price) -- kalau
    # rencananya masuk pas pullback, risk/reward juga dihitung dari titik itu.
    stop_loss, stop_basis = _compute_stop_loss(planned_entry, atr, support)

    if stop_loss is None or stop_loss == planned_entry:
        return _empty('Unable to determine entry/stop due to insufficient data')

    risk_per_share = abs(planned_entry - stop_loss)
    if risk_per_share <= 0 or planned_entry <= 0:
        return _empty('Invalid risk calculation')

    suggested_size = round(RISK_PER_TRADE_PCT / (risk_per_share / planned_entry) * 100, 2)

    tp_list = []
    for i, rr in enumerate(R_MULTIPLIERS):
        tp = _f(planned_entry + risk_per_share * rr)
        tp_list.append({'level': i + 1, 'price': tp, 'rr_ratio': rr})

    invalidation = _invalidation_note(stop_loss, signal, stop_basis)
    entry_note = _entry_note(entry_type, ref_used, current_price, planned_entry)

    return {
        'bias': 'buy',
        'entry_type': entry_type,  # 'market' (harga sekarang OK) atau 'pullback' (tunggu turun dulu)
        'current_price': _f(current_price),
        'entry_price': _f(planned_entry),
        'entry_zone': {'low': _f(zone_low), 'high': _f(zone_high)},
        'entry_note': entry_note,
        'stop_loss': _f(stop_loss),
        'stop_loss_basis': stop_basis,
        'targets': tp_list,
        'suggested_position_size_pct': suggested_size,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': invalidation,
        'disclaimer': _disclaimer(),
    }


def _compute_entry_plan(current_price, atr, support, sma20):
    """
    Cari level referensi teknikal TERDEKAT di bawah current_price (support
    atau SMA20, mana yang lebih deket -- bukan selalu support duluan).
    Ukur seberapa jauh current_price sudah lari dari level itu (dalam
    satuan ATR):
      - Kalau <= EXTENDED_ATR_THRESHOLD: harga masih wajar, entry-nya
        zona simetris di sekitar current_price ('market'), lebar zona
        ikut ATR (lihat _entry_zone_half_width).
      - Kalau > EXTENDED_ATR_THRESHOLD: harga udah extended, entry_zone
        jadi rentang pullback antara level referensi dan harga sekarang,
        BUKAN entry di harga sekarang ('pullback').
      - Kalau nggak ada level referensi valid sama sekali (support/SMA20
        nggak ada atau semuanya di atas current_price): fallback 'market'
        dengan zona simetris ikut ATR juga, nggak bisa dinilai extended
        atau nggak.
    """
    if atr is None or atr == 0:
        half_width = current_price * ENTRY_ZONE_MIN_PCT
        zone_low, zone_high = _entry_zone(current_price, half_width)
        return current_price, zone_low, zone_high, 'market', None

    candidates = [c for c in [support, sma20] if c is not None and c < current_price]
    if not candidates:
        half_width = _entry_zone_half_width(current_price, atr)
        zone_low, zone_high = _entry_zone(current_price, half_width)
        return current_price, zone_low, zone_high, 'market', None

    reference = max(candidates)  # yang PALING DEKET ke current_price
    distance_atr = (current_price - reference) / atr

    if distance_atr <= EXTENDED_ATR_THRESHOLD:
        half_width = _entry_zone_half_width(current_price, atr)
        zone_low, zone_high = _entry_zone(current_price, half_width)
        return current_price, zone_low, zone_high, 'market', reference

    # Target entry-nya level DEKAT referensi (support/SMA20) + buffer, BUKAN
    # direntang dari referensi sampai harga sekarang -- itu yang bikin zona
    # lebar banget (bisa ratusan poin) kalau harga udah lari jauh dari referensi.
    planned_entry = reference + atr * PULLBACK_ZONE_BUFFER_ATR
    half_width = _entry_zone_half_width(planned_entry, atr)
    zone_low, zone_high = _entry_zone(planned_entry, half_width)
    # jaga-jaga: zona jangan sampai nyentuh/ngelewatin harga sekarang
    if zone_high >= current_price:
        zone_high = current_price * 0.995
        if zone_low >= zone_high:
            zone_low = zone_high - half_width * 2
    return planned_entry, zone_low, zone_high, 'pullback', reference


def _entry_zone_half_width(price, atr):
    """Lebar zona entry per sisi: ikut ATR, tapi dibatasi ENTRY_ZONE_MIN_PCT..ENTRY_ZONE_MAX_PCT dari harga."""
    half_width = atr * ENTRY_ZONE_ATR_FRACTION
    min_half = price * ENTRY_ZONE_MIN_PCT
    max_half = price * ENTRY_ZONE_MAX_PCT
    return min(max(half_width, min_half), max_half)


def _entry_zone(price, half_width):
    return price - half_width, price + half_width


def _entry_note(entry_type, ref_used, current_price, planned_entry):
    if entry_type == 'market':
        if ref_used is not None:
            return f'Harga saat ini ({_f(current_price)}) masih wajar relatif ke level teknikal terdekat ({_f(ref_used)}), entry di kisaran harga sekarang masih masuk akal.'
        return f'Harga saat ini ({_f(current_price)}) dipakai sebagai acuan entry -- tidak ada level support/SMA20 valid untuk dijadikan pembanding.'
    return (
        f'Harga saat ini ({_f(current_price)}) sudah cukup jauh dari level teknikal terdekat ({_f(ref_used)}) -- '
        f'disarankan TUNGGU pullback ke zona entry sebelum masuk, bukan beli di harga sekarang.'
    )


def _compute_stop_loss(entry, atr, support):
    """
    Basis stop loss (prioritas):
      1) structure: level support - buffer ATR, kalau ada support valid di bawah entry.
      2) atr_fallback: ATR_MULT_SL x ATR di bawah entry, kalau nggak ada support valid.
    Jarak stop lalu di-clamp DUA ARAH:
      - minimal MIN_STOP_ATR_MULT x ATR (biar stop nggak ke-trigger noise harian)
      - maksimal MAX_STOP_LOSS_PCT dari entry (batas keras risk per trade --
        ini yang sebelumnya kebobolan jadi 25% pas ATR saham lagi lebar)
    """
    buffer = atr * STRUCTURE_BUFFER_ATR_MULT if atr else 0
    min_distance = atr * MIN_STOP_ATR_MULT if atr else None
    max_distance = entry * MAX_STOP_LOSS_PCT

    basis = None
    distance = None

    if support is not None and support < entry:
        distance = entry - (support - buffer)
        basis = 'structure'
    elif atr:
        distance = atr * ATR_MULT_SL
        basis = 'atr_fallback'

    if distance is None:
        return None, None

    if min_distance and distance < min_distance:
        distance = min_distance

    if distance > max_distance:
        distance = max_distance
        basis = f'{basis}_capped'

    return _f(entry - distance), basis


def _avoid(entry, signal):
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
        'entry_type': None,
        'current_price': _f(entry),
        'entry_price': None,
        'entry_zone': None,
        'entry_note': None,
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
        'bias': 'hold', 'entry_type': None, 'current_price': None,
        'entry_price': None, 'entry_zone': None, 'entry_note': None,
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
    basis_txt = {
        'structure': 'support level + ATR buffer',
        'atr_fallback': 'ATR-based (no valid support nearby)',
        'structure_capped': f'support level, jarak dipangkas ke maks {int(MAX_STOP_LOSS_PCT * 100)}% dari entry',
        'atr_fallback_capped': f'ATR-based, jarak dipangkas ke maks {int(MAX_STOP_LOSS_PCT * 100)}% dari entry',
    }.get(stop_basis, stop_basis)
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


def _position_entry_note(buy_price, current_price, pnl_pct, entry_zone, stop_loss):
    note = (
        f"Kamu sudah hold di harga {_f(buy_price)}. "
        f"Harga saat ini {_f(current_price)} ({pnl_pct}% dari avg). "
        f"Stop loss teknikal di {_f(stop_loss)}."
    )
    if entry_zone and entry_zone.get('low') and entry_zone.get('high'):
        note += (
            f" Entry zone teknikal {_f(entry_zone['low'])} — {_f(entry_zone['high'])} "
            f"tersedia untuk averaging down jika sesuai risk management."
        )
    return note


def _position_targets(avg_price, stop_loss):
    if stop_loss is None or avg_price <= stop_loss:
        return []
    risk_per_share = avg_price - stop_loss
    targets = [{'level': 0, 'price': _f(avg_price), 'rr_ratio': 0}]
    for i, mult in enumerate(R_MULTIPLIERS, 1):
        targets.append({'level': i, 'price': _f(avg_price + risk_per_share * mult), 'rr_ratio': mult})
    return targets


def _pos_plan_buy(plan, buy_price, current_price, unrealized_pnl_pct):
    return {
        **plan,
        'avg_price': _f(buy_price),
        'current_vs_avg_pct': unrealized_pnl_pct,
        'entry_note': _position_entry_note(
            buy_price, current_price, unrealized_pnl_pct,
            plan.get('entry_zone'), plan.get('stop_loss'),
        ),
        'targets': _position_targets(buy_price, plan.get('stop_loss')),
    }


def compute_position_review(ind, signal, buy_price, lot, buy_date):
    close = ind.get('_close')
    last_close = close.iloc[-1] if close is not None else None
    if last_close is None:
        return _position_empty(buy_price, lot, buy_date, 'no price data')

    current_price = float(last_close)
    unrealized_pnl = (current_price - buy_price) * lot * 100
    unrealized_pnl_pct = round((current_price - buy_price) / buy_price * 100, 2)
    holding_days = _holding_days(buy_date)

    plan = compute(ind, signal)

    if plan['bias'] == 'avoid':
        recommendation = 'sell'
        suggested_exit_price = _f(current_price)
        suggested_stop = None
        reason = (
            f"Sinyal berubah bearish (score {signal.get('score')}). Kamu sudah hold saham ini -- "
            f"pertimbangkan review/keluar posisi sesuai rencana manajemen risiko kamu sendiri."
        )
        pos_plan = {**plan, 'avg_price': _f(buy_price), 'current_vs_avg_pct': unrealized_pnl_pct}
    elif plan['bias'] == 'buy':
        recommendation = 'hold'
        suggested_exit_price = None
        suggested_stop = plan['stop_loss']
        reason = f"Sinyal masih {signal.get('overall', 'neutral')}. Hold selama harga di atas level invalidasi {suggested_stop}."
        pos_plan = _pos_plan_buy(plan, buy_price, current_price, unrealized_pnl_pct)
    else:
        recommendation = 'hold'
        suggested_exit_price = None
        suggested_stop = None
        reason = plan['invalidation_note'] or 'Data tidak cukup untuk rekomendasi.'
        pos_plan = {**plan, 'avg_price': _f(buy_price), 'current_vs_avg_pct': unrealized_pnl_pct}

    return {
        'ticker': signal.get('ticker'),
        'buy_price': _f(buy_price),
        'current_price': _f(current_price),
        'lot': lot,
        'unrealized_pnl': round(unrealized_pnl, 2),
        'unrealized_pnl_pct': unrealized_pnl_pct,
        'holding_days': holding_days,
        'recommendation': recommendation,
        'suggested_exit_price': suggested_exit_price,
        'suggested_stop': suggested_stop,
        'reason': reason,
        'trading_plan': pos_plan,
        'disclaimer': _disclaimer(),
    }


def _position_empty(buy_price, lot, buy_date, reason):
    return {
        'buy_price': _f(buy_price), 'current_price': None, 'lot': lot,
        'unrealized_pnl': None, 'unrealized_pnl_pct': None,
        'holding_days': _holding_days(buy_date),
        'recommendation': None, 'suggested_exit_price': None,
        'suggested_stop': None, 'reason': reason,
        'disclaimer': _disclaimer(),
    }


def _holding_days(buy_date_str):
    from datetime import datetime
    try:
        buy_dt = datetime.strptime(buy_date_str, '%Y-%m-%d')
        return (datetime.now() - buy_dt).days
    except (ValueError, TypeError):
        return None