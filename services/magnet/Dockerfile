FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o kubron-collider-magnet .

FROM scratch

COPY --from=builder /app/kubron-collider-magnet /kubron-collider-magnet

EXPOSE 8080
ENTRYPOINT ["/kubron-collider-magnet"]