import argparse
import csv
import re
from collections import defaultdict
from datetime import datetime, timedelta
from itertools import combinations
from pathlib import Path

from dedup.data import make_pair, write_jsonl

CANDIDATE_WINDOW = timedelta(days=7)
EXPORT_SQL = "\\copy (SELECT id, incident_id, reporter_id, house_id, title, description, severity, entrance, riser, outcome, dedup_version, created_at FROM incident_report ORDER BY id) TO 'incident_report.csv' CSV HEADER"
CANDIDATE_COLUMNS = ["pair_id", "house_id", "report_a", "report_b", "incident_a", "incident_b", "created_at_a", "created_at_b", "entrance_a", "riser_a", "severity_a", "title_a", "description_a", "entrance_b", "riser_b", "severity_b", "title_b", "description_b", "kind", "weak_label", "label"]


def parse_ts(s):
    m = re.fullmatch(r"(\d{4}-\d\d-\d\d)[ T](\d\d:\d\d:\d\d)(?:\.(\d+))?([+-]\d\d)(?::?(\d\d))?", s.strip())
    if not m:
        raise ValueError(f"bad timestamp: {s!r}")
    date, time, frac, tz_h, tz_m = m.groups()
    return datetime.fromisoformat(f"{date}T{time}.{(frac or '0')[:6].ljust(6, '0')}{tz_h}:{tz_m or '00'}")


def load(path):
    rows = []
    with open(path, encoding="utf-8") as f:
        for r in csv.DictReader(f):
            r["created_at"] = parse_ts(r["created_at"]).isoformat()
            for k in ("entrance", "riser"):
                r[k] = r[k] or None
            rows.append(r)
    return rows


def pid(a, b):
    x, y = sorted([int(a["id"]), int(b["id"])])
    return f"r{x}-{y}"


def mine(rows):
    by_house = defaultdict(list)
    for r in rows:
        by_house[r["house_id"]].append(r)
    weak, candidates = [], []
    for house, rs in by_house.items():
        rs.sort(key=lambda r: r["created_at"])
        for a, b in combinations(rs, 2):
            dt = abs(datetime.fromisoformat(b["created_at"]) - datetime.fromisoformat(a["created_at"]))
            same_inc = a["incident_id"] == b["incident_id"]
            if same_inc:
                weak.append(make_pair(pid(a, b), f"real-house-{house}", house, a, b, 1, "weak_joined", a["dedup_version"], None, a["incident_id"], b["incident_id"]))
                kind = "verify_weak_joined"
            elif a["outcome"] == "created" and b["outcome"] == "created" and dt <= CANDIDATE_WINDOW:
                kind = "created_x_created"
            else:
                continue
            candidates.append({
                "pair_id": pid(a, b), "house_id": house, "report_a": a["id"], "report_b": b["id"],
                "incident_a": a["incident_id"], "incident_b": b["incident_id"],
                "created_at_a": a["created_at"], "created_at_b": b["created_at"],
                **{f"{k}_a": a[k] or "" for k in ("entrance", "riser", "severity", "title", "description")},
                **{f"{k}_b": b[k] or "" for k in ("entrance", "riser", "severity", "title", "description")},
                "kind": kind, "weak_label": 1 if same_inc else "", "label": "",
            })
    return weak, candidates


def main():
    ap = argparse.ArgumentParser(description=f"Export first: psql \"$DATABASE_URL\" -c \"{EXPORT_SQL}\"")
    ap.add_argument("--reports", default="data/raw/incident_report.csv")
    ap.add_argument("--tag", default=datetime.now().strftime("%Y%m%d"))
    args = ap.parse_args()
    weak, candidates = mine(load(args.reports))
    Path("data/labels").mkdir(parents=True, exist_ok=True)
    Path("data/labeling").mkdir(parents=True, exist_ok=True)
    write_jsonl(f"data/labels/weak_joined_{args.tag}.jsonl", weak)
    out = Path(f"data/labeling/candidates_{args.tag}.csv")
    with open(out, "w", encoding="utf-8", newline="") as f:
        w = csv.DictWriter(f, CANDIDATE_COLUMNS)
        w.writeheader()
        w.writerows(candidates)
    print(f"weak_joined={len(weak)} candidates={len(candidates)} -> {out}")


if __name__ == "__main__":
    main()
