import argparse
import hashlib
import json
import shutil
import subprocess
import time
from pathlib import Path

import numpy as np
import torch
from onnxruntime.quantization import QuantType, quantize_dynamic
from transformers import AutoModelForSequenceClassification, AutoTokenizer

from dedup.data import read_jsonl
from dedup.onnx_infer import OnnxDeduper
from dedup.scorers import CrossEncoder

TOLERANCE_FP32 = 1e-4
TOLERANCE_INT8 = 0.05


def sha(p):
    return hashlib.sha256(Path(p).read_bytes()).hexdigest()


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--run", required=True)
    args = ap.parse_args()
    run = Path(args.run)
    cfg = json.loads((run / "config" / "train_config.json").read_text())
    ds = (run / "dataset_version.txt").read_text().strip()
    bundle = run / "onnx"
    bundle.mkdir(exist_ok=True)

    tok = AutoTokenizer.from_pretrained(run / "model")
    model = AutoModelForSequenceClassification.from_pretrained(run / "model").eval()
    x = tok(["нет воды"], ["воды нет"], return_tensors="pt")
    torch.onnx.export(
        model, (x["input_ids"], x["attention_mask"], x["token_type_ids"]), bundle / "model.onnx",
        input_names=["input_ids", "attention_mask", "token_type_ids"], output_names=["logits"],
        dynamic_axes={k: {0: "batch", 1: "seq"} for k in ("input_ids", "attention_mask", "token_type_ids")} | {"logits": {0: "batch"}},
        opset_version=14,
    )
    quantize_dynamic(bundle / "model.onnx", bundle / "model.int8.onnx", weight_type=QuantType.QInt8)
    tok.backend_tokenizer.save(str(bundle / "tokenizer.json"))
    shutil.copy(run / "threshold.json", bundle / "threshold.json")
    commit = subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], text=True).strip()
    meta = {
        "model_version": f"{cfg['name']}@{ds}@{commit}",
        "base_model": cfg["base_model"],
        "format": cfg["format"],
        "max_length": cfg["max_length"],
        "pad_id": tok.pad_token_id,
        "dataset_version": ds,
        "git_commit": commit,
        "score": "mean(sigmoid(logit(a,b)), sigmoid(logit(b,a)))",
    }
    (bundle / "bundle.json").write_text(json.dumps(meta, ensure_ascii=False, indent=2))

    pairs = read_jsonl(Path("data/versions") / ds / "test_hw.jsonl.gz") + read_jsonl(Path("data/versions") / ds / "val_hw.jsonl.gz")
    ref = CrossEncoder(run / "model", cfg["max_length"], cfg["format"]).scores(pairs)
    thr = json.loads((run / "threshold.json").read_text())["threshold"]
    check = {"pairs": len(pairs)}
    for f in ("model.onnx", "model.int8.onnx"):
        d = OnnxDeduper(bundle, f, threads=2)
        s = d.scores([p["a"] for p in pairs], [p["b"] for p in pairs])
        diff = np.abs(s - ref)
        t = time.time()
        for p in pairs[:100]:
            d.scores([p["a"]], [p["b"]])
        single_ms = (time.time() - t) / 100 * 1000
        t = time.time()
        for _ in range(5):
            d.scores([pairs[0]["a"]] * 20, [p["b"] for p in pairs[:20]])
        batch20_ms = (time.time() - t) / 5 * 1000
        tol = TOLERANCE_FP32 if f == "model.onnx" else TOLERANCE_INT8
        check[f] = {
            "sha256": sha(bundle / f),
            "size_mb": round((bundle / f).stat().st_size / 1e6, 1),
            "max_abs_diff_vs_torch": round(float(diff.max()), 6),
            "mean_abs_diff_vs_torch": round(float(diff.mean()), 6),
            "decision_disagreements": int(np.sum((s >= thr) != (ref >= thr))),
            "within_tolerance": bool(diff.max() <= tol),
            "tolerance": tol,
            "latency_ms_1pair_2threads": round(single_ms, 2),
            "latency_ms_20candidates_2threads": round(batch20_ms, 2),
        }
    check["tokenizer_sha256"] = sha(bundle / "tokenizer.json")
    (bundle / "export_check.json").write_text(json.dumps(check, ensure_ascii=False, indent=2))
    print(json.dumps(check, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
