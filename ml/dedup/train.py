import argparse
import json
import random
import re
import subprocess
import time
from collections import Counter
from pathlib import Path

import numpy as np
import torch
from transformers import AutoModelForSequenceClassification, AutoTokenizer

from dedup.data import FORMATS, read_jsonl
from dedup.metrics import summary
from dedup.scorers import DEVICE, CrossEncoder


def set_seed(seed):
    random.seed(seed)
    np.random.seed(seed)
    torch.manual_seed(seed)


def sample_train(pairs, easy_neg_ratio, rng):
    pos = [p for p in pairs if p["label"] == 1]
    hard = [p for p in pairs if p["label"] == 0 and p["source"] == "hard_negative"]
    easy = [p for p in pairs if p["label"] == 0 and p["source"] != "hard_negative"]
    k = min(len(easy), int(easy_neg_ratio * len(pos)))
    return pos + hard + rng.sample(easy, k)


def paraphrase_pairs(cfg, rng):
    import csv
    from huggingface_hub import hf_hub_download

    def pp(split):
        rows = [json.loads(line) for line in open(hf_hub_download("merionum/ru_paraphraser", f"{split}.jsonl", repo_type="dataset"), encoding="utf-8")]
        return [(r["text_1"], r["text_2"], 1 if r["class"] == "1" else 0) for r in rows if r["class"] in ("1", "-1")]

    nmt = []
    for split in ("val", "test"):
        rows = list(csv.DictReader(open(hf_hub_download("cointegrated/ru-paraphrase-NMT-Leipzig", f"{split}.csv", repo_type="dataset"), encoding="utf-8")))
        pos = [(r["original"], r["ru"], 1) for r in rows if float(r["p_good"]) > cfg["nmt_min_p_good"]]
        texts = [r["ru"] for r in rows]
        neg = [(a, rng.choice(texts), 0) for a, _, _ in pos]
        nmt += pos + neg
    return pp("train") + nmt, pp("test")


KEEP_FIELDS = {"diff_entrance", "diff_riser"}


def augment(r, hn_class, rng):
    r = dict(r)
    if hn_class not in KEEP_FIELDS:
        for k in ("entrance", "riser"):
            if rng.random() < 0.3:
                r[k] = None
    d = r["description"]
    if rng.random() < 0.3:
        parts = [x.strip() for x in d.split(",") if x.strip()]
        rng.shuffle(parts)
        d = ", ".join(parts)
    if rng.random() < 0.3:
        d, r["title"] = d.lower(), r["title"].lower()
    if rng.random() < 0.2:
        d = re.sub(r"[^\w\s]", " ", d)
    r["description"] = d
    return r


def to_pair(a, b, label):
    return {"a": {"title": a, "description": ""}, "b": {"title": b, "description": ""}, "label": label, "source": "paraphrase"}


def git_commit():
    try:
        return subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    except Exception:
        return "unknown"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--config", required=True)
    args = ap.parse_args()
    cfg = json.loads(Path(args.config).read_text())
    set_seed(cfg["seed"])
    rng = random.Random(cfg["seed"])

    ds = Path("data/versions") / (cfg.get("dataset_version") or Path("data/versions/LATEST").read_text().strip())
    run = Path("runs") / cfg["name"]
    (run / "model").mkdir(parents=True, exist_ok=True)
    (run / "config").mkdir(exist_ok=True)

    if cfg.get("task") == "paraphrase":
        tr, va = paraphrase_pairs(cfg, rng)
        train = [to_pair(*t) for t in tr]
        val = [to_pair(*t) for t in va]
        fmt = lambda r: r["title"]
    else:
        train = sample_train(read_jsonl(ds / "train.jsonl.gz"), cfg["easy_neg_ratio"], rng)
        val = read_jsonl(ds / "val_syn.jsonl.gz")
        val = rng.sample(val, min(len(val), cfg["val_sample"]))
        val += read_jsonl(ds / "val_hw.jsonl.gz")
        fmt = FORMATS[cfg["format"]]
    print(f"train {len(train)} {Counter((p['label'], p['source']) for p in train)}")

    torch.set_num_threads(cfg.get("threads", 6))
    init = cfg.get("init_from") or cfg["base_model"]
    tok = AutoTokenizer.from_pretrained(init)
    model = AutoModelForSequenceClassification.from_pretrained(init, num_labels=1).to(DEVICE)
    opt = torch.optim.AdamW(model.parameters(), lr=cfg["lr"], weight_decay=0.01)
    steps = cfg["epochs"] * (len(train) // cfg["batch_size"])
    amp = bool(cfg.get("amp")) and DEVICE == "cuda"
    scaler = torch.cuda.amp.GradScaler(enabled=amp)
    sched = torch.optim.lr_scheduler.LambdaLR(opt, lambda s: min(1.0, (s + 1) / (0.06 * steps)) * max(0.0, (steps - s) / steps))

    log = []
    step = 0
    t0 = time.time()
    for epoch in range(cfg["epochs"]):
        model.train()
        rng.shuffle(train)
        for i in range(0, len(train) - cfg["batch_size"] + 1, cfg["batch_size"]):
            batch = train[i:i + cfg["batch_size"]]
            a, b = [], []
            for p in batch:
                ra, rb = p["a"], p["b"]
                if rng.random() < cfg.get("aug", 0):
                    ra, rb = augment(ra, p.get("hn_class"), rng), augment(rb, p.get("hn_class"), rng)
                x, y = fmt(ra), fmt(rb)
                if rng.random() < 0.5:
                    x, y = y, x
                a.append(x)
                b.append(y)
            x = tok(a, b, truncation=True, max_length=cfg["max_length"], padding=True, return_tensors="pt").to(DEVICE)
            y = torch.tensor([float(p["label"]) for p in batch], device=DEVICE)
            with torch.autocast("cuda", dtype=torch.float16, enabled=amp):
                logits = model(**x).logits.squeeze(-1)
            loss = torch.nn.functional.binary_cross_entropy_with_logits(logits.float(), y)
            scaler.scale(loss).backward()
            scaler.unscale_(opt)
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            scaler.step(opt)
            scaler.update()
            sched.step()
            opt.zero_grad()
            step += 1
            if step % 200 == 0:
                print(f"epoch {epoch} step {step}/{steps} loss {loss.item():.4f} {time.time() - t0:.0f}s", flush=True)
        model.eval()
        model.save_pretrained(run / "model")
        tok.save_pretrained(run / "model")
        ce = CrossEncoder(run / "model", cfg["max_length"], cfg["format"])
        if cfg.get("task") == "paraphrase":
            ce.fmt = fmt
        entry = {"epoch": epoch, "elapsed_s": round(time.time() - t0)}
        groups = {}
        for p in val:
            groups.setdefault("hw" if p["source"] == "synthetic_handwritten" else "syn" if cfg.get("task") != "paraphrase" else "paraphrase", []).append(p)
        for g, vs in sorted(groups.items()):
            entry[f"val_{g}"] = summary([p["label"] for p in vs], ce.scores(vs), 0.5)
        log.append(entry)
        print(json.dumps(log[-1], ensure_ascii=False), flush=True)

    (run / "config" / "train_config.json").write_text(json.dumps(cfg, ensure_ascii=False, indent=2))
    (run / "dataset_version.txt").write_text(ds.name + "\n")
    (run / "train_log.json").write_text(json.dumps({"git_commit": git_commit(), "train_pairs": len(train), "epochs": log}, ensure_ascii=False, indent=2))
    print("saved", run)


if __name__ == "__main__":
    main()
