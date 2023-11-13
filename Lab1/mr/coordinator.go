package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type CoordinatorTaskStatus int
const (
	Idle CoordinatorTaskStatus = iota
	InProgress
	Completed
)
type State int // Coordinator的状态
const (
	Map State = iota
	Reduce
	Exit
	Wait
)
type Task struct {
	Input         string
	TaskState     State
	NReducer      int
	TaskNumber    int
	Intermediates []string
	Output        string
}

// CoordinatorTask 就是把Task再封装一层，带上Task开始的时间，Task的状态
// 那么Task自己本身就不用知道自己是在Map还是在reduce了
// 用的时候直接用CoordinatorTask
type CoordinatorTask struct{
	TaskStatus    CoordinatorTaskStatus 
	StartTime     time.Time
	TaskReference *Task
}


type Coordinator struct {
	// Your definitions here.
	TaskQueue     chan *Task          // 等待执行的task
	TaskMeta      map[int]* CoordinatorTask // 当前所有task的信息
	CoorPhase     State               // 当前Coordinator的阶段（Map/Reduce）
	NReduce       int
	InputFiles    []string
	Intermediates [][]string // Map任务产生的R个中间文件的信息
}

var mu sync.Mutex

// Your code here -- RPC handlers for the worker to call.

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) GetTask(args *ExampleArgs, reply *Task) error {
	// args:
	// ExampleArgs就是worker丢过来的东西，也没啥用，知道一个数字就行了
	// Task是一个返回型参数，就是要给Worker第一个Task
	// worker会调用这个func
	mu.Lock()
	defer mu.Unlock()
	if len(c.TaskQueue) > 0 {
		// 如果TaskQueue里面有东西，就把东西丢出去
		// 如果TaskQueue没东西，你来获取任务也没用
		*reply = *<-c.TaskQueue
		// 记录task的启动时间
		c.TaskMeta[reply.TaskNumber].TaskStatus = InProgress
		c.TaskMeta[reply.TaskNumber].StartTime = time.Now()
	} else if c.CoorPhase == Exit {
		// 如果此时c的状态是退出了，你来获取任务也没用
		*reply = Task{TaskState: Exit}
	} else {
		// 任务队列里面如果没有任务，就让worker阻塞等待
		*reply = Task{TaskState: Wait}
	}
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
//
func (c *Coordinator) Done() bool {
	// Your code here.
	mu.Lock()
	defer mu.Unlock()
	ret := (c.CoorPhase == Exit)
	// time.Sleep(3000 * time.Millisecond)
	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		TaskQueue: make(chan *Task, max(nReduce, len(files))),
		// 最少也会创建file个位置给你的，当然，如果reduce数目很大
		// 当然如果reduce数量很大，为了保证reduce的时候不阻塞，也会最少创建nReduce个任务
		TaskMeta: make(map[int]* CoordinatorTask), // 保留所有Task的引用
		// TaskMeta有什么用呢
		// 这样可以得到TaskID找到整个Task了
		CoorPhase: Map, // 表示状态，现在是在Map还是在reduce? 一开始肯定是先Map
		NReduce: nReduce,
		InputFiles: files,
		Intermediates: make([][]string, nReduce), // 中间
	}

	// Step1. 启动MapReduce, 将输入文件切分成大小在16-64MB之间的文件
	c.createMapTask()

	// Step2. 在一组多个机器上启动用户程序
	c.server()
	// 启动一个goroutine检查超时任务
	go c.catchTimeout()
	return &c
}

// Step3. worker向领导者索取任务，领导者给worker指定任务，领导者选择空闲的worker基于map/reduce任务


// ----------------- Utils ----------------- //
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// task和master的状态应该一致。如果在Reduce阶段收到了迟来MapTask结果，应该直接丢弃。
func (c* Coordinator) createMapTask() {
	// 很简单，根据传入的filename，每个文件对应一个maptask
	// for循环就行了
	for idx, file_name := range c.InputFiles{
		// InputFiles里面是多个文件
		taskMeta := Task { // 创建一个Task
			Input: file_name,
			TaskState: Map,
			NReducer: c.NReduce,
			TaskNumber: idx,
		}
		c.TaskQueue <- &taskMeta // 把这个任务丢到队列里面去
		c.TaskMeta[idx] = &CoordinatorTask {
			TaskStatus : Idle, // 状态表示为未执行
			TaskReference: &taskMeta,
		}
	}
}

func (c* Coordinator) createReduceTask() {
	// 做的事情和之前的创建MapTask是一样的
	c.TaskMeta = make(map[int]* CoordinatorTask) // 不确定这一句是干嘛的，应该是要刷新一下TaskMeta
	for idx, files := range c.Intermediates {
		taskMeta := Task {
			TaskState: Reduce,
			NReducer: c.NReduce,
			TaskNumber: idx,
			Intermediates: files,
		}
		c.TaskQueue <- &taskMeta
		c.TaskMeta[idx] = &CoordinatorTask{
			TaskStatus: Idle,
			TaskReference: &taskMeta,
		}
	}
}

func (c* Coordinator) TaskCompleted(task* Task, reply *ExampleReply) error {
	// 更新task的状态
	mu.Lock()
	defer mu.Unlock()
	if task.TaskState != c.CoorPhase || c.TaskMeta[task.TaskNumber].TaskStatus == Completed{
		return nil
	}
	c.TaskMeta[task.TaskNumber].TaskStatus = Completed
	go c.processTaskResult(task) // 有一个任务完成了，领导人肯定要做相关记录啊
	return nil
}

func (c* Coordinator) processTaskResult(task* Task) {
	mu.Lock()
	defer mu.Unlock()
	switch task.TaskState {
	case Map:
		// 收集intermediate（这个完成了的Task）信息
		for reduceTaskId, filePath := range task.Intermediates {
			// 这里面就可以拿到这个Task的id和文件的路径了
			c.Intermediates[reduceTaskId] = append(c.Intermediates[reduceTaskId], filePath)
		}
		if c.allTaskDone() {
			// 如果把这个task处理好之后，发现已经Map完了
			// 进入reduce阶段
			c.createReduceTask()
			c.CoorPhase = Reduce // 设置阶段
		}
	case Reduce:
		// 如果收到的这个完成的task的状态是reduce，那就不用提取信息了
		// 因为已经是最后一步了，提取了没人要
		// 直接判断是不是最后一个task即可
		if c.allTaskDone() {
			c.CoorPhase = Exit
		}
	}
}

func (c* Coordinator) allTaskDone() bool {
	for _, task := range c.TaskMeta {
		if task.TaskStatus != Completed {
			// 如果有一个还没完成那就是还没完成
			return false
		}
	}
	return true
}

// 检查时间
func (c *Coordinator) catchTimeout() {
	for {
		time.Sleep(5 * time.Second)
		mu.Lock()
		if c.CoorPhase == Exit {
			mu.Unlock()
			return
		}
		for _, CoordinatorTask := range c.TaskMeta {
			if CoordinatorTask.TaskStatus == InProgress && time.Now().Sub(CoordinatorTask.StartTime) > 20 * time.Second {
				c.TaskQueue <- CoordinatorTask.TaskReference
				CoordinatorTask.TaskStatus = Idle
			}
		}
		mu.Unlock()
	}
}