from datetime import datetime, timedelta

import numpy as np
import torch
from transformers import AutoModel, AutoModelForSequenceClassification, AutoTokenizer

from dedup.data import FORMATS, normalize

PROD_WINDOW = timedelta(hours=2)
BATCH = 64
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"


def _within_window(p):
    return datetime.fromisoformat(p["created_at_b"]) - datetime.fromisoformat(p["created_at_a"]) <= PROD_WINDOW


def exact_scores(pairs):
    return np.array([float(p["a"]["title"] == p["b"]["title"] and p["a"]["riser"] == p["b"]["riser"] and _within_window(p)) for p in pairs])


def normalized_exact_scores(pairs):
    return np.array([float(normalize(p["a"]["title"]) == normalize(p["b"]["title"]) and normalize(p["a"]["riser"]) == normalize(p["b"]["riser"]) and _within_window(p)) for p in pairs])


class Embedder:
    def __init__(self, name, pooling):
        self.tok = AutoTokenizer.from_pretrained(name)
        self.model = AutoModel.from_pretrained(name).to(DEVICE).eval()
        self.pooling = pooling

    @torch.no_grad()
    def encode(self, texts):
        out = []
        for i in range(0, len(texts), BATCH):
            x = self.tok(texts[i:i + BATCH], truncation=True, max_length=128, padding=True, return_tensors="pt").to(DEVICE)
            h = self.model(**x).last_hidden_state
            if self.pooling == "cls":
                v = h[:, 0]
            else:
                m = x["attention_mask"].unsqueeze(-1).float()
                v = (h * m).sum(1) / m.sum(1)
            out.append(torch.nn.functional.normalize(v, dim=-1).cpu().numpy())
        return np.concatenate(out)

    def scores(self, pairs, fmt="text"):
        f = FORMATS[fmt]
        texts = sorted({f(p[k]) for p in pairs for k in ("a", "b")})
        idx = {t: i for i, t in enumerate(texts)}
        emb = self.encode(texts)
        return np.array([float(emb[idx[f(p["a"])]] @ emb[idx[f(p["b"])]]) for p in pairs])


class CrossEncoder:
    def __init__(self, path, max_length=128, fmt="text"):
        self.tok = AutoTokenizer.from_pretrained(path)
        self.model = AutoModelForSequenceClassification.from_pretrained(path).to(DEVICE).eval()
        self.max_length = max_length
        self.fmt = FORMATS[fmt]

    @torch.no_grad()
    def logits(self, a, b):
        out = []
        for i in range(0, len(a), BATCH):
            x = self.tok(a[i:i + BATCH], b[i:i + BATCH], truncation=True, max_length=self.max_length, padding=True, return_tensors="pt").to(DEVICE)
            lg = self.model(**x).logits
            out.append(lg[:, -1].cpu().numpy() if lg.shape[-1] > 1 else lg[:, 0].cpu().numpy())
        return np.concatenate(out)

    def directional(self, pairs):
        a = [self.fmt(p["a"]) for p in pairs]
        b = [self.fmt(p["b"]) for p in pairs]
        sig = lambda z: 1 / (1 + np.exp(-z))
        return sig(self.logits(a, b)), sig(self.logits(b, a))

    def scores(self, pairs):
        ab, ba = self.directional(pairs)
        return (ab + ba) / 2
