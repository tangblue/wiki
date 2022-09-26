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

## GCP
Conect Cloud SQL with private IP from local develop environment.  [ref](https://medium.com/google-cloud/cloud-sql-with-private-ip-only-the-good-the-bad-and-the-ugly-de4ac23ce98a)
```
gcloud compute scp /local/path/to/cloud_sql_proxy <instanceName>:/tmp
gcloud compute ssh <instance Name> --zone=<Your zone>
gcloud compute start-iap-tunnel <instance Name> 22 \
  --zone=<Your zone> --local-host-port=localhost:4226
ssh -L 3306:localhost:3306 \
  -i ~/.ssh/google_compute_engine \
  -p 4226 localhost \
  -- /tmp/cloud_sql_proxy instances=<connection_name>=tcp:3306
```
