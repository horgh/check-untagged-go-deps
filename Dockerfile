# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /check-untagged-go-deps .

# Runtime stage
FROM golang:1.25-alpine

# The tool needs 'go' command to query module versions
COPY --from=builder /check-untagged-go-deps /check-untagged-go-deps

# GitHub Actions mounts the workspace at /github/workspace
WORKDIR /github/workspace

ENTRYPOINT ["/check-untagged-go-deps"]
