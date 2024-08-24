Version=v1
BACKEND_APP_NAME=ds
BACKEND_SERVERJOB_NAME=ds-job

DOCKER_BUILD_IMAGE_TAG=ds-build

DOCKER_RUN_GAME_IMAGE_TAG=ds-game-run
DOCKER_RUN_DSC_IMAGE_TAG=ds-dsc-run
DOCKER_RUN_DSA_IMAGE_TAG=ds-dsa-run

K8S_WEB_URL=""

GitCount= $(shell git rev-list --count HEAD)
GitSHA1= $(shell git log -n1 --format=format:"%H")
branch=$(shell git symbolic-ref --short HEAD)
Branch=$(subst /,_,$(branch))

.PHONY: build  buildimage test clean vet deployimage dockertest deployimagedesign gitversion assetsmd5 pb
gitversion:
	rm -rf common/civersion.go
	echo 'package common '  >>common/civersion.go
	echo 'const GitVersion = "${Branch}_${GitCount}_${GitSHA1}"' >>common/civersion.go
	echo 'const DSServerName = "${BACKEND_SERVERJOB_NAME}"'  >>common/civersion.go

assetsmd5:
	tool/gen_assets_md5.py ${SlotsConfigPath} ${SystemConfigPath}

build:
	make gitversion
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.buildbinary .
	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) > ${DOCKER_BUILD_IMAGE_TAG}_${Version}_${Branch}_${GitCount}.tar.gz

buildimage:
	make gitversion
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
			-f ./Dockerfile.buildbinary .

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile.run -
	docker tag $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_GAME_IMAGE_TAG):latest

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_DSC_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile_DSC.run -
	docker tag $(DOCKER_RUN_DSC_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_DSC_IMAGE_TAG):latest

	docker run --rm $(DOCKER_BUILD_IMAGE_TAG) \
		| docker build -t $(DOCKER_RUN_DSA_IMAGE_TAG):${Version}_${Branch}_${GitCount} -f Dockerfile_DSA.run -
	docker tag $(DOCKER_RUN_DSA_IMAGE_TAG):${Version}_${Branch}_${GitCount} $(DOCKER_RUN_DSA_IMAGE_TAG):latest


push_webserver:
	echo "Tagging docker image $(DOCKER_RUN_GAME_IMAGE_TAG) to $(K8S_WEB_URL)"
	docker tag $(DOCKER_RUN_GAME_IMAGE_TAG):${Version}_${Branch}_${GitCount} \
		$(K8S_WEB_URL):${Version}_${Branch}_${GitCount}
	echo "Start pushing"
	docker push $(K8S_WEB_URL):${Version}_${Branch}_${GitCount}


test: export DS_LOGLEVEL=info
test:
	go test  -cover -race -parallel 1 -p 1 -count=1 -v ./... | sed '/PASS/s//$(shell printf "\033[32mPASS\033[0m")/' | sed '/FAIL/s//$(shell printf "\033[31mFAIL\033[0m")/' | sed '/coverage/s//$(shell printf "\033[32mcoverage\033[0m")/'

unit_test:
	docker build --build-arg appname=$(BACKEND_APP_NAME) -t $(DOCKER_BUILD_IMAGE_TAG) \
		-f ./Dockerfile.test .
	docker-compose up -d --build
	docker logs -f ${DOCKER_BUILD_IMAGE_TAG} > unit_test.log
	docker-compose down

vet :
	go vet ./...

deployimage:
	make buildimage
	docker-compose down
	docker-compose up -d
pb:
	cd tools &&sed -i "s/\r//" gen_proto.sh &&chmod +x gen_proto.sh && ./gen_proto.sh
	
clean:
