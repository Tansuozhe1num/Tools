# 简易画图软件

一个基于HTML5 Canvas的简易画图软件，模仿Windows画图工具的基本功能。

## 功能特性

- 🎨 **画笔工具**：铅笔和橡皮擦
- 🎯 **颜色选择**：支持自定义颜色选择
- 📏 **画笔粗细**：可调节画笔大小（1-50px）
- 🧹 **清除画布**：一键清除所有内容
- 📱 **移动端支持**：支持触摸屏设备
- ⌨️ **快捷键**：Ctrl+S保存图片

## 使用方法

### 方法一：一键启动（推荐）
- **macOS/Linux**: 在终端中运行 `./launch.sh`
- **Windows**: 双击运行 `launch.bat`
- **Python**: 运行 `python launch_app.py`

### 方法二：手动打开
1. 用浏览器打开 `index.html` 文件
2. 选择画笔工具（铅笔或橡皮擦）
3. 调整颜色和画笔粗细
4. 在画布上开始绘制
5. 使用清除按钮重置画布
6. 按 Ctrl+S 保存图片

## 文件结构

```
.
├── index.html         # 主页面
├── style.css          # 样式文件
├── script.js          # 核心功能脚本
├── launch_app.py      # Python启动脚本
├── launch.sh          # Shell启动脚本（macOS/Linux）
└── README.md          # 说明文档
```

## 技术栈

- HTML5 Canvas
- CSS3
- JavaScript (ES6+)

## 浏览器兼容性

支持所有现代浏览器（Chrome、Firefox、Safari、Edge等）