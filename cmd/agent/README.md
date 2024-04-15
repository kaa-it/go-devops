# cmd/agent

В данной директории будет содержаться код Агента, который скомпилируется в бинарное приложение

## Переменные среды 

| Name              | Description                       | Default Value    |
|-------------------|-----------------------------------|------------------|
| `ADDRESS`         | Base address for metrics server   | `127.0.0.1:8080` |
| `POLL_INTERVAL`   | Poll metric interval in seconds   | `2`              |
| `REPORT_INTERVAL` | Report metric interval in seconds | `10`             |