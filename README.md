# ClaudeInsight

<div align="center"> 
<i>One command to find slowest requests and identify the reasons.</i>
 <br/>
</div>

![](docs/public/ClaudeInsight-demo.gif)

<div align="center">  
 
[![GitHub last commit](https://img.shields.io/github/last-commit/hengyoush/ClaudeInsight)](#) 
[![GitHub release](https://img.shields.io/github/v/release/hengyoush/ClaudeInsight)](#) 
[![Test](https://github.com/hengyoush/ClaudeInsight/actions/workflows/test.yml/badge.svg)](https://github.com/hengyoush/ClaudeInsight/actions/workflows/test.yml) 

<a href="https://trendshift.io/repositories/12330" target="_blank"><img src="https://trendshift.io/api/badge/repositories/12330" alt="hengyoush%2FClaudeInsight | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>
[![Featured on Hacker News](https://hackerbadge.now.sh/api?id=42154583)](https://news.ycombinator.com/item?id=42154583)
<a href="https://hellogithub.com/repository/9e20a14a45dd4cd5aa169acf0e21fc45" target="_blank"><img src="https://abroad.hellogithub.com/v1/widgets/recommend.svg?rid=9e20a14a45dd4cd5aa169acf0e21fc45&claim_uid=temso5CUu6fB7wb" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" /></a>

</div>

[简体中文](./README_CN.md) | English

- [English Document](https://ClaudeInsight.io/)

## Table of Contents

- [ClaudeInsight](#ClaudeInsight)
  - [Table of Contents](#table-of-contents)
  - [What is ClaudeInsight](#what-is-ClaudeInsight)
  - [Examples](#examples)
  - [❗ Requirements](#-requirements)
  - [🎯 How to get ClaudeInsight](#-how-to-get-ClaudeInsight)
  - [📝 Documentation](#-documentation)
  - [⚙ Usage](#-usage)
  - [🏠 How to build](#-how-to-build)
  - [Roadmap](#roadmap)
  - [🤝 Feedback and Contributions](#-feedback-and-contributions)
  - [🙇‍ Special Thanks](#-special-thanks)
  - [🗨️ Contacts](#️-contacts)
  - [Star History](#star-history)

## What is ClaudeInsight

claudeinsight is an **eBPF-based** network issue analysis tool that enables you to
capture network requests, such as HTTP and other protocol requests.
It also helps you analyze abnormal network issues and quickly troubleshooting
without the complex steps of packet capturing, downloading, and analysis.

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
   ![ClaudeInsight find big response](docs/public/whatClaudeInsight.gif)

3. **In-Depth Kernel-Level Latency Details**: In real-world, slow queries to
   remote services can be challenging to diagnose precisely. claudeinsight provides
   kernel trace points from the arrival of requests/responses at the network
   card to the kernel socket buffer, displaying these details in a visual
   format. This allows you to identify exactly which stage is causing delays.

![ClaudeInsight time detail](docs/public/timedetail.jpg)

4. **Lightweight and Dependency-Free**: Almost zero dependencies—just a single
   binary file and one command, with all results displayed in the command line.

5. **Automatic SSL Traffic Decryption** : All captured requests and responses
   are presented in plaintext.

## Examples

**Capture HTTP Traffic with Latency Details**

Run the command:

```bash
./claudeinsight watch http
```

The result is as follows:

![ClaudeInsight quick start watch http](docs/public/qs-watch-http.gif)

**Identify the Slowest Requests in the Last 5 Seconds**

Run the command:

```bash
 ./claudeinsight stat --slow --time 5
```

The result is as follows:

![ClaudeInsight stat slow](docs/public/qs-stat-slow.gif)

## ❗ Requirements

claudeinsight currently supports kernel versions 3.10(from 3.10.0-957) and 4.14 or
above (with plans to support versions between 4.7 and 4.14 in the future).

> You can check your kernel version using `uname -r`.

## 🎯 How to get ClaudeInsight

You can download a statically linked binary compatible with amd64 and arm64
architectures from the
[release page](https://github.com/hengyoush/ClaudeInsight/releases):

```bash
tar xvf ClaudeInsight_vx.x.x_linux_amd64.tar.gz
```

Then, run claudeinsight with **root privilege**:

```bash
sudo ./claudeinsight watch
```

If the following table appears:
![ClaudeInsight quick start success](docs/public/quickstart-success.png) 🎉
Congratulations! claudeinsight has started successfully.

## 📝 Documentation

[English Document](https://ClaudeInsight.io/)

## ⚙ Usage

The simplest usage captures all protocols currently supported by ClaudeInsight:

```bash
sudo ./claudeinsight watch
```

Each request-response record is stored as a row in a table, with each column
capturing basic information about that request. You can use the arrow keys or
`j/k` to move up and down through the records:
![ClaudeInsight watch result](docs/public/watch-result.jpg)

Press `Enter` to access the details view:

![ClaudeInsight watch result detail](docs/public/watch-result-detail.jpg)

In the details view, the first section shows **Latency Details**. Each block
represents a "node" that the data packet passes through, such as the process,
network card, and socket buffer.  
Each block includes a time value indicating the time elapsed from the previous
node to this node, showing the process flow from the process sending the request
to the network card, to the response being copied to the socket buffer, and
finally read by the process, with each step’s duration displayed.

The second section provides **Detailed Request and Response Content**, split
into Request and Response parts, and truncates content over 1024 bytes.

For targeted traffic capture, such as HTTP traffic:

```bash
./claudeinsight watch http
```

You can narrow it further to capture traffic for a specific HTTP path:

```bash
./claudeinsight watch http --path /abc
```

Learn more: [ClaudeInsight Docs](https://ClaudeInsight.io/)

## 🏠 How to build

👉 [COMPILATION.md](./COMPILATION.md)

## Roadmap

The claudeinsight Roadmap shows the future plans for ClaudeInsight. If you have feature
requests or want to prioritize a specific feature, please submit an issue on
GitHub.

_1.6.0_

1. Support for postgresql protocol parsing.
2. Support for HTTP2 protocol parsing.
3. Support for DNS protocol parsing.
4. Support for GnuTLS.

## 🤝 Feedback and Contributions

> [!IMPORTANT]
>
> If you encounter any issues or bugs while using the tool, please feel free to
> ask questions in the issue tracker.

## 🙇‍ Special Thanks

During the development of ClaudeInsight, some code was borrowed from the following
projects:

- [eCapture](https://ecapture.cc/zh/)
- [pixie](https://github.com/pixie-io/pixie)
- [ptcpdump](https://github.com/mozillazg/ptcpdump)

## 🗨️ Contacts

For more detailed inquiries, you can use the following contact methods:

- **My Email:** [hengyoush1@163.com](mailto:hengyoush1@163.com)
- **My Blog:** [http://blog.deadlock.cloud](http://blog.deadlock.cloud/)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=hengyoush/ClaudeInsight&type=Date)](https://star-history.com/#hengyoush/ClaudeInsight&Date)

[Back to top](#top)
