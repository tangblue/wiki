penssl genrsa -out ExampleRootCA.key 4096
openssl req -new -x509 -days 365 -subj "/C=US/ST=State/O=organization/CN=Example Root CA" -extensions v3_ca -key ExampleRootCA.key -out ExampleRootCA.crt
openssl x509 -noout -text -in ExampleRootCA.crt

openssl genrsa -out ExampleIntermediateCA.key 4096
openssl req -new -subj "/C=US/ST=State/O=organization/CN=Example Intermediate CA" -key ExampleIntermediateCA.key -out ExampleIntermediateCA.csr
openssl x509 -req -days 365 -extfile v3CA.ext -in ExampleIntermediateCA.csr -CA ExampleRootCA.crt -CAkey ExampleRootCA.key -CAcreateserial -out ExampleIntermediateCA.crt
openssl x509 -noout -text -in ExampleIntermediateCA.crt

openssl genrsa -out ExampleServer.key 2048
openssl req -new -subj "/C=US/ST=State/O=organization/CN=*.example.com" -key ExampleServer.key -out ExampleServer.csr
openssl x509 -req -days 1000 -extfile v3.ext -in ExampleServer.csr -CA ExampleIntermediateCA.crt -CAkey ExampleIntermediateCA.key -set_serial 0101 -out ExampleServer.crt -sha1
openssl x509 -noout -text -in ExampleServer.crt
