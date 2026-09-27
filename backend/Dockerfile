FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY russian_trusted_root_ca.crt /usr/local/share/ca-certificates/
RUN update-ca-certificates
COPY --from=build /api /api
EXPOSE 8080
CMD ["/api"]
