# 项目名称

开放策略代理（OPA）是一个开源的通用策略引擎。可参考[官网OPA](https://www.openpolicyagent.org/)

OPA 作为由 CNCF 托管的一个孵化项目，开放策略代理（OPA）是一个策略引擎，可用于应用程序实现细粒度的访问控制。 OPA
作为通用策略引擎，可以与微服务一起部署为独立服务。为了保护应用程序，必须先授权对微服务的每个请求，然后才能对其进行处理。 为了检查授权，微服务对 OPA 进行 API 调用，以确定请求是否被授权。

## opa

### opa sidecar 注入

namespace 命名空间级别、工作负载级别注入 opa-sidecar：csmopa-injection=enabled

工作负载级别添加注解：sidecar.csmopa.io/csmopa-injection="false"，可以在 ns 级别已配置 opa-sidecar 自动注入的前提条件下，不注入 opa-sidecar。

### API

opa 策略支持下发 configmap 与 bundle。

configmap 类型：

| annotation               |             示例值              |            说明            |
| ------------------------ |            ---------------- |  ---------------- |
| sidecar.csmopa.io/csmopa-injection    |    可选值，"true"、"false"           |    "false" 表示不注入 opa-sidecar    |
| sidecar.csmopa.io/rego_source_type    |    必填值，configmap              |       策略来源类型              |
| sidecar.csmopa.io/opa_image_url  |     可选值，openpolicyagent/opa:latest-istio    |     opa 镜像地址            |
| sidecar.csmopa.io/rego_source_cm  |     必填值，inject-policy    |       策略数据 configmap 名称               |
| sidecar.csmopa.io/rego_source_cm_key  |    必填值，policy.rego    |         configmap 策略数据 key 名称             |
| sidecar.csmopa.io/grpc_policy_path  |    必填值，istio/authz/allow    |     opa 代理暴露 grpc 服务依据的策略路径        |
| sidecar.csmopa.io/grpc_addr  |              可选值，:9898(默认值)    |    opa 代理暴露 grpc 服务地址                |
| sidecar.csmopa.io/addr_port  |              可选值，9899(默认值)    |    opa 自身监听的端口地址               |
| sidecar.csmopa.io/diagnostic_addr_port  |   可选值，9900(默认值)    |    opa 自身健康检查的端口地址                |

bundle 类型：

| annotation               |             示例值              |            说明            |
| ------------------------ |            ---------------- |  ---------------- |
| sidecar.csmopa.io/csmopa-injection    |    可选值，"true"、"false"           |    "false" 表示不注入 opa-sidecar    |
| sidecar.csmopa.io/rego_source_type    |    必填值，bundle              |       策略来源类型              |
| sidecar.csmopa.io/opa_image_url  |     可选值，openpolicyagent/opa:latest-istio    |     opa 镜像地址            |
| sidecar.csmopa.io/grpc_policy_path  |    必填值，istio/authz/allow    |     opa 代理暴露 grpc 服务依据的策略路径        |
| sidecar.csmopa.io/grpc_addr  |              可选值，:9898(默认值)    |    opa 代理暴露 grpc 服务地址                |
| sidecar.csmopa.io/addr_port  |              可选值，9899(默认值)    |    opa 自身监听的端口地址               |
| sidecar.csmopa.io/diagnostic_addr_port  |   可选值，9900(默认值)    |    opa 自身健康检查的端口地址                |
| sidecar.csmopa.io/rego_source_bundle_config  |   必填值    |    必须提供, opa 启动时 config-file 文件内容                |
| sidecar.csmopa.io/rego_source_bundle_files  |   可选值    |       token 值           |

### 构建 istio-opa bundle 镜像

参考 [脚本](build_opa_images.sh)

### opa-webhook 证书管理(测试环境)

执行 sh sh/cert.sh 命令生成 opa webhook 所使用的证书(演示demo使用，生产环境不推荐使用)。

1、使用 ssl/ca-base64 替换 deployment/webhook.yaml MutatingWebhookConfiguration opa-sidecar-webhook 中的 caBundle 值(base64编码)

2、使用 ssl/cert-base64 替换 deployment/webhook.yaml Secret opa-sidecar-webhook-certs 中的 cert.pem 值(base64编码)

2、使用 ssl/key-base64 替换 deployment/webhook.yaml Secret opa-sidecar-webhook-certs 中的 key.pem 值(base64编码)

### opa 测试案例

#### opa 支持 configmap

部署 opa-webhook，替换 deployment/webhook.yaml 中的镜像地址

```
kubectl apply -f deployment/webhook.yaml
kubectl delete -f deployment/webhook.yaml
```

查看 opa-webhook 的日志：

```
kubectl -n opa-istio logs -f $(kubectl -n opa-istio get pod -l app=opa-sidecar  -o jsonpath={.items..metadata.name})
```

1、设置 istio 自动注入标签

```
kubectl label namespace default istio-injection="enabled"
kubectl label namespace default istio-injection-
```

2、部署 bookinfo 案例

```
kubectl apply -f deployment/bookinfo-opa.yaml
kubectl delete -f deployment/bookinfo-opa.yaml
```

3、部署 productpage 应用

```
kubectl apply -f  deployment/configmap/bookinfo-productpage-v1-configmap.yaml
kubectl delete -f  deployment/configmap/bookinfo-productpage-v1-configmap.yaml
```

其中 productpage 应用带有以下特定注解与标签

```
template:
  metadata:
    annotations:
      sidecar.csmopa.io/opa_image_url: "xxxxx/opa:latest-istio"
      sidecar.csmopa.io/rego_source_type: "configmap"
      sidecar.csmopa.io/rego_source_cm: "inject-policy"
      sidecar.csmopa.io/rego_source_cm_key: "policy.rego"
      sidecar.csmopa.io/grpc_policy_path: "istio/authz/allow"
    labels:
      csmopa-injection: "enabled"
```

如果 Pod 注解缺失必选值，则 Pod 会启动失败。

4、部署网关

```
kubectl apply -f deployment/bookinfo-gateway.yaml
kubectl delete -f deployment/bookinfo-gateway.yaml
```

5、通过 Istio Ingress 网关(GATEWAY_URL)地址访问 bookinfo。

查看 https://istio.io/latest/docs/tasks/traffic-management/ingress/#determining-the-ingress-ip-and-ports 配置
ingress-gateway 的服务对外 ip。

export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT

6、测试

默认情况下 alice 可以同时访问 /productpage 与 /api/v1/products

```
curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage
curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products
```

测试结果：

```
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  4183  100  4183    0     0   6178      0 --:--:-- --:--:-- --:--:--  6178
200%
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   395  100   395    0     0  15192      0 --:--:-- --:--:-- --:--:-- 15192
200%
```

7、开启 opa-sidecar 流量劫持开关，使其 opa 策略生效，本次测试是通过 envoyfilter 开启流量劫持到 opa。

```
kubectl apply -f deployment/opa-envoyfilter-productpage.yaml
kubectl delete -f deployment/opa-envoyfilter-productpage.yaml
```

8、测试在该策略下 alice 可以访问 /productpage，但是不能访问 /api/v1/products。

```
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage

% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
Dload  Upload   Total   Spent    Left  Speed
100  5183  100  5183    0     0   101k      0 --:--:-- --:--:-- --:--:--  101k
200%
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products

% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
Dload  Upload   Total   Spent    Left  Speed
0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
403%
```

#### opa 支持 bundle

部署 opa-webhook，替换 deployment/webhook.yaml 中的镜像地址

```
kubectl apply -f  deployment/webhook.yaml
kubectl delete -f deployment/webhook.yaml
```

查看 opa-webhook 的日志：

```
kubectl -n opa-istio logs -f $(kubectl -n opa-istio get pod -l app=opa-sidecar  -o jsonpath={.items..metadata.name})
```

部署 bundle server

```
kubectl create ns nginx
kubectl apply -f deployment/bundle/bundle-server.yaml -n nginx
kubectl -n nginx cp deployment/bundle/authz.gz $(kubectl -n nginx get pod -l app=nginx -o jsonpath={.items..metadata.name}):/usr/share/nginx/html/authz.gz

kubectl delete ns nginx
```

1、设置 istio 自动注入标签

```
kubectl label namespace default istio-injection="enabled"
```

2、部署 bookinfo 案例

```
kubectl apply -f deployment/bookinfo-opa.yaml
kubectl delete -f deployment/bookinfo-opa.yaml
```

3、部署 productpage 应用

```
kubectl apply -f  deployment/bundle/productpage-v1-bundle.yaml
kubectl delete -f  deployment/bundle/productpage-v1-bundle.yaml
```

查看 bookinfo 的日志：

```
kubectl logs -f $(kubectl get pod -l app=productpage -o jsonpath={.items..metadata.name}) -c csmopainjected-opa-istio
kubectl logs -f $(kubectl get pod -l app=productpage -o jsonpath={.items..metadata.name}) -c istio-proxy

kubectl exec -it $(kubectl get pod -l app=productpage -o jsonpath={.items..metadata.name}) -c csmopainjected-opa-istio bash
kubectl exec -it $(kubectl get pod -l app=ratings -o jsonpath={.items..metadata.name}) -c istio-proxy -- curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://productpage:9080/productpage
kubectl exec -it $(kubectl get pod -l app=ratings -o jsonpath={.items..metadata.name}) -c istio-proxy -- curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://productpage:9080/api/v1/products
```

其中 productpage 应用带有以下特定注解与标签

```
template:
  metadata:
    annotations:
      sidecar.csmopa.io/opa_image_url: "xxx/opa:latest-istio-bundle"
      sidecar.csmopa.io/rego_source_type: "bundle"
      sidecar.csmopa.io/rego_source_bundle_resource_path: "authz.gz"
      sidecar.csmopa.io/grpc_policy_path: "istio/authz/allow"
    labels:
      csmopa-injection: "enabled"
```

如果 Pod 注解缺失必选值，则 Pod 会启动失败。

4、部署网关

```
kubectl apply -f deployment/bookinfo-gateway.yaml
kubectl delete -f deployment/bookinfo-gateway.yaml
```

5、通过 Istio Ingress 网关(GATEWAY_URL)地址访问 bookinfo。

查看 https://istio.io/latest/docs/tasks/traffic-management/ingress/#determining-the-ingress-ip-and-ports 配置
ingress-gateway 的服务对外 ip。

export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT

6、测试

默认情况下 alice 可以同时访问 /productpage 与 /api/v1/products

```
curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage
curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products
```

测试结果：

```
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  4183  100  4183    0     0   6178      0 --:--:-- --:--:-- --:--:--  6178
200%
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   395  100   395    0     0  15192      0 --:--:-- --:--:-- --:--:-- 15192
200%
```

7、开启 opa-sidecar 流量劫持开关，使其 opa 策略生效，本次测试是通过 envoyfilter 开启流量劫持到 opa。

```
kubectl apply -f deployment/opa-envoyfilter-productpage.yaml
kubectl delete -f deployment/opa-envoyfilter-productpage.yaml
```

8、测试在该策略下 alice 可以访问 /productpage，但是不能访问 /api/v1/products。

```
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/productpage

% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
Dload  Upload   Total   Spent    Left  Speed
100  5183  100  5183    0     0   101k      0 --:--:-- --:--:-- --:--:--  101k
200%
➜  opa-webhook git:(master) ✗ curl  -w '%{http_code}' -o /dev/null  --user alice:password -i http://$GATEWAY_URL/api/v1/products

% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
Dload  Upload   Total   Spent    Left  Speed
0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
403%
```

## opa-webhook 版本发布记录

| 版本               |             主要功能              |            说明            |
| ------------------------ |            ---------------- |  ---------------- |
| 1.0     |    支持 configmap 下发 opa 策略         |        |
| 1.1     |    opa 支持 arm istio 集群           |                     |
| 1.2     |    重命名部分注解名称与优化文档           |                     |
| 1.3     |    1. 新增注解 sidecar.csmopa.io/addr_port </br> 2. 新增注解 sidecar.csmopa.io/diagnostic_addr_port 可配置监听与监控检查地址      |                     |
| 1.4     |    opa webhook 支持 bundle 下发策略    |                     |
| 1.5     |    opa webhook 下发 bundle 策略时，增加环境变量 ENV pod_name,pod_namespace    |                     |
| 1.6     |    opa webhook config 与 file 支持 base64   |                     |

## FAQ

