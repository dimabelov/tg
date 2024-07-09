# Цель

Концептуально любой сервис можно разбить на слои:

1. Описание контракта
2. Реализация бизнес-логики
3. Логирование запросов
4. Метрики сервиса
5. Трассировка запросов
6. Транспорт
7. Клиент для интеграции с сервисом

Генератор `tg` предназначен для того, чтобы избавить разработчика от необходимости заниматься рутиной.

Для реализации сервиса разработчику достаточно описать лишь будущий контракт в виде интерфейса на языке `Go` и снабдить
аннотациями в виде специфичных для `tg` комментариев.
Остальная рутинная работа по генерации всех слоёв будет выполнена `tg`, что позволяет сосредоточиться на реализации
единственной ценности сервиса - `бизнес-логике`.

В данный момент для `tg` основным видом транспорта является [jsonRPC 2.0](https://www.jsonrpc.org/specification), но
поддерживается также генерация простого `HTTP` транспорта.
В качестве основы для транспортного уровня, был выбран [go-fiber](https://docs.gofiber.io), основанный
на [fasthttp](https://github.com/valyala/fasthttp), как альтернативе стандартной библиотеки `net/http`, превосходящий
оригинал по скорости более чем в 10 раз.

# Шаблон сервиса

Инициализация через шаблон не является обязательным шагом для использования `tg`, но позволяет упростить работу, в
случае создания сервиса с нуля.

Для генерации сервиса из шаблона с нуля, можно воспользоваться командой `init`:

```bash
tg init -module <go module name> -service <service name> <project name>
```

В результате, в папке `<project name>` будет сгенерирован работоспособный шаблон проект сервиса.

# Описание контракта

Источником истины для `tg` является интерфейс на языке `Go`, снабжённый аннотациями.

К методам интерфейса предъявляются следующие требования:

1. Все аргументы и возвращаемые значения методов интерфейса должны быть именованными. Эти имена, по-умолчанию, будут
   использованы как ключи на транспортном уровне.
2. Первым аргументом метода должен быть `context`, а последним возвращаемым значением - `error`.

```Go
// @tg jsonRPC-server log metrics trace  
type Some interface {
Method(ctx context.Context, arg1 string, arg2 int) (ret1 int, ret2 float64, err error)
}

```

Этого описания достаточно, что генерации сервиса, предоставляющего публичный метода `Some.Method`, посредством [jsonRPC 2.0](https://www.jsonrpc.org/specification) транспорта.

Запрос `jsonRPC` для метода `Method` интерфейса `Some` будет выглядеть следующим образом:

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "some.method",
  "params": {
    "arg1": "v",
    "arg2": 2
  }
}
```

Ответ:

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "result": {
    "ret1": 2,
    "ret2": 2
  }
}
```

# Сервер

## Генерация кода

Для генерации транспорта, необходимо выполнить команду:

```bash
tg transport --services . --out ../internal/transport
```

Для генерации документации в формате [openAPI](https://swagger.io/docs/specification/about/), необходимо выполнить
команду:

```bash
tg swagger --services . --outFile ../api/swagger.yaml
```

Где,

`services` - путь до папки с интерфейсом (в норме для `tg` эта папка является рабочей)
`outPath` - путь, где будет сохранён результат
`outPackage` - путь, где будет сохранён `package.json` с описанием `npm` пакета

Хорошей практикой считается использование утилиты `goimports`, после генерации:

```bash
goimports -l -w ../internal/transport
```

## Инициализация сервера

Для инициализации сервера, необходимо перечислить сервисы (интерфейсы, описанные [ранее](/#Описание контракта)), которые
он будет обслуживать и запустить его любым доступным способом, согласно
документации [go-fiber](https://docs.gofiber.io/api/app#listen)

Пример инициализации:

```Go
...

svcSome := some.New()

options := []transport.Option{
   transport.Use(cors.New()),
   transport.WithRequestID("X-Request-Id"),
   transport.Some(transport.NewSome(svcSome)),
}

srv := transport.New(log.Logger, options...).WithMetrics().WithLog()

srv.ServeHealth(config.Service().HealthBind, "OK")
srv.ServeMetrics(log.Logger, "/", config.Service().MetricsBind)

go func () {
   log.Info().Str("bind", config.Service().Bind).Msg("listen on")
   if err := srv.Fiber().Listen(config.Service().Bind); err != nil {
        log.Panic().Err(err).Msg("server error")
   }  
}()

...
```

Как видно из примера, в списке опций можно передавать не только сервисы, сгенерированные из интерфейсов, но и
вспомогательные обработчики.
Метод `transport.Use` поддерживает все возможности, предоставляемые [go-fiber](https://docs.gofiber.io/api/app). С
перечнем готовых мидлвар можно ознакомиться [здесь](https://docs.gofiber.io/category/-middleware).

Дополнительно можно указать следующие опции:

#### SetFiberCfg(cfg fiber.Config)

Опция позволяет управлять конфигурацией `go-fiber`, согласно [документации](https://docs.gofiber.io/api/fiber#config).

Пример:

```Go
fiberConfig := fiber.Config{
   Prefork: true,
   CaseSensitive: true,
   StrictRouting: true,
   ServerHeader: "Fiber",
   AppName: "Some Test App v1.0.1",
}

...

options := []transport.Option{
   transport.SetFiberCfg(fiberConfig),
   transport.Use(cors.New()),
   transport.WithRequestID("X-Request-Id"),
   transport.Some(transport.NewSome(svcSome)),
}

srv := transport.New(log.Logger, options...).WithMetrics().WithLog()

...

```

#### SetReadBufferSize(size int)

Опция позволяет указать размер буфера чтения в байтах (по умолчанию 4096).

#### SetWriteBufferSize(size int)

Опция позволяет указать размер буфера записи в байтах (по умолчанию 4096).

#### MaxBodySize(max int)

Опция позволяет указать максимальный размер тела запроса в байтах (по умолчанию 4 194 304).

#### MaxBatchSize(size int)

Опция позволяет указать максимальное количество запросов, которые можно передать за раз
в [батче](https://www.jsonrpc.org/specification#batch) (по умолчанию 100).

#### MaxBatchWorkers(size int)

Опция позволяет указать максимальное количество обработчиков, которые будут запускаться параллельно для каждого батч
запроса (по умолчанию 10).

#### ReadTimeout(timeout time.Duration)

Опция позволяет указать таймаут чтения для запросов (по умолчанию `unlimited`).

#### WriteTimeout(timeout time.Duration)

Опция позволяет указать таймаут записи для запросов (по умолчанию `unlimited`).

#### WithRequestID(headerName string)

Опция позволяет указать заголовок из которого будет извлекаться идентификатор запроса. Его будет логироваться с
ключом `requestID`, передаваться в трассировку и транслироваться в ответе с тем же заголовком.

# Клиент

## Генерация кода

Для генерации `Go` клиента, необходимо выполнить команду (поддерживается генерация клиента для [jsonRPC 2.0](https://www.jsonrpc.org/specification)) :

```bash
tg client -go --services . --outPath ../pkg/clients/go
```

Для генерации `javaScript` клиента, необходимо выполнить команду:

```bash
tg client client -js --services . --outPath ../pkg/clients/js --outPackage ../
```

Где,

`services` - путь до папки с интерфейсом (в норме для `tg` эта папка является рабочей)
`outPath` - путь, где будет сохранён результат
`outPackage` - путь, где будет сохранён `package.json` с описанием `npm` пакета

Хорошей практикой считается использование утилиты `goimports`, после генерации:

```bash
goimports -l -w ../pkg/clients/go
```

## Инициализация клиента

Для инициализации клиента, необходимо указать адрес сервера.

Пример инициализации:

```Go
...

cli := some.New("http://127.0.0.1:9000")

...
```

Где, `cli` будет общим клиентом для всех интерфейсов, которые участвовали в генерации.

Чтобы получить клиента для конкретного интерфейса, необходимо его извлечь соответствующим методом, как указано ниже:

```Go
...

cli := some.New("http://127.0.0.1:9000")
someCli := cli.Some()

...
```

При инициализации клиента можно указать следующие опции:

#### DecodeError(decoder ErrorDecoder)

Опция позволяет указать декодер, с помощью которого можно получить нужные типы ошибок. Хорошей практикой является
экспорт декодера из репозитория, предоставляющего клиента.

`ErrorDecoder` представляет собой функцию со следующей сигнатурой:

```Go
type ErrorDecoder func (errData json.RawMessage) error
```

По умолчанию, если не указал декодер явно, ошибки преобразуются к структуре, имплементирующей интерфейс `error` вида:

```Go
type errorJsonRPC struct {
   Code    int         `json:"code"`
   Message string      `json:"message"`
   Data    interface{} `json:"data,omitempty"`
}

func (err errorJsonRPC) Error() string {
    return err.Message
}
```

#### LogRequest()

Опция, включающая логирование всех запросов клиента в формате `curl`.

#### LogOnError()

Опция, включающая логирование всех запросов клиента в формате `curl`, если в ответ получена ошибка.

#### Headers(headers ...any)

Опция, позволяющая извлечь данные из контекста запроса и передать их через перечисленные заголовки.
В качестве параметров принимаются ключи контекста. Это могут быть как простые строки, так и любые типы, имплементирующие
интерфейс `fmt.Stringer`.

#### ConfigTLS(tlsConfig \*tls.Config)

Опция, позволяющая установить собственную конфигурацию `TLS` для клиента.
Может понадобиться, например, когда на сервере используется самоподписанный сертификат и нужно выключить его проверку.

## clientWithCB

- интерфейс

Включает генерацию `circuit breaker` для методов интерфейса.

#### CircuitBreaker(cfg cb.Settings)

Опция, позволяющая установить собственную конфигурацию для `circuit breaker`:

```Go
type Settings struct {
   MaxRequests   uint32
   Interval      time.Duration
   Timeout       time.Duration
   ReadyToTrip   func (counts Counts) bool
   OnStateChange func (name string, from State, to State)
   IsSuccessful  func (err error) bool
}
```

Где,

`MaxRequests` — это максимальное количество запросов, которым разрешено пройти, когда `circuit breaker` полуоткрыт.
Если `MaxRequests` равно `0`, `circuit breaker` разрешает только `1` запрос.

`Interval` — это циклический период закрытого состояния, в течение которого прерыватель цепи очищает внутренние
счетчики, описанные далее в этом разделе. Если `Interval` равен `0`, `circuit breaker` не очищает внутренние счетчики во
время закрытого состояния.

`Timeout` — это период открытого состояния, по истечении которого состояние `circuit breaker` становится полуоткрытым.
Если `Timeout` равен `0`, значение тайм-аута  `circuit breaker` устанавливается равным 60 секундам.

`ReadyToTrip` вызывается с `Counts` всякий раз, когда запрос завершается сбоем в закрытом состоянии. Если `ReadyToTrip`
возвращает `true`, `circuit breaker` будет переведен в открытое состояние. Если `ReadyToTrip` равен `nil`,
используется `ReadyToTrip` по умолчанию. `ReadyToTrip` по умолчанию возвращает `true`, когда количество последовательных
сбоев превышает `5`.

`OnStateChange` вызывается всякий раз, когда изменяется состояние `circuit breaker`.

`IsSuccessful` вызывается с ошибкой, возвращенной из запроса. Если `IsSuccessful` возвращает `true`, ошибка считается
нормальным поведением. В противном случае ошибка засчитывается как сбой. Если `IsSuccessful` равен `nil`,
используется `IsSuccessful` по умолчанию, который возвращает `false` для всех не нулевых ошибок.

#### Cache(cache cache)

Опция, позволяющая включить `fallback` кэширование для `circuit breaker`. В качестве параметра принимается любой
объект, имплементирующий интерфейс:

```Go
type cache interface {
   SetTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) (err error)
   GetTTL(ctx context.Context, key string, value interface{}) (createdAt time.Time, ttl time.Duration, err error)
}
```

При установленной опции, каждый успешный запрос кэшируется с ключом равным хэшу от параметров запроса. Таким образом,
при срабатывании `fallBack` обработчика `circuit breaker`, в ответе клиента вернётся результат последнего удачного
запроса, вместо ошибки.

#### FallbackTTL(ttl time.Duration)

Опция, устанавливающая время, на которое кэшируется последний успешный ответ, для `fallback` (по умолчанию 24 часа).

# # Аннотация

Аннотацией в терминах `tg` называется комментарий, оформленный специальным образом.
Целью аннотаций является указание генератору параметров и настроек, специфичных для конкретного сервиса.

Аннотации могут быть определены на разных уровнях - пакет, интерфейс, метод интерфейса и на уровне типов.

Аннотации имеют следующие уровни определения:

- на уровне пакета, действуют на все методы всех интерфейсов в этом пакете
- на уровне интерфейса действуют на все методы этого интерфейса
- на уровне метода, действуют только на этот метод

В случае конфликтов, приоритет имеют аннотации с наименьшей зоной действия.

Аннотации имею следующий формат:

```Go
// @tg <имя>=<значение>
```

В случае, когда аннотации имею смысл флагов, значение может не указываться. Несколько аннотаций может быть сгруппировано
в одной строке (разделитель пробел). Например:

```Go
// @tg http-prefix=v1 jsonRPC-server log metrics trace
```

Следующая запись синонимична пред идущей:

```Go
// @tg http-prefix=v1 
// @tg jsonRPC-server 
// @tg log metrics trace
```

## log

- модуль
- интерфейс

Включает генерацию логирования.

## trace

- модуль
- интерфейс

Включает генерацию трассировку методов интерфейсов.

## metrics

- модуль
- интерфейс

Включает генерацию метрик для методов интерфейсов.

## desc=\`краткое описание \`

- модуль
- интерфейс
- метод
- тип

Добавляет краткое описание той сущности, на уровне которой определён.
Используется, в том числе, при генерации документации в формате [openAPI](https://swagger.io/specification).
В случае генерации `web` клиента, описание на уровне пакета используется в `package.json` для описания `npm` пакета.

## summary=\`Детальное описание метода <br/> Вторая строка описания с **жирным тестом**.\`

- метод

Детальное писание метода в генерируемой
документации [openAPI](https://swagger.io/docs/specification/paths-and-operations).
Поддерживает перенос строки и прочие возможности форматирования `openAPI`.

## <имя переменой в сигнатуре функции>.tags=<тэг>:<значение>|<тэг>:<значение>

Позволяет указать дополнительные теги в `exchange` структурах метода интерфейса или переопределить существующие.
Типичный пример - сокрытие чувствительных данных поля и логах:

```Go
...
// @tg token.tags=dumper:hide,md
Login(ctx context.Context, token string) (cookie *types.Cookie, err error)
...
```

В результате, в логах, середина строки `token` будет заменена на символы `*`.

## type=<тип>

- тип

Указывает тип поля в генерируемой документации, согласно
спецификации [openAPI](https://swagger.io/docs/specification/data-models/data-types/)

## enums=val1,val2,val3

- тип

Для поля можно перечислить список возможных значений.

## format=uuid

- тип

Указывает формат поля в генерируемой документации, согласно
спецификации [openAPI](https://swagger.io/docs/specification/data-models/data-types/)

## example=someExampleValue

- тип

Указывает пример значения поля в генерируемой документации, согласно
спецификации [openAPI](https://swagger.io/docs/specification/data-models/data-types/)

## http-args=<имя переменой в сигнатуре функции>|<имя ключа в URL>

- метод

Определяет маппинг параметров, переданных в параметрах `URL`, в аргументы метода.

## http-path=/<URL путь>/:<имя переменой в сигнатуре функции>

- метод

Определяет маппинг параметров, переданных в пути `URL`, в аргументы метода.
Переменные, которые попали в маппинг, исключаются из `exchange` структур.

## http-prefix=<префикс пути в URL>

- модуль
- интерфейс

Задаёт префикс к пути `URL` методов.
Формула пути, по которому доступен метода выглядит следующим образом:

`/globalPrefix/prefix/methodPath`

Где,

`globalPrefix` - префикс, объявленный на уровне пакета
`prefix` - префикс, объявленный на уровне интерфейса
`methodPath` - имя интерфейса/имя метода, но может быть переопределён через аннотацию `http-path` v

## http-headers=<имя переменой в сигнатуре функции>|<заголовок>

- метод

Определяет маппинг параметров, переданных в заголовках запроса, в аргументы/результаты метода.
Переменные, которые попали в маппинг, исключаются из `exchange` структур.

## http-cookies=<имя переменой в сигнатуре функции>|<заголовок>

- метод

Определяет маппинг параметров, переданных в `cookie` запроса, в аргументы/результаты метода.
Переменные, которые попали в маппинг, исключаются из `exchange` структур.

## http-method=<HTTP метод>

- метод

Указывает `HTTP` метод, который будет использован для доступа к методу интерфейса.

## http-success=<HTTP код>

- метод

Указывает `HTTP` код ответа, который будет считаться успешным, при доступе к методу интерфейса.

## packageJSON=\`<имя пакета>\`

- модуль

Переопределяет пакет, который будет использоваться для кодирования/декодирования `JSON`.
Используется для случаев, когда нужно особое поведение кодека или есть более оптимальный кодек, предоставляющий тот же
интерфейс, что и стандартный `encoding/json`.

Например, `github.com/seniorGolang/json` возвращает пустые срезы как `[]`, а не как `nil`, в стандартном `encoding/json`
и имеет ряд других оптимизаций по скорости работы.

## uuidPackage=\`<имя пакета>\`

- модуль

Переопределяет пакет, который будет использоваться для кодирования/декодирования `UUID`, при конвертации.
В замещающем пакете должен быть определён метод `Parse(s string) (UUID, error)`.
По умолчанию используется пакет `github.com/google/uuid`.

## swaggerTags=<тэг1,тэг2>

- интерфейс

Указывает теги для описания интерфейса в
формате [openAPI](https://swagger.io/docs/specification/grouping-operations-with-tags).

## log-skip=<имя переменой в сигнатуре функции>,<имя переменой в сигнатуре функции>

- метод

Указывает какие переменных из сигнатуры метода нужно исключить из логирования.

## deprecated

- метод

Помечает метод как - `deprecated` в документации [openAPI](https://swagger.io/docs/specification/paths-and-operations).

## tagNoOmitempty

- модуль
- интерфейс

По умолчанию для всех полей методов включен тег `omitempty`, что исключает пустые поля из ответа.
Может существенно сэкономить трафик, но не всегда `fronend` готов к такому поведение.
Но иногда такое поведение может быть неожиданным для `fronend`.

## handler=<модуль Go>:<Тип>

- метод

Переключает работу метода в так называемый `кастомный` режим.
Это означает, что для этого метода не генерируется никаких обработчиков, а используется тот, который указан в аннотации.

Кастомный обработчик должен иметь следующую сигнатуру:

```Go
CustomHandler(ctx *fiber.Ctx, svc <тип интерфеса, к которому принадлежит метод>) (err error)
```

Рекомендуется использовать кастомные обработчики только в крайнем случае, когда невозможно имплементировать метод
другими способами.
Т.к. то, что происходит в этом обработчик никак не формализовано, то логи, метрики и прочее нужно реализовать
самостоятельно.

## requestContentType=<mime тип>

- метод

Позволяет указать `mime` тип, который ожидается в запросе.
По умолчанию `application/ json`.

## responseContentType=<mime тип>

- метод

Позволяет указать `mime` тип, который ожидается в ответе.
По умолчанию `application/ json`.

## security=\`bearer\`

- модуль

Позволяет указать в документации [openAPI](https://swagger.io/docs/specification/authentication), что используется
авторизация.

## servers=`<адрес>;<имя>|<адрес>;<имя>`

- модуль

Указывает генератору документации список адресов, по которым доступен сервис и их человеко читаемые имена.

## version=<версия сервиса>

- модуль

Указывает генератору документации текущую версию сервиса.

## title=\`<заголовок документации к сервису>\`

- модуль

Указывает генератору документации заголовок к документации сервиса.

## author=\`автор сервиса\`

- модуль

Указывает генератору `NPM` модуля автора сервиса.

## npmRegistry=<адрес репозитория NPM>

- модуль

Указывает генератору `NPM` модуля адрес репозитория, где будет опубликован клиент.

## npmName=<имя пакета NPM>

- модуль

Указывает генератору `NPM` модуля имя пакета, под которым будет опубликован клиент.

## npmPrivate=<true|false>

- модуль

Указывает генератору `NPM` модуля является ли он публичным или приватным.

## license=<вид лицензии>

- модуль

Указывает генератору `NPM` модуля под какой лицензией он распространяется.

## http-server

- интерфейс

Включает генерацию `HTTP` сервера на базе интерфейса.

## jsonRPC-server

- интерфейс

Включает генерацию [jsonRPC 2.0](https://www.jsonrpc.org/specification) сервера на базе интерфейса.

# Метрики

## RequestCount Counter

RequestCount = prometheus.NewCounterFrom(prometheus.CounterOpts{  
Help:      "Number of requests received",  
Name:      "count",  
Namespace: "service",  
Subsystem: "requests",  
}, []string{"method", "service", "success"})

## RequestCountAll Counter

RequestCountAll = prometheus.NewCounterFrom(prometheus.CounterOpts{  
Help:      "Number of all requests received",  
Name:      "all_count",  
Namespace: "service",  
Subsystem: "requests",  
}, []string{"method", "service"})

## RequestLatency Histogram

RequestLatency = prometheus.NewHistogramFrom(prometheus.HistogramOpts{  
Help:      "Total duration of requests in microseconds",  
Name:      "latency_microseconds",  
Namespace: "service",  
Subsystem: "requests",  
}, []string{"method", "service", "success"})