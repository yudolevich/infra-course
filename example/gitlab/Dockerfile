FROM golang:1.22-alpine as builder

WORKDIR /usr/src

COPY . .
RUN go build -o /usr/src/app

FROM scratch
COPY --from=builder /usr/src/app /app

CMD ["/app"]
