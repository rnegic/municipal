import numpy as np
from sklearn.metrics import average_precision_score, precision_recall_curve, roc_auc_score

PRECISION_TARGETS = (0.95, 0.98, 0.99)


def at_threshold(y, s, thr):
    y, s = np.asarray(y), np.asarray(s)
    pred = s >= thr
    tp = int(np.sum(pred & (y == 1)))
    fp = int(np.sum(pred & (y == 0)))
    fn = int(np.sum(~pred & (y == 1)))
    tn = int(np.sum(~pred & (y == 0)))
    p = tp / (tp + fp) if tp + fp else 1.0
    r = tp / (tp + fn) if tp + fn else 0.0
    return {
        "threshold": float(thr),
        "precision": round(p, 4),
        "recall": round(r, 4),
        "f1": round(2 * p * r / (p + r), 4) if p + r else 0.0,
        "tp": tp, "fp": fp, "fn": fn, "tn": tn,
        "fpr": round(fp / (fp + tn), 5) if fp + tn else 0.0,
        "fnr": round(fn / (fn + tp), 4) if fn + tp else 0.0,
    }


def curve_points(y, s):
    p, r, t = precision_recall_curve(y, s)
    return p[:-1], r[:-1], t


def recall_at_precision(y, s, target):
    p, r, t = curve_points(y, s)
    ok = p >= target
    if not ok.any():
        return 0.0, None
    i = int(np.argmax(np.where(ok, r, -1)))
    return float(r[i]), float(t[i])


def precision_at_recall(y, s, target):
    p, r, _ = curve_points(y, s)
    ok = r >= target
    return float(p[ok].max()) if ok.any() else 0.0


def select_thresholds(y, s):
    p, r, t = curve_points(y, s)
    f1 = np.where(p + r > 0, 2 * p * r / np.maximum(p + r, 1e-9), 0)
    out = {"f1": float(t[int(np.argmax(f1))])}
    for target in PRECISION_TARGETS:
        _, thr = recall_at_precision(y, s, target)
        out[f"p{int(target * 100)}"] = thr if thr is not None else float(np.max(s)) + 1e-6
    return out


def summary(y, s, thr):
    y, s = np.asarray(y), np.asarray(s)
    both = len(set(y.tolist())) == 2
    distinct = len(np.unique(s)) > 2
    out = {
        "n": int(len(y)),
        "positives": int(y.sum()),
        "roc_auc": round(float(roc_auc_score(y, s)), 4) if both else None,
        "pr_auc": round(float(average_precision_score(y, s)), 4) if both else None,
        "recall_at_p95": round(recall_at_precision(y, s, 0.95)[0], 4) if both else None,
        "precision_at_r80": round(precision_at_recall(y, s, 0.80), 4) if both else None,
        "binary_scorer": not distinct,
        "at_threshold": at_threshold(y, s, thr),
    }
    return out


def by_class(pairs, s, thr):
    out = {}
    for p, score in zip(pairs, s):
        key = "positive" if p["label"] else p["hn_class"]
        d = out.setdefault(key, {"n": 0, "errors": 0})
        d["n"] += 1
        d["errors"] += int((score >= thr) != bool(p["label"]))
    for d in out.values():
        d["error_rate"] = round(d["errors"] / d["n"], 4)
    return dict(sorted(out.items()))


def simulate_merges(pairs, s, thr):
    reports = {}
    score = {}
    for p, sc in zip(pairs, s):
        ka = (p["group_id"], p["created_at_a"], p["text_a"])
        kb = (p["group_id"], p["created_at_b"], p["text_b"])
        reports[ka] = p["incident_a"]
        reports[kb] = p["incident_b"]
        score[(ka, kb)] = score[(kb, ka)] = sc
    by_group = {}
    for k, inc in reports.items():
        by_group.setdefault(k[0], []).append((k, inc))
    res = {"reports": 0, "true_incidents": 0, "predicted_incidents": 0, "correct_merges": 0, "false_merges": 0, "missed_merges": 0}
    for items in by_group.values():
        items.sort(key=lambda x: x[0][1])
        clusters = []
        seen_true = set()
        for k, inc in items:
            res["reports"] += 1
            best, best_s = None, -1.0
            for c in clusters:
                sc = score.get((k, c["first"]), -1.0)
                if sc >= thr and sc > best_s:
                    best, best_s = c, sc
            is_new_true = inc not in seen_true
            seen_true.add(inc)
            if best is None:
                clusters.append({"first": k, "incident": inc})
                if not is_new_true:
                    res["missed_merges"] += 1
            elif best["incident"] == inc:
                res["correct_merges"] += 1
            else:
                res["false_merges"] += 1
        res["true_incidents"] += len(seen_true)
        res["predicted_incidents"] += len(clusters)
    return res
