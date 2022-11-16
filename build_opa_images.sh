#!/bin/bash

set -o errexit

BUNDLE=${BUNDLE:-"true"}
REPO=${REPO:-"ccr-2ql0kyd9-pub.cnc.bj.baidubce.com"}
PREFIX=${PREFIX:-"bms-mesh"}
VERSION=${VERSION:-"latest-istio"}
BUNDLE_NAME=${BUNDLE_NAME:-"bundle"}
APP_NAME=${APP_NAME:-"opa"}
OPA_VERSION=${OPA_VERSION:-"v0.38.0"}
RUN_SH_PATH="${PWD}"/deployment/bundle

opa_envoy_plugin_name=${opa_envoy_plugin_name:-"opa-envoy-plugin"}
TEMP_ROOT=${PWD}/output-opa-istio
rm -rf "${TEMP_ROOT}"
mkdir -p "${TEMP_ROOT}"
BIN_DIR=${TEMP_ROOT}/bin
mkdir -p "${BIN_DIR}"

FILE_NAME="${OPA_VERSION}"-envoy.tar.gz
wget -P  "${TEMP_ROOT}" https://github.com/open-policy-agent/opa-envoy-plugin/archive/refs/tags/"${FILE_NAME}"

TARGET="${TEMP_ROOT}"/"${opa_envoy_plugin_name}"
mkdir -p "${TARGET}"
tar -zxvf "${TEMP_ROOT}/${FILE_NAME}" -C  "${TARGET}" --strip-components 1

BUILD_COMMIT=$("${TARGET}"/build/get-build-commit.sh)
BUILD_TIMESTAMP=$("${TARGET}"/build/get-build-timestamp.sh)
BUILD_HOSTNAME=$("${TARGET}"/build/get-build-hostname.sh)

LDFLAGS="-X github.com/open-policy-agent/opa/version.Version=${OPA_VERSION} \
  -X github.com/open-policy-agent/opa/version.Vcs=${BUILD_COMMIT} \
  -X github.com/open-policy-agent/opa/version.Timestamp=${BUILD_TIMESTAMP} \
  -X github.com/open-policy-agent/opa/version.Hostname=${BUILD_HOSTNAME}"

LINUX_AMD_NAME=amd64
LINUX_ARM_NAME=arm64
LINUX_AMD_NAME_OPA=opa
LINUX_ARM_NAME_OPA=opa

cd "${TARGET}"
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go generate ./... && go build -o "${BIN_DIR}"/"${LINUX_AMD_NAME}"/"${LINUX_AMD_NAME_OPA}"  -ldflags "${LDFLAGS}"  ./cmd/opa-envoy-plugin/...

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm64
go generate ./... && go build -o "${BIN_DIR}"/"${LINUX_ARM_NAME}"/"${LINUX_ARM_NAME_OPA}"  -ldflags "${LDFLAGS}"  ./cmd/opa-envoy-plugin/...
cd "${TEMP_ROOT}"

unset GOOS
unset GOARCH

if [ "${BUNDLE}" == "true" ]; then
  VERSION=${VERSION}-${BUNDLE_NAME}
  cp "${RUN_SH_PATH}"/run.sh "${TEMP_ROOT}"
  cat > "${TEMP_ROOT}"/Dockerfile_"${LINUX_AMD_NAME}" <<EOF
FROM alpine:3.15

RUN apk update \
 && apk upgrade \
 && apk add --no-cache bash \
 bash-doc \
 bash-completion \
 && rm -rf /var/cache/apk/* \
 && /bin/bash

WORKDIR /app
COPY  ./bin/${LINUX_AMD_NAME}/${LINUX_AMD_NAME_OPA} /app/${LINUX_AMD_NAME_OPA}
COPY  ./run.sh /app/run.sh
RUN chmod a+x /app/run.sh
CMD ["/app/run.sh"]
EOF

cat > "${TEMP_ROOT}"/Dockerfile_"${LINUX_ARM_NAME}" <<EOF
FROM alpine:3.15

RUN apk update \
 && apk upgrade \
 && apk add --no-cache bash \
 bash-doc \
 bash-completion \
 && rm -rf /var/cache/apk/* \
 && /bin/bash

WORKDIR /app
COPY  ./bin/${LINUX_ARM_NAME}/${LINUX_ARM_NAME_OPA} /app/${LINUX_ARM_NAME_OPA}
COPY  ./run.sh /app/run.sh
RUN chmod a+x /app/run.sh
CMD ["/app/run.sh"]
EOF

else
  cat > "${TEMP_ROOT}"/Dockerfile_"${LINUX_AMD_NAME}" <<EOF
FROM alpine:3.15

RUN apk update \
 && apk upgrade \
 && apk add --no-cache bash \
 bash-doc \
 bash-completion \
 && rm -rf /var/cache/apk/* \
 && /bin/bash

WORKDIR /app
COPY  ./bin/${LINUX_AMD_NAME}/${LINUX_AMD_NAME_OPA} /app/${LINUX_AMD_NAME_OPA}
ENTRYPOINT ["/app/${LINUX_AMD_NAME_OPA}"]
CMD ["run"]
EOF

cat > "${TEMP_ROOT}"/Dockerfile_"${LINUX_ARM_NAME}" <<EOF
FROM alpine:3.15

RUN apk update \
 && apk upgrade \
 && apk add --no-cache bash \
 bash-doc \
 bash-completion \
 && rm -rf /var/cache/apk/* \
 && /bin/bash

WORKDIR /app
COPY  ./bin/${LINUX_ARM_NAME}/${LINUX_ARM_NAME_OPA} /app/${LINUX_ARM_NAME_OPA}
ENTRYPOINT ["/app/${LINUX_ARM_NAME_OPA}"]
CMD ["run"]
EOF
fi

BASE_IMAGE="${REPO}/${PREFIX}/${APP_NAME}:${VERSION}"
LINUX_AMD_NAME_IMAGE=${BASE_IMAGE}-${LINUX_AMD_NAME}
LINUX_ARM_NAME_IMAGE=${BASE_IMAGE}-${LINUX_ARM_NAME}

echo "${BASE_IMAGE}"
echo "${LINUX_AMD_NAME_IMAGE}"
echo "${LINUX_ARM_NAME_IMAGE}"

docker buildx build -t "${LINUX_AMD_NAME_IMAGE}" .  --platform linux/amd64  -f Dockerfile_"${LINUX_AMD_NAME}"  --load
docker buildx build -t "${LINUX_ARM_NAME_IMAGE}" .  --platform linux/arm64  -f Dockerfile_"${LINUX_ARM_NAME}"  --load

docker push "${LINUX_AMD_NAME_IMAGE}"
docker push "${LINUX_ARM_NAME_IMAGE}"

docker  manifest create  "${BASE_IMAGE}" "${LINUX_AMD_NAME_IMAGE}" "${LINUX_ARM_NAME_IMAGE}" --amend
docker  manifest push   "${BASE_IMAGE}"

# kubectl   get pods | grep Evicted |awk '{print$1}'|xargs kubectl   delete pods
# 如果遇到 manifest 创建镜像 layer 层不存在，则先删除该 manifest 镜像
# docker manifest rm ccr-2ql0kyd9-pub.cnc.bj.baidubce.com/bms-mesh/opa:latest-istio-bundle
# 批量清除 opa 镜像
# docker rmi $(docker images | grep opa |awk '{print $3}')

cd ../
rm -rf output-opa-istio

