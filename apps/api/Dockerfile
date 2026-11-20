# syntax=docker/dockerfile:1

# Build stage: pinned Go toolchain image (see KI-003 in PROJECT_LEDGER.md: pin by digest once CI exists).
FROM golang:1.27 AS build
WORKDIR /src

# Dependencies first so this layer is cached until go.mod/go.sum change.
# When dependencies exist (M0.2+), change the next line to: COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
    -o /out/api ./cmd/api

# Runtime stage: no shell, no package manager, non-root user.
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
