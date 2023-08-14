# nss_http

Name Service Switch Service that uses an http JSON backend to authenticate users and groups.


## Quick Setup
1. Create a sample [users.json](users.json) and [groups.json](groups.json).
   1. To create a hashed password you can use `openssl passwd -6`
2. Spin up a http server that hosts these files
   1. e.g. `python -m SimpleHTTPServer 8000` or `python3 -m http.server 8000`
3. Compile the library for your system or use a prebuilt version from the [Releases](releases) page.
   1. To compile for your system you need `go1.20` and `make`.
   2. Run `make install` to build and install the library.
4. Make sure you placed the library correctly at `/lib/libnss_http.so.2`
5. Create a new config at `/etc/nss_http.json`:
   ```json
   {
     "Providers": [
       {
         "Name": "http_static",
         "UsersURL": "http://localhost:8000/users.json",
         "GroupsURL": "http://localhost:8000/groups.json",
         "RequestTimeout": "1m",
         "Headers": {}
       }
     ],
     "Cache": {
       "Name": "disabled"
     }
   }
   ```
6. Adjust `/etc/nsswitch.conf` to include `http` for `passwd`, `shadow` and `group`:
   ```
   # /etc/nsswitch.conf
   passwd:         compat http
   group:          compat http
   shadow:         compat http
   gshadow:        files
   ...
   ```
7. Test the functionality using getent passwd <username>

### SSH Authentication
It is possible to add ssh authentication to the system by altering the `sshd_config`:
```
AuthorizedKeysCommand /sbin/nss_http_ssh
AuthorizedKeysCommandUser nobody
```

### Troubleshooting
nss_http will write a log file to `/var/log/nss_http.log`.  
You can change the log file path by specifying `NSS_HTTP_LOG_FILE`,  
setting it to `disabled` will disable the log file.
(notice you could also use `/dev/stdout` or `/dev/stderr` as a value).  
You can also enable debug logging by setting `NSS_HTTP_DEBUG` to `true`.

## Configuration
Configuration lives at `/etc/nss_http.json`.  
It is an ordinary json file that specifies the providers to use to look up 
users and groups.  
Notice that you can specify multiple providers, but only one cache provider.  
nss_http will always follow the order of the specified providers to look up users and groups, however, it will always
try to lookup users and groups by cache first.
```json
{
  "Providers": [
    Provider...
  ],
  "Cache": Cache
}
```

## Providers
### http_static
Fetch [users.json](users.json) and [groups.json](groups.json) from a server and
then use that to lookup users and groups.
#### Example Configuration
```json
{
  "Providers": [
    {
      "Name": "http_static",
      "UsersURL": "http://localhost:8000/users.json", 
      "GroupsURL": "http://localhost:8000/groups.json",
      "RequestTimeout": "1m",
      "Headers": {}
    }
  ]
}
```

### http_rest
Lookup a specific user or group by a REST api.

#### Endpoints that will be called
Based on the use case following endpoints will be called:

| use case               | request url                   |
|------------------------|-------------------------------|
| lookup a user by name  | GET `<url>/user/name/<name>`  |
| lookup a user by uid   | GET `<url>/user/uid/<uid>`    |
| lookup a group by name | GET `<url>/group/name/<name>` |
| lookup a group by uid  | GET `<url>/group/uid/<uid>`   |

#### user response
When a user is requested, make sure to respond with http status code `200` and return json content:
```json
{
  "User": "joe",
  "Passwd": "$6$.WdgkoyPbvxIDDKU$mOVy8BlNvGssTojiLDyo37S7/puNMBx53S4VAp1nhxSnV5G7bzZw42QxbcYiq4TJwReY0cBLQGc5Dt6Mnk4lg1",
  "Name": "Joe Doe",
  "Dir": "/home/joe",
  "Shell": "/bin/bash",
  "Uid": 3000,
  "Gid": 3000
}
```
If you want to signal that the requested user does not exist return the http status code `404`.

#### group response
When a group is requested, make sure to respond with http status code `200` and return json content:
```json
{
  "User": "joe",
  "Passwd": "$6$.WdgkoyPbvxIDDKU$mOVy8BlNvGssTojiLDyo37S7/puNMBx53S4VAp1nhxSnV5G7bzZw42QxbcYiq4TJwReY0cBLQGc5Dt6Mnk4lg1",
  "Name": "Joe Doe",
  "Dir": "/home/joe",
  "Shell": "/bin/bash",
  "Uid": 3000,
  "Gid": 3000
}
```
If you want to signal that the requested group does not exist return the http status code `404`.

#### Example Configuration
```json
{
  "Providers": [
    {
      "Name": "http_rest",
      "URL": "http://localhost:8000/",
      "RequestTimeout": "1m",
      "Headers": {}
    }
  ]
}
```

### redis
Lookup a specific user or group by using redis key value.

#### Organization of data
Based on the use case following keys will be used:

| use case               | key                 |
|------------------------|---------------------|
| lookup a user by name  | `user/name/<name>`  |
| lookup a user by uid   | `user/uid/<uid>`    |
| lookup a group by name | `group/name/<name>` |
| lookup a group by uid  | `group/uid/<uid>`   |

Notice that the data must be encoded in json.

#### Example Configuration
```json
{
  "Providers": [
    {
      "Name": "redis",
      "URL": "redis://myusername:mypassword@localhost:6379"
    }
  ]
}
```

## Cache
### disabled
Disable cache entirely, no data will be cached.
#### Example Configuration
```json
{
  "Cache": {
     "Name": "disabled"
  }
}
```

### intern
Use internal memory cache.  
This is still not sufficient for production systems since nss_http's memory will be freed after a application is done.
#### Example Configuration
```json
{
  "Cache": {
     "Name": "intern",
     "TTL": "1m"
  }
}
```

### redis
Cache using redis.  

### Organization of data
Uses the same format as the redis provider.

#### Example Configuration
```json
{
  "Cache": {
    "Name": "redis",
     "URL": "redis://myusername:mypassword@localhost:6379",
     "TTL": "1m"
  }
}
```

