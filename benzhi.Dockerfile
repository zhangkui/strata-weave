FROM golang:1.22-bookworm
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/strataweave ./cmd/strataweave
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/strataweave"]
