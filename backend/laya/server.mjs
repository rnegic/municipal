import http from "node:http";
import { Laya } from "@receptron/laya";
import { createClassifier, loadOptions } from "./questions.mjs";

const MAX_BODY_BYTES = 64 * 1024;

let classify = null;

const send = (res, code, body) => {
  res.writeHead(code, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
};

const server = http.createServer(async (req, res) => {
  if (req.method === "GET" && req.url === "/healthz") return send(res, classify ? 200 : 503, { ready: !!classify });
  if (req.method !== "POST" || req.url !== "/classify") return send(res, 404, { error: "not found" });
  if (!classify) return send(res, 503, { error: "model loading" });
  let raw = "";
  for await (const chunk of req) {
    raw += chunk;
    if (raw.length > MAX_BODY_BYTES) return send(res, 413, { error: "body too large" });
  }
  try {
    const { text } = JSON.parse(raw);
    if (typeof text !== "string" || !text.trim()) return send(res, 400, { error: "text required" });
    send(res, 200, { category: await classify(text) });
  } catch (err) {
    send(res, 500, { error: String(err) });
  }
});

server.listen(Number(process.env.PORT ?? 8090));
const laya = await Laya.load(loadOptions());
classify = await createClassifier(laya);
console.log("laya ready", laya.modelDir);
