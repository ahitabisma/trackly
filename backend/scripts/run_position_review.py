import argparse, json, sys
import pandas as pd
import numpy as np

from indicators import compute_all
from signals import compute as compute_signals
from trading_plan import compute_position_review


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


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--ohlcv', required=True)
    parser.add_argument('--ticker', required=True)
    parser.add_argument('--buy-price', type=float, required=True)
    parser.add_argument('--lot', type=float, required=True)
    parser.add_argument('--buy-date', required=True)
    args = parser.parse_args()

    with open(args.ohlcv) as f:
        rows = json.load(f)

    df = pd.DataFrame(rows)
    ind = compute_all(df)
    signal = compute_signals(ind, args.ticker)

    result = compute_position_review(ind, signal, args.buy_price, args.lot, args.buy_date)
    sys.stdout.write(json.dumps(result, cls=_NumpyEncoder))


if __name__ == '__main__':
    main()
