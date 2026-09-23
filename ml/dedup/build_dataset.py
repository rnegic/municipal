import argparse
import hashlib
import json
from collections import Counter, defaultdict
from pathlib import Path

from dedup.data import handwritten_pairs, read_jsonl, write_jsonl

SPLIT_RATIOS = (0.70, 0.15, 0.15)
MIN_TEXT_LEN = 8


def bucket(group_id, salt):
    h = int(hashlib.sha256(f"{salt}:{group_id}".encode()).hexdigest()[:8], 16) / 0xFFFFFFFF
    if h < SPLIT_RATIOS[0]:
        return "train"
    if h < SPLIT_RATIOS[0] + SPLIT_RATIOS[1]:
        return "val"
    return "test"


def pair_key(p):
    a = (p["text_a"], p["a"].get("entrance") or "", p["a"].get("riser") or "")
    b = (p["text_b"], p["b"].get("entrance") or "", p["b"].get("riser") or "")
    return tuple(sorted([a, b]))


def sha(path):
    return hashlib.sha256(Path(path).read_bytes()).hexdigest()


def quality(pairs):
    keys = Counter(pair_key(p) for p in pairs)
    labels = defaultdict(set)
    for p in pairs:
        labels[pair_key(p)].add(p["label"])
    ordered = Counter((p["text_a"], p["text_b"]) for p in pairs)
    reverse = sum(1 for (x, y) in ordered if x != y and (y, x) in ordered) // 2
    return {
        "pairs": len(pairs),
        "empty_text": sum(1 for p in pairs if not p["a"]["title"].strip() or not p["b"]["title"].strip() or not p["a"]["description"].strip() or not p["b"]["description"].strip()),
        "short_text": sum(1 for p in pairs if min(len(p["text_a"]), len(p["text_b"])) < MIN_TEXT_LEN),
        "identical_a_b": sum(1 for p in pairs if pair_key(p)[0] == pair_key(p)[1]),
        "identical_a_b_label0": sum(1 for p in pairs if pair_key(p)[0] == pair_key(p)[1] and p["label"] == 0),
        "duplicate_pairs": sum(c - 1 for c in keys.values() if c > 1),
        "reverse_order_pairs": reverse,
        "conflicting_keys": sum(1 for k in labels if len(labels[k]) > 1),
        "positive": sum(p["label"] for p in pairs),
        "negative": sum(1 - p["label"] for p in pairs),
        "positive_rate": round(sum(p["label"] for p in pairs) / max(len(pairs), 1), 4),
        "by_source": dict(Counter(p["source"] for p in pairs)),
        "by_hn_class": dict(Counter(p["hn_class"] for p in pairs if p["hn_class"])),
        "groups": len({p["group_id"] for p in pairs}),
        "houses": len({p["house_id"] for p in pairs}),
        "incidents": len({p[k] for p in pairs for k in ("incident_a", "incident_b") if p[k]}),
        "time_min": min((p["created_at_a"] for p in pairs), default=None),
        "time_max": max((p["created_at_b"] for p in pairs), default=None),
    }


def clean_train(pairs):
    labels = defaultdict(set)
    for p in pairs:
        labels[pair_key(p)].add(p["label"])
    seen, out, dropped_conflict, dropped_dup = set(), [], 0, 0
    for p in pairs:
        k = pair_key(p)
        if len(labels[k]) > 1:
            dropped_conflict += 1
            continue
        if k in seen:
            dropped_dup += 1
            continue
        seen.add(k)
        out.append(p)
    return out, {"dropped_conflicting": dropped_conflict, "dropped_duplicates": dropped_dup}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--synthetic", default="data/raw/synthetic_template_v3.jsonl.gz")
    ap.add_argument("--handwritten-test", default="data/raw/handwritten_eval_v3.json")
    ap.add_argument("--handwritten-val", nargs="*", default=["data/raw/handwritten_val_v1.json", "data/raw/handwritten_eval_v1.json", "data/raw/handwritten_eval_v2.json"])
    ap.add_argument("--real", nargs="*", default=sorted(str(p) for p in Path("data/labels").glob("*.jsonl")))
    ap.add_argument("--out", default="data/versions")
    ap.add_argument("--salt", default="split-v1")
    args = ap.parse_args()

    inputs = [args.synthetic, args.handwritten_test] + list(args.handwritten_val) + list(args.real)
    version = "ds-" + hashlib.sha256("".join(sha(p) for p in inputs).encode() + args.salt.encode()).hexdigest()[:10]

    splits = defaultdict(list)
    for p in read_jsonl(args.synthetic):
        splits[{"train": "train", "val": "val_syn", "test": "test_syn"}[bucket(p["group_id"], args.salt)]].append(p)
    splits["test_hw"] = handwritten_pairs(args.handwritten_test)
    splits["val_hw"] = [p for path in args.handwritten_val for p in handwritten_pairs(path)]
    real_by_id = {}
    for p in (p for path in args.real for p in read_jsonl(path)):
        if p["pair_id"] not in real_by_id or p["source"] == "manual":
            real_by_id[p["pair_id"]] = p
    real = list(real_by_id.values())
    for p in real:
        b = bucket(p["group_id"], args.salt)
        if p["source"] == "weak_joined" and b != "train":
            continue
        splits[{"train": "train", "val": "val_real", "test": "test_real"}[b]].append(p)

    raw_train = splits["train"]
    splits["train"], train_clean = clean_train(raw_train)

    groups = {name: {p["group_id"] for p in ps} for name, ps in splits.items()}
    names = list(groups)
    group_leaks = {f"{x}&{y}": len(groups[x] & groups[y]) for i, x in enumerate(names) for y in names[i + 1:] if groups[x] & groups[y]}
    train_keys = {pair_key(p) for p in splits["train"]}
    train_texts = {t for p in splits["train"] for t in (p["text_a"], p["text_b"])}
    text_overlap = {
        name: {
            "pairs_seen_in_train": sum(1 for p in ps if pair_key(p) in train_keys),
            "texts_seen_in_train": sum(1 for p in ps for t in (p["text_a"], p["text_b"]) if t in train_texts),
            "texts_total": 2 * len(ps),
        }
        for name, ps in splits.items() if name != "train"
    }

    report = {
        "dataset_version": version,
        "inputs": {p: sha(p) for p in inputs},
        "split_salt": args.salt,
        "split_ratios": SPLIT_RATIOS,
        "train_cleaning": train_clean,
        "raw_train_quality": quality(raw_train),
        "splits": {name: quality(ps) for name, ps in splits.items()},
        "group_leakage": group_leaks,
        "text_overlap_with_train": text_overlap,
        "real_pairs_total": len(real),
        "manual_pairs_total": sum(1 for p in real if p["source"] == "manual"),
        "warnings": [],
    }
    if not any(p["source"] in ("manual", "weak_joined") for ps in splits.values() for p in ps):
        report["warnings"].append("Insufficient real-world evaluation data: 0 real/manual pairs; every split is synthetic")
    for name, q in report["splits"].items():
        syn = sum(v for k, v in q["by_source"].items() if k in ("synthetic", "hard_negative", "synthetic_handwritten", "augmented"))
        q["synthetic_share"] = round(syn / max(q["pairs"], 1), 4)
    if group_leaks:
        report["warnings"].append(f"group leakage between splits: {group_leaks}")

    out = Path(args.out) / version
    out.mkdir(parents=True, exist_ok=True)
    for name, ps in splits.items():
        for p in ps:
            p["dataset_version"] = version
        write_jsonl(out / f"{name}.jsonl.gz", ps)
    report["files"] = {f.name: sha(f) for f in sorted(out.glob("*.jsonl.gz"))}
    (out / "dataset_report.json").write_text(json.dumps(report, ensure_ascii=False, indent=2))
    (out / "dataset_report.md").write_text(render_md(report))
    Path(args.out, "LATEST").write_text(version)
    print(version)
    for name, q in report["splits"].items():
        print(f"{name:10s} pairs={q['pairs']:7d} pos={q['positive']:6d} groups={q['groups']:5d} syn_share={q['synthetic_share']}")
    for w in report["warnings"]:
        print("WARNING:", w)


def render_md(r):
    lines = [f"# Dataset report `{r['dataset_version']}`", ""]
    for w in r["warnings"]:
        lines.append(f"> **WARNING:** {w}")
    lines += ["", "## Splits", "", "| split | pairs | pos | neg | pos rate | groups | houses | incidents | synthetic share | sources |", "|---|---|---|---|---|---|---|---|---|---|"]
    for name, q in r["splits"].items():
        lines.append(f"| {name} | {q['pairs']} | {q['positive']} | {q['negative']} | {q['positive_rate']} | {q['groups']} | {q['houses']} | {q['incidents']} | {q['synthetic_share']} | {q['by_source']} |")
    lines += ["", "## Hard-negative classes", "", "| split | " + " | ".join(["class", "count"]) + " |", "|---|---|---|"]
    for name, q in r["splits"].items():
        for cls, c in sorted(q["by_hn_class"].items()):
            lines.append(f"| {name} | {cls} | {c} |")
    lines += ["", "## Quality checks (raw train before cleaning)", ""]
    q = r["raw_train_quality"]
    for k in ("empty_text", "short_text", "identical_a_b", "identical_a_b_label0", "duplicate_pairs", "reverse_order_pairs", "conflicting_keys"):
        lines.append(f"- {k}: {q[k]}")
    lines.append(f"- train cleaning: {r['train_cleaning']}")
    lines += ["", "## Leakage", "", f"- group overlap between splits: {r['group_leakage'] or 'none'}", "- text overlap with train:", ""]
    for name, o in r["text_overlap_with_train"].items():
        lines.append(f"  - {name}: pairs {o['pairs_seen_in_train']}, texts {o['texts_seen_in_train']}/{o['texts_total']}")
    lines += ["", f"Real pairs: {r['real_pairs_total']}, manual: {r['manual_pairs_total']}", "", "## Inputs", ""]
    for p, h in r["inputs"].items():
        lines.append(f"- `{p}` sha256 `{h[:16]}`")
    return "\n".join(lines) + "\n"


if __name__ == "__main__":
    main()
