---
prev:
  text: "ClaudeInsight 是什么"
  link: "./what-is-ClaudeInsight"
next: false
---

# 快速开始

## 安装要求

**内核版本要求**

For amd64:

- 3.x: 3.10.0-957 版本及以上内核
- 4.x: 4.14 版本以上内核
- 5.x, 6.x: 全部支持

For arm64:

- 5.5 以上

## 安装并运行 {#prerequire}

你可以从 [release page](https://github.com/hengyoush/ClaudeInsight/releases)
中下载以静态链接方式编译的适用于 amd64 和 arm64 架构的二进制文件：

```bash
tar xvf ClaudeInsight_vx.x.x_linux_amd64.tar.gz
```

然后以 **root** 权限执行如下命令：

```bash
sudo ./claudeinsight watch
```

如果显示了下面的表格： ![ClaudeInsight quick start success](/quickstart-success.png)
🎉 恭喜你，ClaudeInsight 启动成功了。

> [!TIP]
>
> 如果上面的命令执行失败了？没关系，在这个 [FAQ](./faq)
> 里看看有没有符合你的情况，如果没有欢迎提出
> [github issue](https://github.com/hengyoush/ClaudeInsight/issues) !

## 常见问题

请查看：[常见问题](./faq)

## 下一步

- 快速了解 claudeinsight 的使用方法，请查看：[5 分钟学会使用 ClaudeInsight](./how-to)
