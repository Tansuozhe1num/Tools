#!/bin/bash
# 简易画图软件启动脚本

echo "=================================================="
echo "简易画图软件启动器"
echo "=================================================="

# 检查HTML文件是否存在
if [ -f "index.html" ]; then
    echo "检测到画图软件文件"
    
    # 获取当前目录绝对路径
    CURRENT_DIR=$(pwd)
    HTML_FILE="$CURRENT_DIR/index.html"
    
    echo "正在打开: $HTML_FILE"
    
    # 根据不同操作系统使用不同命令打开
    case "$(uname -s)" in
        Darwin)    # macOS
            open "$HTML_FILE"
            ;;
        Linux)     # Linux
            xdg-open "$HTML_FILE" 2>/dev/null || \
            gnome-open "$HTML_FILE" 2>/dev/null || \
            echo "请手动在浏览器中打开 index.html"
            ;;
        CYGWIN*|MINGW32*|MINGW64*|MSYS*)  # Windows
            start "$HTML_FILE"
            ;;
        *)
            echo "未知操作系统，请手动在浏览器中打开 index.html"
            ;;
    esac
    
    echo "画图软件已启动！"
    echo ""
    echo "使用说明："
    echo "- 选择铅笔或橡皮擦工具"
    echo "- 调整颜色和画笔粗细" 
    echo "- 在画布上自由绘制"
    echo "- 按Ctrl+S保存图片"
    echo "- 点击清除按钮重置画布"
else
    echo "错误：找不到 index.html 文件"
    echo "请确保脚本与HTML文件在同一目录下"
fi

# 等待用户按键
read -p "按回车键退出..."