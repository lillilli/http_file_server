SERVICE_NAME   := http_file_server
VERSION        := $(shell git describe --tags --always --dirty="-dev")
DATE           := $(shell date -u '+%Y-%m-%d-%H:%M UTC')
VERSION_FLAGS  := -ldflags='-X "main.Version=$(VERSION)" -X "main.BuildTime=$(DATE)"'
PKG            := github.com/lillilli/http_file_server
PKG_LIST       := $(shell go list ${PKG}/... | grep -v /vendor/)
CONFIG         := $(wildcard local.yml)
NAMESPACE	   := "default"

# Проверка, задана ли переменная окружения REGISTRY
ifdef REGISTRY
	REGISTRY := $(REGISTRY)/
endif

# Флаг отвечающий за подробный вывод результатов работы make, раскоментируйте чтобы включить
V := 1

.PHONY: clean test

.PHONY: all
all: setup test build

# Установка необходимых для сборки или тестирования утилит и зависимостей
.PHONY: setup
setup: clean
	@echo "Setup..."
	# e.g. sudo apt-get install clang

# Сборка исполняемого файла сервиса
.PHONY: build
build:
	@echo "Building..."
	$Q cd src/cmd/$(SERVICE_NAME) && go build

# Запуск сервиса с локальным конфигом
.PHONY: run
run: build
	@echo "Running..."
	$Q cd src/cmd/$(SERVICE_NAME) && ./$(SERVICE_NAME) -config=../../../local.yml

# Запуск тестов
.PHONY: test
test:
	@echo "Testing..."
	$Q go test -race ${PKG_LIST}

# Отображение списка существующих тэгов
.PHONY: tags
tags:
	@echo "Listing tags..."
	$Q @git tag

# Сборка Docker-образа
.PHONY: image
image:
	@echo "Docker Image Build..."
	$Q docker build -t $(REGISTRY)$(SERVICE_NAME):$(VERSION) .

.PHONY: push
push:
	@echo "Pushing Docker image..."
	$Q docker push $(REGISTRY)$(SERVICE_NAME):$(VERSION)

.PHONY: deploy
deploy:
	@echo "Deploy Docker image..."
	$Q cd deploy && helm upgrade --set image.tag=${VERSION} --namespace=${NAMESPACE} --install ${SERVICE_NAME} .

# Очистка окружения от временных файлов и т.п.
.PHONY: clean
clean:
	@echo "Clean..."
	$Q rm -f src/cmd/$(SERVICE_NAME)/$(SERVICE_NAME)

# Подсчет покрытия кода тестами
.PHONY: coverage
coverage:
	@echo "Calculating coverage..."
	$Q PKG=$(PKG) ./tools/coverage.sh;

Q := $(if $V,,@)