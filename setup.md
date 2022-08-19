## HOME
```
echo "tmpfs		/tmp		tmpfs	defaults,noatime,nosuid,nodev,mode=1777	0	1" >> /etc/fstab 
echo "sudo -u tangxf bash -c 'mkdir -p /tmp/\${USER}/{cache,go/src}'" >> /etc/rc.local 
rm -rf ${HOME}/.cache && ln -sf /tmp/${USER}/cache ${HOME}/.cache
```

## screenrc
```
diff -u0 /etc/screenrc ~/.screenrc 
--- /etc/screenrc       2018-04-27 20:24:38.588362853 +0900
+++ ~/.screenrc   2018-05-13 12:18:16.769370294 +0900
@@ -69 +69,2 @@
-hardstatus string "%h%? users: %u%?"
+hardstatus string "%{= kw} %n %t %h"
+
@@ -89 +90 @@
-#termcapinfo xterm|xterms|xs|rxvt ti@:te@
+termcapinfo xterm|xterms|xterm-256color|xs|rxvt ti@:te@
```

## GO
### Install
```
(VER=1.17.2; [ -d ~/opt ] || mkdir ~/opt; rm -rf ~/opt/go; curl https://dl.google.com/go/go${VER}.linux-amd64.tar.gz | tar xz -C ~/opt/)
export PATH="$HOME/opt/go/bin:$HOME/go/bin:$PATH"
git clone https://github.com/fatih/vim-go.git ~/.vim/pack/plugins/start/vim-go
go get -u github.com/jstemmer/gotags
```
### kubernetes
```
cat << 'EOF'  >> ~/.bashrc
source <(kubectl completion bash)
EOF
```

## docker
```
curl -L https://github.com/docker/compose/releases/download/"$(curl --silent https://api.github.com/repos/docker/compose/releases/latest | jq .name -r)"/docker-compose-$(uname -s)-$(uname -m) -o ~/.local/bin/docker-compose && chmod a+x ~/.local/bin/docker-compose
```
