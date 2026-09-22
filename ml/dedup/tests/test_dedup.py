import sys
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from dedup.data import normalize, struct_text
from dedup.metrics import at_threshold, recall_at_precision, select_thresholds, simulate_merges
from mine_pairs import parse_ts

BUNDLE = Path(__file__).resolve().parents[1] / "runs/ft-rubert-tiny2-struct/onnx"


def test_metrics():
    y = [1, 1, 0, 0, 0]
    s = [0.9, 0.4, 0.8, 0.1, 0.2]
    m = at_threshold(y, s, 0.5)
    assert (m["tp"], m["fp"], m["fn"], m["tn"]) == (1, 1, 1, 2)
    r, thr = recall_at_precision(y, s, 0.95)
    assert r == 0.5 and thr == 0.9
    t = select_thresholds(y, s)
    assert t["p95"] == 0.9


def test_merge_simulation():
    def pair(a, b, ia, ib, t_a, t_b):
        return {"group_id": "h", "created_at_a": t_a, "created_at_b": t_b, "text_a": a, "text_b": b, "incident_a": ia, "incident_b": ib}
    pairs = [pair("x1", "x2", "X", "X", "1", "2"), pair("x1", "y1", "X", "Y", "1", "3"), pair("x2", "y1", "X", "Y", "2", "3")]
    r = simulate_merges(pairs, [0.9, 0.8, 0.1], 0.5)
    assert r["correct_merges"] == 1 and r["false_merges"] == 1 and r["predicted_incidents"] == 1


def test_preprocessing():
    assert normalize("Нет  Горячей воды!!!") == "нет горячей воды"
    assert normalize("Ёлка") == "елка"
    assert struct_text({"title": "t", "description": "d", "entrance": None, "riser": "3", "severity": "critical"}) == "подъезд: ?; стояк: 3; срочность: critical; t. d"
    assert parse_ts("2026-09-20 10:00:00.5+00").isoformat() == "2026-09-20T10:00:00.500000+00:00"
    assert parse_ts("2026-09-20 10:00:00+03").utcoffset().total_seconds() == 3 * 3600


def test_symmetry():
    if not BUNDLE.exists():
        print("skip test_symmetry: no bundle")
        return
    from dedup.onnx_infer import OnnxDeduper
    d = OnnxDeduper(BUNDLE)
    a = {"title": "Нет горячей воды", "description": "стояк 3, с утра", "entrance": "2", "riser": "3", "severity": "critical"}
    b = {"title": "Горячей нет", "description": "кв 27, ледяная", "entrance": None, "riser": None, "severity": "critical"}
    ab = d.scores([a], [b])[0]
    ba = d.scores([b], [a])[0]
    assert abs(ab - ba) < 1e-6, (ab, ba)
    raw_ab, raw_ba = d.directional([a], [b])
    print(f"raw directional asymmetry {abs(raw_ab[0] - raw_ba[0]):.4f} (symmetrized score removes it)")


def test_hard_negative_smoke():
    if not BUNDLE.exists():
        return
    from dedup.onnx_infer import OnnxDeduper
    d = OnnxDeduper(BUNDLE)
    hot = {"title": "Нет горячей воды", "description": "стояк 3, горячей нет", "riser": "3", "severity": "critical"}
    cold = {"title": "Нет холодной воды", "description": "стояк 3, холодной нет, горячая есть", "riser": "3", "severity": "critical"}
    assert not d.score(hot, cold)["decision"]
    assert np.isfinite(d.score(hot, hot)["duplicate_probability"])


if __name__ == "__main__":
    for name, fn in list(globals().items()):
        if name.startswith("test_"):
            fn()
            print("ok", name)
