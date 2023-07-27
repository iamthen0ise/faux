FROM golang:latest as builder
ARG buildplatform

LABEL maintainer="Evgenii Uvarov <e.uvarov@me.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN make build

EXPOSE 8080

RUN chmod +x /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["faux"]
