import argparse, json, sys
import pandas as pd
from indicators import compute_all
from signals import compute as compute_signal
from trading_plan import compute as compute_plan

class _NumpyEncoder(json.JSONEncoder):
    def default(self, o):
        import numpy as np
        if isinstance(o, (np.integer,)): return int(o)
        if isinstance(o, (np.floating,)): return float(o)
        if isinstance(o, np.bool_): return bool(o)
        return super().default(o)

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--input", required=True)
    ap.add_argument("--ticker", required=True)
    args = ap.parse_args()

    with open(args.input) as f:
        raw = json.load(f)

    df = pd.DataFrame(raw)
    required = {"date", "open", "high", "low", "close", "volume"}
    missing = required - set(df.columns)
    if missing:
        print(json.dumps({"error": f"missing columns: {missing}"}), file=sys.stderr)
        sys.exit(1)

    df["date"] = pd.to_datetime(df["date"])
    df.sort_values("date", inplace=True)

    if len(df) < 30:
        print(json.dumps({"error": f"need >=30 bars, got {len(df)}"}), file=sys.stderr)
        sys.exit(1)

    ind = compute_all(df)
    sig = compute_signal(ind, args.ticker)
    plan = compute_plan(ind, sig)

    out = {
        "indicators": ind,
        "signal": sig,
        "trading_plan": plan,
    }
    print(json.dumps(out, cls=_NumpyEncoder))

if __name__ == "__main__":
    main()
