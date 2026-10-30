# --- build -------------------------------------------------------------------
FROM golang:1.22-alpine AS build

WORKDIR /src

# Baixa dependências em camada separada para aproveitar o cache do Docker
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migration ./cmd/migration

# --- runtime -----------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -H -u 10001 app

WORKDIR /app

COPY --from=build /out/api /out/migration ./

USER app

EXPOSE 8080

# Por padrão sobe a API. Para rodar as migrations:
#   docker compose run --rm migrate          (ou)   /app/migration -seed
CMD ["/app/api"]
