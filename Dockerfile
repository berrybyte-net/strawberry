FROM golang:1.18-alpine as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download \
    && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w" -o /app/strawberry cmd/strawberry/main.go

FROM gcr.io/distroless/static-debian11:latest
COPY --from=build /app/strawberry /
ENTRYPOINT ["/strawberry"]