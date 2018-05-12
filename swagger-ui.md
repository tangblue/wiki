## SWAGGER UI(API document)
1. Get an example yaml
    ```
    wget https://raw.githubusercontent.com/openservicebrokerapi/servicebroker/v2.13/openapi.yaml
    ```
1. Run docker container to serve the example yaml
    ```
    (SWAGGER_JSON=/tmp/openapi.yaml; \
    docker run -d -p 8080:8080 \
    --mount type=bind,source=$(pwd)/openapi.yaml,target=${SWAGGER_JSON} \
    -e SWAGGER_JSON=${SWAGGER_JSON} swaggerapi/swagger-ui)
    ```
1. Visit sample page  
    http://127.0.0.1:8080/

Note: The docker bind mount share the same inode between host and docker contain. After editing the example yaml file by vim, save file to the same inode by:
```
set backupcopy=yes
```

## [kube-openapi](https://github.com/kubernetes/kube-openapi)
```
go get github.com/kubernetes/kube-openapi
cd github.com/kubernetes/kube-openapi
dep init
mkdir example/model
touch example/model/header.txt
cat << EOF > example/model/model.go
package model

// MinimalPod is a minimal pod.
// +k8s:openapi-gen=true
type MinimalPod struct {
        Name string \`json:"name"\`
}
EOF

(MODEL_PATH=github.com/kubernetes/kube-openapi/example/model; \
go run example/openapi-gen/main.go \
-h example/model/header.txt \
-i ${MODEL_PATH} \
-p ${MODEL_PATH})

less example/model/openapi_generated.go
```
