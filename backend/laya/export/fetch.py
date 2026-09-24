import os
import shutil
import sys

from huggingface_hub import snapshot_download

revision, dest = sys.argv[1], sys.argv[2]
src = snapshot_download("convaiinnovations/laya", revision=revision,
                        allow_patterns=["multilingual/*", "multilingual/**", "rl_common.py", "rl_agent_api.py", "email_utils.py"])
shutil.copytree(os.path.join(src, "multilingual"), dest, dirs_exist_ok=True)
for f in ["rl_common.py", "rl_agent_api.py", "email_utils.py"]:
    shutil.copy(os.path.join(src, f), dest)
