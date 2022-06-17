# AnyLANServer设计



# 1.文档介绍

## 1.1. 文档说明

本文档旨在说明Metadesk系统建议连接的过程

## 1.2. 参考文档

- STUN协议文档，https://www.rfc-editor.org/rfc/rfc5389

## 1.3. 其它资料

### 1.3.1. 测试可用的stun服务器

* turn2.l.google.com



# 2.AnyLANServer设计

## 2.1. 总体流程

### 2.1.1. 内网AnyLAN与AnyLANServer建立常连接设计

```mermaid
sequenceDiagram
participant client as AnyLAN(远端)
participant signal as AnyLANServer
participant server as AnyLAN(内网)

server ->> signal: 注册服务
signal -->> server: 返回注册成功，并维持心跳

opt 开始连接
client ->> signal: 请求连接服务端
signal ->> server: 有新的连接请求
server -->> signal: 返回准备就绪
signal -->> client: 返回可以连接，并返回会话ID
end

```

### 2.1.2. 会话ID交互过程

```mermaid
sequenceDiagram
participant client as AnyLAN(远端)
participant stun as Stun-Server
participant signal as AnyLANServer
participant server as AnyLAN(内网)

client ->> stun: 获取外网地址
stun -->> client: 返回外网地址
server ->> stun: 请求外网地址
stun -->> server: 返回外网地址
client ->> signal: 开始连接, 传入会话ID
server ->> signal: 开始连接, 传入会话ID
signal -->> client: 回复对方信息
signal -->> server: 回复对方信息

loop 开始尝试打洞
server ->> client: TCP交换信息,同步双方信息
client --> server: UDP对发消息
client ->> server: 回复发送完成，并报告自己最新的出口地址
end
```



## 2.2. TCP协议设计

TCP协议采用单行json数据封装格式；不同的指令以`\r\n`区分

### 2.2.1. 通用字段说明

| json字段 | 说明     | 数据类型 | 备注                                                         |
| -------- | -------- | -------- | ------------------------------------------------------------ |
| cmd      | 消息类型 | 字符串   | req:register : 注册<br>res:register : 回复注册<br>req:hert : 心跳<br>res:hert : 回复心跳<br>req:connect : 连接<br>res:connect : 回复连接<br>req:notify : 有新连接加入<br>res:notify : 回复新连接加入<br>req:contact : 开始握手<br>res:contact : 回复握手<br>req:punch : 开始尝试互相连接<br>res:punch : 回复尝试互相连接 |
| sid      | 会话ID   | 字符串   | 相同的链路，会话ID必须不相同                                 |
| err      | 错误消息 | 字符串   | 出错时，显示的字符串                                         |