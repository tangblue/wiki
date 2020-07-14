## bash
* Run multiple commands as another user
```
cat << 'EOF' | sudo -i -u $(whoami) bash
echo '$HOME'
echo "$HOME"
EOF
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

* env for command(Bash Command Substitution)
```
(export $(cat << EOF
A=B
B=C
EOF
); set)
```

* env file for docker(Bash Process Substitution)
```
docker run -it --rm --env-file <(cat << EOF
A=B
B=C
EOF
) bash
```
