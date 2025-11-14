FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o kubron-experiment-cake .

FROM scratch

COPY --from=builder /app/kubron-experiment-cake /kubron-experiment-cake

EXPOSE 8080
ENTRYPOINT ["/kubron-experiment-cake"]
