# Remote Desk

## 目前处于开发状态中

---

### 项目概述

客户端使用Go语言编写, UI界面采用Fyne框架, 得益于Go语言良好的跨平台性 支持多种操作系统 Win/Linux/MacOS/Android/iOS

服务端同样采用Go语言编写, 采用WebRtc框架, 集信令服务, ice服务, 验证服务于一体

# 核心功能

## 客户端

- 支持控制端和被控端两种模式
    - 简洁直观的工具栏
        - 全屏显示
        - 状态显示
        - 画面质量调节
        - 帧率显示
        - 显示器选择
- 实时屏幕共享和远程控制
    - 可调节画面质量和帧率
- 内置性能监控
    - 实时显示 CPU 使用率
    - 内存占用监控
    - 网络延迟统计
    - 运行时间统计
    - Goroutine 数量监控
- 心跳机制 维持连接 减少重新打洞 提升稳定性

## 服务端

### 服务端采用WebRtc框架

参考教程为 [WebRtc教程](https://tonybai.com/2024/12/14/webrtc-first-lesson-how-connection-estabish/)
这个教程有很多篇文章 都可以看下

其流程图为: (图片截取自上述教程)  
<img src="/png/webRtc_step.png"  width = "30%" />