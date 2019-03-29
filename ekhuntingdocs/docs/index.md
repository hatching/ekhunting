# Cuckoo Mass URL analysis

Cuckoo Mass URL analysis was developed by Hatching International and CERT-EE during the Exploit Kit Hunting project. The project was co-financed by the Connecting Europe Facility of the European Union.

## What is it?

Cuckoo Mass URL analysis (MassURL hereafter) is a new addition to Cuckoo Sandbox. It is aimed analyzing large amounts (100k+) of URLs a day, without needing multiple servers.
With MassURL, it is possible to create large 'URL groups' and schedule these to be analyzed every X days on the environment matching the configured 'analysis profile(s)'.
The Mass URL component of Cuckoo will take care of creating tasks and tracking them. URLs are opened in batches inside VMs, while a Windows kernel driver tracks the behavior of all opened
browsers and will malicious processes.

Collected data is processed and reported real-time. Collected HTTP requests and executed Javascript are attributed to the webpage/URL that caused them. This data is stored in a separate 'URL diary' for each URL.
Every new analysis, a new URL diary is created for each analyzed URL. The full URL diary history can be searched and signatures can be created that raise alerts if data in a URL diary matches it.

### The operator dashboard
Cuckoo MassURL comes with a operator dashboard which an operator can use to fully use MassURL. It is where URL groups are created, results are viewed, and most important: where the live events will be shown in cause of a possible malware infection of one of the analyzed webpages.

### Setting up