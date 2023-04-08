# Build stage
FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /app/src .
EXPOSE 8080
CMD ["./app"]
