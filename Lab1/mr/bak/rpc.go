package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import "os"
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
// RPC的意思其实就和protobuf，json一样，就是打包
// 我写网络服务器的时候也要定义Request和Respones啊，序列化和反序列化嘛
type TaskRequest struct{
	X int
}
type TaskRespone struct{
	FileName string // MapTask
	NumMapworkers int
	NumReduceWorkers int
}


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
