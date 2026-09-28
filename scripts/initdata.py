import argparse
import hashlib
import hmac
import json
import os
import time
from urllib.parse import urlencode


def sign(bot_token, user_id, name, start_param, auth_date):
    fields = {"auth_date": str(auth_date), "user": json.dumps({"id": user_id, "first_name": name}, ensure_ascii=False)}
    if start_param:
        fields["start_param"] = start_param
    check = "\n".join(f"{k}={fields[k]}" for k in sorted(fields))
    secret = hmac.new(b"WebAppData", bot_token.encode(), hashlib.sha256).digest()
    fields["hash"] = hmac.new(secret, check.encode(), hashlib.sha256).hexdigest()
    return urlencode(fields)


if __name__ == "__main__":
    p = argparse.ArgumentParser(description="Заголовок Authorization жителя для локальной проверки API")
    p.add_argument("--user", type=int, default=900000101)
    p.add_argument("--name", default="Проверяющий")
    p.add_argument("--house", default="h_1", help="start_param: h_<id> привязывает дом, как QR-наклейка")
    p.add_argument("--token", default=os.environ.get("MAX_BOT_TOKEN", ""))
    a = p.parse_args()
    if not a.token:
        p.error("нужен --token или MAX_BOT_TOKEN — тот же, что у api")
    print("tma " + sign(a.token, a.user, a.name, a.house, int(time.time())))
