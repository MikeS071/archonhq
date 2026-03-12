.PHONY: help up down api-run repo-tree check-boundaries

help:
	@echo "Targets:"
	@echo "  make up        # start local infra + api via docker compose"
	@echo "  make down      # stop local stack"
	@echo "  make api-run   # run api locally (no build)"
	@echo "  make repo-tree # print top-level tree"
	@echo "  make check-boundaries # verify package dependency rules"

up:
	docker compose up -d

down:
	docker compose down

api-run:
	go run ./apps/api/cmd/server

repo-tree:
	find . -maxdepth 2 -type d | sort

check-boundaries:
	./scripts/check-import-boundaries.sh
