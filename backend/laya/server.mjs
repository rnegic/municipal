import http from "node:http";
import { Laya } from "@receptron/laya";
import { questions, toState, loadOptions } from "./questions.mjs";

let laya = null;

const send = (res, code, body) => {
  res.writeHead(code, { "Content-Type": "application/json" });
  res.end(JSON.stringify(body));
};

const server = http.createServer(async (req, res) => {
  if (req.method === "GET" && req.url === "/healthz") return send(res, laya ? 200 : 503, { ready: !!laya });
  if (req.method !== "POST" || req.url !== "/classify") return send(res, 404, { error: "not found" });
  if (!laya) return send(res, 503, { error: "model loading" });
  let raw = "";
  for await (const chunk of req) raw += chunk;
  try {
    const { text } = JSON.parse(raw);
    if (typeof text !== "string" || !text.trim()) return send(res, 400, { error: "text required" });
    const { answers } = await laya.systemOne(toState(text), questions);
    send(res, 200, { category: answers.category });
  } catch (err) {
    send(res, 500, { error: String(err) });
  }
});

server.listen(Number(process.env.PORT ?? 8090));
laya = await Laya.load(loadOptions());
console.log("laya ready", laya.modelDir);
