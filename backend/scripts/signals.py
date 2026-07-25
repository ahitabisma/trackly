def compute(ind, ticker):
    breakdown = []
    total = 0
    close = ind.get('_close')
    last_close = close.iloc[-1] if close is not None else None

    # Trend
    sma20, sma50 = ind.get('sma20'), ind.get('sma50')
    if sma20 and sma50:
        if sma20 > sma50:
            breakdown.append(_sig('SMA20/50', 'bullish', 'SMA20 above SMA50 (golden cross)'))
            total += 1
        else:
            breakdown.append(_sig('SMA20/50', 'bearish', 'SMA20 below SMA50 (death cross)'))
            total -= 1
    adx, di_p, di_m = ind.get('adx'), ind.get('di_plus'), ind.get('di_minus')
    if adx and di_p and di_m:
        if adx >= 25:
            if di_p > di_m:
                breakdown.append(_sig('ADX', 'bullish', f'Strong trend ({adx}), DI+ > DI-'))
                total += 1
            else:
                breakdown.append(_sig('ADX', 'bearish', f'Strong trend ({adx}), DI- > DI+'))
                total -= 1
        else:
            breakdown.append(_sig('ADX', 'neutral', f'Weak trend ({adx})'))

    # Momentum
    rsi = ind.get('rsi')
    if rsi is not None:
        if rsi < 30:
            breakdown.append(_sig('RSI', 'bullish', f'oversold ({rsi})'))
            total += 1
        elif rsi > 70:
            breakdown.append(_sig('RSI', 'bearish', f'overbought ({rsi})'))
            total -= 1
        elif rsi < 45:
            breakdown.append(_sig('RSI', 'bullish', f'leaning oversold ({rsi})'))
            total += 0.5
        elif rsi > 55:
            breakdown.append(_sig('RSI', 'bearish', f'leaning overbought ({rsi})'))
            total -= 0.5
        else:
            breakdown.append(_sig('RSI', 'neutral', f'mid-range ({rsi})'))

    macd, macd_sig, macd_hist = ind.get('macd'), ind.get('macd_signal'), ind.get('macd_histogram')
    if macd is not None and macd_sig is not None:
        if macd > macd_sig:
            breakdown.append(_sig('MACD', 'bullish', 'MACD above signal line'))
            total += 1
        else:
            breakdown.append(_sig('MACD', 'bearish', 'MACD below signal line'))
            total -= 1
    if macd_hist is not None:
        if macd_hist > 0:
            breakdown.append(_sig('MACD Hist', 'bullish', 'histogram positive'))
            total += 0.5
        else:
            breakdown.append(_sig('MACD Hist', 'bearish', 'histogram negative'))
            total -= 0.5

    stoch_k, stoch_d = ind.get('stoch_k'), ind.get('stoch_d')
    if stoch_k is not None and stoch_d is not None:
        if stoch_k < 20 and stoch_d < 20:
            breakdown.append(_sig('Stochastic', 'bullish', 'oversold'))
            total += 1
        elif stoch_k > 80 and stoch_d > 80:
            breakdown.append(_sig('Stochastic', 'bearish', 'overbought'))
            total -= 1
        elif stoch_k < 30:
            breakdown.append(_sig('Stochastic', 'bullish', 'leaning oversold'))
            total += 0.5
        elif stoch_k > 70:
            breakdown.append(_sig('Stochastic', 'bearish', 'leaning overbought'))
            total -= 0.5
        else:
            breakdown.append(_sig('Stochastic', 'neutral', 'mid-range'))

    # Volatility — BB position
    bb_u, bb_l = ind.get('bb_upper'), ind.get('bb_lower')
    if last_close is not None and bb_u and bb_l:
        bb_range = bb_u - bb_l
        if bb_range > 0:
            pos = (last_close - bb_l) / bb_range
            if pos < 0.2:
                breakdown.append(_sig('Bollinger', 'bullish', 'price near lower band'))
                total += 0.5
            elif pos > 0.8:
                breakdown.append(_sig('Bollinger', 'bearish', 'price near upper band'))
                total -= 0.5
            else:
                breakdown.append(_sig('Bollinger', 'neutral', 'mid-band'))

    # Volume
    if ind.get('volume_spike'):
        breakdown.append(_sig('Volume', 'confirming', 'volume spike detected'))

    max_score = sum(abs(b['score']) for b in breakdown if b['score'] != 0) or 1
    normalised = total / max_score if max_score > 0 else 0

    if normalised > 0.3:
        overall = 'bullish'
    elif normalised < -0.3:
        overall = 'bearish'
    else:
        overall = 'neutral'

    return {
        'overall': overall,
        'score': round(normalised, 2),
        'breakdown': breakdown,
        'ticker': ticker,
    }


def _sig(name, direction, note):
    score_map = {'bullish': 1, 'bearish': -1, 'neutral': 0, 'confirming': 0}
    return {'indicator': name, 'signal': direction, 'note': note, 'score': score_map.get(direction, 0)}
