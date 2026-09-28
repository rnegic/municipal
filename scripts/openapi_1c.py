import copy
import sys

import yaml

SRC = "backend/openapi.yaml"
DST = "docs/uk-1c-openapi.yaml"

DESCRIPTION = """Контракт для 1С управляющей компании: операции основного API с тегом `integration`.
Сгенерирован из backend/openapi.yaml командой `make openapi-1c` — правьте исходник.

Авторизация: `Authorization: Bearer uk_live_…` — ключ выпускается в кабинете УК и действует
как диспетчер своей УК: видны только её дома и заявки. Не больше 10 запросов в секунду на ключ.
Истёкшая лицензия УК или отозванный ключ — 401.

Все запросы и ответы — JSON в UTF-8, ошибки — `{ "code": "...", "message": "..." }`.
Гайд по подключению — docs/uk-1c-integration.md.
"""


def refs(node, found):
    if isinstance(node, dict):
        ref = node.get("$ref")
        if isinstance(ref, str) and ref.startswith("#/components/"):
            _, _, kind, name = ref.split("/", 3)
            found.add((kind, name))
        for v in node.values():
            refs(v, found)
    elif isinstance(node, list):
        for v in node:
            refs(v, found)


def build(api):
    paths = {}
    for path, item in api["paths"].items():
        ops = {m: op for m, op in item.items() if isinstance(op, dict) and "integration" in op.get("tags", [])}
        for op in ops.values():
            op["tags"] = ["integration"]
            op["security"] = [{"UkApiKey": []}]
        if ops:
            paths[path] = ops

    needed, seen = set(), set()
    refs(paths, needed)
    while needed - seen:
        kind, name = (needed - seen).pop()
        seen.add((kind, name))
        refs(api["components"][kind][name], needed)

    components = {}
    for kind, name in sorted(seen):
        components.setdefault(kind, {})[name] = api["components"][kind][name]
    components["securitySchemes"] = {"UkApiKey": api["components"]["securitySchemes"]["UkApiKey"]}

    return {
        "openapi": api["openapi"],
        "info": {"title": "Дом.Пульс — API для 1С", "version": api["info"]["version"], "description": DESCRIPTION},
        "servers": [{"url": "https://nerionapp.ru"}],
        "security": [{"UkApiKey": []}],
        "tags": [{"name": "integration", "description": "Обмен заявками и статусами с 1С УК"}],
        "paths": paths,
        "components": components,
    }


if __name__ == "__main__":
    with open(SRC, encoding="utf-8") as f:
        spec = build(copy.deepcopy(yaml.safe_load(f)))
    out = yaml.safe_dump(spec, allow_unicode=True, sort_keys=False, width=120)
    if "--check" in sys.argv:
        with open(DST, encoding="utf-8") as f:
            if f.read() != out:
                sys.exit(f"{DST} устарел: выполните make openapi-1c")
    else:
        with open(DST, "w", encoding="utf-8") as f:
            f.write(out)
