# Simple chat

## Для запуска потребуются следующие библиотеки:

1. https://github.com/gin-gonic/gin
2. https://github.com/olahol/melody
3. https://github.com/gorilla/websocket

## Как запустить

После установки библиотек, делаем следующее:

1. Для запуска сервера: `go run . -type=server -addr=:8088
   `
2. Для запуска polling клиента: `go run . -type=polling   `
3. Для запуска websocket клиента: `go run . -type=client `

Так же можно установить websocket соединение через postman(https://blog.postman.com/postman-supports-websocket-apis/)
`ws://localhost:8088/ws` в качестве адреса, в параметры например:
`
{
"type": 1,
"name": "Anton",
"text": "hello world! 2"
}
`
