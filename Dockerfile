FROM python:3.9

RUN apt update -y

RUN apt install golang -y

RUN go get github.com/go-telegram-bot-api/telegram-bot-api

RUN go get github.com/go-resty/resty/v2

RUN go get github.com/go-chi/chi/v5

RUN go get github.com/go-chi/render

RUN go get github.com/dustin/go-humanize

RUN go get github.com/mvdan/unidecode

ENTRYPOINT ["go", "run", "main.go"]
