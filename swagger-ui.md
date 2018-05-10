## SWAGGER UI(API document)
1. Get an example yaml
```
wget https://raw.githubusercontent.com/openservicebrokerapi/servicebroker/v2.13/openapi.yaml
```

1. Run docker container to serve the example yaml
```
docker run -d -p 8080:8080 \
--mount type=bind,source=$(pwd)/openapi.yaml,target=/tmp/openapi.yaml \
-e SWAGGER_JSON=/tmp/openapi.yaml swaggerapi/swagger-ui
```

1. Visit sample page  
   http://127.0.0.1:8080/

Note: The docker bind mount share the same inode between host and docker contain. After editing the example yaml file by vim, save file to the same inode by:
```
set backupcopy=yes
```
