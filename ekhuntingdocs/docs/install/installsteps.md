# Setting up Mass URL analysis
Mass URL analysis uses multiple components that interact with each other. These components all need to be configured.
Not all components are in the Mass URL Cuckoo package. All supporting tools and components are located in [the ekhunting repository](https://github.com/hatching/ekhunting){target=_blank}.

These steps are meant to get the minimum setup. Follow the steps in order.

!!! note "Note"
    When specifying `$CWD`, this refers to the Cuckoo working directory that is used.

    When specifying `EKTOOLS`, this refers to the directory of the cloned supporting components.

**1. Clone the EKHunting supporting components repository**

```bash
$ git clone git@github.com:hatching/ekhunting.git
```

**2. Create a new virtualenv for Cuckoo Mass URL and install the Cuckoo package**

```bash
$ git clone git@github.com:hatching/cuckoo-ekhunting.git && cd cuckoo-ekhunting
$ pip install .
```

**3. Create a Cuckoo working directory.**

```bash
$ cuckoo init
```

**4. Installing and configuring Elasticsearch 6.*.**

Install Elasticsearch 6.*. A guide can be found [here](https://www.elastic.co/guide/en/elasticsearch/reference/current/deb.html){target=_blank}.

After setting up Elasticsearch, open `$CWD/conf/massurl.conf`. Fill in the configuration values.

Enable massurl by setting enabled to ‘yes’.

Under [elasticsearch], add the Elasticsearch server: hosts = example.com:9200.

```
[massurl]
enabled = yes

[elasticsearch]
hosts = example.com:9200
```

**5. Start the Cuckoo event messaging server.**

The event messaging server is used by multiple components to send messages/exchange information.

```bash
$ cuckoo --debug eventserver
```

At this point, it is possible to start the Mass URL web dashboard.

```bash
$ cuckoo massurl -H 127.0.0.1 -p 8080
```

**6. Setting up network capture replay for replay analyses.**

Install ‘**EKTOOLS/tools/mitmproxy-0.18.2.tar.gz**‘ in the same virtualenv as Cuckoo.

```bash
$ pip install EKTOOLS/tools/mitmproxy-0.18.2.tar.gz
```

Create a new virtualenv for mitmproxy3 and install **mitmproxy-4.0.0.tar.gz** in there:

```bash
$ virtualenv mitmproxy3
$ source mitmproxy3/bin/activate
$ pip install EKTOOLS/tools/mitmproxy-4.0.0.tar.gz
```

Run mitmdump once, so the ‘mitmproxy-ca-cert.p12’ file is generated in ~/.mitmproxy.

```bash
$ mitmdump
```

Copy ‘**~/mitmproxy-ca-cert.p12**’ to ‘**$CWD/analyzer/windows/bin/cert.p12**’

Open ‘**$CWD/conf/auxiliary.conf**’ and enable replay analysis. Add the path of the mitmdump binary of the mitmproxy3 virtualenv to ‘**mitmdump**’

```
[replay]
enabled = yes
mitmdump = /tmp/mitmproxy3/mitmproxy3/bin/mitmdump
port_base = 51000
```

Start Cuckoo DNSserve. It is required to redirect all requests to mitmdump.

```bash
$ cuckoo --debug dnsserve --hardcode 1.2.3.4 -H 192.168.56.1 --sudo
```

**7. Start the Cuckoo rooter.**

```bash
 cuckoo --debug rooter --sudo --group <cuckoo user>
```

**8. Install Golang, and download, compile, and run the real-time Onemon processor.**

```bash
$ go get github.com/hatching/ekhunting/realtime
$ go install github.com/hatching/ekhunting/cmd/realtime
$ realtime localhost:42037 </home/<cuckoo user>/.cuckoo>
```

**9. Install VMCloak and create Windows 7 analysis VMs with VMCloak.**

```bash
$ pip install vmcloak
```

Use the Windows 7 ISO found here: http://cuckoo.sh/win7ultimate.iso. The kernel monitor has been tested to work with this version.

Create a Windows 7 VM and install Firefox 41.0.2, Internet Explorer 11 and any other software such as Flash and Java.

```bash
$ vmcloak init --win7x64 --ramsize 6144 --cpus 2 w7x64
```
```bash
$ vmcloak install cuckoo1 firefox
$ vmcloak install cuckoo1 ie11
```

<u>After installing the software</u>, use modify mode (vmcloak modify) and disable PatchGuard by opening **EKTOOLS/tools/patchandgo.exe** inside the VM. This will disable PatchGuard on Windows, allowing the kernel monitor to be loaded.

The file can be posted to the VM by using the agent.

```bash
$ curl -F "filepath=c:\patchandgo.exe" -F "file=@EKTOOLS/tools/patchandgo.exe" http://192.168.56.101:8000/store
```

Create the snapshots

```bash
$ vmcloak snapshot w7x64 cuckoo1 192.168.56.101
```

Lastly, add the VMs to Cuckoo

```bash
$ cuckoo machine --add cuckoo1 192.168.56.101
```

**10. Start Cuckoo and the Mass URL dashboard**

```bash
$ cuckoo --debug
$ cuckoo massurl -H 127.0.0.1 -p 8080
```
