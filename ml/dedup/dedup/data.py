import gzip
import json
import re
from datetime import datetime, timedelta
from itertools import combinations

FIELDS = ("title", "description", "entrance", "riser", "severity", "created_at")


def read_jsonl(path):
    op = gzip.open if str(path).endswith(".gz") else open
    with op(path, "rt", encoding="utf-8") as f:
        return [json.loads(line) for line in f if line.strip()]


def write_jsonl(path, rows):
    op = gzip.open if str(path).endswith(".gz") else open
    with op(path, "wt", encoding="utf-8") as f:
        for r in rows:
            f.write(json.dumps(r, ensure_ascii=False) + "\n")


def plain_text(r):
    return f"{r['title']}. {r['description']}"


def struct_text(r):
    return f"подъезд: {r.get('entrance') or '?'}; стояк: {r.get('riser') or '?'}; срочность: {r.get('severity') or '?'}; {r['title']}. {r['description']}"


FORMATS = {"text": plain_text, "struct": struct_text}


def normalize(s):
    s = (s or "").lower().replace("ё", "е")
    s = re.sub(r"[^\w\s]", " ", s)
    return re.sub(r"\s+", " ", s).strip()


def make_pair(pair_id, group_id, house_id, a, b, label, source, method, hn_class=None, incident_a=None, incident_b=None):
    if a["created_at"] > b["created_at"]:
        a, b, incident_a, incident_b = b, a, incident_b, incident_a
    return {
        "pair_id": pair_id,
        "group_id": group_id,
        "house_id": house_id,
        "incident_a": incident_a,
        "incident_b": incident_b,
        "a": {k: a.get(k) for k in FIELDS},
        "b": {k: b.get(k) for k in FIELDS},
        "text_a": plain_text(a),
        "text_b": plain_text(b),
        "created_at_a": a["created_at"],
        "created_at_b": b["created_at"],
        "label": label,
        "source": source,
        "generation_method": method,
        "hn_class": None if label else (hn_class or "same_house_diff_problem"),
        "original_pair_id": None,
    }


def handwritten_pairs(path):
    d = json.load(open(path, encoding="utf-8"))
    base = datetime.fromisoformat(d["base_time"])
    keys = d["report_fields"]
    out = []
    for h_i, h in enumerate(d["houses"]):
        rel = {}
        for x, y, cls in h["relations"]:
            rel[(x, y)] = rel[(y, x)] = cls
        reports = []
        for row in h["reports"]:
            r = dict(zip(keys, row))
            r["created_at"] = (base + timedelta(days=3 * h_i, hours=r.pop("hours"))).isoformat()
            r["incident"] = f"{h['id']}-{r['incident']}"
            reports.append(r)
        for i, (a, b) in enumerate(combinations(reports, 2)):
            ia, ib = a["incident"], b["incident"]
            same = ia == ib
            cls = rel.get((ia.split("-")[1], ib.split("-")[1]))
            out.append(make_pair(f"{h['id']}-p{i}", h["id"], h["id"], a, b, int(same), d["source"], d["version"], cls, ia, ib))
    return out
