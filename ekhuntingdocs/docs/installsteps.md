# Setting up Mass URL analysis
Mass URL analysis uses multiple components that interact with each other. These components all need to be configured.

**1. Installing the Cuckoo package**

`$ git clone git@github.com:hatching/cuckoo-ekhunting.git && cd cuckoo-ekhunting`

`$ pip install .`

**2. Creating a Cuckoo working directory.**

`$ cuckoo init`

**3. Installing and configuring Elasticsearch 6.*.**

After setting up Elasticsearch, open `massurl.conf` in the Cuckoo cwd config directory. Fill in the configuration values.

Enabled massurl by setting enabled to ‘yes’.

Under [elasticsearch], add the Elasticsearch server to hosts = example.com:9200.

```
[massurl]
enabled = yes

[elasticsearch]
hosts = example.com:9200
```

**4. Starting the Cuckoo event messaging server.**
`$ cuckoo --debug eventserver`

At this point, it is possible to start the Mass URL web dashboard.

`$ cuckoo massurl -H 127.0.0.1 -p 8080`

**5. Setting up network capture replay for replay analyses.**

Install ‘**mitmproxy-0.18.2.tar.gz**‘ in the same virtualenv as Cuckoo.

`$ pip install mitmproxy-0.18.2.tar.gz`

Create a new virtualenv for mitmproxy3 and install **mitmproxy-4.0.0.tar.gz** in there:

`$ virtualenv mitmproxy3`

`$ source mitmproxy3/bin/activate`

`$ pip install mitmproxy-4.0.0.tar.gz`

Run mitmdump once, so the ‘cert.p12’ file is generated in ~/.mitmproxy.

`$ mitmdump`

Copy ‘**mitmproxy-ca-cert.p12**’ to ‘**$CWD/analyzer/windows/bin/cert.p12**’

Open ‘**$CWD/conf/auxiliary.conf**’ and enable replay analysis. Add the path of the mitmdump binary of the mitmproxy3 virtualenv to ‘**mitmdump**’

```
[replay]
enabled = yes
mitmdump = /tmp/mitmproxy3/mitmproxy3/bin/mitmdump
port_base = 51000
```

Start Cuckoo DNSserve. It is required to redirect all requests to mitmdump.

`$ cuckoo --debug dnsserve --hardcode 1.2.3.4 -H 192.168.56.1 --sudo`

**6. Start the Cuckoo rooter.**

`$ cuckoo --debug rooter --sudo --group <cuckoo user>`

**7. Install Golang, and download, compile, and run the real-time Onemon processor.**

`$ go get github.com/hatching/ekhunting/realtime`

`$ go install github.com/hatching/ekhunting/cmd/realtime`

`$ realtime localhost:42037 </home/<cuckoo user>/.cuckoo>`

**8. Install the provided VMCloak version and create Windows 7 analysis VMs with VMCloak.**

`$ pip install vmcloak`

Use the Windows 7 ISO found here: **http://cuckoo.sh/win7ultimate.iso**

Create a Windows 7 VM and install Firefox 41.0.2 and Internet Explorer 11.

`$ vmcloak install cuckoo1 firefox`

`$ vmcloak install cuckoo1 ie11`

<u>After installing the software</u>, use modify mode (vmcloak modify) and apply the patch by running **patchandgo.exe** inside the VM. The file can be posted to the VM by using the agent.

`$ curl -F "filepath=c:\patchandgo.exe" -F "file=@patchandgo.exe" http://192.168.56.101:8000/store`

Create the snapshots

`vmcloak snapshot w7x64 cuckoo1 192.168.56.101`

Lastly, add the VMs to Cuckoo

`$ cuckoo machine --add cuckoo1 192.168.56.101`

**9. Start Cuckoo.**

`$ cuckoo --debug`

**10. Start the Mass URL dashboard.**

`$ cuckoo massurl -H 127.0.0.1 -p 8080`