## Admission ValidatingWebhook
```shell
# 安装 cert maneger
kubectl apply -f deploy/cert-manager-1.5.3.yaml

# 创建证书 secret 
kubectl apply -f deploy/tls-secret.yaml

# 创建 configuration
kubectl apply -f deploy/validatingWebhookConfiguration.yaml

# 部署应用
kubectl apply -f deploy/deployment.yaml

# 卸载全部
make clear
```

## 构建镜像
```shell
docker build --platform=linux/amd64 -t dierbei/mutating_webhook:202304272210 .
```

## 命名空间添加 label
```shell
kubectl label namespace testing debug=true
```

## 测试
```shell
kubectl run test1 --image=nginx:1.18-alpine --namespace=admission
kubectl run test2 --image=busybox --namespace=admission
```