CI_DOCKER_IMAGE ?= serjk/opensuse-go-pacemaker
WD_PATH ?=/opt/go/src/github.com/serjk/go-pacemaker/
PWD = $(shell pwd)

docker:
	docker build -t $(CI_DOCKER_IMAGE)  -f Dockerfile . | cat

push: docker
	docker push $(CI_DOCKER_IMAGE) | cat

test:
	docker run --rm -v $(PWD):$(WD_PATH) $(CI_DOCKER_IMAGE) make --debug=b -C $(WD_PATH) dockertest

dockertest:
	dep ensure
	go test ./...


.PHONY: test docker push dockertest