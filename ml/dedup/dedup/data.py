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


def struct2_text(r):
    loc = " ".join(f"{k} {r[f]}" for k, f in (("подъезд", "entrance"), ("стояк", "riser")) if r.get(f))
    return f"{loc + '; ' if loc else ''}{r['title']}. {r['description']}"


ORDINALS = {"перв": 1, "втор": 2, "трет": 3, "четв": 4, "пят": 5, "шест": 6, "седьм": 7, "восьм": 8, "девят": 9, "десят": 10, "двушк": 2, "трёшк": 3, "четвёрк": 4}
NUM = r"(\d{1,3}|" + "|".join(ORDINALS) + r")[а-яё]*"


def _num(tok):
    return tok if tok.isdigit() else next((str(v) for k, v in ORDINALS.items() if tok.startswith(k)), None)


def _find(text, words):
    t = text.lower()
    m = re.search(rf"(?:{words})[\s№:.-]*{NUM}", t) or re.search(rf"{NUM}[\s-]*(?:{words})", t)
    return _num(m.group(1)) if m else None


def extract_location(r):
    text = f"{r['title']} {r['description']}"
    apts = re.findall(r"(?:кв\.?|квартир[аеыу]?)\s*№?\s*(\d{1,4})|\b(\d{1,4})-?(?:я|й|ю)\s+(?:кв|квартир)|я (?:из|в) (\d{1,4})", text.lower())
    return {
        "entrance": r.get("entrance") or _find(text, r"подъезд\w*|под\.|п\b|п\."),
        "riser": r.get("riser") or _find(text, r"стояк\w*"),
        "apartment": next((x for m in apts for x in m if x), None),
    }


def struct3_text(r):
    loc = extract_location(r)
    head = " ".join(f"{k} {loc[f]}" for k, f in (("подъезд", "entrance"), ("стояк", "riser"), ("кв", "apartment")) if loc[f])
    return f"{head + '; ' if head else ''}{r['title']}. {r['description']}"


FORMATS = {"text": plain_text, "struct": struct_text, "struct2": struct2_text, "struct3": struct3_text}


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
