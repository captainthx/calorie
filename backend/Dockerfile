FROM golang:1.26.4-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/api /app/api

ENV MODE=release
ENV PORT=8080

EXPOSE 8080

CMD ["/app/api"]
