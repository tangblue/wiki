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

* load variables from a file
```
$ cat env_db 
db_host=xxxx
db_port=443
db_user=xxxx
db_password=xxxx
db_database=xxxx
db_schema=xxxx
$ . env_db; PGPASSWORD=${db_password} psql -h $db_host -p $db_port -U $db_user -d $db_database
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


## Docker

### Dockerfile for python
```
FROM python:3.9-slim as builder

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc

COPY requirements.txt .
RUN pip wheel --no-cache-dir --no-deps --wheel-dir /app/wheels -r requirements.txt


FROM python:3.9-slim

WORKDIR /app

COPY --from=builder /app/wheels /wheels
COPY --from=builder /app/requirements.txt .

RUN pip install --no-cache /wheels/*

RUN addgroup --gid 1001 --system app && \
    adduser --no-create-home --shell /bin/false --disabled-password --uid 1001 --system --group app

USER app

COPY src /app/src

CMD ["python","./src/main.py"]
```

### Install python modules for lambda
```
docker run --rm --platform linux/amd64 \
    -v $(pwd)/src/requirements.txt:/app/src/requirements.txt \
    -v $(pwd)/build/lambda_layer:/work -w /work \
    python:3.9 pip install -r /app/src/requirements.txt -t ./python
pushd ./build/lambda_layer; zip -r ${OLDPWD}/build/python_libs.zip .; popd
```
