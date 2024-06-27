FROM golang:1.22 as buildenv

WORKDIR /src
ADD . /src

RUN apt-get install openssl
RUN go build -ldflags="-w -s" -o event ./cmd/main.go
RUN chmod +x event

# ////////////////////////////////////////////////////////////////

FROM ubuntu:latest
# не alpine, в момент когда я тестировал сборку, apline тупит и не стартует бинарник который видит
WORKDIR /usr/local/app

ENV APP_ENV=dev

COPY --from=buildenv /src/event .

EXPOSE 8080
EXPOSE 80

CMD ["sh", "-c", "./event"]