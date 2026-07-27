import argparse, json, sys, os
import pandas as pd

from indicators import compute_all
from signals import compute as compute_signals
from trading_plan import compute as compute_plan
from chart_renderer import render as render_chart

REQUIRED_COLS = ['date', 'open', 'high', 'low', 'close', 'volume']
MIN_BARS = 30


def _validate(df):
    missing = [c for c in REQUIRED_COLS if c not in df.columns]
    if missing:
        return f'missing required columns: {", ".join(missing)}'
    if len(df) < MIN_BARS:
        return f'minimum {MIN_BARS} data bars required, got {len(df)}'
    dates = pd.to_datetime(df['date'], errors='coerce')
    if dates.isna().any():
        return 'column "date" contains invalid or unparseable values'
    if not dates.is_monotonic_increasing:
        return 'dates must be in ascending order'
    return None


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True)
    parser.add_argument('--ticker', required=True)
    parser.add_argument('--out', required=True)
    args = parser.parse_args()

    try:
        with open(args.input) as f:
            rows = json.load(f)

        if not rows:
            json.dump({'error': 'no data'}, sys.stdout)
            return

        df = pd.DataFrame(rows)
        err = _validate(df)
        if err:
            json.dump({'error': err}, sys.stdout)
            return

        ind = compute_all(df)
        signal = compute_signals(ind, args.ticker)
        plan = compute_plan(ind, signal)

        render_chart(df, ind, signal, plan, args.out, args.ticker)

        output = {
            'indicators': {k: v for k, v in ind.items() if not k.startswith('_')},
            'signal': signal,
            'trading_plan': plan,
        }
        json.dump(output, sys.stdout)

    except Exception as e:
        json.dump({'error': f'unexpected error: {e}'}, sys.stdout)


if __name__ == '__main__':
    main()
