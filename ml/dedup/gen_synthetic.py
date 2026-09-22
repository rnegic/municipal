import argparse
import gzip
import json
import random
from datetime import datetime, timedelta, timezone
from itertools import combinations

from dedup.vocab import ENTRANCE_OBJECTS, OBJECTS, PLUMBING, WATER_KEYWORD_OBJECTS

GENERATION_METHOD = "template_v3"
SALIENT = {"diff_entrance": "entrance", "diff_riser": "riser", "diff_apartment": "apartment"}


class Scene:
    def __init__(self, rng, scene_id):
        self.rng = rng
        self.id = scene_id
        self.house_id = rng.randint(1000, 99999)
        self.entrances = rng.randint(2, 8)
        self.floors = rng.choice([5, 9, 9, 12, 16])
        self.t0 = datetime(2026, 1, 1, tzinfo=timezone.utc) + timedelta(minutes=rng.randint(0, 60 * 24 * 240))
        self.incidents = []
        self.relations = {}

    def risers_of(self, e):
        return [str(2 * e - 1), str(2 * e)]

    def random_location(self, scope):
        rng = self.rng
        e = rng.randint(1, self.entrances)
        riser = rng.choice(self.risers_of(e))
        floor = rng.randint(1, self.floors)
        apt = (e - 1) * self.floors * 4 + (floor - 1) * 4 + rng.randint(1, 4)
        loc = {"scope": scope, "entrance": None, "riser": None, "apartment": None, "floor": None}
        if scope in ("entrance", "riser", "apartment"):
            loc["entrance"] = e
        if scope in ("riser", "apartment"):
            loc["riser"] = riser
        if scope == "apartment":
            loc["apartment"], loc["floor"] = apt, floor
        return loc

    def overlaps(self, obj, problem, loc):
        for x in self.incidents:
            if x["object"] != obj or x["problem"] != problem:
                continue
            if "house" in (x["scope"], loc["scope"]):
                return True
            if "entrance" in (x["scope"], loc["scope"]):
                if x["entrance"] == loc["entrance"]:
                    return True
            elif "riser" in (x["scope"], loc["scope"]):
                if x["riser"] == loc["riser"]:
                    return True
            elif x["apartment"] == loc["apartment"]:
                return True
        return False

    def add(self, obj, problem, loc, t):
        if self.overlaps(obj, problem, loc):
            return None
        inc = {"id": f"{self.id}-i{len(self.incidents)}", "object": obj, "problem": problem, "t": t, **loc}
        for x in self.incidents:
            if x["object"] == obj and x["problem"] == problem:
                attr = {"entrance": "entrance", "riser": "riser", "apartment": "apartment"}.get(loc["scope"])
                if attr:
                    x.setdefault("salient", attr)
                    inc["salient"] = attr
        self.incidents.append(inc)
        return inc

    def base_incident(self):
        rng = self.rng
        obj = rng.choice(list(OBJECTS))
        spec = OBJECTS[obj]
        problem = rng.choice(list(spec["problems"]))
        scope = rng.choice(spec["scopes"])
        t = self.t0 + timedelta(minutes=rng.randint(0, 60 * 24 * 7))
        for _ in range(20):
            inc = self.add(obj, problem, self.random_location(scope), t)
            if inc:
                return inc
            obj = rng.choice(list(OBJECTS))
            problem = rng.choice(list(OBJECTS[obj]["problems"]))
            scope = rng.choice(OBJECTS[obj]["scopes"])
        return None

    def sibling(self, base):
        rng = self.rng
        options = []
        obj, loc = base["object"], {k: base[k] for k in ("scope", "entrance", "riser", "apartment", "floor")}
        if obj in ("hot_water", "cold_water"):
            other = "cold_water" if obj == "hot_water" else "hot_water"
            options.append(("hot_vs_cold", other, "none", loc))
        if base["scope"] == "riser":
            others = [r for e2 in range(1, self.entrances + 1) for r in self.risers_of(e2) if r != base["riser"]]
            r2 = rng.choice(others)
            options.append(("diff_riser", obj, base["problem"], {**loc, "riser": r2, "entrance": (int(r2) + 1) // 2}))
        if base["scope"] == "entrance" and self.entrances > 1:
            e2 = rng.choice([e for e in range(1, self.entrances + 1) if e != base["entrance"]])
            options.append(("diff_entrance", obj, base["problem"], {**loc, "entrance": e2}))
        if base["scope"] == "apartment":
            new = self.random_location("apartment")
            if new["apartment"] != base["apartment"]:
                options.append(("diff_apartment", obj, base["problem"], new))
        problems = [p for p in OBJECTS[obj]["problems"] if p != base["problem"]]
        if problems:
            options.append(("same_object_diff_cause", obj, rng.choice(problems), loc))
        if obj in ENTRANCE_OBJECTS:
            o2 = rng.choice(sorted(ENTRANCE_OBJECTS - {obj}))
            options.append(("same_location_diff_object", o2, rng.choice(list(OBJECTS[o2]["problems"])), {**loc, "scope": "entrance"}))
        if obj in WATER_KEYWORD_OBJECTS:
            o2 = rng.choice(sorted(WATER_KEYWORD_OBJECTS - {obj}))
            scope2 = OBJECTS[o2]["scopes"][0]
            loc2 = self.random_location(scope2)
            if base["entrance"] and loc2["entrance"]:
                loc2["entrance"] = base["entrance"]
            options.append(("same_keywords", o2, rng.choice(list(OBJECTS[o2]["problems"])), loc2))
        if obj in ("roof_leak", "basement_flood"):
            o2 = "basement_flood" if obj == "roof_leak" else "roof_leak"
            options.append(("roof_vs_basement", o2, "flood" if o2 == "basement_flood" else "leak", self.random_location(OBJECTS[o2]["scopes"][0])))
        if obj == "heating":
            o2 = rng.choice(sorted(PLUMBING - {"pipe_leak"}))
            p2 = rng.choice(list(OBJECTS[o2]["problems"]))
            options.append(("heating_vs_plumbing", o2, p2, {**loc, "scope": "riser" if loc["riser"] else OBJECTS[o2]["scopes"][-1]}))
        if obj in PLUMBING and obj != "pipe_leak":
            options.append(("heating_vs_plumbing", "heating", rng.choice(list(OBJECTS["heating"]["problems"])), loc if loc["scope"] != "apartment" else self.random_location("riser")))
        if obj in ("elevator", "lighting"):
            o2 = "lighting" if obj == "elevator" else "elevator"
            options.append(("elevator_vs_lighting", o2, rng.choice(list(OBJECTS[o2]["problems"])), loc))
        if not options:
            return None
        cls, o2, p2, loc2 = rng.choice(options)
        if loc2["scope"] not in OBJECTS[o2]["scopes"]:
            loc2 = {**self.random_location(OBJECTS[o2]["scopes"][0]), **({"entrance": loc2["entrance"]} if loc2.get("entrance") and "entrance" in OBJECTS[o2]["scopes"] else {})}
        t = base["t"] + timedelta(minutes=rng.randint(-60 * 24, 60 * 24))
        sib = self.add(o2, p2, loc2, t)
        if sib is None:
            return None
        self.relations[(base["id"], sib["id"])] = cls
        if cls in SALIENT:
            base["salient"] = sib["salient"] = SALIENT[cls]
        return sib


def typo(rng, s):
    if len(s) < 6:
        return s
    i = rng.randrange(1, len(s) - 1)
    op = rng.choice(["drop", "swap", "dup"])
    if op == "drop":
        return s[:i] + s[i + 1:]
    if op == "swap":
        return s[:i - 1] + s[i] + s[i - 1] + s[i + 1:]
    return s[:i] + s[i] + s[i:]


def style(rng, s):
    if rng.random() < 0.15:
        s = typo(rng, s)
    if rng.random() < 0.2:
        s = s.lower()
    if rng.random() < 0.15:
        s = s.replace("ё", "е")
    if rng.random() < 0.1:
        s = s.rstrip(".!") + rng.choice(["!!!", "!", "...", " срочно", " сделайте что-нибудь"])
    return s


def location_phrase(rng, scene, inc, reporter):
    parts = []
    salient = inc.get("salient")
    force = {k: salient == k and rng.random() < 0.85 for k in ("entrance", "riser", "apartment")}
    if inc["scope"] == "house" and rng.random() < 0.45:
        parts.append(rng.choice(["во всём доме", "по всему дому", "у всех соседей", "во всех подъездах"]))
    if reporter["entrance"] and (force["entrance"] or rng.random() < 0.45):
        parts.append(rng.choice(["в {e} подъезде", "в подъезде №{e}", "{e} подъезд", "под. {e}", "в {e}-м подъезде"]).format(e=reporter["entrance"]))
    if reporter["riser"] and (force["riser"] or (inc["object"] in PLUMBING | {"heating", "radiator_leak"} and rng.random() < 0.3)):
        parts.append(rng.choice(["по стояку {r}", "стояк {r}", "на стояке {r}"]).format(r=reporter["riser"]))
    if reporter["apartment"] and (force["apartment"] or (inc["scope"] in ("apartment", "riser") and rng.random() < 0.4)):
        parts.append(rng.choice(["кв. {a}", "в квартире {a}", "у нас в {a}", "квартира {a}"]).format(a=reporter["apartment"]))
    if inc["object"] == "roof_leak" and rng.random() < 0.4:
        parts.append(f"на {scene.floors} этаже")
    elif reporter["floor"] and inc["scope"] == "apartment" and rng.random() < 0.3:
        parts.append(f"на {reporter['floor']} этаже")
    rng.shuffle(parts)
    return ", ".join(parts)


def make_report(rng, scene, inc):
    spec = OBJECTS[inc["object"]]
    prob = spec["problems"][inc["problem"]]
    rep_loc = scene.random_location("apartment")
    reporter = {
        "entrance": inc["entrance"] or rep_loc["entrance"],
        "riser": inc["riser"] or (rep_loc["riser"] if inc["scope"] == "house" else None),
        "apartment": inc["apartment"] or rep_loc["apartment"],
        "floor": inc["floor"] or rep_loc["floor"],
    }
    if inc["scope"] == "entrance":
        reporter["riser"] = None
    loc = location_phrase(rng, scene, inc, reporter)
    desc = rng.choice(prob["descs"]).replace("{loc}", loc).replace("  ", " ").strip()
    if not loc:
        desc = desc.replace(" ,", ",").strip()
    title = style(rng, rng.choice(prob["titles"]))
    desc = style(rng, desc[0].upper() + desc[1:])
    t = inc["t"] + timedelta(minutes=rng.randint(0, 60 * 36))
    sev = spec["severity"] if rng.random() < 0.8 else rng.choice(["critical", "warning"])
    return {
        "incident": inc["id"],
        "title": title,
        "description": desc,
        "entrance": str(reporter["entrance"]) if rng.random() < (0.9 if inc.get("salient") == "entrance" else 0.6) else None,
        "riser": reporter["riser"] if reporter["riser"] and rng.random() < (0.9 if inc.get("salient") == "riser" else 0.35) else None,
        "severity": sev,
        "created_at": t.isoformat(),
    }


def build_scene(rng, idx):
    scene = Scene(rng, f"syn{idx:05d}")
    for _ in range(rng.randint(3, 6)):
        base = scene.base_incident()
        if base is None:
            continue
        for _ in range(rng.choices([0, 1, 2], [0.3, 0.5, 0.2])[0]):
            scene.sibling(base)
    reports = []
    for inc in scene.incidents:
        for _ in range(rng.choices([1, 2, 3, 4], [0.45, 0.3, 0.15, 0.1])[0]):
            reports.append(make_report(rng, scene, inc))
    return scene, reports


def plain(r):
    return f"{r['title']}. {r['description']}"


def scene_pairs(scene, reports):
    out = []
    for i, (a, b) in enumerate(combinations(reports, 2)):
        if a["created_at"] > b["created_at"]:
            a, b = b, a
        same = a["incident"] == b["incident"]
        rel = scene.relations.get((a["incident"], b["incident"])) or scene.relations.get((b["incident"], a["incident"]))
        out.append({
            "pair_id": f"{scene.id}-p{i}",
            "group_id": scene.id,
            "house_id": scene.house_id,
            "incident_a": a["incident"],
            "incident_b": b["incident"],
            "a": {k: a[k] for k in ("title", "description", "entrance", "riser", "severity", "created_at")},
            "b": {k: b[k] for k in ("title", "description", "entrance", "riser", "severity", "created_at")},
            "text_a": plain(a),
            "text_b": plain(b),
            "created_at_a": a["created_at"],
            "created_at_b": b["created_at"],
            "label": int(same),
            "source": "hard_negative" if rel else "synthetic",
            "generation_method": GENERATION_METHOD,
            "hn_class": None if same else (rel or "same_house_diff_problem"),
            "original_pair_id": None,
        })
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--scenes", type=int, default=1500)
    ap.add_argument("--seed", type=int, default=13)
    ap.add_argument("--out", default="data/raw/synthetic_template_v3.jsonl.gz")
    args = ap.parse_args()
    rng = random.Random(args.seed)
    n = 0
    with gzip.open(args.out, "wt", encoding="utf-8") as f:
        for i in range(args.scenes):
            scene, reports = build_scene(rng, i)
            for p in scene_pairs(scene, reports):
                f.write(json.dumps(p, ensure_ascii=False) + "\n")
                n += 1
    print(f"{n} pairs -> {args.out}")


if __name__ == "__main__":
    main()
