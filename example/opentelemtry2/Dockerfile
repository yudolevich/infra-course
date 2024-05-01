FROM golang:1.21 as build

WORKDIR /src
COPY . /src/
RUN go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/app ./main.go

FROM scratch
COPY --from=build /bin/app /app
ENTRYPOINT ["/app"]
