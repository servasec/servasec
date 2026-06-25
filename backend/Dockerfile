# ============================================
# Build main application
# ============================================
FROM golang:1.25-alpine AS go_builder

WORKDIR /app

COPY go.mod go.sum ./
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o servasec

# ============================================
# Dev stage
# ============================================
FROM golang:1.25-alpine AS dev

WORKDIR /app

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 -s /sbin/nologin app && \
    mkdir -p /home/app/.ssh && \
    chown -R app:app /home/app/.ssh && \
    chmod 700 /home/app/.ssh && \
    chown -R app:app /app

USER app

RUN mkdir -p /app/tmp
RUN go install github.com/air-verse/air@v1.61.7

COPY --chown=app:app . .
RUN go mod download

COPY entrypoint.dev.sh /entrypoint.sh

EXPOSE 8080

CMD ["sh", "/entrypoint.sh"]

# ============================================
# Prod stage
# ============================================
FROM alpine:3.22 AS prod

WORKDIR /app

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1000 -s /sbin/nologin app && \
    chown -R app:app /app

USER app

COPY --from=go_builder /app/servasec /app/servasec
COPY config/casbin_model.conf /app/config/
COPY config/casbin_policies.csv /app/config/
COPY entrypoint.prod.sh /entrypoint.sh

EXPOSE 8080

CMD ["sh", "/entrypoint.sh"]
