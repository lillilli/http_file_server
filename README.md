# HTTP File Server

Test project.

## Description

HTTP File Server - server, providing http API for downloading, uploading and removing files by their md5 hash string.
All files store in directory, which is setting up from the config file.

## HTTP API

### Health

```http
GET /health
```

Response:

```http
200 OK
```

### Upload

Uploading file to the server. File is setting up by the form, field "upload".

```http
POST /upload
Content-Type: multipart/form-data
```

Response:

```http
200 OK

"file hash"
```

### Download

Downloading file from the server.

```http
POST /download/{file_hash}
```

Response:

```http
200 OK

"file data"
```

### Delete

Removing file by his md5 hash.

```http
POST /delete/{file_hash}
```

Response:

```http
200 OK

ok
```

## Local launch

### Requirements

You need to have vgo installed.

```bash
go get -u golang.org/x/vgo
```

### Launch

1. Clone the repository.
2. Install dependencies, create the config file.
3. Create static files directory, based on config file.
4. Launch the project.

```bash
git clone https://github.com/lillilli/http_file_server.git && cd http_file_server
make setup && make config
mkdir -p shared/static
make run
```

### Docker

1. Clone the repository.
2. Make image (need some time).
3. Launch image (will be available on localhost:8080).

```bash
git clone https://github.com/lillilli/http_file_server.git && cd http_file_server
make image
make run:image
```
