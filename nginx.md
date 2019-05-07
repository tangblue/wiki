## nginx.conf
```
server{
    listen         80;
    server_name    localhost;
    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }
    location /api/ {
        proxy_pass    http://ip:port/api/;
        proxy_set_header    Host    $host;
        proxy_set_header    X-Real-IP    $remote_addr;
        proxy_set_header    X-Forwarded-Host      $host;
        proxy_set_header    X-Forwarded-Server    $host;
        proxy_set_header    X-Forwarded-For    $proxy_add_x_forwarded_for;
    }
}
```

## Start nginx
```
docker run --name test-nginx \
            -v /path/static/:/usr/share/nginx/html:ro \
            -v /path/nginx.conf:/etc/nginx/conf.d/default.conf:ro \
            -p 80:80 \
            -d nginx
```
