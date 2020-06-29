QuickCache
======

Fast, concurrency-safe, evicting in-memory cache implemented in Golang.

### Feature

- Memory Limit
- LFU
- Concurrency-safe
- Low GC
- Low Race Conflict


- 内存占用上限
- 缓存淘汰策略
- 并发安全，低冲突
- Low GC


### 设计

之前写过LRU cache，类比着LRU Cache来看，根据特点选择合适的数据结构：
1. 内存占用上限  
    控制内存占用上限，防止OOM：
    []byte类型存储，可计算上限
2. 缓存淘汰策略  
    容量不足时淘汰掉老数据、不常使用的数据。
    使用ringbuf做存储，kv对排列起来存储，为了识别，需要定义一个header结构来存储kv对的信息。
    外加索引结构方便查询，第一反应当然是hash map
3. 并发安全，低冲突  
    分段锁
4. gc代价低  
    索引结构如果用普通的hash map，gc时会遍历kv，性能较差，需要特殊处理。因为已经求出了hash，又需要控制gc，可以考虑自行实现hash map。
    传统的拉链法用的是链表，gc不友好，可以考虑将链表并入数组中，比如i,2i,4i...ki算一个链表，或者是1...i这段连续内存算链表。
    考虑到局部性原理，连续内存更优。
    > 当 map 的 key 和 value 都不是指针，并且 size 都小于 128 字节的情况下，会把 bmap 标记为不含指针，这样可以避免 gc 时扫描整个 hmap
5. 其他  
    缓存优化：内存对齐，缓存行padding
    查询优化：利用hash进行查询加速
                                                                                         

整体设计就出来了：
1. 哈希分段，每段独占一个锁
2. 每段包含索引结构和存储结构
    1. 其中存储结构是一个固定大小的ringBuf，ringBuf中按kv对排列存储
    2. 索引结构是一个自行实现的hash map，将数组分段模拟链表，自带扩容机制
3. ringBuf的存储格式：header|key|value  
    其中header：keyLen|hash|valLen|valCap|deleted|bucketId

