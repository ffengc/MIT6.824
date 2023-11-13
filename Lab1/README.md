# MapReduce

## 0.Reference
**代码参考仓库: [https://github.com/s09g/mapreduce-go/](https://github.com/s09g/mapreduce-go/)**
这一位前辈实现的是旧版本的课程实验，我是根据他的代码进行学习的，在我自己的基础上改成了新的版本
然后注释中也加入了我自己的解释和理解
**这位前辈对应的学习视频: [MIT 6.824 (2020 Spring) Lab1 - MapReduce 实现](https://www.bilibili.com/video/BV1Sy4y1j76o/?spm_id_from=333.999.0.0&vd_source=c92992dd205127af7566e47096862953)**


## 1. 运行方法
克隆仓库
```bash
git clone https://github.com/Yufccode/MIT6.824.git
```
进入Lab1的main目录
```bash
cd ./MIT6.824/Lab1/main
```
运行测试脚本
```bash
sh test-mr.sh
```

## 2. 遗留问题
### 2.1 Dialing错误的修复
输出结果:
```bash
parallels@ubuntu-linux-22-04-desktop:~/Project/MIT6.824/MIT6.824/Lab1/main$ bash test-mr.sh 
*** Starting wc test.
--- wc test: PASS
2023/11/13 09:10:06 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
*** Starting indexer test.
2023/11/13 09:10:13 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
2023/11/13 09:10:23 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
*** Starting reduce parallelism test.
2023/11/13 09:10:37 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
--- reduce parallelism test: PASS
*** Starting job count test.
2023/11/13 09:10:57 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
2023/11/13 09:10:57 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
2023/11/13 09:10:57 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
--- job count test: PASS
*** Starting early exit test.
2023/11/13 09:11:03 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
2023/11/13 09:11:06 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
--- early exit test: PASS
*** Starting crash test.
2023/11/13 09:13:21 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
2023/11/13 09:13:21 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
--- crash test: PASS
*** PASSED ALL TESTS
```
所有的测试结果都是PASS的，但是会出现这个问题:
```bash
2023/11/13 09:13:21 dialing:dial unix /var/tmp/824-mr-1000: connect: connection refused
```
这个报错出现的原因是worker调用askForTask发送RPC时coordinator已经退出。
```go
func (c *Coordinator) Done() bool {
    c.slock.Lock()
    defer c.slock.Unlock()

    time.Sleep(200 * time.Millisecond) // <- +
    return c.status == 2
}
```
在coordinator中的Done方法中增加一段睡眠，这是为了等待那些正处于轮询周期中短暂休眠状态的worker，这样一来这些worker会获取到coordinator已经退出的状态从而安全退出。

但是通过测试，我发现睡眠时间越久，出现Dial错误的概率越低，但是测试不通过的概率会变高

如果有解决这个方法的伙伴们可以直接私信与我讨论，谢谢。