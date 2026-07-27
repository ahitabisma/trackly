import math

RISK_PER_TRADE_PCT = 0.01
ATR_MULT_SL = 2.5
R_MULTIPLIERS = [1.5, 2.5, 3.5]
TIME_STOP_DAYS = 20


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
    support, resistance = ind.get('support'), ind.get('resistance')
    overall = signal.get('overall', 'neutral')

    last_close_f = float(last_close)

    if overall == 'bullish' or overall == 'neutral':
        bias = 'buy'
        entry = last_close_f
        stop_loss = _f(entry - atr * ATR_MULT_SL) if atr else (_f(support) if support else None)
        targets = []
        for mult in R_MULTIPLIERS:
            tp = _f(entry + atr * mult) if atr else (_f(resistance * (1 + 0.02 * mult)) if resistance else None)
            if tp:
                targets.append(tp)
    elif overall == 'bearish':
        bias = 'sell'
        entry = last_close_f
        stop_loss = _f(entry + atr * ATR_MULT_SL) if atr else (_f(resistance) if resistance else None)
        targets = []
        for mult in R_MULTIPLIERS:
            tp = _f(entry - atr * mult) if atr else (_f(support * (1 - 0.02 * mult)) if support else None)
            if tp:
                targets.append(tp)
    else:
        return _empty('No clear directional bias — wait for more confirmation')

    if stop_loss is None or stop_loss == entry:
        return _empty('Unable to determine entry/stop due to insufficient data')

    risk_per_share = abs(entry - stop_loss)
    if risk_per_share <= 0 or entry <= 0:
        return _empty('Invalid risk calculation')

    suggested_size = round(RISK_PER_TRADE_PCT / (risk_per_share / entry) * 100, 2)

    tp_list = []
    for i, tp in enumerate(targets):
        reward = abs(tp - entry)
        rr = round(reward / risk_per_share, 2) if risk_per_share > 0 else 0
        tp_list.append({'level': i + 1, 'price': tp, 'rr_ratio': rr})

    entry_zone = _f(entry * 0.99) if bias == 'buy' else _f(entry * 1.01)
    invalidation = _invalidation_note(bias, stop_loss, signal)

    return {
        'bias': bias,
        'entry_zone': entry_zone,
        'entry_price': _f(entry),
        'stop_loss': _f(stop_loss),
        'targets': tp_list,
        'suggested_position_size_pct': suggested_size,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': invalidation,
        'disclaimer': _disclaimer(),
    }


def _empty(reason):
    return {
        'bias': 'hold', 'entry_zone': None, 'entry_price': None,
        'stop_loss': None, 'targets': [],
        'suggested_position_size_pct': 0,
        'suggested_lots': None,
        'time_stop_days': TIME_STOP_DAYS,
        'invalidation_note': reason,
        'disclaimer': _disclaimer(),
    }


def _invalidation_note(bias, stop, signal):
    parts = []
    if signal.get('overall') == 'neutral':
        parts.append('Overall signal is neutral — plan has lower confidence')
    parts.append(f'Stop loss at {stop}')
    parts.append('Invalidated if opposite signal confluence >= 3 indicators')
    parts.append(f'Time stop: {TIME_STOP_DAYS} trading days (~4 weeks)')
    return '. '.join(parts)


def _disclaimer():
    return (
        'Analisis dan trading plan ini bersifat informatif dan tidak mengandung ajakan untuk membeli atau menjual '
        'saham. Keputusan investasi sepenuhnya berada di tangan investor. Kinerja masa lalu tidak menjamin hasil '
        'di masa depan. Selalu lakukan due diligence sebelum bertransaksi.'
    )
