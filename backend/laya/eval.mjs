import { readFileSync } from "node:fs";
import { Laya } from "@receptron/laya";
import { createClassifier, loadOptions } from "./questions.mjs";

const { labeled, noise } = JSON.parse(readFileSync(new URL("./eval.json", import.meta.url)));
const laya = await Laya.load(loadOptions());
console.log("model", laya.modelDir);
const classify = await createClassifier(laya);

const run = async (text) => {
  const t = Date.now();
  const a = await classify(text);
  return { label: a.choice, p: a.probabilities[a.choice], ms: Date.now() - t };
};

await run("прогрев");
const rows = [];
for (const ex of labeled) {
  const r = await run(ex.text);
  const ok = r.label === ex.category;
  rows.push({ ...r, ok });
  console.log(`${r.label}:${r.p.toFixed(2)} ${r.ms}ms`, ok ? "" : `≠${ex.category}`, "|", ex.text);
}
const noiseRows = [];
for (const text of noise) {
  const r = await run(text);
  noiseRows.push(r);
  console.log(`NOISE ${r.label}:${r.p.toFixed(2)} | ${text}`);
}

console.log(`\naccuracy=${(rows.filter((r) => r.ok).length / rows.length).toFixed(2)}`);
const ms = rows.map((r) => r.ms).sort((a, b) => a - b);
console.log(`latency p50=${ms[Math.floor(ms.length / 2)]}ms max=${ms.at(-1)}ms`);
console.log("\nthreshold  coverage  accuracy  noise_passed");
for (const t of [0.5, 0.6, 0.7, 0.8, 0.9]) {
  const cov = rows.filter((r) => r.p >= t);
  const acc = cov.length ? cov.filter((r) => r.ok).length / cov.length : 0;
  console.log(`${t}        ${(cov.length / rows.length).toFixed(2)}      ${acc.toFixed(2)}      ${noiseRows.filter((r) => r.p >= t).length}/${noiseRows.length}`);
}
