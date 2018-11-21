## bash
* Run multiple commands as another user with single quote escape
```
sudo -i -u another bash -c '
echo '\''$HOME'\'' &&
echo '\''$HOME'\''
'
```
