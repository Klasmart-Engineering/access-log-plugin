all: clean build

build:
	go build -buildmode=plugin -o access-log-plugin.so ./src/main.go

clean:
	go clean

login:
	echo ${GH_PAT} | docker login ghcr.io -u USERNAME --password-stdin

b:
	docker buildx build -t gateway .

r:
	docker-compose down --remove-orphans && \
    docker-compose rm -f -v && \
	docker-compose up && \
	docker-compose down --remove-orphans && \
	docker-compose rm -f -v

br: b r

test-unit:
	go test -v ./test/unit/...

test-integration:
	docker-compose -f docker-compose.yaml down --remove-orphans && \
    docker-compose -f docker-compose.yaml rm -fv && \
    rm -rf ./postgres-integration-data && \
    make b && \
    echo "print images" && \
    docker images ls && \
    docker-compose -f docker-compose.yaml up -d && \
    echo "Starting integration tests" && \
    go clean -testcache && \
    go test -v ./test/integration/... && \
    echo "Finished integration tests" && \
    docker-compose -f docker-compose.yaml down --remove-orphans && \
    docker-compose -f docker-compose.yaml rm -fv