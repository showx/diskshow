# diskshow

终端里的 [SpaceSniffer](http://www.udc.cz/spacesniffer/)：用 **Go TUI** 把目录占用画成可点击的嵌套矩形树图，一眼看出谁在吃磁盘。

## 体验

- **面积 = 占用**：越大的文件/目录，矩形越大
- **嵌套树图**：目录内部继续画出子项，不用点进去也能看到里面的大文件
- **点击下钻**：左键选中最内层矩形，双击或 Enter 进入该目录
- **实时扫描**：扫描过程中矩形会随结果长大，不用等全部扫完
- **列表视图**：Tab 切换到按大小排序的明细列表，带占比条

## 运行

需要 Go 1.21+。Windows 建议在 [Windows Terminal](https://aka.ms/terminal) 中运行，以获得鼠标和真彩色支持。

```bash
go build -o diskshow.exe .
.\diskshow.exe
.\diskshow.exe D:\code
.\diskshow.exe C:\Users
```

```text
diskshow [选项] [目录]
  -skip-hidden    跳过以 . 开头的隐藏项
  -all            包含回收站等系统特殊目录
  -version        显示版本
```

## 快捷键

| 操作 | 作用 |
|------|------|
| 鼠标左键 | 选中矩形（命中最内层文件/目录） |
| 双击 / Enter | 进入目录，查看其子项占用 |
| 右键 / Backspace | 返回上一级 |
| Tab | 树图 ↔ 列表 |
| ↑ ↓ ← → / hjkl | 同级项目间移动 |
| o | 在资源管理器中定位选中项 |
| ? | 帮助 |
| q | 退出 |

默认会跳过 `System Volume Information`、`$Recycle.Bin` 以及符号链接/联接，避免循环和权限噪声。
