FROM golang:1.22 as buildenv

WORKDIR /src
ADD . /src

RUN go mod tidy
RUN go build -ldflags="-w -s" -o event ./cmd/main.go
RUN chmod +x event

# ////////////////////////////////////////////////////////////////

FROM golang:1.22
# не alpine, в момент когда я тестировал сборку, apline тупит и не стартует бинарник который видит
# взял именно этот образ так как можно юзать пакетный мэнэджер
WORKDIR /usr/local/app

ENV APP_ENV=dev

# для https запросов
RUN apt-get install -y ca-certificates 

COPY --from=buildenv /src/event .

EXPOSE 8080

CMD ["sh", "-c", "./event"]