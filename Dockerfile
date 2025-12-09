FROM golang:1.25.4-alpine AS build
WORKDIR /app

COPY go.mod ./
COPY main.go ./

# Build a static binary for the target platform (defaults to linux/amd64).
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/shopify-webhook-faker

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=build /out/shopify-webhook-faker /app/shopify-webhook-faker

EXPOSE 8080
ENTRYPOINT ["/app/shopify-webhook-faker"]
