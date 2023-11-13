package mr

import "log"
import "net"
import "os"
import "net/rpc"
import "net/http"


// 一个Task应该是一个filename
// 然后Coordinator里面应该有一系列等待被完成的Task
// Coordinator处理的Task有两种，一种是MapTask，一种是ReduceTask
type Task struct {
	FileName string
	Status int // 0 start 1 running 2 finish
}


type Coordinator struct {
	// Your definitions here.
	Status int // 0表示启动状态 1表示Map 2表示Reduce
	NumMapWorkers int
	NumReduceWorkers int
	MapTask chan Task
	ReduceTask chan Task
	NumMapTasks int // task的个数
	NumReduceTasks int // task的个数
	MapTaskFinished chan bool
	ReduceTaskFinished chan bool
}

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

func (c *Coordinator) GetTask(args *TaskRequest, reply *TaskRespone) error {
	if c.Status == 0 {
		reply.FileName = 
	} else if c.Status == 1 {
		// 所有的Mapfinished
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
	ret := false

	// Your code here.


	return ret
}

//
// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{Status: 0, 
		NumMapWorkers:3 ,
		NumReduceWorkers: nReduce,
		MapTask: make(chan Task),
		ReduceTask: make(chan Task)}
	// Your code here.

	c.server()
	return &c
}
