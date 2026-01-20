FROM golang:1.23.1-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/solback ./cmd

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/solback /app/solback
COPY --from=build /src/config.json /app/config.json
COPY --from=build /src/index.html /app/index.html

ENV CONFIG_PATH=/app/secrets.json
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/solback"]
