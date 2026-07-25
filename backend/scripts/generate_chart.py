import argparse, json, sys, os
import pandas as pd

from indicators import compute_all
from signals import compute as compute_signals
from trading_plan import compute as compute_plan
from chart_renderer import render as render_chart


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True)
    parser.add_argument('--ticker', required=True)
    parser.add_argument('--out', required=True)
    args = parser.parse_args()

    with open(args.input) as f:
        rows = json.load(f)

    if not rows:
        json.dump({'error': 'no data'}, sys.stdout)
        return

    df = pd.DataFrame(rows)

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


if __name__ == '__main__':
    main()
