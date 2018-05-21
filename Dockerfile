FROM golang:1.10.2 as builder
ARG version
WORKDIR /app
ADD ./main.go .
ADD ./main_test.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o main -ldflags "-s -w -X main.version=$version" .
RUN go test -timeout 30s

FROM scratch
WORKDIR /app
COPY --from=builder /app/main .
CMD ["/app/main"]
EXPOSE 3000
