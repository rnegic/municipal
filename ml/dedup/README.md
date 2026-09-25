# ml/dedup — модель дедупликации заявок

Pairwise Cross-Encoder (текущая модель — `sergeyzh/BERTA`, `runs/ftJ-berta-s14`), дообученный на задаче
«две заявки сообщают об одной и той же конкретной аварии?». Вход: две заявки
(title, description, подъезд, стояк, срочность). Выход: `P(duplicate)`.

Итоговый отчёт и статус: [`docs/ml/dedup-report.md`](../../docs/ml/dedup-report.md).

## Окружение

```bash
cd ml/dedup
uv venv -p 3.10 .venv
uv pip install -p .venv/bin/python -r requirements.txt --index-strategy unsafe-best-match
export HF_HOME=$PWD/cache/hf TOKENIZERS_PARALLELISM=false
```

В `requirements.txt` CPU-сборка torch для сервера. Для обучения на GPU поставьте CUDA-сборку (`torch 2.4.1+cu121`), `dedup/scorers.py` сам выберет cuda.

## Полный цикл

```bash
.venv/bin/python gen_synthetic.py
.venv/bin/python build_dataset.py
.venv/bin/python train.py --config configs/ft-text.json
.venv/bin/python train.py --config configs/pp-pretrain.json
.venv/bin/python gen_synthetic.py --extra-vocab --out data/raw/synthetic_template_v4.jsonl.gz && .venv/bin/python build_dataset.py --synthetic data/raw/synthetic_template_v4.jsonl.gz
.venv/bin/python build_dataset.py
.venv/bin/python train.py --config configs/pp2-berta.json
.venv/bin/python train.py --config configs/ftJ-berta-s14.json
.venv/bin/python evaluate.py --models runs/ftJ-berta-s14 --threshold-split val_hw
.venv/bin/python export.py --run runs/ftJ-berta-s14
```

| Шаг | Что делает | Результат |
|---|---|---|
| `gen_synthetic.py` | шаблонные «сцены дома» с латентными инцидентами и hard-negative сиблингами | `data/raw/synthetic_template_v1.jsonl.gz` |
| `build_dataset.py` | split по `group_id`, проверки качества, версия | `data/versions/<ds>/*.jsonl.gz`, `dataset_report.{json,md}` |
| `train.py` | fine-tune, BCE, фиксированный seed | `runs/<name>/{model,config,dataset_version.txt,train_log.json}` |
| `evaluate.py` | 5 бейзлайнов + модели на одних и тех же срезах; порог только по val | `runs/eval/metrics.json`, `runs/<name>/{metrics,threshold}.json` |
| `export.py` | ONNX fp32 + int8, сверка с PyTorch, latency на 2 потоках | `runs/<name>/onnx/` |

Порог в коде модели не зашит: он лежит в `threshold.json` рядом с бандлом.

## Реальные данные (цикл обновления)

```bash
psql "$DATABASE_URL" -c "\copy (SELECT id, incident_id, reporter_id, house_id, title, description, severity, entrance, riser, outcome, dedup_version, created_at FROM incident_report ORDER BY id) TO 'data/raw/incident_report.csv' CSV HEADER"
.venv/bin/python mine_pairs.py
.venv/bin/python label.py data/labeling/candidates_<tag>.csv
.venv/bin/python label.py data/labeling/candidates_<tag>.csv --export data/labels/manual_<tag>.jsonl
.venv/bin/python build_dataset.py
```

- `weak_joined` попадают только в train и никогда в val/test.
- `manual` делятся по дому (`group_id = real-house-<id>`) на train/val_real/test_real.
- Если есть ручная метка, она перекрывает weak-метку той же пары.

## Inference

```bash
.venv/bin/python serve.py --bundle runs/ftJ-berta-s14/onnx --port 8090
curl -s localhost:8090/score -d '{"request_a":{"title":"Нет горячей воды","description":"с утра, стояк 3","riser":"3"},"request_b":{"title":"Горячей нет","description":"кв 27, 2 подъезд","riser":"3"}}'
```

Ответ: `{"duplicate_probability": …, "decision": …, "model_version": …, "threshold": …}`.
`POST /match` принимает `{request, candidates[]}` и возвращает лучшего кандидата выше порога.
Скор симметричный: среднее по порядкам (A,B) и (B,A). По умолчанию используется `model.onnx` (fp32); `model.int8.onnx` не прошёл сверку с PyTorch (см. `export_check.json`).

Из Python:

```python
from dedup.onnx_infer import OnnxDeduper
d = OnnxDeduper("runs/ftJ-berta-s14/onnx")
d.score(a, b)
```
