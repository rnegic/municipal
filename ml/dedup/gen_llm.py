import argparse
import json
import os
import random
import re
import sys
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

API = "https://integrate.api.nvidia.com/v1/chat/completions"
FIELDS = ["incident", "hours", "entrance", "riser", "severity", "title", "description"]
HN = ["hot_vs_cold", "diff_apartment", "diff_entrance", "diff_riser", "same_object_diff_cause",
      "same_location_diff_object", "same_keywords", "roof_vs_basement", "heating_vs_plumbing",
      "elevator_vs_lighting", "same_house_diff_problem"]
OBJECTS = ["горячая вода", "холодная вода", "отопление", "канализация", "лифт", "освещение в подъезде",
           "электричество в квартире", "протечка с потолка", "подвал", "крыша", "домофон", "мусоропровод",
           "газ", "двор и уборка", "входная дверь подъезда", "вентиляция", "стояк", "окна в подъезде"]
STYLES = ["разговорный, с эмоциями и лишними подробностями", "очень короткий, 3–6 слов, без знаков препинания",
          "пожилой человек, вежливо и многословно", "раздражённый, с претензиями к УК", "с опечатками и сокращениями",
          "деловой, по существу", "сбивчивый, путает детали, но суть понятна", "молодёжный, строчными буквами"]
EVAL_FILES = ["handwritten_eval_v1.json", "handwritten_eval_v2.json", "handwritten_eval_v3.json",
              "handwritten_val_v1.json", "handwritten_eval_v4_ai1.json", "handwritten_eval_v4_ai2.json"]

RULE = """Правило «одна авария»: одна авария — это одна причина, которую устраняет одна бригада за один выезд.
- Разные симптомы одной поломки в одном месте и в одно время — одна авария («горячая еле тёплая» и «горячей нет» в одном стояке в одно утро).
- Две независимые поломки, даже в одной квартире, — разные аварии (течь стояка и отдельно подтекающий сифон).
- Авария масштаба дома (отключили горячую воду, нет света во всём доме) — одна авария, даже если пишут из разных подъездов.
- Одинаковая поломка в разных местах (лифт в 1-м и в 3-м подъезде, течь в кв. 12 и в кв. 58 разных стояков) — разные аварии."""

GEN = """Ты пишешь реалистичные заявки жильцов многоквартирного дома в управляющую компанию для обучения модели поиска дублей.

{rule}

Напиши {n} домов. В каждом доме 3–6 аварий (инциденты a, b, c, …), на каждую аварию 1–4 заявки от разных жильцов.
Обязательно делай трудные случаи: аварии-«близнецы», которые отличаются одним аспектом (горячая/холодная, другой подъезд,
другой стояк, другая квартира, та же вещь но другая причина, те же слова но другая проблема), и дубли, написанные
совсем разными словами (разные симптомы одной поломки, разный уровень подробностей).
Темы для этих домов: {objects}. Стиль заявок смешивай, в том числе: {styles}.
Каждая заявка — отдельный живой человек: свои слова, своя длина, свои подробности. Описание не повторяет заголовок
и не начинается с него. Дубли не должны быть перефразом друг друга слово в слово: один пишет про симптом, другой про
последствия, третий коротко. Не используй английские слова.
Поля entrance (подъезд) и riser (стояк) — строки с номером или null; жильцы часто их не заполняют.
severity — "critical", "warning" или "info". hours — через сколько часов от начала недели подана заявка (число 0–160),
заявки одной аварии обычно близко по времени.
В relations для каждой пары аварий-«близнецов» укажи класс отличия из списка: {hn}.

Ответь только JSON без пояснений:
{{"houses": [{{"relations": [["a", "b", "diff_entrance"]],
  "reports": [["a", 0.5, "2", null, "critical", "заголовок", "текст заявки"], ...]}}]}}
Порядок полей в reports: incident, hours, entrance, riser, severity, title, description."""

VERIFY = """Ниже заявки жильцов одного дома в управляющую компанию. Сгруппируй их по авариям.

{rule}

Заявки:
{reports}

Ответь только JSON: {{"groups": [[номера заявок одной аварии], ...]}}. Каждая заявка ровно в одной группе."""


def call(model, prompt, key, max_tokens, temperature, retries=6, **extra):
    body = json.dumps({"model": model, "max_tokens": max_tokens, "temperature": temperature,
                       "messages": [{"role": "user", "content": prompt}], **extra}).encode()
    for attempt in range(retries):
        req = urllib.request.Request(API, body, {"Authorization": f"Bearer {key}", "Content-Type": "application/json"})
        try:
            with urllib.request.urlopen(req, timeout=600) as r:
                content = json.load(r)["choices"][0]["message"].get("content")
            if content:
                return content
            raise ValueError("empty content")
        except (urllib.error.URLError, TimeoutError, KeyError, ValueError) as e:
            if attempt == retries - 1:
                raise
            print(f"  {model}: {e}, retry", file=sys.stderr)
            slow = isinstance(e, urllib.error.HTTPError) and e.code == 429
            time.sleep((60 if slow else 15) * (attempt + 1))


def parse_json(text):
    text = re.sub(r"<think>.*?</think>", "", text, flags=re.S)
    start, end = text.find("{"), text.rfind("}")
    return json.loads(text[start:end + 1])


def norm(s):
    return re.sub(r"[^а-яa-z0-9]+", " ", (s or "").lower().replace("ё", "е")).strip()


def valid_house(h):
    reps = h.get("reports")
    if not isinstance(reps, list) or len(reps) < 3:
        return False
    for r in reps:
        if not (isinstance(r, list) and len(r) == 7 and isinstance(r[0], str) and isinstance(r[1], (int, float))
                and r[4] in ("critical", "warning", "info") and isinstance(r[5], str) and isinstance(r[6], str)):
            return False
        if re.search(r"[؀-ۿ一-鿿぀-ヿ]|[a-zA-Z]{4,}", r[5] + r[6]) or not re.search(r"[а-яА-Я]", r[6]):
            return False
        if len(norm(r[5])) > 12 and norm(r[6]).startswith(norm(r[5])):
            return False
    incidents = {r[0] for r in reps}
    return len(incidents) >= 2 and all(isinstance(x, list) and len(x) == 3 and x[2] in HN for x in h.get("relations", []))


def clean_house(h):
    for r in h["reports"]:
        r[2] = None if r[2] in (None, "", "null") else str(r[2])
        r[3] = None if r[3] in (None, "", "null") else str(r[3])
    ids = {r[0] for r in h["reports"]}
    h["relations"] = [x for x in h.get("relations", []) if x[0] in ids and x[1] in ids and x[0] != x[1]]
    return h


def partition(labels):
    groups = {}
    for i, lab in enumerate(labels):
        groups.setdefault(lab, set()).add(i)
    return {frozenset(g) for g in groups.values()}


def verify(h, model, key, rng):
    order = list(range(len(h["reports"])))
    rng.shuffle(order)
    lines = []
    for n, i in enumerate(order, 1):
        r = h["reports"][i]
        lines.append(f"{n}. подъезд: {r[2] or '—'}, стояк: {r[3] or '—'}, через {r[1]} ч. «{r[5]}» — {r[6]}")
    out = parse_json(call(model, VERIFY.format(rule=RULE, reports="\n".join(lines)), key, 8000, 0.0, reasoning_effort="low"))
    got = [None] * len(order)
    for g_i, g in enumerate(out["groups"]):
        for n in g:
            if 1 <= int(n) <= len(order):
                got[order[int(n) - 1]] = g_i
    return None if None in got else got


def salvage(reports, got):
    alive = list(range(len(reports)))
    def bad(i, j):
        return (reports[i][0] == reports[j][0]) != (got[i] == got[j])
    while True:
        counts = {i: sum(bad(i, j) for j in alive if j != i) for i in alive}
        worst = max(alive, key=lambda i: counts[i])
        if counts[worst] == 0:
            return alive
        alive.remove(worst)


def job(args, key, call_i, banned):
    rng = random.Random(f"{args.seed}-{args.model}-{call_i}")
    prompt = GEN.format(rule=RULE, n=args.per_call, objects=", ".join(rng.sample(OBJECTS, 5)),
                        styles="; ".join(rng.sample(STYLES, 3)), hn=", ".join(HN))
    for attempt in range(3):
        try:
            houses = parse_json(call(args.model, prompt, key, 12000, 0.9, reasoning_effort="low"))["houses"]
            break
        except (json.JSONDecodeError, KeyError, TypeError) as e:
            if attempt == 2:
                raise
            print(f"  {args.model}: bad json {e}, retry", file=sys.stderr)
    stats = {"generated": len(houses), "invalid": 0, "contaminated": 0, "disagree": 0, "dropped_reports": 0}
    kept = []
    for h in houses:
        if not valid_house(h):
            stats["invalid"] += 1
            continue
        h = clean_house(h)
        if any(norm(r[6]) in banned or norm(r[5]) + "|" + norm(r[6]) in banned for r in h["reports"]):
            stats["contaminated"] += 1
            continue
        if args.verifier:
            got = verify(h, args.verifier, key, rng)
            if got is None:
                stats["disagree"] += 1
                continue
            alive = salvage(h["reports"], got)
            dropped = [h["reports"][i] for i in range(len(h["reports"])) if i not in alive]
            if dropped:
                with open(args.out + ".disputes.jsonl", "a", encoding="utf-8") as f:
                    f.write(json.dumps({"house": h, "verifier_groups": got, "dropped": dropped}, ensure_ascii=False) + "\n")
            stats["dropped_reports"] += len(dropped)
            h["reports"] = [h["reports"][i] for i in alive]
            if len(h["reports"]) < 3 or len({r[0] for r in h["reports"]}) < 2:
                stats["disagree"] += 1
                continue
            h = clean_house(h)
        kept.append({"relations": h["relations"], "reports": h["reports"]})
    return kept, stats


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", required=True)
    ap.add_argument("--verifier")
    ap.add_argument("--houses", type=int, default=50)
    ap.add_argument("--per-call", type=int, default=2)
    ap.add_argument("--workers", type=int, default=4)
    ap.add_argument("--prefix", required=True)
    ap.add_argument("--version", default="handwritten_train_v2")
    ap.add_argument("--out", required=True)
    ap.add_argument("--seed", default="1")
    args = ap.parse_args()
    key = os.environ["NVIDIA_API_KEY"]

    banned = set()
    for f in EVAL_FILES:
        for h in json.load(open(Path("data/raw") / f, encoding="utf-8"))["houses"]:
            for r in h["reports"]:
                banned.add(norm(r[6]))
                banned.add(norm(r[5]) + "|" + norm(r[6]))
    banned.discard("")

    out = Path(args.out)
    doc = json.load(open(out, encoding="utf-8")) if out.exists() else {
        "version": args.version, "source": "synthetic_handwritten", "base_time": "2028-01-01T06:00:00+00:00",
        "author_model": args.model, "verifier_model": args.verifier, "report_fields": FIELDS, "houses": []}
    total = {"generated": 0, "invalid": 0, "contaminated": 0, "disagree": 0, "dropped_reports": 0, "failed_calls": 0}
    call_i = len(doc["houses"]) * 10
    with ThreadPoolExecutor(args.workers) as ex:
        pending = set()
        while len(doc["houses"]) < args.houses or pending:
            while len(doc["houses"]) + len(pending) * args.per_call < args.houses and len(pending) < args.workers:
                pending.add(ex.submit(job, args, key, call_i, banned))
                call_i += 1
            done = next(as_completed(pending))
            pending.discard(done)
            try:
                kept, stats = done.result()
            except Exception as e:
                total["failed_calls"] += 1
                print(f"call failed: {e}", file=sys.stderr)
                if total["failed_calls"] > 3 * args.houses // args.per_call + 5:
                    break
                continue
            for k, v in stats.items():
                total[k] += v
            for h in kept[:max(0, args.houses - len(doc["houses"]))]:
                h["id"] = f"{args.prefix}{len(doc['houses']) + 1:03d}"
                doc["houses"].append({"id": h["id"], "relations": h["relations"], "reports": h["reports"]})
            tmp = out.with_suffix(".tmp")
            tmp.write_text(json.dumps(doc, ensure_ascii=False, indent=1), encoding="utf-8")
            os.replace(tmp, out)
            print(f"{len(doc['houses'])}/{args.houses} {total}", flush=True)
    print(json.dumps(total, ensure_ascii=False))


if __name__ == "__main__":
    main()
