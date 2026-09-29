# nss_http

Name Service Switch Service that uses an http JSON backend to authenticate users and groups.

> **Requires glibc.** NSS modules are a glibc mechanism, so this library only
> works on glibc based distributions (Debian, Ubuntu, Fedora, Arch, ...).
> It **cannot** work on musl based distributions such as Alpine: musl has no
> `nss.h`/`gshadow.h`, exposes no `__nss_*` loader hooks, and reads
> `/etc/passwd`, `/etc/group` and `/etc/shadow` directly. Alpine's own
> `/etc/nsswitch.conf` documents that "musl itself does not support NSS" and
> only honours a `hosts:` line. Installing `gcompat` does not help, since it
> provides glibc symbol shims rather than the NSS plugin mechanism.
>
> On musl systems the `nss_http_sshkey` helper still works, because it is an
> ordinary executable: see [SSH Authentication](#ssh-authentication). sshd will
> however still require the user to exist in `/etc/passwd`.

## Quick Setup
1. Create a sample [users.json](users.json) and [groups.json](groups.json).
   1. To create a hashed password you can use `openssl passwd -6`
2. Spin up a http server that hosts these files
   1. e.g. `python -m SimpleHTTPServer 8000` or `python3 -m http.server 8000`
3. Compile the library for your system or use a prebuilt version from the [Releases](releases) page.
   1. To compile for your system you need `go1.21`, `make` and the glibc
      development headers (`libc6-dev` on Debian/Ubuntu).
   2. Run `make install` to build and install the library.
4. Make sure you placed the library correctly at `/lib/libnss_http.so.2`
5. Create a new config at `/etc/nss_http.json`:
   ```json
   {
     "Providers": [
       {
         "Name": "http",
         "URLs": {
           "Users": "http://localhost:8000/users.json",
           "Groups": "http://localhost:8000/groups.json"
         },
         "RequestTimeout": "1m",
         "Headers": {}
       }
     ],
     "Cache": {
       "Name": "disabled"
     },
     "AllowListingOfUsers": false,
     "AllowListingOfGroups": false
   }
   ```
6. Adjust `/etc/nsswitch.conf` to include `http` for `passwd`, `shadow` and `group`:
   ```
   # /etc/nsswitch.conf
   passwd:         compat http
   group:          compat http
   shadow:         compat http
   gshadow:        files http
   ...
   ```
7. Test the functionality using `getent passwd <username>`

### Groups and group membership
A user's *primary* group comes from the `Gid` field on the user. *Supplementary*
group membership comes from the `GroupMembers` list on a group:

```json
[
  { "Name": "joe",    "Passwd": "", "Gid": 3000, "GroupMembers": [] },
  { "Name": "admins", "Passwd": "", "Gid": 6000, "GroupMembers": ["joe"] }
]
```

Verify both with:

```console
$ getent group admins
admins:x:6000:joe
$ id joe
uid=3000(joe) gid=3000(joe) groups=3000(joe),6000(admins)
```

Note that `/etc/nsswitch.conf` is consulted in order, so a group that also
exists in `/etc/group` (for example `staff`, gid 50 on Debian) will resolve to
the local entry instead of the HTTP one.

### SSH Authentication
It is possible to add ssh authentication to the system by altering the `sshd_config`:
```
AuthorizedKeysCommand /sbin/nss_http_sshkey
AuthorizedKeysCommandUser nobody
```

### Troubleshooting
_nss_http_ will write a log file to `/var/log/nss_http.log`.  
You can change the log file path by specifying `NSS_HTTP_LOG_FILE`,  
setting it to `disabled` will disable the log file.
(notice you could also use `/dev/stdout` or `/dev/stderr` as a value).  
You can also enable debug logging by setting `NSS_HTTP_DEBUG` to `true`.

## Configuration
Configuration lives at `/etc/nss_http.json`.  
It is an ordinary json file that specifies the providers to use to look up 
users and groups.
```json
{
  "Providers": [
    Provider...
  ],
  "Cache": Cache,
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
}
```
Notice that you can specify multiple providers, but only one cache provider.  
_nss_http_ will always follow the order of the specified providers to look up users and groups, however, it will always
try to lookup users and groups by cache first.

You can disallow the listing of all users and groups, this is especially useful if
you deal with a lot of users and groups. Or only have a http server that returns individual
users and groups.

By default, all passwords need to be hashed upfront in the [crypt(3)](https://en.wikipedia.org/wiki/Crypt_(C)) format:
```
$<id>[$<param>=<value>(,<param>=<value>)*][$<salt>[$<hash>]]
```

Depending on your system, you can use `openssl passwd -6` to hash the passwords upfront.

You could also use plain text mode by setting `DisableShadow` to `true`, but this is not recommended.

## Providers
### http
The http provider allows you to lookup users and groups using http.  
You can either specify only two endpoints pointing to a [users.json](users.json) and [groups.json](groups.json) file.  
Or point directly to individual resources. (more in the example section).

You could also specify headers in the `headers` section, use that to specify a token using
the `Authorization` header or similar.

#### Example Configuration
##### A list of users and groups
In this example we only point to a list of users and groups,
_nss_http_ will automatically pull the correct user and group outside of this list.
```json
{
  "Providers": [
    {
      "Name": "http",
      "URLs": {
         "Users": "http://localhost:800/users.json",
         "Groups": "http://localhost:800/groups.json"
      },
      "RequestTimeout": "1m",
      "Headers": {}
    }
  ],
  "Cache": {
    "Name": "disabled"
  },
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
}
```

##### Specific endpoints for individual users/groups
A more resource efficient variant is to directly point to individual resources.  
When a lookup will be performed the id or name of the user/group will be appended.
In this example `getent passwd joe` will result calling `http://localhost:800/user/name/joe`.
```json
{
  "Providers": [
    {
      "Name": "http",
      "URLs": {
         "UserUID": "http://localhost:800/user/uid/",
         "UserName": "http://localhost:800/user/name/",
         "GroupGID": "http://localhost:800/group/gid/",
         "GroupName": "http://localhost:800/group/name/"
      },
      "RequestTimeout": "1m",
      "Headers": {}
    }
  ],
  "Cache": {
    "Name": "disabled"
  },
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
}
```
> Notice that a list endpoint is required when you want to allow listing of users and groups
> using `AllowListingOfUsers` and `AllowListingOfGroups`.

##### Specify both, a list and individual endpoints
You could also specify the list and individual endpoints.  
_nss_http_ will lookup specific users/groups via the individual endpoints,
list requests will go directly to the users/groups endpoint.
```json
{
  "Providers": [
    {
      "Name": "http",
      "URLs": {
         "Users": "http://localhost:800/users",
         "UserUID": "http://localhost:800/user/uid/",
         "UserName": "http://localhost:800/user/name/",
         "Groups": "http://localhost:800/groups",
         "GroupGID": "http://localhost:800/group/gid/",
         "GroupName": "http://localhost:800/group/name/"
      },
      "RequestTimeout": "1m",
      "Headers": {}
    }
  ],
  "Cache": {
    "Name": "disabled"
  },
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
}
```
> Notice that a list endpoint is required when you want to allow listing of users and groups
> using `AllowListingOfUsers` and `AllowListingOfGroups`.

#### Required Server Responses
##### User List Response
For the `Users` endpoint a list of users is expected as a result.
You can see an example in the [users.json](users.json) file.

##### Groups List Response
For the `Groups` endpoint a list of groups is expected as a result.
You can see an example in the [groups.json](groups.json) file.

#### UserUID / UserName Response
When a specific user is requested, make sure to respond with http status code `200` and return json content:
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

#### GroupUID / GroupName Response
When a group is requested, make sure to respond with http status code `200` and return json content:

```json
{
   "Name": "admins",
   "Passwd": "",
   "Gid": 6000,
   "GroupMembers": ["joe"]
}
```
If you want to signal that the requested group does not exist return the http status code `404`.

> Notice that in `GroupMembers` you only specify users which will have this group as a secondary group, the primary
> group information is already present in the user data.

### redis
Lookup a specific user or group by using redis key value.

#### Organization of data
Based on the use case following keys will be used:

| use case               | key                  |
|------------------------|----------------------|
| lookup a user by name  | `users/name/<name>`  |
| lookup a user by uid   | `users/id/<uid>`     |
| lookup a group by name | `groups/name/<name>` |
| lookup a group by uid  | `groups/id/<gid>`    |

Notice that the data must be encoded in json.

#### Example Configuration
```json
{
  "Providers": [
    {
      "Name": "redis",
      "URL": "redis://myusername:mypassword@localhost:6379"
    }
  ],
  "Cache": {
    "Name": "disabled"
  },
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
}
```

### file
The file provider allows you to lookup users and groups using a json file.  
You need to specify a [users.json](users.json) and [groups.json](groups.json) file.  

#### Example Configuration
```json
{
  "Providers": [
    {
      "Name": "file",
      "Users": "/etc/nss_http/users.json",
      "Groups": "/etc/nss_http/groups.json"
    }
  ],
  "Cache": {
    "Name": "disabled"
  },
  "AllowListingOfUsers": false,
  "AllowListingOfGroups": false,
  "DisableShadow": false
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

## Partial Usage
You don't have to use the users and groups functionality of _nss_http_.
Simply remove `http` from the `nsswitch.conf` file where you don't want to use it. 
And remove the urls in the `http` provider.
