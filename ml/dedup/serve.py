import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from dedup.onnx_infer import OnnxDeduper

REQUIRED = ("title", "description")


def valid(r):
    return isinstance(r, dict) and all(isinstance(r.get(k), str) and r[k].strip() for k in REQUIRED)


def handler(deduper):
    class H(BaseHTTPRequestHandler):
        def _send(self, code, body):
            data = json.dumps(body, ensure_ascii=False).encode()
            self.send_response(code)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        def do_GET(self):
            if self.path == "/healthz":
                return self._send(200, {"model_version": deduper.model_version, "threshold": deduper.threshold})
            self._send(404, {"error": "not found"})

        def do_POST(self):
            try:
                body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", 0))) or b"{}")
            except (ValueError, json.JSONDecodeError):
                return self._send(400, {"error": "invalid json"})
            if self.path == "/score":
                a, b = body.get("request_a"), body.get("request_b")
                if not (valid(a) and valid(b)):
                    return self._send(400, {"error": "request_a and request_b need non-empty title and description"})
                return self._send(200, deduper.score(a, b))
            if self.path == "/match":
                req, cands = body.get("request"), body.get("candidates", [])
                if not valid(req) or not isinstance(cands, list) or not all(valid(c) for c in cands):
                    return self._send(400, {"error": "request and candidates need non-empty title and description"})
                return self._send(200, deduper.match(req, cands, int(body.get("top_k", 20))))
            self._send(404, {"error": "not found"})

        def log_message(self, *args):
            pass

    return H


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--bundle", required=True)
    ap.add_argument("--port", type=int, default=8090)
    ap.add_argument("--threads", type=int, default=2)
    args = ap.parse_args()
    d = OnnxDeduper(args.bundle, threads=args.threads)
    print(f"serving {d.model_version} threshold={d.threshold} on :{args.port}", flush=True)
    ThreadingHTTPServer(("127.0.0.1", args.port), handler(d)).serve_forever()


if __name__ == "__main__":
    main()
