BIN_DIR=output/bin
RELEASE_DIR=output/release
REL_OS_ARCH=amd64 arm64
REL_OS=linux
IMAGE_REPO?=ccr-2ql0kyd9-pub.cnc.bj.baidubce.com/bms-mesh
IMAGE_NAME?=opa-sidecar-webhook
IMAGE_TAG?=$(shell cat version/VERSION)

all: export GOPROXY="https://goproxy.cn"
all: clean push_multi_arch_images clean

clean:
	rm -rf output

bin-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o=${BIN_DIR}/amd64/${REL_OS}/${IMAGE_NAME} ./cmd/

bin-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o=${BIN_DIR}/arm64/${REL_OS}/${IMAGE_NAME} ./cmd/

bin: bin-linux-amd64 bin-linux-arm64

bin-linux-amd64-image:
	docker buildx build -t ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-amd64 --platform linux/amd64  -f build/Dockerfile.amd64 . --load

bin-linux-arm64-image:
	docker buildx build -t ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-arm64 --platform linux/arm64  -f build/Dockerfile.arm64 . --load

images: bin bin-linux-amd64-image bin-linux-arm64-image

image-linux-amd64:
	docker push ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-amd64

image-linux-arm64:
	docker push ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-arm64

push: images image-linux-amd64 image-linux-arm64

release: images
	mkdir -p ${RELEASE_DIR}
	docker save -o ${RELEASE_DIR}/${IMAGE_NAME}-${REL_OS}-amd64.tar ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-amd64
	docker save -o ${RELEASE_DIR}/${IMAGE_NAME}-${REL_OS}-arm64.tar ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-arm64

create_multi_arch_images: push
	docker manifest create ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-arm64 ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}-${REL_OS}-amd64 --amend

push_multi_arch_images: create_multi_arch_images
	docker manifest push ${IMAGE_REPO}/${IMAGE_NAME}:${IMAGE_TAG}
