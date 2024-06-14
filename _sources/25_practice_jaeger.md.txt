# Jaeger
В данном практическом занятии познакомимся базовому взаимодействию
с инструментом трассировки [jaeger][].

## Vagrant
```ruby
Vagrant.configure("2") do |config|
  config.vm.define "jaeger" do |c|
    c.vm.box = "ubuntu/lunar64"
    c.vm.hostname = "jaeger"
    c.vm.network "forwarded_port", guest: 8888, host: 8888
    c.vm.network "forwarded_port", guest: 8889, host: 8889
    c.vm.provision "shell", inline: <<-SHELL
      apt-get update -q
      apt-get install -yq docker.io docker-compose-v2
      usermod -a -G docker vagrant
    SHELL
  end
end
```
Данная конфигурация установит на виртуальную машину [docker][] и
[docker compose][docker-compose], с помощью которых в дальнейшем будут
развернуты остальные компоненты.

## Jaeger
Развернем `jaeger` с помощью следующего `compose.yaml`:
```yaml
services:
  jaeger:
    container_name: jaeger
    image: jaegertracing/all-in-one
    ports:
      - "8889:14268"
      - "8888:16686"
```
```console
$ docker compose up -d
[+] Running 2/2
 ✔ Network vagrant_default  Created                                                   0.1s
 ✔ Container jaeger         Started                                                   0.3s
```

После чего по адресу [localhost:8888/search](http://localhost:8888/search) будет
доступен интерфейс.

![](img/jaeger1.png)

## Simple Trace
Создадим простое приложение, которое будет отправлять трейсы в jaeger:
```golang
package main

import (
        "context"
        "errors"
        "log"
        "time"

        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
        "go.opentelemetry.io/otel/sdk/resource"
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
        "go.opentelemetry.io/otel/trace"
        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
)

const (
        ServiceName = "service"
)

var (
        prv *sdktrace.TracerProvider
        tr trace.Tracer
        errSpan = errors.New("span error")
)

func initTracer(ctx context.Context) error{
        conn, err := grpc.NewClient("jaeger:4317",
                grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
                return err
        }

        exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
        if err != nil {
                return err
        }

        res, err := resource.New(ctx,
                resource.WithAttributes(
                        semconv.ServiceName(ServiceName),
                ),
        )
        if err != nil {
                return err
        }

        prv = sdktrace.NewTracerProvider(
                sdktrace.WithBatcher(exp),
                sdktrace.WithResource(res),
        )

        otel.SetTracerProvider(prv)
        tr = prv.Tracer("tracer")

        return nil
}

func main() {
        ctx := context.Background()
        if err := initTracer(ctx);err != nil {
                log.Fatal("init tracer", err)
        }

        ctx, span := tr.Start(ctx, "span")
        time.Sleep(time.Second)
        span.End()

        if err := prv.Shutdown(ctx); err != nil {
                log.Fatal("failed shutdown", err)
        }
}
```
Данное приложение инициализирует отправку трейсов: создает grpc подключение, атрибуты,
провайдер и экспортер. После чего создает спан и после ожидания закрывает его.

Добавим `Dockerfile` для него:

```dockerfile
FROM golang:1.21 as build

WORKDIR /src
COPY main.go /src/main.go
RUN go mod init example \
  && go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/app ./main.go

FROM scratch
COPY --from=build /bin/app /app
CMD ["/app"]
```

А также обновим `compose.yaml`:
```yaml
services:
  jaeger:
    container_name: jaeger
    image: jaegertracing/all-in-one
    ports:
      - "8889:14268"
      - "8888:16686"
  app:
    container_name: test
    image: test
    build: .
```

После чего запустим:
```console
$ docker compose up -d --build
[+] Running 2/2
 ✔ Container test    Started                                                          0.4s
 ✔ Container jaeger  Running                                                          0.0s
```

Когда приложение отработает, то в интерфейсе `jaeger` можно будет увидеть наш новый
сервис с названием `service`:

![](img/jaeger2.png)

Нажав `Find Traces` можем увидеть список последних трейсов:

![](img/jaeger3.png)

А кликнув по конкретному трейсу увидим подробности о нем:

![](img/jaeger4.png)

Наш трейс состоит из одного спана, кликнув по нему можно узнать подробности и о нем:

![](img/jaeger5.png)

## Multiple Spans
Добавим в наше приложение генерацию нескольких спанов, через которые можно отслеживать
выполнение нескольких операций. Для этого напишем функцию `test`, которая при вызове
будет создавать новый спан используя контекст корневого и выполняться со случайной
задержкой:
```golang
package main

import (
        "context"
        "errors"
        "log"
        "time"
        "fmt"
        "math/rand"

        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
        "go.opentelemetry.io/otel/sdk/resource"
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
        "go.opentelemetry.io/otel/trace"
        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
)

const (
        ServiceName = "service"
)

var (
        prv *sdktrace.TracerProvider
        tr trace.Tracer
        errSpan = errors.New("span error")
)

func initTracer(ctx context.Context) error{
        conn, err := grpc.NewClient("jaeger:4317",
                grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
                return err
        }

        exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
        if err != nil {
                return err
        }

        res, err := resource.New(ctx,
                resource.WithAttributes(
                        semconv.ServiceName(ServiceName),
                ),
        )
        if err != nil {
                return err
        }

        prv = sdktrace.NewTracerProvider(
                sdktrace.WithBatcher(exp),
                sdktrace.WithResource(res),
        )

        otel.SetTracerProvider(prv)
        tr = prv.Tracer("tracer")

        return nil
}

func main() {
        ctx := context.Background()
        if err := initTracer(ctx);err != nil {
                log.Fatal("init tracer", err)
        }

        ctx, span := tr.Start(ctx, "root span")
        test(ctx, 1)
        test(ctx, 2)
        span.End()

        if err := prv.Shutdown(ctx); err != nil {
                log.Fatal("failed shutdown", err)
        }
}

func test(ctx context.Context, count int) {
        ctx, span := tr.Start(ctx, fmt.Sprintf("span-%d", count))
        defer span.End()
        num := rand.Intn(5)+1
        time.Sleep(time.Duration(num)*time.Second)
}
```
И запустим:
```console
$ docker compose up -d --build
[+] Running 2/2
 ✔ Container jaeger  Running                                                          0.0s
 ✔ Container test    Started                                                          0.5s
```

После чего в jaeger увидим новый трейс:

![](img/jaeger6.png)

В результате видно какое время выполнялась каждая функция.

## Error Span
Как видно было при детальном рассмотрении информации о спане - в нем можно хранить
множество дополнительных атрибутов. Одним из удобных объектов для хранения в спане -
это информация об ошибке. Добавим в нашу функцию `test` возможность возникновения
ошибки:
```golang
package main

import (
        "context"
        "errors"
        "log"
        "time"
        "fmt"
        "math/rand"

        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/codes"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
        "go.opentelemetry.io/otel/sdk/resource"
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
        "go.opentelemetry.io/otel/trace"
        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
)

const (
        ServiceName = "service"
)

var (
        prv *sdktrace.TracerProvider
        tr trace.Tracer
        errSpan = errors.New("span error")
)

func initTracer(ctx context.Context) error{
        conn, err := grpc.NewClient("jaeger:4317",
                grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
                return err
        }

        exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
        if err != nil {
                return err
        }

        res, err := resource.New(ctx,
                resource.WithAttributes(
                        semconv.ServiceName(ServiceName),
                ),
        )
        if err != nil {
                return err
        }

        prv = sdktrace.NewTracerProvider(
                sdktrace.WithBatcher(exp),
                sdktrace.WithResource(res),
        )

        otel.SetTracerProvider(prv)
        tr = prv.Tracer("tracer")

        return nil
}

func main() {
        ctx := context.Background()
        if err := initTracer(ctx);err != nil {
                log.Fatal("init tracer", err)
        }

        ctx, span := tr.Start(ctx, "root span")
        test(ctx, 1)
        test(ctx, 2)
        span.End()

        if err := prv.Shutdown(ctx); err != nil {
                log.Fatal("failed shutdown", err)
        }
}

func test(ctx context.Context, count int) {
        ctx, span := tr.Start(ctx, fmt.Sprintf("span-%d", count))
        defer span.End()
        num := rand.Intn(5)+1
        time.Sleep(time.Duration(num)*time.Second)
        if num%2 == 0 {
                span.SetStatus(codes.Error, errSpan.Error())
                span.RecordError(errSpan)
        }
}
```
И запустим:
```console
$ docker compose up -d --build
[+] Running 2/2
 ✔ Container test    Started                                                          0.6s
 ✔ Container jaeger  Running                                                          0.0s
```

После чего может появиться трейс и спан с ошибкой:

![](img/jaeger7.png)

![](img/jaeger8.png)

А в деталях спана можно увидеть информацию об ошибке:

![](img/jaeger9.png)

## Nested Spans
Трассировку можно также производить по стеку вызовов, в `golang` это делается передачей
информации о спане через контекст. Добавим в нашу функцию `test` возможность рекурсивного
запуска с передачей контекста.
```golang
package main

import (
        "context"
        "errors"
        "fmt"
        "log"
        "math/rand"
        "time"

        "go.opentelemetry.io/otel"
        "go.opentelemetry.io/otel/codes"
        "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
        "go.opentelemetry.io/otel/sdk/resource"
        sdktrace "go.opentelemetry.io/otel/sdk/trace"
        semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
        "go.opentelemetry.io/otel/trace"
        "google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
)

const (
        ServiceName = "service"
)

var (
        prv *sdktrace.TracerProvider
        tr trace.Tracer
        errSpan = errors.New("span error")
)

func initTracer(ctx context.Context) error{
        conn, err := grpc.NewClient("jaeger:4317",
                grpc.WithTransportCredentials(insecure.NewCredentials()),
        )
        if err != nil {
                return err
        }

        exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
        if err != nil {
                return err
        }

        res, err := resource.New(ctx,
                resource.WithAttributes(
                        semconv.ServiceName(ServiceName),
                ),
        )
        if err != nil {
                return err
        }

        prv = sdktrace.NewTracerProvider(
                sdktrace.WithBatcher(exp),
                sdktrace.WithResource(res),
        )

        otel.SetTracerProvider(prv)
        tr = prv.Tracer("tracer")

        return nil
}

func main() {
        ctx := context.Background()
        if err := initTracer(ctx);err != nil {
                log.Fatal("init tracer", err)
        }

        ctx, span := tr.Start(ctx, "root span")

        test(ctx, 1)

        test(ctx, 3)

        test(ctx, 1)

        span.End()

        if err := prv.Shutdown(ctx); err != nil {
                log.Fatal("failed shutdown", err)
        }
}

func test(ctx context.Context, count int) {
        if count < 1 {
                return
        }
        ctx, span := tr.Start(ctx, fmt.Sprintf("span-%d", count))
        defer span.End()
        test(ctx, count - 1)
        num := rand.Intn(5)+1
        time.Sleep(time.Duration(num)*time.Second)
        if num%2 == 0 {
                span.SetStatus(codes.Error, errSpan.Error())
                span.RecordError(errSpan)
        }
        log.Println("called test", count, num)
}
```
И запустим:
```console
$ docker compose up -d --build
[+] Running 2/2
 ✔ Container test    Started                                                          0.5s
 ✔ Container jaeger  Running                                                          0.0s
```

После чего можем наблюдать в трейсе вложенность спанов:
![](img/jaeger10.png)

Черной полосой в трейсе указывается `critical path`, который показывает в каких
местах при обработке вносится задержка. Также в информации об ошибке можно увидеть
на какой секунде она произошла при обработке данного запроса.

![](img/jaeger11.png)

Таким образом jaeger позволяет анализировать работу приложения, находя проблемные
и узкие места.


[jaeger]:https://www.jaegertracing.io/docs/1.56/
[docker]:https://docs.docker.com/engine/
[docker-compose]:https://docs.docker.com/compose/
