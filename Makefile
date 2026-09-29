REPO := rnegic/municipal
DEDUP_RELEASE := dedup-ftM2-berta-s16
LAYA_RELEASE := laya-onnx-55cf4c4

.PHONY: up stop down reset logs

up: .env ml/dedup/bundle.tar.gz backend/laya/model.tar.gz
	docker compose up -d --build

stop:
	docker compose stop

down:
	docker compose down

reset:
	docker compose down -v

logs:
	docker compose logs -f

.env:
	cp .env.example .env
	@echo "Создан .env: заполните POSTGRES_PASSWORD и MAX_BOT_TOKEN и повторите make up"
	@exit 1

ml/dedup/bundle.tar.gz:
	gh release download $(DEDUP_RELEASE) -R $(REPO) -p bundle.tar.gz -D ml/dedup

backend/laya/model.tar.gz:
	gh release download $(LAYA_RELEASE) -R $(REPO) -p model.tar.gz -D backend/laya
