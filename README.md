# HTTP File Server

## Общее описание сервера

HTTP File Server - сервер, предостваляющий http api для загрузки, получения и удаления файла по хэшу (md5).
Все файлы хранятся в директории, которая задаётся через конфигурационный файл.

## HTTP API

### Health

```http
GET /health
```

Ответ вида:

```http
200 OK
```

### Upload

Загрузка файла на сервер. Файл передается через форму в поле upload.

```http
POST /upload
Content-Type: multipart/form-data
```

Ответ вида:

```http
200 OK

"file hash"
```

### Download

Получение файла с сервера.

```http
POST /download/{file_hash}
```

Ответ вида:

```http
200 OK

"file data"
```

### Delete

Удаление файла с сервера.

```http
POST /delete/{file_hash}
```

Ответ вида:

```http
200 OK

ok
```

## Сборка

1. Склонировать репозиторий.
2. Собрать проект

```bash
git clone ssh://git@gitlab2.sqtools.ru:10022/internals/sqcdn/api.git
make build
```

### Локальный запуск

1. Отредактировать `local.yml`. (Скорее всего потребуется сменить только порт, по-умолчанию порт 8080)
2. Запустить сервис.

```bash
vim local.yml
make run
```

P.s.: local.yml и shared/static не в .gitignore для удобства.
