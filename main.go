package main

import (
	"flag"
	"fmt"
	"os"

	"diskshow/internal/scan"
	"diskshow/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

func main() {
	skipHidden := flag.Bool("skip-hidden", false, "跳过名称以 . 开头的隐藏项")
	includeSpecial := flag.Bool("all", false, "包含系统特殊目录（回收站、System Volume Information 等）")
	showVer := flag.Bool("version", false, "显示版本")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `diskshow %s — 终端里的 SpaceSniffer 风格磁盘分析

用法:
  diskshow [选项] [目录]

选项:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
示例:
  diskshow
  diskshow D:\code
  diskshow C:\Users

操作:
  点击矩形选中，双击或 Enter 进入目录
  右键或 Backspace 返回上一级
  Tab 切换树图 / 列表
  ? 查看全部快捷键
`)
	}
	flag.Parse()
	if *showVer {
		fmt.Println("diskshow", version)
		return
	}

	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开 %s: %v\n", path, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "不是目录: %s\n", path)
		os.Exit(1)
	}

	m := ui.New(path, scan.Options{
		SkipHidden:  *skipHidden,
		SkipSpecial: !*includeSpecial,
	})
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
		os.Exit(1)
	}
}
