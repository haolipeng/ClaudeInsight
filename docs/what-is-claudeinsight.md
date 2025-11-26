---
next:
  text: "Quickstart"
  link: "./quickstart"
prev: false
---

# What is claudeinsight ？{#what-is-ClaudeInsight}

claudeinsight is a Network Traffic Analyzer that provides real-time, packet-level to
protocol-level visibility into a host's internal network, capturing and
analyzing all inbound and outbound traffic.

## Why ClaudeInsight?

> There are already many network troubleshooting tools available, such as
> tcpdump, iftop, and netstat. So, what benefits does claudeinsight offer?

### Drawbacks of Traditional Packet Capture with tcpdump

1. **Difficulty filtering based on protocol-specific information**: For example,
   in the case of the HTTP protocol, it's challenging to capture packets based
   on a specific HTTP path, requiring tools like Wireshark/tshark for secondary
   filtering.
2. **Difficulty filtering packets based on the sending or receiving process**:
   especially when multiple processes are deployed on a single machine and you
   only need to capture packets for a specific process.
3. **Low troubleshooting efficiency**: The typical troubleshooting process
   involves using tcpdump in the production environment to capture packets and
   generate a pcap file, then downloading it locally for analysis with tools
   like Wireshark/tshark, often consuming a significant amount of time.
4. **Limited analysis capabilities**: Tcpdump only provides basic packet capture
   capabilities with minimal advanced analysis, requiring pairing with
   Wireshark. Traditional network monitoring tools like iftop and netstat offer
   only coarse-grained monitoring, making it challenging to identify root
   causes.
5. **Lacking the functionality to analyze encrypted traffic**: such as SSL
   protocol requests, cannot be viewed in plain text.

### What claudeinsight Can Offer You

1. **Powerful Traffic Filtering**: Not only can filter based on traditional
   IP/port information, can also filter by process, L7 protocol information,
   request/response byte size, latency, and more.

```bash
# Filter by pid
./claudeinsight watch --pids 1234
# Filter by response byte size
./claudeinsight watch --resp-size 10000
```

2. **Advanced Analysis Capabilities** : Unlike tcpdump, which only provides
   fine-grained packet capture, claudeinsight supports aggregating captured packet
   metrics across various dimensions, quickly providing the critical data most
   useful for troubleshooting.  
   Imagine if the bandwidth of your HTTP service is suddenly maxed out—how would
   you quickly analyze `which IPs` and `which  requests` are causing it?  
   With ClaudeInsight, you just need one command: `claudeinsight stat http --bigresp` to find
   the largest response byte sizes sent to remote IPs and view specific data on
   request and response metrics.  
   ![ClaudeInsight find big response](/whatClaudeInsight.gif)

3. **In-Depth Kernel-Level Latency Details**: In real-world, slow queries to
   remote services can be challenging to diagnose precisely. claudeinsight provides
   kernel trace points from the arrival of requests/responses at the network
   card to the kernel socket buffer, displaying these details in a visual
   format. This allows you to identify exactly which stage is causing delays.

![ClaudeInsight time detail](/timedetail.jpg)

4. **Lightweight and Dependency-Free**: Almost zero dependencies—just a single
   binary file and one command, with all results displayed in the command line.

5. **Automatic SSL Traffic Decryption** : All captured requests and responses
   are presented in plaintext.

## When to Use claudeinsight {#use-cases}

- **Capture Request and Response**

claudeinsight provides the **watch** command, allowing you to filter and capture
various traffic types. It supports filtering based on process ID, as well as IP
and port. Additionally, you can filter based on protocol-specific fields, such
as HTTP paths.
The captured traffic includes not only the request and response content but also
detailed timing information, such as the time taken for requests to go from
system calls to the network card and for responses to travel from the network
card to the socket buffer and then to the process.

- **Analyze Abnormal Flow Path**

ClaudeInsight’s stat command can help you quickly identify abnormal links. The stat
command supports aggregation across multiple dimensions.

For example, it can aggregate by remote IP, allowing you to quickly analyze
which remote IP is slower. claudeinsight also supports various metrics, such as
request-response latency and request-response size. With these features, you can
resolve 80% of network issues quickly.

- **Global Dependency Analysis** <Badge type="tip" text="beta" />

Sometimes, you may need to know which external resources a machine depends on.
claudeinsight offers the `overview` command to capture all external resources a machine
relies on and their latency in a single command.

## Basic Examples

**Capture HTTP Traffic with Latency Details**

Run the command:

```bash
./claudeinsight watch http
```

The result is as follows:

![ClaudeInsight quick start watch http](/qs-watch-http.gif)

**Identify the Slowest Requests in the Last 5 Seconds**

Run the command:

```bash
 ./claudeinsight stat --slow --time 5
```

The result is as follows:

![ClaudeInsight stat slow](/qs-stat-slow.gif)
