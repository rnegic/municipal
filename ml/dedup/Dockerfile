FROM python:3.10-slim
WORKDIR /app
COPY requirements-serve.txt .
RUN pip install --no-cache-dir -r requirements-serve.txt
ARG BUNDLE_SHA256
COPY bundle.tar.gz /tmp/bundle.tar.gz
RUN echo "$BUNDLE_SHA256  /tmp/bundle.tar.gz" | sha256sum -c - \
 && mkdir /bundle && tar -xzf /tmp/bundle.tar.gz -C /bundle && rm /tmp/bundle.tar.gz
COPY serve.py .
COPY dedup/__init__.py dedup/data.py dedup/onnx_infer.py dedup/
ENV DEDUP_THREADS=2
EXPOSE 8090
CMD python serve.py --bundle /bundle --host 0.0.0.0 --port 8090 --threads $DEDUP_THREADS
