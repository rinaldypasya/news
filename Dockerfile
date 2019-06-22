FROM golang:latest AS compiler
LABEL maintainer="Rinaldy Pasya <rinaldypasya@gmail.com>"

RUN apk update && \
    apk upgrade && \
    apk add bash git

RUN go get github.com/markbates/refresh
RUN go get github.com/gin-gonic/gin
RUN go get github.com/go-redis/cache
RUN go get github.com/go-redis/redis
RUN go get github.com/olivere/elastic
RUN go get github.com/vmihailenco/msgpack
RUN go get github.com/streaday/ampq
RUN go get github.com/itsjamie/gin-cors
RUN go get github.com/jinzhu/gorm
RUN go get github.com/jinzhu/gorm/dialects/postgres