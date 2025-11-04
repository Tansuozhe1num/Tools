#!/usr/bin/env python3
"""
简易画图软件启动脚本
一键打开HTML页面
"""

import webbrowser
import os
import sys
from pathlib import Path

def open_html_file():
    """打开HTML文件"""
    # 获取当前脚本所在目录
    current_dir = Path(__file__).parent
    html_file = current_dir / "index.html"
    
    # 检查文件是否存在
    if not html_file.exists():
        print(f"错误：找不到文件 {html_file}")
        print("请确保index.html文件与脚本在同一目录下")
        return False
    
    # 获取文件的绝对路径
    file_path = html_file.absolute()
    
    # 使用默认浏览器打开文件
    try:
        # 使用file://协议打开本地文件
        url = f"file://{file_path}"
        print(f"正在打开画图软件: {file_path}")
        webbrowser.open(url)
        print("画图软件已启动！")
        return True
    except Exception as e:
        print(f"打开文件时出错: {e}")
        return False

def main():
    """主函数"""
    print("=" * 50)
    print("简易画图软件启动器")
    print("=" * 50)
    
    # 检查当前目录文件
    current_dir = Path(__file__).parent
    files = list(current_dir.glob("*.html"))
    
    if files:
        print(f"检测到HTML文件: {', '.join(f.name for f in files)}")
    else:
        print("警告：未检测到HTML文件")
    
    # 打开HTML文件
    success = open_html_file()
    
    if success:
        print("\n启动成功！")
        print("功能说明：")
        print("- 选择铅笔或橡皮擦工具")
        print("- 调整颜色和画笔粗细")
        print("- 在画布上自由绘制")
        print("- 按Ctrl+S保存图片")
        print("- 点击清除按钮重置画布")
    else:
        print("\n启动失败，请检查文件是否存在")
    
    # 等待用户按键退出（在命令行中运行时有意义）
    try:
        input("\n按回车键退出...")
    except:
        pass  # 在非交互式环境中忽略

if __name__ == "__main__":
    main()