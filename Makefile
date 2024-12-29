include .env
export
up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f
.PHONY: up

down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: down

rm-volume: ### remove docker volume
	docker volume rm redis_data
.PHONY: rm-volume

lint: ### check by golangci linter
	golangci-lint run
.PHONY: lint