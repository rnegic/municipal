import argparse
import hashlib
import json
import os
import shutil
import subprocess
import tarfile
import time
from collections import deque
from pathlib import Path

DEDUP = Path(__file__).resolve().parent.parent
REPO = DEDUP.parent.parent


def now():
    return time.strftime("%Y-%m-%dT%H:%M:%S%z")


def write_json(path, data):
    tmp = path.with_suffix(path.suffix + ".tmp")
    tmp.write_text(json.dumps(data, ensure_ascii=False, indent=2))
    tmp.replace(path)


def sha256(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def pending(root):
    done = {p.name.removesuffix(".status.json") for p in (root / "logs").glob("*.status.json")}
    jobs = [p for p in (root / "jobs").glob("*.json") if p.stem not in done]
    return sorted(jobs, key=lambda p: p.stat().st_mtime)


def git(*args):
    subprocess.run(["git", *args], cwd=REPO, check=True, capture_output=True, text=True)


def checkout(job):
    git("fetch", "--quiet", "origin")
    git("checkout", "--quiet", "--force", job.get("ref", "origin/main"))
    git("clean", "--quiet", "-fd")
    if job.get("patch"):
        subprocess.run(["git", "apply", "--whitespace=nowarn", "-"], cwd=REPO, check=True,
                       input=job["patch"], capture_output=True, text=True)


def run_cmd(cmd, log_path, tail_path, env):
    tail = deque(maxlen=200)
    last_flush = 0.0
    with open(log_path, "a") as log:
        proc = subprocess.Popen(["bash", "-lc", cmd], cwd=DEDUP, env=env, stdout=subprocess.PIPE,
                                stderr=subprocess.STDOUT, text=True, bufsize=1)
        for line in proc.stdout:
            log.write(line)
            tail.append(line)
            if time.time() - last_flush > 15:
                log.flush()
                tail_path.write_text("".join(tail))
                last_flush = time.time()
        rc = proc.wait()
    tail_path.write_text("".join(tail))
    return rc


def collect(job, out):
    out.mkdir(parents=True, exist_ok=True)
    files = []
    for pattern in job.get("collect", []):
        for src in DEDUP.glob(pattern):
            if src.is_file():
                dst = out / str(src.relative_to(DEDUP)).replace("/", "__")
                shutil.copy2(src, dst)
                files.append(dst.name)
    bundle = None
    if job.get("bundle"):
        onnx = DEDUP / "runs" / job["bundle"] / "onnx"
        tar = out / "bundle.tar.gz"
        with tarfile.open(tar, "w:gz") as t:
            for p in sorted(onnx.iterdir()):
                t.add(p, arcname=p.name)
        bundle = {"file": tar.name, "sha256": sha256(tar), "bytes": tar.stat().st_size}
    return files, bundle


def save_runs(job, root):
    for name in job.get("save_runs", []):
        src, dst = DEDUP / "runs" / name, root / "runs" / name
        if dst.exists():
            shutil.rmtree(dst)
        shutil.copytree(src, dst)


def restore_runs(root):
    runs = DEDUP / "runs"
    runs.mkdir(exist_ok=True)
    for src in (root / "runs").glob("*"):
        if src.is_dir() and not (runs / src.name).exists():
            shutil.copytree(src, runs / src.name)


def process(job_path, root, env):
    job = json.loads(job_path.read_text())
    jid = job_path.stem
    logs = root / "logs"
    status_path = logs / f"{jid}.status.json"
    status = {"id": jid, "state": "running", "started": now(), "cmd": job["cmd"]}
    write_json(status_path, status)
    log_path, tail_path = logs / f"{jid}.log", logs / f"{jid}.tail.txt"
    t0 = time.time()
    try:
        checkout(job)
        status["commit"] = subprocess.run(["git", "rev-parse", "HEAD"], cwd=REPO, capture_output=True,
                                          text=True).stdout.strip()
        write_json(status_path, status)
        rc = run_cmd(job["cmd"], log_path, tail_path, env)
        status["rc"] = rc
        if rc == 0:
            files, bundle = collect(job, root / "results" / jid)
            save_runs(job, root)
            status.update(files=files, bundle=bundle)
        status["state"] = "done" if rc == 0 else "failed"
    except Exception as e:
        status.update(state="failed", error=f"{type(e).__name__}: {getattr(e, 'stderr', '') or e}")
    status.update(finished=now(), seconds=round(time.time() - t0))
    write_json(status_path, status)
    return status


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", required=True)
    ap.add_argument("--venv", default=str(DEDUP / ".venv"))
    ap.add_argument("--poll", type=float, default=20)
    ap.add_argument("--once", action="store_true")
    args = ap.parse_args()
    root = Path(args.root)
    for d in ("jobs", "logs", "results", "runs"):
        (root / d).mkdir(parents=True, exist_ok=True)
    restore_runs(root)
    env = dict(os.environ, VIRTUAL_ENV=args.venv, PATH=f"{args.venv}/bin:{os.environ['PATH']}",
               HF_HOME=str(DEDUP / "cache" / "hf"), TOKENIZERS_PARALLELISM="false", PYTHONUNBUFFERED="1")
    heartbeat = root / "worker.json"
    while True:
        write_json(heartbeat, {"alive": now(), "pid": os.getpid()})
        jobs = pending(root)
        if jobs:
            s = process(jobs[0], root, env)
            print(f"{now()} {s['id']}: {s['state']} rc={s.get('rc')} {s.get('seconds')}s", flush=True)
        elif args.once:
            return
        else:
            time.sleep(args.poll)


if __name__ == "__main__":
    main()
