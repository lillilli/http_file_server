# build stage
FROM golang:1.11 as builder

ENV SERVICE_NAME http_file_server
ENV PKG github.com/lillilli/http_file_server

RUN mkdir -p /go/src/${PKG}
WORKDIR /go/src/${PKG}

COPY . .

RUN go get -u golang.org/x/vgo

RUN make setup && make config
RUN cd cmd/${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -a -installsuffix cgo -o ${SERVICE_NAME}

FROM alpine:3.7
RUN apk --no-cache add ca-certificates

WORKDIR /root/

ENV SERVICE_NAME http_file_server
ENV PKG github.com/lillilli/http_file_server

COPY --from=builder /go/src/${PKG}/cmd/${SERVICE_NAME}/${SERVICE_NAME} .

RUN echo -e "StaticDir: shared/static" >> local.yml
RUN cat local.yml
RUN ls
RUN mkdir -p shared/static
RUN ls

# ENTRYPOINT ["./http_file_server -config=local.yml"]
