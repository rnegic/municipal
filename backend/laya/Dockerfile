FROM python:3.12-slim AS export
RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*
RUN pip install --no-cache-dir --index-url https://download.pytorch.org/whl/cpu torch \
 && pip install --no-cache-dir transformers safetensors onnx onnxscript onnxruntime huggingface_hub
ARG LAYA_HF_REVISION=55cf4c4ebb4ebe31b2550e8bdf3bd21b99753851
ARG LAYA_SRC_REVISION=6478649e723122ca24bbf5fb69ed1010023c9750
RUN git init -q /src && cd /src && git fetch -q --depth 1 https://github.com/receptron/laya.git ${LAYA_SRC_REVISION} && git checkout -q FETCH_HEAD
COPY export/ /tools/
RUN python /tools/fetch.py ${LAYA_HF_REVISION} /checkpoint \
 && cd /src/export && python export_onnx.py /checkpoint /model \
 && python /tools/alias_tokens.py /model/tokenizer/tokenizer.json

FROM node:20-slim
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --omit=dev
COPY *.mjs eval.json ./
COPY --from=export /model /model
ENV LAYA_MODEL_DIR=/model
EXPOSE 8090
CMD ["node", "server.mjs"]
