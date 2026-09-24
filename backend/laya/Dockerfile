FROM python:3.12-slim@sha256:2f17fc044b579bab302c2e8054d3a686e2cb9a83de48e70534b94cd8ebbe06a9 AS export
RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
COPY export/requirements.txt /tools/requirements.txt
RUN pip install --no-cache-dir --extra-index-url https://download.pytorch.org/whl/cpu -r /tools/requirements.txt
ARG LAYA_HF_REVISION=55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851
ARG LAYA_SRC_REVISION=6478649e723122ca24bbf5fb69ed1010023c9750
RUN git init -q /src && cd /src && git fetch -q --depth 1 https://github.com/receptron/laya.git ${LAYA_SRC_REVISION} && git checkout -q FETCH_HEAD
COPY export/ /tools/
RUN python /tools/fetch.py ${LAYA_HF_REVISION} /checkpoint \
 && cd /src/export && python export_onnx.py /checkpoint /model \
 && python /tools/alias_tokens.py /model/tokenizer/tokenizer.json

FROM node:20-slim@sha256:2cf067cfed83d5ea958367df9f966191a942351a2df77d6f0193e162b5febfc0
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev
COPY *.mjs eval.json ./
COPY --from=export /model /model
ENV LAYA_MODEL_DIR=/model
EXPOSE 8090
CMD ["node", "server.mjs"]
