import argparse, json, sys, os
import pandas as pd
import numpy as np

from indicators import compute_all
from signals import compute as compute_signals
from trading_plan import compute as compute_plan
from chart_renderer import render as render_chart

REQUIRED_COLS = ['date', 'open', 'high', 'low', 'close', 'volume']
MIN_BARS = 30


class _NumpyEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (np.integer,)):
            return int(obj)
        if isinstance(obj, (np.floating,)):
            return float(obj)
        if isinstance(obj, (np.bool_,)):
            return bool(obj)
        if isinstance(obj, np.ndarray):
            return obj.tolist()
        return super().default(obj)


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
            sys.stdout.write(json.dumps({'error': 'no data'}, cls=_NumpyEncoder))
            return

        df = pd.DataFrame(rows)
        err = _validate(df)
        if err:
            sys.stdout.write(json.dumps({'error': err}, cls=_NumpyEncoder))
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
        sys.stdout.write(json.dumps(output, cls=_NumpyEncoder))

    except Exception as e:
        sys.stdout.write(json.dumps({'error': f'unexpected error: {e}'}, cls=_NumpyEncoder))


if __name__ == '__main__':
    main()
