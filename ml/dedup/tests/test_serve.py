import json
import sys
import threading
import urllib.error
import urllib.request
from http.server import ThreadingHTTPServer
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from serve import handler


class FakeDeduper:
    model_version = "fake@1"
    threshold = 0.5

    def score(self, a, b):
        return {"duplicate_probability": 0.9, "decision": True, "model_version": self.model_version, "threshold": self.threshold}

    def match(self, request, candidates, top_k=20):
        return {"match": candidates[0]["id"] if candidates else None, "duplicate_probability": 0.9,
                "model_version": self.model_version, "threshold": self.threshold, "scores": []}


def post(url, body):
    req = urllib.request.Request(url, json.dumps(body).encode(), {"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req) as r:
            return r.status, json.load(r)
    except urllib.error.HTTPError as e:
        return e.code, json.load(e)


def test_serve():
    srv = ThreadingHTTPServer(("127.0.0.1", 0), handler(FakeDeduper()))
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    base = f"http://127.0.0.1:{srv.server_address[1]}"
    try:
        with urllib.request.urlopen(base + "/healthz") as r:
            assert json.load(r)["model_version"] == "fake@1"
        req = {"title": "Нет воды", "description": "с утра", "entrance": None, "riser": None, "severity": "critical", "created_at": "2026-09-25T10:00:00Z"}
        code, out = post(base + "/match", {"request": req, "candidates": [dict(req, id=7)], "top_k": 1})
        assert code == 200 and out["match"] == 7, out
        code, _ = post(base + "/match", {"request": {"title": "", "description": "x"}, "candidates": []})
        assert code == 400
    finally:
        srv.shutdown()


if __name__ == "__main__":
    test_serve()
    print("ok test_serve")
