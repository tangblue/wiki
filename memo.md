## bash
* Run multiple commands as another user with single quote escape
```
sudo -i -u another bash -c '
echo '\''$HOME'\'' &&
echo '\''$HOME'\''
'
```

* Append multiple lines to script
```
cat << 'EOF' >> ~/.bashrc
export PYENV_ROOT="$HOME/.pyenv"
export PATH="$PYENV_ROOT/bin:$PATH"
if command -v pyenv 1>/dev/null 2>&1; then
    eval "$(pyenv init -)"
fi
eval "$(pyenv virtualenv-init -)"
EOF
```
