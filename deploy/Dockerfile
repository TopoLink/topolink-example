FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -o /out/consumer ./cmd/consumer

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/consumer /usr/local/bin/consumer
EXPOSE 8080 9090
ENTRYPOINT ["/usr/local/bin/api"]
