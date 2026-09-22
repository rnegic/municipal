import argparse
import json
import random
import subprocess
import time
from collections import Counter
from pathlib import Path

import numpy as np
import torch
from transformers import AutoModelForSequenceClassification, AutoTokenizer

from dedup.data import FORMATS, read_jsonl
from dedup.metrics import summary
from dedup.scorers import CrossEncoder


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

    train = sample_train(read_jsonl(ds / "train.jsonl.gz"), cfg["easy_neg_ratio"], rng)
    val = read_jsonl(ds / "val_syn.jsonl.gz")
    val = rng.sample(val, min(len(val), cfg["val_sample"]))
    fmt = FORMATS[cfg["format"]]
    print(f"train {len(train)} {Counter((p['label'], p['source']) for p in train)}")

    torch.set_num_threads(cfg.get("threads", 6))
    tok = AutoTokenizer.from_pretrained(cfg["base_model"])
    model = AutoModelForSequenceClassification.from_pretrained(cfg["base_model"], num_labels=1)
    opt = torch.optim.AdamW(model.parameters(), lr=cfg["lr"], weight_decay=0.01)
    steps = cfg["epochs"] * (len(train) // cfg["batch_size"])
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
                x, y = fmt(p["a"]), fmt(p["b"])
                if rng.random() < 0.5:
                    x, y = y, x
                a.append(x)
                b.append(y)
            x = tok(a, b, truncation=True, max_length=cfg["max_length"], padding=True, return_tensors="pt")
            y = torch.tensor([float(p["label"]) for p in batch])
            loss = torch.nn.functional.binary_cross_entropy_with_logits(model(**x).logits.squeeze(-1), y)
            loss.backward()
            torch.nn.utils.clip_grad_norm_(model.parameters(), 1.0)
            opt.step()
            sched.step()
            opt.zero_grad()
            step += 1
            if step % 200 == 0:
                print(f"epoch {epoch} step {step}/{steps} loss {loss.item():.4f} {time.time() - t0:.0f}s", flush=True)
        model.eval()
        model.save_pretrained(run / "model")
        tok.save_pretrained(run / "model")
        ce = CrossEncoder(run / "model", cfg["max_length"], cfg["format"])
        s = ce.scores(val)
        m = summary([p["label"] for p in val], s, 0.5)
        log.append({"epoch": epoch, "val_syn_sample": m, "elapsed_s": round(time.time() - t0)})
        print(json.dumps(log[-1], ensure_ascii=False), flush=True)

    (run / "config" / "train_config.json").write_text(json.dumps(cfg, ensure_ascii=False, indent=2))
    (run / "dataset_version.txt").write_text(ds.name + "\n")
    (run / "train_log.json").write_text(json.dumps({"git_commit": git_commit(), "train_pairs": len(train), "epochs": log}, ensure_ascii=False, indent=2))
    print("saved", run)


if __name__ == "__main__":
    main()
