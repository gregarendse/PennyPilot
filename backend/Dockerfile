FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /out/pennypilot ./cmd/server

FROM alpine:3.20
RUN adduser -D -H pennypilot
USER pennypilot
WORKDIR /app
COPY --from=build /out/pennypilot /app/pennypilot
EXPOSE 8080
ENTRYPOINT ["/app/pennypilot"]
