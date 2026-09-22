import json
from pathlib import Path

import numpy as np
import onnxruntime as ort
from tokenizers import Tokenizer

from dedup.data import FORMATS


class OnnxDeduper:
    def __init__(self, bundle, model_file="model.int8.onnx", threads=2):
        bundle = Path(bundle)
        self.meta = json.loads((bundle / "bundle.json").read_text())
        thr = json.loads((bundle / "threshold.json").read_text())
        self.threshold = thr["threshold"]
        self.model_version = self.meta["model_version"]
        self.fmt = FORMATS[self.meta["format"]]
        self.tok = Tokenizer.from_file(str(bundle / "tokenizer.json"))
        self.tok.enable_truncation(self.meta["max_length"], strategy="longest_first")
        self.tok.enable_padding(pad_id=self.meta["pad_id"])
        opts = ort.SessionOptions()
        opts.intra_op_num_threads = threads
        opts.inter_op_num_threads = 1
        self.sess = ort.InferenceSession(str(bundle / model_file), opts, providers=["CPUExecutionProvider"])
        self.inputs = {i.name for i in self.sess.get_inputs()}

    def _logits(self, a, b):
        enc = self.tok.encode_batch(list(zip(a, b)))
        feed = {
            "input_ids": np.array([e.ids for e in enc], dtype=np.int64),
            "attention_mask": np.array([e.attention_mask for e in enc], dtype=np.int64),
            "token_type_ids": np.array([e.type_ids for e in enc], dtype=np.int64),
        }
        return self.sess.run(None, {k: v for k, v in feed.items() if k in self.inputs})[0][:, 0]

    def directional(self, reqs_a, reqs_b):
        a = [self.fmt(r) for r in reqs_a]
        b = [self.fmt(r) for r in reqs_b]
        sig = lambda z: 1 / (1 + np.exp(-z))
        return sig(self._logits(a, b)), sig(self._logits(b, a))

    def scores(self, reqs_a, reqs_b):
        ab, ba = self.directional(reqs_a, reqs_b)
        return (ab + ba) / 2

    def score(self, a, b):
        p = float(self.scores([a], [b])[0])
        return {"duplicate_probability": round(p, 6), "decision": p >= self.threshold, "model_version": self.model_version, "threshold": self.threshold}

    def match(self, request, candidates, top_k=20):
        cands = sorted(candidates, key=lambda c: c.get("created_at") or "", reverse=True)[:top_k]
        if not cands:
            return {"match": None, "model_version": self.model_version, "threshold": self.threshold, "scores": []}
        s = self.scores([request] * len(cands), cands)
        i = int(np.argmax(s))
        return {
            "match": cands[i].get("id") if s[i] >= self.threshold else None,
            "duplicate_probability": round(float(s[i]), 6),
            "model_version": self.model_version,
            "threshold": self.threshold,
            "scores": [{"id": c.get("id"), "p": round(float(x), 6)} for c, x in zip(cands, s)],
        }
