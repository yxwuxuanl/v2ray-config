# v2ray-config

从`Just My Socks`的节点订阅链接中获取节点配置并生成v2ray配置文件

### Docker使用

```shell
docker run --rm -v $PWD:/config -w /config lin2ur/v2ray-config --subscribe=$YOUR_SUBSCRIBE_LINK
# cat ./v2ray.json 
```

### 部署到k8s集群

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: v2ray
  labels:
    app: v2ray
spec:
  replicas: 1
  selector:
    matchLabels:
      app: v2ray
  template:
    metadata:
      name: v2ray
      labels:
        app: v2ray
    spec:
      enableServiceLinks: false
      automountServiceAccountToken: false
      volumes:
        - name: config
          emptyDir: { }
      initContainers:
        - name: config
          image: lin2ur/v2ray-config
          imagePullPolicy: IfNotPresent
          workingDir: /config
          volumeMounts:
            - mountPath: /config
              name: config
          args:
            - --subscribe
            - $YOUR_SUBSCRIBE_LINK
      containers:
        - name: v2ray
          image: v2fly/v2fly-core:v5.2.1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 1087
              name: http
            - containerPort: 1080
              name: socks
          args:
            - run
            - -config
            - /v2ray.json
          volumeMounts:
            - mountPath: /v2ray.json
              name: config
              subPath: v2ray.json
      restartPolicy: Always

---
apiVersion: v1
kind: Service
metadata:
  name: v2ray
spec:
  selector:
    app: v2ray
  ports:
    - port: 1087
      targetPort: 1087
      name: http
    - port: 1080
      targetPort: 1080
      name: socks
  type: NodePort
```