run:
	go run ./cmd/app

build:
	go build -o bin/app ./cmd/app

# ──────────────────────────────────────────────────
# Authentication & Database Commands
# ──────────────────────────────────────────────────

db-migrate:
	supabase db push

db-status:
	supabase db push --dry-run

db-reset:
	supabase db reset

db-seed:
	@echo "Seeding database... (not yet implemented)"

# ──────────────────────────────────────────────────
# Development Commands
# ──────────────────────────────────────────────────

dev:
	go run ./cmd/app

logs:
	@tail -f storage/logs/*.log

clean:
	rm -rf bin/

.PHONY: run build db-migrate db-status db-reset dev logs clean