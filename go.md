## Tiny Docker image
```
GO_ENABLED=0 go build -ldflags="-s -w" -o app main.go && tar c app | docker import - --change 'CMD ["/app"]' tinyimage:latest
docker run -it --rm tinyimage:latest
```
