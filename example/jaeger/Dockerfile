FROM golang:1.21 as build

WORKDIR /src
COPY main.go /src/main.go
RUN go mod init example \
  && go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/app ./main.go

FROM scratch
COPY --from=build /bin/app /app
CMD ["/app"]
