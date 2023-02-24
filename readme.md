## v0.1
目标： 实现简单的tcp服务功能
1. 实现了简单的tcp
2. 完善tcp 的细节
3. echoHandler 实现
4. main 函数编写
5. 实现了原子操作的bool 类型

功能： 
1. 实现信号接收，可防止手动关闭时资源未释放问题
2. 实现每个连接由协程来处理
3. 增加了超时等待组，实现增加和删除客户端计量操作，另外设置也可以设置业务场景超时时间
4. 有log 信息


## v0.2
目标： 实现redis 协议解析器



## aof落盘功能测试命令
set 添加指令
`*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`
select切换数据库指令
`*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`

测试命令
这时数据库0和1都有key value 数据
`*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`
`*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
`*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`



## 单机版本增加集群功能
一致性哈希 ： 为了每增加1个或者2个机器时（微小增删），数据迁移比较方便（如果不采用此技术，增加1个会导致原先数据存储不一致问题）

用go实现一致性哈希



### 测试项目命令
set 添加： `*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`  set key value 
select 切换数据库：`*2\r\n$6\r\nselect\r\n$1\r\n1\r\n`
get 查询：`*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n`
ping 命令：`*1\r\n$4\r\nPING\r\n`
