# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /check-untagged-go-deps .

# Runtime stage
FROM golang:1.25-alpine

# git is required for 'go list' to query module versions from repositories
RUN apk add --no-cache git

COPY --from=builder /check-untagged-go-deps /check-untagged-go-deps

# GitHub Actions mounts the workspace at /github/workspace
WORKDIR /github/workspace

ENTRYPOINT ["/check-untagged-go-deps"]
