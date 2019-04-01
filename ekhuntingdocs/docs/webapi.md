# The web API

The Mass URL web API allows for all features that the dashboard can do. It can be used to control and interact with the Mass URL component of
Cuckoo Sandbox.


#### /api/group/add
###### POST /api/group/add

Create a new URL group

**Form parameters**:

* `name` *Required* *(string)* - The name of the new URL group
* `description` *Required* *(string)* - A description for the URL group

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `409` -  Group name exists

**CURL Example**

```bash
curl http://localhost:8080/api/group/add -F "name=newgroup" -F "description=An example group"
```

**Response example**

```
{
  "group_id": 1
}
```


#### /api/group/add/url

###### POST /api/group/add/url

**Form parameters**:

Add a small amount of URLs to an existing URL group. Use the bulk URL addition API for 1000+ URLs.

A group name or id must be provided.

* `name` *Optional* *(string)* - The id of a group
* `group_id` *Optional* *(string)* - The id of a URL group
* `urls` *Required* *(string)* - A list of one or more URLs
* `separator` *Optional* *(string)* - A single character to split to given list of URLs on. Uses a newline if no separator is provided.

**Status codes**

* `200` - Success
* `400` - Invalid or missing parameter
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/add/url -F "group_id=1" -F "urls=url1,url2,url3,url4" -F "separator=,"
```

**Response example**

```
{
  "info": "Added new URLs to group 1",
  "message": "success"
}
```

#### /api/group/(group)/url/add
###### POST /api/group/(integer:group_id)/url/add

Bulk add URLs to a group from a file. Use this if the amount of URLs is 1000+.

**Form parameters**:

* `urls` *Required* *(File)* - A file filled with URLs split on newline or the given separator character.

**URL parameters**

* `separator` *Optional* *(string)* - A single character to split to given file with URLs on. Uses a newline if no separator is provided.

**Status codes**

* `200` - Success
* `400` - Invalid URL list
* `404` - Group does not exist

**CURL Example**

```bash
curl "http://localhost:8080/api/group/1/url/add?separator=," -F "urls=@100kalexa.txt"
```

**Response example**

```
{
  "message": "Added 100000 URLs to group newgroup"
}
```


#### /api/group/view/(group)
###### GET /api/group/view/*(string:name)* or *(integer:group_id)*

Retrieve information about the specific group.

**URL parameters**:

* `details` *Optional* *(bool)* - Add additional URL count, amount of unread alerts, and if any alerts are high level.

**Status codes**

* `200` - Success
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/view/1?details=1
```

**Response example**

```
{
  "batch_size": 5,
  "batch_time": 40,
  "completed": true,
  "description": "An example group",
  "highalert": 0,
  "id": 1,
  "max_parallel": 50,
  "name": "newgroup",
  "profiles": [],
  "progress": null,
  "run": 1,
  "schedule": null,
  "schedule_next": null,
  "status": "completed",
  "unread": 0,
  "urlcount": 100000
}

```


#### /api/group/view/(group)/urls
###### GET /api/group/view/*(string:name)* or *(integer:group_id)*/urls

Retrieve all URLs for the specified group.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of URLs to return
* `offset` *Optional* *Integer* - Offset of the URL list

**Status codes**

* `200` - Success
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/view/1/urls
```

**Response example**

```
{
  "group_id": 1,
  "name": "newgroup",
  "urls": [
    {
      "id": "5c79bf2476b63a5908261aaad68719c6a2e62e5a3177f8ab97f82da348ba6e37",
      "url": "url4"
    },
    {
      "id": "2b9a40694179883a0dd41b2b16be242746cff1ac8cfd0fdfb44b7279bfc56362",
      "url": "url1"
    },
    {
      "id": "46b8f9c6f6468a93cc46729022ba9625e4d22167392b2e54fb0ad0f312868ad5",
      "url": "url3"
    },
    {
      "id": "86729d96320481bc7f78a334b8c81f216631fec96b0ef19040537c4144384068",
      "url": "url2"
    }
  ]
}
```

#### /api/group/delete
###### POST /api/group/delete

Delete a specified group. A group id or name must be provided.

!!! note "Note"
    This will not delete the URLs from the database. URLs keep existing, as they exist in multiple groups.


**Form parameters**:

* `name` *Optional* *(string)* - The name of the new URL group
* `group_id` *Optional* *(string)* - A description for the URL group

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/delete -F "group_id=1"
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/groups/list
###### GET /api/groups/list

Retrieve a list of all existing groups.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of groups to return
* `offset` *Optional* *Integer* - Offset of the group list
* `details` *Optional* *(bool)* - Add additional URL count, amount of unread alerts, and if any alerts are high level.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

```bash
curl http://localhost:8080/api/group/list
```

**Response example**

```
[
  {
    "batch_size": 5, 
    "batch_time": 25, 
    "completed": true, 
    "description": "Another example group", 
    "id": 2, 
    "max_parallel": 50, 
    "name": "newgroup2", 
    "profiles": [], 
    "progress": null, 
    "run": 0, 
    "schedule": null, 
    "schedule_next": null, 
    "status": null
  }, 
  {
    "batch_size": 5, 
    "batch_time": 40, 
    "completed": true, 
    "description": "An example group", 
    "id": 1, 
    "max_parallel": 50, 
    "name": "newgroup", 
    "profiles": [], 
    "progress": null, 
    "run": 1, 
    "schedule": null, 
    "schedule_next": null, 
    "status": "completed"
  }
]
```

#### /api/group/delete/url
###### POST /api/group/delete/url

Delete the specfied URLs from the specified group. A group name or group_id must be given.

**Form parameters**:

* `name` *Optional* *(string)* - The id of a group
* `group_id` *Optional* *(string)* - The id of a URL group
* `urls` *Required* *(string)* - A list of one or more URLs
* `separator` *Optional* *(string)* - A single character to split to given list of URLs on. Uses a newline if no separator is provided.
* `delall` *Optional* *(bool)* - Delete all URLs from the specified group.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/delete/url -F "group_id=1" -F "urls=url1,url2" -F "separator=,"
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/group/(group)/settings
###### POST /api/group/*(integer:group)*/settings

Update group analysis settings for the given group.

**Form parameters**:

* `threshold` *Optional* *(integer)* - The amount of URLs per task when the group is scheduled. Recommended is >= 10 <= 60.
* `batch_size` *Optional* *(integer)* - The amount of URLs uploaded and opened in the VM at the same time. Recommended is >= 4 <= 8.
* `batch_time` *Optional* *(integer)* - The amount of seconds each batch of URLs should stay opened inside the VM. Recommended is >= 30 <= 60.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

```bash
curl http://localhost:8080/api/group/1/settings -F "batch_size=5"
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/group/(group)/profiles
###### POST /api/group/*(integer:group)*/profiles

Replace the configured analysis profiles for the specified group to the specified list of profiles.
If an empty list is sent, it will removed the currently configured profiles.

**Form parameters**:

* `profile_ids` *Required* *(integer)* - A comma-separated list of analysis profile IDs

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

```bash
curl http://localhost:8080/api/group/1/profiles -F "profile_ids=4,7"
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/group/schedule/(group)
###### POST /api/group/schedule/*(integer:group)*

Create or update a schedule for the specified group. A group will only be analyzed if it has a schedule.

!!! warning "Note"
    A schedule can only be added if the group has URLs and one or more analysis profiles.

**Form parameters**:

* `schedule` *Required* *(string)* - A schedule format string identifying when a group should be analyzed. The format is either Xd@24hourtime or day@24hourtime.
Examples are: 1d@08:00 to start analysis daily at 08:00 or monday@08:00 to start the analysis every monday at 08:00.

This field also accepts `now`. This will cause the scheduler to immediately schedule and analyze the group.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `404` - Group does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/group/schedule/1 -F "schedule=now"

```

**Response example**

```
{
  "message": "Scheduled at 2019-03-01 15:32:47"
}
```


#### /api/profile/add
###### POST /api/profile/add

Create an analysis profile. Profiles represent an analysis environment on which a group is analyzed.

**Form parameters**:

* `name` *Required* *(string)* - The name of the new analysis profile
* `browser` *Required* *(string)* - The browser to use. Options are: `ie`, `ff`, or `edge`. Use the tags field to specify tags of the analysis machines that have the specified browser installed.
* `route` *Required* *(string)* - A route for the analyses to use. Can `internet`, `vpn`, or `socks5`. Only configured and enabled routes can be specified.
* `country` *optional* *(string)* - A country for the `vpn` or `socks5` route. The specified country will be used, if available.
* `tags` *Required* *(integer)* - A comma-separated list of tag ids to use.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `409` - Profile name already exists

**CURL Example**

```bash
curl http://localhost:8080/api/profile/add -F "name=newprofile" -F "browser=ie" -F "route=internet" -F "tags=1,4"
```

**Response example**

```
{
  "profile_id": 1
}
```


#### /api/profile/update/(profile)
###### POST /api/profile/update/*(integer:profile_id)*

Replace the specified parameters for the specified profile.

**Form parameters**:

* `browser` *Required* *(string)* - The browser to use. Options are: `ie`, `ff`, or `edge`. Use the tags field to specify tags of the analysis machines that have the specified browser installed.
* `route` *Required* *(string)* - A route for the analyses to use. Can `internet`, `vpn`, or `socks5`. Only configured and enabled routes can be specified.
* `country` *optional* *(string)* - A country for the `vpn` or `socks5` route. The specified country will be used, if available.
* `tags` *Required* *(integer)* - A comma-separated list of tag ids to use.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

```bash
curl http://localhost:8080/api/profile/update/1 -F "browser=ff" -F "route=vpn" -F "tags=1,4,8"
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/profile/(profile)
###### GET /api/profile/*(integer:profile_id)* or *(string:name)*

Retrieve the specified analysis profile.

**Status codes**

* `200` - Success
* `404` - Profile does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/profile/1
```

**Response example**

```
{
  "browser": "ff", 
  "country": "", 
  "id": 1, 
  "name": "newprofile", 
  "route": "vpn", 
  "tags": [
    {
      "id": 1, 
      "name": "windows7"
    }, 
    {
      "id": 4, 
      "name": "flash2000228"
    },
    {
      "id": 8, 
      "name": "firefox42"
    }
  ]
}

```


#### /api/profile/list
###### GET /api/profile/list

Retrieve a list of all existing analysis profiles.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of profiles to return
* `offset` *Optional* *Integer* - Offset of the profile list

**Status codes**

* `200` - Success
* `400` -  Invalid parameter

**CURL Example**

```bash
curl http://localhost:8080/api/profile/list
```

**Response example**

```
[
  {
    "browser": "ie", 
    "country": "", 
    "id": 2, 
    "name": "newprofile2", 
    "route": "internet", 
    "tags": [
      {
        "id": 1, 
        "name": "windows7"
      }, 
      {
        "id": 4, 
        "name": "flash2000228"
      }
    ]
  }, 
  {
    "browser": "ff", 
    "country": "", 
    "id": 1, 
    "name": "newprofile", 
    "route": "vpn", 
    "tags": [
      {
        "id": 1, 
        "name": "windows7"
      }, 
      {
        "id": 4, 
        "name": "flash2000228"
      },
      {
      "id": 8, 
      "name": "firefox42"
     }
    ]
  }
]
```


#### /api/profile/delete/(profile)
###### POST /api/profile/delete/*(integer:profile_id)* or *(string:name)*

Delete the specified profile.

**Status codes**

* `200` - Success

**CURL Example**

```bash
curl -XPOST http://localhost:8080/api/profile/delete/1
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/signature/add
###### POST /api/signature/add

Create a new URL diary signature. A JSON document must be sent.

**JSON keys**:

* `name` *Required* *(string)* - The name of the new signature
* `content` *Required* *(dictionary)* - A JSON dictionary according to the signature format.
* `level` *Required* *(integer)* - The level of alert that should be raised if a URL diary match is found.
* `enabled` *Required* *(bool)* - Enable or disable the signature.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `409` -  Signature name already exists

**CURL Example**

```bash
curl http://localhost:8080/api/signature/add -H "Content-Type: application/json" -d '
{  
  "name":"FindIframeEval",
  "level":3,
  "enabled":true,
  "content":{  
    "responsedata":[  
      {  
        "must":[  
          "<iframe>"
        ]
      }
    ],
    "javascript":[  
      {  
        "must":[  
          "eval("
        ]
      }
    ]
  }
}'
```

**Response example**

```
{
  "signature_id": 1
}
```


#### /api/signature/update/(signature)
###### POST /api/signature/update/*(integer:signature_id)*

Update an existing signature. A full signature must always be sent. Sent values replace existing values.

**JSON keys**:

* `content` *Required* *(dictionary)* - A JSON dictionary according to the signature format.
* `level` *Required* *(integer)* - The level of alert that should be raised if a URL diary match is found.
* `enabled` *Required* *(bool)* - Enable or disable the signature.

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter
* `404` -  Signature does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/signature/update/1 -H "Content-Type: application/json" -d '
{  
  "level":3,
  "enabled":false,
  "content":{  
    "responsedata":[  
      {  
        "must":[  
          "<iframe>",
          "<script>"
        ]
      }
    ],
    "javascript":[  
      {  
        "must":[  
          "eval("
        ]
      }
    ]
  }
}'
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/signature/list
###### GET /api/signature/list

Retrieve a list of all existing signatures.

**Status codes**

* `200` - Success

**CURL Example**

```bash
curl http://localhost:8080/api/signatures/list
```

**Response example**

```
[
  {
    "content": {
      "requestdata": [
        {
          "must": [
            "content-encoding: gzip"
          ]
        }
      ]
    }, 
    "enabled": true, 
    "id": 2, 
    "last_run": "2019-03-01 20:15:16", 
    "level": 1, 
    "name": "uselesssignature"
  }, 
  {
    "content": {
      "javascript": [
        {
          "must": [
            "eval("
          ]
        }
      ], 
      "responsedata": [
        {
          "must": [
            "<iframe>",
            "<script>"
          ]
        }
      ]
    }, 
    "enabled": false, 
    "id": 1, 
    "last_run": "2019-03-01 20:12:20", 
    "level": 3, 
    "name": "FindIframeEval"
  }
]

```


#### /api/signature/(signature)
###### GET /api/signature/*(integer:signature_id)*

Retrieve the specified signature

**Status codes**

* `200` - Success
* `404` - Signature does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/signature/1
```

**Response example**

```
{
  "content": {
    "javascript": [
      {
        "must": [
          "eval("
        ]
      }
    ], 
    "responsedata": [
      {
        "must": [
          "<iframe>",
          "<script>"
        ]
      }
    ]
  }, 
  "enabled": false, 
  "id": 3, 
  "last_run": "2019-03-31 20:12:20", 
  "level": 3, 
  "name": "FindIframeEval"
}
```


#### /api/signature/run/(signature)
###### POST /api/signature/run/*(integer:signature_id)*

Run the specified signature and find all matching URL diaries.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of matches to return
* `offset` *Optional* *Integer* - Offset of the matches list

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

```bash
curl -XPOST http://localhost:8080/api/signature/run/1
```

**Response example**

```
[
  {
    "datetime": 1553781290519, 
    "id": "0dfa0a00-5161-11e9-9a68-94b86d9621f2", 
    "url": "example.com", 
    "version": 1
  }, 
  {
    "datetime": 1553781290477, 
    "id": "0dfa09d8-5161-11e9-9a68-94b86d9621f2", 
    "url": "example.net", 
    "version": 1
  }, 
  {
    "datetime": 1553781290140, 
    "id": "0dfa09a6-5161-11e9-9a68-94b86d9621f2", 
    "url": "example.org", 
    "version": 1
  }
]

```


#### /api/signature/delete/(signature)
###### /api/signature/delete/*(integer:signature_id)*

Delete the specified signature.

**Status codes**

* `200` - Success

**CURL Example**

```bash
curl -XPOST http://localhost:8080/api/signature/delete/1
```

**Response example**

```
{
  "message": "success"
}
```


#### /api/requestlog/(requestlog)
###### GET /api/requestlog/(uuid:requestlog_id)

Retrieve a specific URL request log. Request logs are created for each requested page by the parent URL. The parent is
the URL dairy for the URL that is in an actual URL group.

**Status codes**

* `200` - Success
* `404` - Request log does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/requestlog/4f5ad86f-53f5-11e9-859b-ac1f6b2d9373
```

**Response example**

```
{
  "datetime": 1554064867530, 
  "log": [
    {
      "request": "GET / HTTP/1.1\r\naccept-language: en-US\r\naccept..truncated", 
      "response": "HTTP/1.1 200 OK\r\ncontent-length: 606\r\nx-cache..truncated", 
      "time": 1554064812.321427
    }
  ], 
  "parent": "4f5ad86e-53f5-11e9-859b-ac1f6b2d9373", 
  "url": "http://example.net/"
}
```


#### /api/diary/(diary)
###### GET /api/diary/(uuid:urldiary_id)

Retrieve a specific URL diary. A URL diary is created each time a URL in a group is analyzed.
It can contain requests that were made, malicious process signatures that were triggered, and executed javascript.

**Status codes**

* `200` - Success
* `404` - URL diary does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/diary/4f5ad86e-53f5-11e9-859b-ac1f6b2d9373
```

**Response example**

```
{
  "browser": "Internet Explorer", 
  "datetime": 1554064867544, 
  "javascript": [], 
  "machine": "w7x64_12", 
  "requested_urls": [
    {
      "len": 19, 
      "request_log": "4f5ad86f-53f5-11e9-859b-ac1f6b2d9373", 
      "url": "http://example.net/"
    }
  ], 
  "signatures": [], 
  "url": "example.net", 
  "url_id": "3daab7cff97925bbd07d11df5dc3b0e37e2d965520175ada0ec62ce72cda5ed2", 
  "version": 1
}
```


#### /api/diary/url/(url)
###### GET /api/diary/url/(string:url_id)

Retrieve a list of all existing URL diary IDs for the specified URL id.
The URL id is a sha256 hash of the URL.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of diary IDs to return
* `offset` *Optional* *Integer* - Offset of the diary IDs list

**Status codes**

* `200` - Success

**CURL Example**

```bash
curl http://localhost:8080/api/diary/url/3daab7cff97925bbd07d11df5dc3b0e37e2d965520175ada0ec62ce72cda5ed2
```

**Response example**

```
[
  {
    "datetime": 1554064867544, 
    "id": "4f5ad86e-53f5-11e9-859b-ac1f6b2d9373", 
    "version": 1
  }
]
```


#### /api/diary/search
###### GET /api/diary/search

Search all URL diaries. Multiple filters are available. If no filters are specified, all content will be searched.
The '*' wildcard can be used, but can cause inaccurate results if the string is short. Multiple filters can be combined by using 'AND' after each filter.
A filter must always be directly followed by a colon.

**URL parameters**:

* `q` *Required* *(string)* - The search query. It is recommended to urlencode the contents of this parameter.
* `limit` *Optional* *Integer* - Maximum amount of diary IDs to return
* `offset` *Optional* *Integer* - Offset of the diary IDs list

**Available search filters**

* `url` - This specific URL
* `requests` - Match string in requested URL/URI strings
* `javascript` - Match string in executed javascript
* `signatures` - Match string in matched real-time signature (Not custom URL diary signatures)

**Status codes**

* `200` - Success
* `400` -  Invalid or missing parameter

**CURL Example**

Find all diaries that contains requests with '.jpg' and that have the executed javascript function 'eval'.

```bash
curl http://localhost:8080/api/diary/search?q=requests:.jpg AND javascript:eval(
```

**Response example**

```
[
  {
    "datetime": 1553795228121, 
    "id": "818e8ccb-5181-11e9-859b-ac1f6b2d9373", 
    "url": "example.com", 
    "version": 3
  }, 
  {
    "datetime": 1553795076786, 
    "id": "276e221c-5181-11e9-859b-ac1f6b2d9373", 
    "url": "example.net", 
    "version": 3
  }
]
```

#### /api/alerts/list
###### GET /api/alerts/list

Retrieve a list of all alerts.

**URL parameters**:

* `limit` *Optional* *Integer* - Maximum amount of alerts to return
* `offset` *Optional* *Integer* - Offset of the alerts list

**Status codes**

* `200` - Success

**CURL Example**

```bash
curl http://localhost:8080/api/alerts/list
```

**Response example**

```
[
  {
    "content": "The analysis of group 'exampledomains' has completed. ", 
    "diary_id": null, 
    "id": 2, 
    "level": 1, 
    "read": false, 
    "signature": null, 
    "target": null, 
    "task_id": null, 
    "timestamp": "2019-03-31 20:41:19", 
    "title": "Group analysis completed", 
    "url_group_name": "exampledomains"
  }, 
  {
    "content": "The analysis of group 'exampledomains' has started", 
    "diary_id": null, 
    "id": 1, 
    "level": 1, 
    "read": false, 
    "signature": null, 
    "target": null, 
    "task_id": null, 
    "timestamp": "2019-03-31 20:40:09", 
    "title": "Group analysis started", 
    "url_group_name": "exampledomains"
  }
]
```


#### /api/alerts/read
###### POST /api/alerts/read

Mark alerts as read.

**Form parameters**:

* `alert` *Optional* *(integer)* - The ID of a specific alert
* `groupname` *Optional* *(string)* - The name of a group. Marks all alerts as read for this group
* `markall` *Optional* *(bool)* - Mark all alertsas read

**Status codes**

* `200` - Success
* `400` - Invalid parameter

**CURL Example**

```bash
curl http://localhost:8080/api/alerts/read -F "alert=1"
```

**Response example**

```
{
  "message": "OK"
}

```


#### /api/alerts/delete
###### POST /api/alerts/delete

Delete alerts. All alerts can be clearerd or this can be done per alert, group, and level. Parameters
can be combined. For example, to delete all level 1 alerts for a specific group.

**Form parameters**:

* `alert` *Optional* *(integer)* - The ID of a specific alert
* `level` *Optional* *(integer)* - A level 1-3
* `groupname` *Optional* *(string)* - The name of a group. Delete all alerts for this group
* `clearall` *Optional* *(bool)* - Delete all alerts.

**Status codes**

* `200` - Success
* `400` - Invalid parameter

**CURL Example**

```bash
curl http://localhost:8080/api/alerts/delete -F "level=1" -F "groupname=newgroup"
```

**Response example**

```
{
  "message": "OK"
}
```


#### /api/pcap/(task)
###### GET /api/pcap/(integer:task_id)

Download the PCAP of a specific task. Alerts with detections contain a task_id. This can be used
to find potentially interesting PCAPs.

**URL parameters**:

* `exists` *optional* *(bool)* - Only check if the PCAP exists, instead of downloading.

**Status codes**

* `200` - Success
* `404` - PCAP does not exist

**CURL Example**

```bash
curl http://localhost:8080/api/pcap/1
```

