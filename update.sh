#!/bin/bash

# 更新项目脚本
# 功能：执行git pull更新代码并清理server/logs目录

echo "开始更新项目..."

# 确保脚本在项目根目录下执行
if [ ! -d ".git" ]; then
  echo "错误：请在项目根目录下运行此脚本"
  exit 1
fi

# 执行git pull更新代码
echo "正在从远程仓库拉取最新代码..."
git pull

# 检查git pull命令是否成功
if [ $? -ne 0 ]; then
  echo "错误：git pull 命令执行失败，请检查网络连接或git配置"
  exit 1
fi

echo "代码更新成功！"

# 删除server/logs目录
echo "正在清理日志文件..."
if [ -d "server/logs" ]; then
  rm -rf server/logs
  echo "server/logs目录已删除"
else
  echo "server/logs目录不存在，无需清理"
fi

echo "项目更新完成！"