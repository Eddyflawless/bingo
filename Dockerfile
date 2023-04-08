# Build stage
FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o app src/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache shadow

RUN apk --no-cache add ca-certificates

# RUN groupadd -g 1000 appuser && \
# useradd -r -u 1000 -g appuser appuser

# # Switch to app user
# USER appuser

WORKDIR /root/
COPY --from=build /app/src .
EXPOSE 8080
CMD ["./app"]
