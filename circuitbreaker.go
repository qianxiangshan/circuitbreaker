package circuitbreaker

import (
	"errors"
	"sync/atomic"
)

//command exec result
var (
	FAIL    = errors.New("command exec fail")
	TIMEOUT = errors.New("command exec timeout")
	REJECT  = errors.New("commnad be reject")
)

type circuitstatus int

//CIRCUIT BREADER status
const (
	OPEN circuitstatus = iota
	HAFLOPEN
	CLOSE
)

//命令实现接口,error返回必须是上述定义的error变量.返回未知的error变量将会导致panic
type Commander interface {
	//正常流程执行的命令
	Run() error
	//run失败后,运行fallback,fallback必须不能阻塞,提供正确的回复.
	FallBack() error
}

// 熔断判断条件可设定, 熔断恢复策略可设定,

//记录每秒的的请求的结果情况
type record struct {
	//请求总和
	total uint32
	//请求成功个数
	success uint32
	//请求失败个数
	fail uint32
	//请求超时个数
	timeout uint32
	//请求被拒绝个数
	reject uint32
	//seq
	seq uint32
}

func (r *record) reset() {
	atomic.StoreUint32(&r.total, 0)
	atomic.StoreUint32(&r.success, 0)
	atomic.StoreUint32(&r.fail, 0)
	atomic.StoreUint32(&r.timeout, 0)
	atomic.StoreUint32(&r.reject, 0)
}

//cycle
