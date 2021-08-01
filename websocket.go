package main

/*
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	logger "nway/utils/log"
	"nway/utils/nway_string"
	"os"
	"strings"
	"unsafe"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

func GetUUId() string {

	u1 := uuid.Must(uuid.NewV4())

	return u1.String()
}

type NwayConn struct {
	Conn     *websocket.Conn
	Serverip string
	Caller   string
	Callee   string
}

var NWAY_WEBSOCKET_VERSION string = "1.1"

func main() {}

var Conns map[string]*NwayConn

//初始化库，创建连接池
//export nway_asr_init
func nway_asr_init() C.int {
	Conns = make(map[string]*NwayConn, 0)
	logger.SetConsole(false)
	os.MkdirAll("/opt/fsgui/log", 777)
	logger.SetRollingFile("/opt/fsgui/log", "nway_websocket.log", 300, 50, logger.MB)
	//logger.SetLevel(logger.ALL)
	//logger.RollingFile
	logger.Info("Starting websocket service\n")
	logger.Debug("Debug info\n")
	logger.Info("This version:", NWAY_WEBSOCKET_VERSION)
	logger.SetLevel(logger.ALL)
	return 0
}

//c_serverip 要连接的服务信息如 ws://10.0.0.120:2800
//sid 由库返回调用者的一个信息
//errmsg 调用过程中产生的错误日志,需调用方申请内存
//caller 主叫号码
//callee 被叫号码
//errmsg_len 返回的错误日志的最大长度
//返回值：正常返回0,不非常返回 -1
//export nway_asr_connect
func nway_asr_connect(c_serverip *C.char, sid *C.char, errmsg *C.char, caller *C.char, callee *C.char, errmsg_len C.int) C.int {
	var err error
	var nway_conn *NwayConn
	nway_conn = new(NwayConn)
	nway_conn.Callee = C.GoString(callee)
	nway_conn.Caller = C.GoString(caller)
	nway_conn.Serverip = C.GoString(c_serverip)
	nway_conn.Conn, _, err = websocket.DefaultDialer.Dial(nway_conn.Serverip, nil)
	if err != nil {
		fmt.Println("dial:", err)
		msg := err.Error()
		s1 := C.CString(msg)
		C.strcpy(errmsg, s1)
		nway_asr_free_var(s1)
		return -1
	}
	//随机生成一个sid
	uuid := GetUUId()
	logger.Debug("connect sid:", uuid)
	s2 := C.CString(uuid)
	//C.strncpy(sid, s2, C.size_t(errmsg_len))
	C.strcpy(sid, s2)
	nway_asr_free_var(s2)
	logger.Debug("connect sid:", sid)
	Conns[uuid] = nway_conn
	return 0
}

//释放通过这里申请的内存空间
//v 为char*
//export nway_asr_free_var
func nway_asr_free_var(v *C.char) C.int {
	C.free(unsafe.Pointer(v))
	return 0
}

//向websocket发送链接
//sid由模块返回调用者的id
//c_data 由调用方按c char*发送的数据包
//c_len 由调用方弄的c_data的真实长度
//c_result 中间过程中产生的识别结果,或者发送时产生的错误日志，需预申请内存
//result_len 返回的c_result的最长长度，避免strcpy异常
//返回值：正常返回0,不非常返回 -1
//export nway_asr_sendmessage
func nway_asr_sendmessage(sid *C.char, c_data *C.void, c_len C.int, c_result *C.char, result_len C.int) C.int {
	//go_data := C.GoString(c_data)
	uuid := C.GoString(sid)
	nway_conn := Conns[uuid]
	if nway_conn != nil {
		logger.Debug("send sid:", uuid, " len:", c_len)
		//go_data := make([]byte, c_len)
		//copy(go_data, (*(*[1024]byte)(unsafe.Pointer(c_data)))[:c_len:c_len])
		go_data := C.GoBytes(unsafe.Pointer(c_data), c_len)
		logger.Debug(go_data)
		err := nway_conn.Conn.WriteMessage(websocket.BinaryMessage, go_data)
		if err != nil {
			fmt.Println("send data :", err)
			logger.Debug("connect sid:", uuid)

			msg := err.Error()
			s := C.CString(msg)
			//C.strncpy(c_result, s, C.size_t(result_len))
			C.strcpy(c_result, s)
			nway_asr_free_var(s)

			return -1
		}
		return 0
	}
	logger.Debug("can not found connect sid:", uuid)
	return -2
}

//需要发送最后的包，如果对端有vad，则只需要获取结果后关闭链接，否则按需要发结束符
//sid由模块返回调用者的id
//c_result 返回结果,需要预先在外部申请内存，尽量大些
//result_len 返回的c_result的最长长度，避免strcpy异常
//返回值：正常返回0,不非常返回 -1
//export nway_asr_stop
func nway_asr_stop(sid *C.char, c_result *C.char, result_len C.int) C.int {
	uuid := C.GoString(sid)
	nway_conn := Conns[uuid]
	if nway_conn != nil {
		nway_conn.Conn.WriteMessage(websocket.BinaryMessage, nway_string.StringByte("{\"eof\" : 1}"))
		_, message, err := nway_conn.Conn.ReadMessage()
		nway_conn.Conn.Close()
		if err != nil {
			logger.Error("read:", err)
			return -1
		}
		logger.Debug("recv message:", string(message))
		txt := gjson.Get(string(message), "text")
		txt2 := strings.ReplaceAll(txt.String(), " ", "")
		// r_len := int(result_len)
		// if len(txt2) > r_len {
		// 	txt2 = txt2[1 : r_len-1]
		// }
		s := C.CString(string(txt2))

		C.strcpy(c_result, s)

		//C.strncpy(c_result, s, C.size_t(result_len))
		nway_asr_free_var(s)

		return 0
	}

	logger.Debug("can not found connect sid:", uuid)
	return -2
}

//释放整个库，同时关闭所有的链接
//export nway_asr_release
func nway_asr_release() C.int {
	for k, v := range Conns {
		v.Conn.Close()
		logger.Debug("close sid:", k)
	}
	return 0
}

//export get_uuid
func get_uuid(id *C.char, id_len C.int) C.int {
	uuid := GetUUId()
	//logger.Debug("connect sid:", uuid)
	fmt.Println("uuid:", uuid)
	//id = (*C.char)(unsafe.Pointer(C.CString(uuid)))
	s := C.CString(uuid)
	C.strncpy(id, s, C.size_t(id_len))
	nway_asr_free_var(s)
	fmt.Println("id:", id)
	return 0
}

// //go 代码,是生成为.so动态库由c语言调用
// //export my_connect
// func my_connect(c_serverip *C.char, sid *C.char, errmsg *C.char) C.int {
// 	var err error
// 	var nway_conn *NwayConn
// 	nway_conn = new(NwayConn)

// 	nway_conn.Serverip = C.GoString(c_serverip)
// 	nway_conn.Conn, _, err = websocket.DefaultDialer.Dial(nway_conn.Serverip, nil)
// 	if err != nil {
// 		fmt.Println("dial:", err)
// 		msg := err.Error()
// 		//连接错误输出给errmsg这个输入的char*变量
// 		errmsg = (*C.char)(unsafe.Pointer(C.CString(msg)))
// 		return -1
// 	}
// 	//随机生成一个sid
// 	uuid := GetUUId()
// 	//通过websocket连接成功后，返回给c语言调用函数，char* sid;一个uuid
// 	sid = (*C.char)(unsafe.Pointer(C.CString(uuid)))
// 	fmt.Println(sid)
// 	Conns[uuid] = nway_conn
// 	return 0
// }

//c代码
// /*
//    char* errmsg=NULL;
//    char* sid=NULL;
//    int recStatus = my_connect("ws://10.1.1.12:1234",sid,errmsg);
//    if (recStatus==-1) printf("errmsg:%s\n",errmsg);
//    else if (recStatus == 0) printf("sid:%s\n",sid);
// /*
//执行了以上调用后，在go代码中输出的sid是正常的，但是通过c语言调用go生成的动态链接库后，输出不管errmsg还是sid都是null,意味着使用C.CString分配的内存没有返回到调用函数处
