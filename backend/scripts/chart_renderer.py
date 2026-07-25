import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import mplfinance as mpf
import pandas as pd
import numpy as np

plt.rcParams['font.family'] = 'sans-serif'
plt.rcParams['font.size'] = 9


def _valid(s):
    return s is not None and isinstance(s, pd.Series) and not s.isna().all() and len(s) > 0


def render(df, ind, signal, trading_plan, out_path, ticker):
    df = df.copy()
    df['date'] = pd.to_datetime(df['date'])
    df.set_index('date', inplace=True)
    df.index.name = None

    series = ind.get('_series', {})
    sw_high = ind.get('_swing_high', [])
    sw_low = ind.get('_swing_low', [])

    colors = mpf.make_marketcolors(up='#26a69a', down='#ef5350', wick='inherit', edge='inherit', volume='inherit')
    style = mpf.make_mpf_style(base_mpf_style='charles', marketcolors=colors,
                               rc={'font.size': 8, 'axes.labelsize': 8, 'xtick.labelsize': 7, 'ytick.labelsize': 7})

    apds = []
    panels = {'main': True, 'volume': True, 'rsi': False, 'macd': False, 'stoch': False}
    p = 1

    # Volume
    vol_ma = series.get('volume_ma20')
    apds.append(mpf.make_addplot(df['volume'], panel=p, color='#78909c', type='bar', width=0.6, alpha=0.4, ylabel='Volume'))
    if _valid(vol_ma):
        apds.append(mpf.make_addplot(vol_ma, panel=p, color='#f57c00', width=0.8))

    # RSI
    rsi = series.get('rsi')
    if _valid(rsi):
        p += 1
        panels['rsi'] = True
        apds.append(mpf.make_addplot(rsi, panel=p, color='#1565c0', width=0.8, ylabel='RSI'))
        n = len(rsi)
        apds.append(mpf.make_addplot(pd.Series([70] * n, index=rsi.index), panel=p, color='#ef5350', width=0.5))
        apds.append(mpf.make_addplot(pd.Series([30] * n, index=rsi.index), panel=p, color='#26a69a', width=0.5))

    # MACD
    macd = series.get('macd')
    macd_sig = series.get('macd_signal')
    macd_hist = series.get('macd_histogram')
    if _valid(macd):
        p += 1
        panels['macd'] = True
        apds.append(mpf.make_addplot(macd, panel=p, color='#1565c0', width=0.8, ylabel='MACD'))
        if _valid(macd_sig):
            apds.append(mpf.make_addplot(macd_sig, panel=p, color='#e65100', width=0.8))
        if _valid(macd_hist):
            bar_colors = ['#ef5350' if v < 0 else '#26a69a' for v in macd_hist]
            apds.append(mpf.make_addplot(macd_hist, panel=p, type='bar', width=0.6, color=bar_colors, alpha=0.5))

    # Stochastic
    sk = series.get('stoch_k')
    sd = series.get('stoch_d')
    if _valid(sk):
        p += 1
        panels['stoch'] = True
        apds.append(mpf.make_addplot(sk, panel=p, color='#1565c0', width=0.8, ylabel='Stoch'))
        if _valid(sd):
            apds.append(mpf.make_addplot(sd, panel=p, color='#e65100', width=0.8))
        n = len(sk)
        apds.append(mpf.make_addplot(pd.Series([80] * n, index=sk.index), panel=p, color='#ef5350', width=0.5))
        apds.append(mpf.make_addplot(pd.Series([20] * n, index=sk.index), panel=p, color='#26a69a', width=0.5))

    num_sub = 1 + (1 if panels['volume'] else 0) + (1 if panels['rsi'] else 0) + (1 if panels['macd'] else 0) + (1 if panels['stoch'] else 0)
    ratios = [4]
    for _ in range(num_sub - 1):
        ratios.append(1)

    fig, axes = mpf.plot(df, type='candle', style=style, addplot=apds,
                         volume=False, panel_ratios=ratios,
                         returnfig=True,
                         figsize=(16, 2 + 2 * num_sub), tight_layout=True,
                         xrotation=0,
                         title=f'{ticker} — Technical Analysis')

    ax_main = axes[0]
    idx = df.index
    for sh in sw_high[-10:]:
        mask = (df['high'] - sh).abs() < sh * 0.005
        if mask.any():
            pt = idx[mask][-1]
            ax_main.scatter(mdates.date2num(pt), sh, marker='v', color='#ef5350', s=30, zorder=5)
    for sl in sw_low[-10:]:
        mask = (df['low'] - sl).abs() < sl * 0.005
        if mask.any():
            pt = idx[mask][-1]
            ax_main.scatter(mdates.date2num(pt), sl, marker='^', color='#26a69a', s=30, zorder=5)

    tp = trading_plan
    if tp and tp.get('bias') != 'hold' and tp.get('entry_price'):
        last_date = idx[-1]
        x = mdates.date2num(last_date)

        entry = tp.get('entry_price')
        stop = tp.get('stop_loss')
        targets = tp.get('targets', [])

        ax_main.axhline(y=entry, color='#1565c0', linestyle='--', linewidth=1, alpha=0.6)
        ax_main.annotate(f'Entry {entry}', xy=(x, entry), xytext=(x + 2, entry),
                         fontsize=8, color='#1565c0', fontweight='bold')

        if stop:
            ax_main.axhline(y=stop, color='#ef5350', linestyle='--', linewidth=1, alpha=0.6)
            ax_main.annotate(f'SL {stop}', xy=(x, stop), xytext=(x + 2, stop),
                             fontsize=8, color='#ef5350', fontweight='bold')

        for t in targets:
            tp_price = t['price']
            ax_main.axhline(y=tp_price, color='#26a69a', linestyle=':', linewidth=0.8, alpha=0.5)
            ax_main.annotate(f'TP{t["level"]} {tp_price} (R:{t["rr_ratio"]})',
                             xy=(x, tp_price), xytext=(x + 2, tp_price),
                             fontsize=7, color='#26a69a')

    fig.savefig(out_path, dpi=150, bbox_inches='tight', facecolor='white')
    plt.close(fig)
