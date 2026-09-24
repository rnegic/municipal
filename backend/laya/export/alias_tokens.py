import json
import sys

path = sys.argv[1]
tok = json.load(open(path, encoding="utf-8"))
vocab = tok["model"]["vocab"]
for alias, token in {"[CLS]": "<bos>", "[SEP]": "<eos>", "[MASK]": "<mask>", "[PAD]": "<pad>"}.items():
    vocab[alias] = vocab[token]
json.dump(tok, open(path, "w", encoding="utf-8"), ensure_ascii=False)
