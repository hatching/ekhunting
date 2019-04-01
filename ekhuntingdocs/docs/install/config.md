# Mass URL configuration

The Mass URL configuration is explained here. The Mass URL configuration file is only available after [setting up](/install/installsteps) Mass URL.

A newly generated `massurl.conf`.

```
[massurl]
# Enable the Mass URL analysis component. This requires the Elasticsearch
# server in this config to also be configured.
enabled = no

# Try to recover TLS keys from collected behavioral log and PCAP.
# Required to decrypt HTTPs traffic. This can slow down the analysis.
extract_tls = yes

[elasticsearch]
# Comma-separated list of ElasticSearch hosts. Format is IP:PORT, if port is
# missing the default port is used.
# Example: hosts = 127.0.0.1:9200, 192.168.1.1:80
hosts = 127.0.0.1:9200

# Increase default timeout from 10 seconds, required when indexing larger
# analysis documents.
timeout = 300

# The unique index name that will be used to store URL diaries.
diary_index = urldiary

# The unique index name that will be used to network requests that are related
# to a specific URL diary.
related_index = related

# How many of the first bytes of a network request and response should be stored?
request_store = 16384

[eventserver]
# The IP the Cuckoo event client should connect to, to receive mass url analysis events
# from the Cuckoo event server
ip = 127.0.0.1

# The port the Cuckoo event server is listening on
port = 42037

[retention]
# Manages how long URL diary entries, tasks, and alerts are stored.
# it is recommended to enable this feature, as large URL groups can causes thousands
# of large tasks be to be created each day.
enabled = yes

# The amount of days a task is kept in the database and on disk. Enter 0 to keep tasks forever.
tasks = 5

# Move the PCAP file of a task when the task is deleted?
keep_pcap = yes

# The amount of days URL diaries should be kept. Upon removal, both the URL diary and the
# request logs for it are removed. Enter 0 to keep url diaries forever.
urldiaries = 365

# The amount of days an alert should be kept. Enter 0 to keep alerts forever.
alerts = 365
```

#### Results cleanup/retention

Cuckoo Mass URL analysis creates a large amount of data, such as URL diaries and Cuckoo tasks. Some data is not used anymore by Cuckoo MassURL after a group analysis is finished. Large URL groups can result in thousands of tasks added each day, increasing the database size, and using up a lot of disk space with unused data.

Analysis retention lets the operator configure how many days specific results should be kept. The retention feature will automatically remove it after the configured amount of days. <u>Analysis retention is enabled by default.</u>

Retention can be configure for:

**Tasks**

Tasks older than X days are removed from the database and from disk. By default, the network dump (dump.pcap) is kept upon task removal. It is moved to $CWD/storage/files/pcap. This can be disabled. The default setting for task removal is 5 days.

**URL Diaries**

Removes URL diaries and any request logs related to a URL diary that is removed. By default, URL diary entries are removed after 365 days.

**Alerts**

Removes alerts older than the specified amount of days. By default, alerts are removed after 365 days.