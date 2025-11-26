import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "ClaudeInsightAsset",
  description: "ClaudeInsightAsset official website",
  head: [["link", { rel: "icon", href: "/claudeinsight.png" }]], // 浏览器标签页logo
  locales: {
    cn: { label: "简体中文", lang: "cn" },
    root: { label: "English", lang: "en" }
  },
  appearance: "dark",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: "/claudeinsight.png",
    nav: [
      { text: "Home", link: "/" },
      { text: "Guide", link: "./what-is-claudeinsight" }
    ],

    sidebar: [
      {
        text: "Introduction",
        items: [
          { text: "What is ClaudeInsight?", link: "./what-is-claudeinsight" },
          { text: "Quickstart", link: "./quickstart" },
          { text: "FAQ", link: "./faq" }
        ]
      },
      {
        text: "Tutorial",
        items: [
          { text: "Learn claudeinsight in 5 minutes", link: "./how-to" },
          { text: "How to use watch", link: "./watch" },
          { text: "How to use stat", link: "./stat" }
        ]
      },
      {
        text: "Reference",
        items: [{ text: "JSON Output Format", link: "./json-output" }]
      },
      {
        text: "Development",
        items: [
          { text: "How to build", link: "./how-to-build" },
          {
            text: "How to add a new protocol",
            link: "./how-to-add-a-new-protocol"
          },
          {
            text: "Debug Tips",
            link: "./debug-tips"
          }
        ]
      }
    ],

    socialLinks: [
      { icon: "github", link: "https://github.com/hengyoush/ClaudeInsight" }
    ],

    footer: {
      message:
        'Released under the <a href="https://github.com/hengyoush/ClaudeInsight/blob/main/LICENSE">Apache-2.0 license.',
      copyright:
        'Copyright © 2024-present <a href="https://github.com/hengyoush">Hengyoush'
    },

    search: {
      provider: "local"
    }
  }
});
