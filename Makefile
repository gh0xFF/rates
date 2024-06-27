# уже устал дебажить/перезапускать этот глючный докер
# при старте сервиса, когда запрашивается курс валюты в докере получаю 404 код
# в то же время запускаю нативно и всё работает как надо

# может баг сам пройдёт через 5 минут, а может нужно перезапустить ноут/докер
# костыльно, но работает с mysql

start:
	export ENV=prod && \
	export PORT=8080 && \
	export CONNECTION_STRING="root:password@tcp(localhost:3306)/rates" && \
	export RATES_URL=https://api.nbrb.by/exrates/rates?periodicity=0 && \
	go mod tidy && \
	docker compose  -f "docker-compose.yml" up -d --build database && \
	sleep 10 && \
	go run ./cmd/main.go
