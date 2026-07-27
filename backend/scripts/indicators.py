import pandas as pd
import numpy as np
import ta
import math


def _safe_series(func, *args, **kwargs):
    try:
        return func(*args, **kwargs)
    except Exception:
        l = len(args[0]) if args else 1
        return pd.Series([None] * l)


def compute_all(df):
    df = df.copy()
    close, high, low, volume = df['close'], df['high'], df['low'], df['volume']
    n = len(close)

    sma20 = ta.trend.sma_indicator(close, 20)
    sma50 = _safe_series(ta.trend.sma_indicator, close, 50) if n >= 50 else pd.Series([None] * n)
    sma200 = _safe_series(ta.trend.sma_indicator, close, 200) if n >= 200 else pd.Series([None] * n)
    ema20 = ta.trend.ema_indicator(close, 20)
    ema50 = _safe_series(ta.trend.ema_indicator, close, 50) if n >= 50 else pd.Series([None] * n)
    adx = _safe_series(ta.trend.adx, high, low, close, 14)
    di_plus = _safe_series(ta.trend.adx_pos, high, low, close, 14)
    di_minus = _safe_series(ta.trend.adx_neg, high, low, close, 14)

    rsi = _safe_series(ta.momentum.rsi, close, 14)
    macd = _safe_series(ta.trend.macd, close)
    macd_signal = _safe_series(ta.trend.macd_signal, close)
    macd_hist = _safe_series(ta.trend.macd_diff, close)

    stoch_k = _safe_series(ta.momentum.stoch, high, low, close, 14, 3) if n >= 14 else pd.Series([None] * n)
    stoch_d = _safe_series(ta.momentum.stoch_signal, high, low, close, 14, 3, 3) if n >= 14 else pd.Series([None] * n)

    bb = ta.volatility.BollingerBands(close, 20, 2)
    bb_upper = bb.bollinger_hband()
    bb_middle = bb.bollinger_mavg()
    bb_lower = bb.bollinger_lband()
    atr = _safe_series(ta.volatility.average_true_range, high, low, close, 14)

    obv = _safe_series(ta.volume.on_balance_volume, close, volume)
    vol_ma20 = ta.trend.sma_indicator(volume, 20)

    bb_mid_v = bb_middle.iloc[-1]
    bb_w = None
    if not pd.isna(bb_mid_v) and bb_mid_v != 0:
        bb_w = round(float((bb_upper.iloc[-1] - bb_lower.iloc[-1]) / bb_mid_v), 4)

    result = _scalar({
        'sma20': sma20, 'sma50': sma50, 'sma200': sma200,
        'ema20': ema20, 'ema50': ema50,
        'adx': adx, 'di_plus': di_plus, 'di_minus': di_minus,
        'rsi': rsi,
        'macd': macd, 'macd_signal': macd_signal, 'macd_histogram': macd_hist,
        'stoch_k': stoch_k, 'stoch_d': stoch_d,
        'bb_upper': bb_upper, 'bb_middle': bb_middle, 'bb_lower': bb_lower,
        'atr': atr,
        'obv': obv, 'volume_ma20': vol_ma20,
    })
    result['bb_width'] = bb_w
    vl = volume.iloc[-1]
    vma = vol_ma20.iloc[-1]
    result['volume_spike'] = bool(vl > vma * 2) if not (pd.isna(vma) or vma == 0) else False

    sw_high, sw_low = _swing_points(high.values, low.values, df.index, 5)
    sw_high_prices = [p for _, p in sw_high]
    sw_low_prices = [p for _, p in sw_low]
    result['support'] = float(round(min(sw_low_prices[-3:]) if len(sw_low_prices) >= 3 else (min(sw_low_prices) if sw_low_prices else min(low.values)), 2))
    result['resistance'] = float(round(max(sw_high_prices[-3:]) if len(sw_high_prices) >= 3 else (max(sw_high_prices) if sw_high_prices else max(high.values)), 2))

    fib_high = max(high.values)
    fib_low = min(low.values)
    if sw_high_prices:
        fib_high = sw_high_prices[-1]
    if sw_low_prices:
        fib_low = sw_low_prices[-1]
    diff = float(fib_high - fib_low)
    if diff > 0:
        result['fib_23_6'] = float(round(fib_high - 0.236 * diff, 2))
        result['fib_38_2'] = float(round(fib_high - 0.382 * diff, 2))
        result['fib_50_0'] = float(round(fib_high - 0.500 * diff, 2))
        result['fib_61_8'] = float(round(fib_high - 0.618 * diff, 2))
    else:
        for k in ['fib_23_6', 'fib_38_2', 'fib_50_0', 'fib_61_8']:
            result[k] = None

    result['_series'] = {
        'sma20': sma20, 'sma50': sma50,
        'ema20': ema20, 'ema50': ema50,
        'bb_upper': bb_upper, 'bb_middle': bb_middle, 'bb_lower': bb_lower,
        'rsi': rsi,
        'macd': macd, 'macd_signal': macd_signal, 'macd_histogram': macd_hist,
        'stoch_k': stoch_k, 'stoch_d': stoch_d,
        'volume_ma20': vol_ma20,
    }
    result['_swing_high'] = sw_high
    result['_swing_low'] = sw_low
    result['_close'] = close
    return result


def _scalar(d):
    return {k: _last(v) for k, v in d.items()}


def _last(s):
    if s is None or (isinstance(s, (pd.Series, pd.DataFrame)) and (s.empty or s.isna().all())):
        return None
    if isinstance(s, pd.Series):
        val = s.iloc[-1]
    else:
        val = s
    try:
        f = float(val)
        return None if math.isnan(f) or math.isinf(f) else round(f, 2)
    except (ValueError, TypeError):
        return None


def _swing_points(highs, lows, idx, window=5):
    local_max, local_min = [], []
    n = len(highs)
    if n < window * 2 + 1:
        return [], []
    for i in range(window, n - window):
        if highs[i] == max(highs[i - window:i + window + 1]):
            local_max.append((idx[i], highs[i]))
        if lows[i] == min(lows[i - window:i + window + 1]):
            local_min.append((idx[i], lows[i]))
    return local_max, local_min
