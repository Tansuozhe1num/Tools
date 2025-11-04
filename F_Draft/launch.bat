@echo off
chcp 65001 >nul
echo ==================================================
echo 简易画图软件启动器
echo ==================================================

REM 检查HTML文件是否存在
if not exist "index.html" (
    echo 错误：找不到 index.html 文件
    echo 请确保批处理文件与HTML文件在同一目录下
    pause
    exit /b 1
)

echo 检测到画图软件文件
echo 正在启动画图软件...

REM 使用默认浏览器打开HTML文件
start "" "index.html"

echo.
echo 画图软件已启动！
echo.
echo 使用说明：
echo - 选择铅笔或橡皮擦工具
echo - 调整颜色和画笔粗细
echo - 在画布上自由绘制
echo - 按Ctrl+S保存图片
echo - 点击清除按钮重置画布
echo.
pause