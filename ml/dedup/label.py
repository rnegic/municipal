import argparse
import csv
from pathlib import Path

from dedup.data import make_pair, write_jsonl

from mine_pairs import CANDIDATE_COLUMNS

PROMPT = "Одна и та же конкретная авария? [1] да  [0] нет  [s] пропустить  [q] выход: "


def side(r, s):
    return {"title": r[f"title_{s}"], "description": r[f"description_{s}"], "entrance": r[f"entrance_{s}"] or None, "riser": r[f"riser_{s}"] or None, "severity": r[f"severity_{s}"], "created_at": r[f"created_at_{s}"]}


def interactive(path):
    rows = list(csv.DictReader(open(path, encoding="utf-8")))
    todo = [r for r in rows if r["label"] == ""]
    print(f"{len(todo)} unlabeled of {len(rows)}")
    for n, r in enumerate(todo, 1):
        print(f"\n[{n}/{len(todo)}] house {r['house_id']} ({r['kind']})")
        for s in ("a", "b"):
            print(f"  {s.upper()}: {r[f'created_at_{s}'][:16]} подъезд={r[f'entrance_{s}'] or '?'} стояк={r[f'riser_{s}'] or '?'}\n     {r[f'title_{s}']} — {r[f'description_{s}']}")
        ans = ""
        while ans not in ("0", "1", "s", "q"):
            ans = input(PROMPT).strip().lower()
        if ans == "q":
            break
        if ans in ("0", "1"):
            r["label"] = ans
            with open(path, "w", encoding="utf-8", newline="") as f:
                w = csv.DictWriter(f, CANDIDATE_COLUMNS)
                w.writeheader()
                w.writerows(rows)


def export(path, out):
    rows = [r for r in csv.DictReader(open(path, encoding="utf-8")) if r["label"] in ("0", "1")]
    pairs = [make_pair(r["pair_id"], f"real-house-{r['house_id']}", r["house_id"], side(r, "a"), side(r, "b"), int(r["label"]), "manual", "human_label", None, r["incident_a"], r["incident_b"]) for r in rows]
    write_jsonl(out, pairs)
    print(f"{len(pairs)} manual pairs -> {out}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("csv")
    ap.add_argument("--export", help="write labeled rows as data/labels/manual_*.jsonl")
    args = ap.parse_args()
    if args.export:
        export(args.csv, args.export)
    else:
        interactive(args.csv)


if __name__ == "__main__":
    main()
