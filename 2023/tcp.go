package main2023

/*
tcp可靠连接,需要3次握手，保证稳定，可靠，无差错，不丢失不重复按顺序到达
udp不可靠，不保证不丢失，不保证按顺序到达
一台主机，能有多少个TCP连接数
端口限制：一台机器对外连接，需要一个fd(文件标识符),一个fd占用一个端口，端口数是16位，其中主动监听端口，如22，21，23这种，基本小于65535个连接数（uint16）；ip_local_port_range参数控制
文件描述限制：每建立一个tcp连接，操作系统就会为其分配一个文件描述符号，打开这个文件可看到3方面限制。系统级，用户级，进程级
内存限制：tcp连接数过大可能会出现内存溢出
cpu限制：每个tcp连接数都会占用cpu资源，占用cpu资源过多，可能会导致死机

一开始都关闭,服务端监听某个端口
1.客户端发送一个FIN包请求建立连接,序号seq=x,客户端进入syn-sent阶段
2.服务端收到请求,返回一个FIN+ACK包,序号seq=y,确认序号ack=x+1,服务端进入syn-rcvd阶段,确认客户端序号有效,服务器能正常接收到客户端发送的数据,并统一创建新连接
3.客户端收到请求,返回一个ACK包,确认序号ack=y+1,表示确认收到服务器同意连接的信号

1.客户端主动请求关闭,发送一段报文,标志位是FIN,序号是seq=x,请求释放连接,之后客户端进入半关闭状态,停止客户端继续发送消息,但任然可以接受消息,进入FIN-WAIT-1阶段
2.服务端收到请求关闭连接请求，进入close-wait半关闭状态,并返回一段报文,状态位是ACK,确认序号ack=x+1,客户端收到消息后进入fin-wait-2阶段
//前两次挥手既然服务端知道了客户端想要释放连接，也让客户端知道了服务端了解了自己想要释放连接的请求
3.服务端发起第三报文,状态位FIN,序号seq=x,ack=y+1,表示消息处理完毕,已经做好了释放连接的准备
4.客户端发最后一次挥手,状态位ACK,确认号x+1,随后进入time-wait阶段,客户端等待完2MSL后结束time-wait阶段,关闭连接,完成四次挥手
为什么time-wait要等待2msl
客户端在发出最后的ack之后并不能知道服务端是否收到了消息，如果服务端在1s内没收到ack，会重新发送fin,计时重置

4层协议，应用层-传输层-网络层-物理层
7层协议,应用层-表示层-会话层-传输层-网络层-数据链路层-物理层
sync攻击-第三次握手不回应,占满半连接队列,网关超时设置

https加密过程
客户端->服务端, 第一随机数,发送自己的SSL/TSL版本号,需要服务端校验客户端的SSL版本
服务端->客户端, 服务器公钥、CA证书、第二随机数
客户端->服务端, 非对称加密(服务器公钥+预主秘钥)
服务端：解密得到预主秘钥, （第1随机数+第2随机数+预主秘钥）= 对称加密秘钥
客户端：得到（第1随机数+第2随机数+预主秘钥）= 对称加密秘钥
开始进行对称加密消息发送

线程池五种状态
runing 运行中
showdown 不接收新任务，但能处理已添加的任务
stop 不处理新任务，不处理已添加的任务，正在处理的任务会被中断
tiding 所有任务终止后，clt记录的任务数为0，接着会调用钩子terminated
terminated 线程池彻底终止

字节序
0x123456789
大端序：先存高位的那一端,0x12,0x34,0x45
小端序：先存低位的那一端,0x89,0x67,0x45

参数调优
tcp_syn_retries(syn重传)
tcp在建立连接的时候会向服务端发送syn包，这时候需要等待服务端的ack响应，如果服务端未响应，就会进行1。2.4.8.16.32这样翻倍重新发送信息，如果一定次数都还未收到响应
则会终止三次握手
tcp_max_syn_backlog（syn半连接队列大小）
tcp在syn后需要记录在服务端的半连接队列里面,如果超出了这个半连接队列，服务端无法再建立新的连接
tcp_syncookies（syc时如果半连接队列满了，则会用syncookie取代半连接）
如果syn的半连接队列满了，不会直接丢弃连接。如果开启了syncookie就可以在不使用syn队列的情况下成功建立连接
服务端计算一个值ack给客户端，客户端返回ack时，取出这个值，验证成功则成功建立连接
tcp_synack_retries（synack重传次数）
服务端向客户端发送syn+ack后，需要得到客户端的ack，如果服务端没有收到客户端的ack，则会进行重新发送syn+ack
tcp_abort_on_overflow(accept连接成功队列)
连接成功后会把连接从半连接队列移到这个accept成功连接队列,等进程调用accept的时候会把连接取出来，如果不能及时取出，会导致连接队列溢出，从而已连接好的tcp被丢弃
tcp_fastopen（cookie代替三次握手）
可绕过3次握手，客户端请求syn，服务端返回自己加密后的cookie，客户端再次请求连接带上这个cookie
tcp_sack（重传机制，开启sack重传方式）
tcp_dsack(重复重传机制,开始后启用d-sack)

什么是孤儿连接？
tcp调用了close后就意味这完全断开连接，完全断开不仅无法传输数据,也不能发送数据。此时，调用了close函数的一方的连接叫做孤儿连接，
tcp_orphan_retries（fin重传次数）
客户端请求fin关闭连接，未收到ack，重新发送fin
tcp_max_orphans（最大孤儿连接数）
遭到恶意攻击，孤儿连接无法重新发送fin包，如果孤儿连接数量大于它，新增的孤儿连接不再走四次挥手，而是直接发送RST报文强制关闭
tcp_fin_timeout(孤儿连接超时)
close函数关闭连接的孤儿连接，状态不可以持续太久
time_wait()
tcp_max_tw_bockets(最大time_wait连接数)
time_wait状态会占用端口并且无法提供新的连接使用，如果time_wait的连接数超过了这个，则新关闭的连接不再经历time_wait,而是直接关闭
tcp_tw_reuse（time_wait连接重用）
tcp_max_tw_buckets的数量不是越大越好，如果开启了这个tcp_tw_reuse，表示新来连接可以复用在time_wait下的端口
close_wait
tcp二次挥手后进入close_wait状态
tcp_tw_recycle


重传机制：
	关键名词：RTT(数据发送时刻接收到确认的时刻的差值) RTO(发送数据-超时重新发送的时间)
	方式1-超时重传：(以时间为驱动)发送方发送数据后，如果在一定时间内没有收到接收方的ack的确认应答报文，就会重新发送该数据。超时重传策略是超时间隔加倍
	方式2-快速重传：(以数据为驱动)发送端发送发送数据后，前面收到3个相同的ack报文时，如果定时器还没过期（可能是超时重传时间还没到）,发送端会重传丢失的报文。（不确定重传哪一个）
	方式3-SACK:选择性重传,只重传丢失的数据。接收方将缓存起来的数据告诉给发送方,这样发送方就知道有哪些数据已经收到了,哪些数据没收到。这个时候只要选择发送没收到的数据就行了
	方式4-Duplicate-SACK,使用SACK重传,如果数据发送成功，但是Ack的时候包丢了。发送方会重复发送，这个时候d-sack就起作用，会告诉发送方说他其实已经收到了数据，但是ack包丢了
滑动窗口
	前言:tcp发送数据不是一个send必须要等待ack后再发送下一个send。而是多批send数据，一次性也不能发送太多，要看接收方的缓冲区的接受能力
	tcp头部的window窗口大小，接收端告诉发送方自己还有多少缓冲区可以接收数据。接收方根据这个接收端的处理能力而发送数据
	发送方的窗口：已发送确认收到ack - 已发送未收到ack - 未发送且接收方准备接收的数据 - 未发送且接收方未具备接收的数据
	接收方的窗口：已成功接收并确认的数据 - 未收到数据但可以接受的数据 - 未收到数据不可以接受的数据
	接收窗口和发送窗口大小不相等（约等于关系,网络延迟关系不可保证窗口大小）
流量控制
	发送方不可能一直发，如果一直发，接收方处理不过来，未能即使的回复ack，就会触发发送方的重传机制，导致网络流量无端的消费
	窗口数据会放在操作系统的缓冲区（缓冲区会被调整），如果无法及时读取缓冲区的内容，窗口会受影响，所以需要流量控制
拥塞控制



=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================

redis的数据类型5种
字符串string、列表list、哈希hash、有序集合zset、无序集合set
redis的数据结构6种
简单动态字符串、双向连接、压缩列表、哈希表、跳表、整数数组
redis为什么用单线程
多线程编程会面临共享资源的并发访问控制问题,redis单线程主要指网络IO和键值读取是由一个线程来完成的,持久化、异步删除、集群数据同步是由其他线程执行的
redis为什么这么快
一方面redis大部分操作都在内存上完成,另一方面redis采用多路复用机制,使其在网络IO操作中能频繁处理大量的客户端请求,实现高吞吐率,
和他本身的数据结构有关系，操作键值最终就是对数据结构进行增删改查操作
aof,先写入磁盘后写入日志,不需要校验语句正确性,每个语句一条记录,为了避免aof重写太慢,aof提供了aof重写机制,将所有指令合并成一条一次性写入
rdb,快照数据,创建一个子进程,借用（写时复制技术）专门写rdb,避免主线程阻塞
redis4.0提出一个混合使用AOF和RDB方法,内存快照以一定的频率执行,在两次快照之间,使用aof日志记录这期间的所有命令操作
雪崩-同一时间内大量缓存数据过期，不失效，分散失效时间(随机)
穿透-不断发请求,redis没有-mysql没有导致数据库压力过大，布隆过滤器
击穿-热数据失效瞬间，大量请求到数据库,互斥锁，让线程回写缓存，数据永不过期
redis主库挂了怎么办，把之前的操作写入缓冲区，看门狗,哨兵集群选主投票
主从的流程
从库和主库建立连接，主库向从库发送offset和runID
主库执行bgsave命令，加载RDB文件，从库接收到RDB文件先清空数据，再加载RDB文件（可能在同步之前从库保存了其他数据，所以这样做）
主观下线：哨兵ping主库和从库，如果ping不通就把该库标记为主观下线
客观下线：如果多个哨兵斗殴ping不通主库和从库,就会把该库标记为客观下线
选主：每个哨兵给从库进行筛选+打分,最优的成为主库
redis慢操作
1.有可能哈希冲突，

=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
一条sql语句是怎么执行的
mysql 分为server层和存储引擎层,server层包括了连接器，查询缓存，分析器，优化器，执行器。

mysql特性
	原子性、要么全部成功要么全部失败
	一直性、数据最终一致
	持久性、落库
	隔离性、每个事务之间相互独立不相互影响
隔离级别
	读未提交
	读已提交
	可重复读
	串行
隔离级别可能会导致的问题
	脏读			读到了未提交的数据
	幻读			第一次读，插入新数据,第二次读读到了第一次没有的东西
	不可重复读	同一个数据,第一次读和第二次读数据不一致
普通索引和唯一索引的区别
	普通索引要回表，唯一索引在检测到后就会停止检测，普通索引要检测到第一个不满足才会停止
聚簇索引(主键索引)和非聚簇索引
	聚簇索引的叶子节点存放行数据，非聚簇索引叶子节点存放的是聚簇索引的ID,如果查找聚簇索引,先需要找到叶子节点的主键ID,再回表回到聚簇索引去找行数据
	聚簇索引一个表只有一个，非聚簇索引可以有很多个
using index 覆盖索引,查询的东西在索引中就有,不需要回表
using index condition 索引下推,查询的条件在索引中有,减少回表次数
using where 在查找使用索引的情况下,需要回表去查询所需的数据
myisam和innodb
	myisam不支持事务,innodb支持事务,myisam只支持表级锁，innodb支持表级别和行级锁,myisam的索引和数据文件是分离的,叶子节点保存的是数据文件的指针,innodb的索引和数据放在一起的,叶子节点存的是数据
全局锁：对整个数据库进行加锁，之后数据库处于只读状态
表锁2种：1.(元数据锁 meta data lock，MDL)：隐式使用 2.表锁(不支持行锁的引擎才会被用到)
什么是间隙锁gap lock
	行锁会产生幻读,用间隙锁解决问题，在行与行之间加个锁就叫间隙锁,行锁+间隙锁合成为next-key-lock,每个锁前开后闭区间
	间隙锁只有在可重复读的情况下才会生效
buffer poll,怎么管理buffer poll中的内存内？
	一个内存,用来缓存磁盘中的数据页；使用淘汰机制，有lru链表，free链表，flush链表
change buffer
	buffer poll中开辟出来的一个内存块，用来缓存插入操作
脏页-内存和磁盘的数据不一致
干净页-内存和磁盘的数据一致
redo log,bin log
	redo log记录所有操作记录，innodb持有,引擎层； 环状，指针循环写
	bin log mysql server层记录操作日志,所有的引擎都可以使用
redo和bin怎么关联
共同id XID,如果redo有prepare和commit，直接提交，如果redo只有prepare，拿XID去binlog找事务
怎么保证数据不丢失？事务是怎么提交的-为什么要2阶段提交
	2阶段提交保证原子性,持久到磁盘,重启后可以从日志文件中恢复数据,2阶段提交先写redo-prepare-binlog-commit
怎么确定binlog日志的完整性
	statement格式,记录的是原生语句,最后commit
	row模式	xid event
当前读和快照读
当前读,在高并发情况下读取最新的记录并且其他记录不可以修改这个记录,当前读加了悲观锁(互斥)
快照读,某一瞬间的数据,记录的是老数据,快照读没有加锁
MVCC
多版本并发控制,一个并发控制的方法,通过数据行的多个版本来实现数据库的并发控制,一个不采用锁来控制事务的方式,可同时解决脏读幻读不可重复读等事务隔离问题
mvcc3个隐含字段，记录事务的id,上个版本数据记录的回滚指针、隐藏主键row_id
怎么判断主库是否出问题
创建一个health表定时更新，自带的performance_schema表判断是否超出时间阈值


分库分表
将单个数据库表拆分成多个数据库表，从而提高数据库的可扩展性和并发性
哈希分库分表，将数据分配到不同的数据库表中，通过哈希算法计算键值的分布情况，并将键值分配到相对应的数据库表中
范围分库分表，将数据按照关键字的范围划分到不同的数据库表中，每个数据库表处理关键字范围内的数据
读写分离分库分表，读写分离分库分表将读操作和写操作分配到不同的数据库表中，从而提高数据库的并发性
使用场景
数据量大，比较难扩展的时候，
搞并发场景，
减少数据库单表数据大小限制

=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================

kafka
leader,follower(参与选举),observer（只读）
副本：集群leader，其他都是follower,leader负责写和读，同时把数据同步给follower
怎么保证消息不丢失
	生产者：producer可以设置acks参数为-1
	消费者：关闭自动提交，改成消息处理完成后手动提交偏移量
怎么保证消息不重复消费
	生产者：不重要(消费端去重就行)
	消费者：幂等性判断（比如简历去重表）
消息接收方式
-1 leader+follower都确认后才发送下一条数据
0 生产者负责发给broker，不关注响应
1 副本应答数大于等于这个数才继续
消息积压
生产速度大于消费速度就会产生积压。2方面考虑，1kafka处理不过来,增加topic partition副本个数。
消费者挂了或者消费者处理不过来，提高分批次拉取的数量


=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
zk
leader（读写），follower（处理读，参与选举），observer（只读）
所有的数据都保存在节点上，znode，多个znode构成一棵树的目录结构
分布式开源协调服务，可以用来用统一配置管理，分布式锁集群管理
持久化机制
1.事务日志，把命令保存在文件中。 2.定时快照
特性
顺序一致性。客户端发起请求，严格按顺序存入
原子性、所有集群中的集齐都应用成功请求，或者全部不成功
单一视图，所有服务数据和模型都一样
可靠，所有服务都一样，持久备份
实时性，所有的数据都是最新的

=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================


框架+优化

数据结构b树
b树和b+树
b树的的非叶子节点包含索引和数据，叶子节点不关联
b+树非叶子节点只包含索引，数据

算法

=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
内存逃逸
变量和数据有些存放在栈上有些存放在堆上，栈上的数据在函数结束的时候就会被销毁，而堆上的数据要考gc算法来处理，一般来说从栈上逃逸到堆上
或者一开始就直接在堆上的变量内存叫做内存逃逸


sync.once底层怎么实现
有个done的标识位，初始是0，执行一次之后cas变更为1

gmp模型
g 协程
p 处理器
m 线程

一个g最后占用cpu10s，防止别的g被饿死
一个p绑定的g了的队列最多可存放多少个g？256个
p的个数在哪里设置，$GOMAXPROCS设置
p在什么时候创建，程序启动时确定了p的最大数量n后，运行系统这个时候就会创建n个p
m在什么时候创建，没有足够的m来关联p的时候并运行其中的g的时候，会去找空闲的m，如果没有空闲的m，则会创建m
当一个m运行一个g遇到阻塞的时候，会找或创建一个m来接手当前的p
g可以无限创建吗？可以无限创建，但是受到机器配置的影响，一个协程4k内存的话，如果无限创建协程超出了内存的配置，就会保存


make和new的区别
make只用于切片，map，管道的初始化
new用于一些类型内存分配（初始化为0）
make返回引用类型本身，new返回指向指针

编程遵循：开闭原则，单一原则
defer和return谁先
return 后的语句和方法先执行，defer的语句后执行

cgo，cgo提供golang和c语言相互的调用的机制，可以通过cgo用golang调用c的接口
操作系统在CPU里加了一层专门用来管理虚拟内存和物理内存映射关系的东西，MMU（memory Management Unit）
TLB(translation lookaside buffer) MMU寻址时的缓存,MMU寻址的时候不会每次查磁盘地址,会先在TLB缓存里查找,如果找到了则直接返回
如果未找到,则查询虚拟缓存页表中查询关系

page(页,每个页有PTE)
span(页表,多个 页组成)
threadCache(freeList) //每个线程的内存缓存  小于256k为小内存 256k-1M中内存
centralCache(freeList)//多个线程共用的内存缓存(针对一二级内存,只分配小内存)
pageHeap(freeList、Large Span Set集合)//针对中大内存(申请内存的时候,跳过12级内存,直接在这里申请)
go内存中分为

正码 反码 补码
计算机中没有减法，在计算中中识别不了正数和负数，早期计算机设置0为正数，1为负数
正码 首位用1和0来区分正负数1表示负数,眼见的就是正码
反码 首位用1和0来区分正负数1表示负数,其余取反
补码 在反码的基础上+1



=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
=======================================================================================
CAP定理
分布式稳定 C一致性 A可用性 P是分区容错性，这三个最多只能同时满足2个
熔断
关闭，未触发熔断
半开，触发错误阈值，进入半开，限流，隔一段时间进行尝试放流，如果ok，则关闭，多次不ok全开
全开，限流，进制反问



日活2w
arpu 7
arppu 120
*/