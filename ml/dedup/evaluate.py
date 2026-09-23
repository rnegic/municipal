import argparse
import json
import random
import time
from pathlib import Path

import numpy as np
import torch

from dedup.data import read_jsonl
from dedup.metrics import by_class, select_thresholds, simulate_merges, summary
from dedup.scorers import CrossEncoder, Embedder, exact_scores, normalized_exact_scores

PRIMARY_POLICY = "p95"
SYN_GROUPS = 60
RUNS = {}


def group_sample(pairs, n, seed):
    groups = sorted({p["group_id"] for p in pairs})
    keep = set(random.Random(seed).sample(groups, min(n, len(groups))))
    return [p for p in pairs if p["group_id"] in keep]


def load_splits(ds, seed):
    return {
        "val_syn": group_sample(read_jsonl(ds / "val_syn.jsonl.gz"), SYN_GROUPS, seed),
        "val_hw": read_jsonl(ds / "val_hw.jsonl.gz"),
        "test_syn": group_sample(read_jsonl(ds / "test_syn.jsonl.gz"), SYN_GROUPS, seed),
        "test_hw": read_jsonl(ds / "test_hw.jsonl.gz"),
        **({"test_real": read_jsonl(ds / "test_real.jsonl.gz")} if (ds / "test_real.jsonl.gz").exists() else {}),
    }


def scorers(models, no_baselines=False):
    out = {
        "exact-match": (lambda ps: exact_scores(ps), None),
        "normalized-exact": (lambda ps: normalized_exact_scores(ps), None),
    }
    if no_baselines:
        return {**out, **model_scorers(models)}
    emb_tiny = Embedder("cointegrated/rubert-tiny2", "cls")
    out["embedding-rubert-tiny2"] = (lambda ps: emb_tiny.scores(ps), None)
    emb_ml = Embedder("sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2", "mean")
    out["embedding-multilingual-minilm"] = (lambda ps: emb_ml.scores(ps), None)
    base = CrossEncoder("cross-encoder/mmarco-mMiniLMv2-L12-H384-v1", 128, "text")
    out["base-cross-encoder-mmarco"] = (base.scores, base)
    return {**out, **model_scorers(models)}


def model_scorers(models):
    out = {}
    for m in models:
        cfg = json.loads((Path(m) / "config" / "train_config.json").read_text())
        ce = CrossEncoder(Path(m) / "model", cfg["max_length"], cfg["format"])
        out[cfg["name"]] = (ce.scores, ce)
        RUNS[cfg["name"]] = Path(m)
    return out


def errors(pairs, s, thr, k=15):
    rows = [(float(sc), p) for sc, p in zip(s, pairs) if (sc >= thr) != bool(p["label"])]
    fps = sorted([r for r in rows if not r[1]["label"]], key=lambda r: -r[0])[:k]
    fns = sorted([r for r in rows if r[1]["label"]], key=lambda r: r[0])[:k]
    fmt = lambda r: {"score": round(r[0], 4), "hn_class": r[1]["hn_class"], "a": f"[{r[1]['a']['entrance']}/{r[1]['a']['riser']}] {r[1]['text_a']}", "b": f"[{r[1]['b']['entrance']}/{r[1]['b']['riser']}] {r[1]['text_b']}"}
    return {"false_positives": [fmt(r) for r in fps], "false_negatives": [fmt(r) for r in fns]}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dataset", default=None)
    ap.add_argument("--models", nargs="*", default=[])
    ap.add_argument("--out", default="runs/eval")
    ap.add_argument("--seed", type=int, default=7)
    ap.add_argument("--threshold-split", default="val_syn", choices=["val_syn", "val_hw"])
    ap.add_argument("--val-only", action="store_true")
    ap.add_argument("--no-baselines", action="store_true")
    args = ap.parse_args()
    torch.set_num_threads(12)
    ds = Path("data/versions") / (args.dataset or Path("data/versions/LATEST").read_text().strip())
    splits = load_splits(ds, args.seed)
    if args.val_only:
        splits = {k: v for k, v in splits.items() if k.startswith("val")}
    out = Path(args.out)
    (out / "scores").mkdir(parents=True, exist_ok=True)
    result = {"dataset_version": ds.name, "primary_policy": PRIMARY_POLICY, "split_sizes": {k: len(v) for k, v in splits.items()}, "models": {}}

    for name, (fn, ce) in scorers(args.models, args.no_baselines).items():
        t = time.time()
        S = {}
        for split, ps in splits.items():
            cache = out / "scores" / f"{name}__{split}.npy"
            S[split] = np.load(cache) if cache.exists() else fn(ps)
            np.save(cache, S[split])
        binary = len(np.unique(np.concatenate(list(S.values())))) <= 2
        thr_syn = {"binary": 0.5} if binary else select_thresholds([p["label"] for p in splits["val_syn"]], S["val_syn"])
        thr_hw = {"binary": 0.5} if binary else select_thresholds([p["label"] for p in splits["val_hw"]], S["val_hw"])
        thr = 0.5 if binary else (thr_syn if args.threshold_split == "val_syn" else thr_hw)[PRIMARY_POLICY]
        m = {"thresholds_val_syn": thr_syn, "thresholds_val_hw": thr_hw, "threshold": thr, "splits": {}}
        for split, ps in splits.items():
            y = [p["label"] for p in ps]
            m["splits"][split] = {
                "at_primary": summary(y, S[split], thr),
                "by_class": by_class(ps, S[split], thr),
                "merges": simulate_merges(ps, S[split], thr),
            }
            if not binary and split.startswith("test"):
                m["splits"][split]["at_val_hw_p95"] = summary(y, S[split], thr_hw["p95"])["at_threshold"]
                m["splits"][split]["at_val_syn_f1"] = summary(y, S[split], thr_syn["f1"])["at_threshold"]
        if ce is not None and "test_hw" in splits:
            sym = {}
            for split in ("test_hw", "test_syn"):
                ab, ba = ce.directional(splits[split])
                d = np.abs(ab - ba)
                sym[split] = {"max_abs_diff": round(float(d.max()), 4), "mean_abs_diff": round(float(d.mean()), 5), "decision_flips_at_threshold": int(np.sum((ab >= thr) != (ba >= thr)))}
            m["symmetry_raw_directional"] = sym
        if "test_hw" in splits:
            m["errors_test_hw"] = errors(splits["test_hw"], S["test_hw"], thr)
        m["elapsed_s"] = round(time.time() - t, 1)
        result["models"][name] = m
        if name in RUNS and not args.val_only:
            (RUNS[name] / "metrics.json").write_text(json.dumps(m, ensure_ascii=False, indent=2))
            (RUNS[name] / "threshold.json").write_text(json.dumps({"threshold": thr, "policy": "max recall s.t. precision >= 0.95", "selected_on": args.threshold_split, "dataset_version": ds.name, "alternatives_val_syn": thr_syn, "alternatives_val_hw": thr_hw}, indent=2))
        line = [f"{name:34s} thr={thr:.4f}"]
        for sp in m["splits"]:
            a = m["splits"][sp]["at_primary"]
            t = a["at_threshold"]
            line.append(f"{sp} AUC={a['roc_auc']} PR={a['pr_auc']} R@P95={a['recall_at_p95']} P={t['precision']} R={t['recall']} FP={t['fp']} FN={t['fn']}")
        print(" | ".join(line), flush=True)

    (out / "metrics.json").write_text(json.dumps(result, ensure_ascii=False, indent=2))
    print("saved", out / "metrics.json")


if __name__ == "__main__":
    main()
