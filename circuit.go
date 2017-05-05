package circuitbreaker

import (
	"container/ring"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type CircuitBreaker struct {
	//最近时间的请求记录,最新时间的指针.
	lastrecord *ring.Ring

	//熔断器状态
	status circuitstatus

	//单位时间内的总请求量
	totalRequstCount uint32

	lock4status sync.Mutex

	//熔断器从close恢复到halfopen状态的间隔
	interval time.Duration

	//单位时间内触发熔断器的阀值. 依据实际请求量设置
	requestThreshold uint32
	//circuit not used
	stop bool
}

//初始化熔断器,需要设置,熔断器记录样本的个数lasttime,开启检测熔断器的总的请求量,和熔断器触发后恢复的时间段,时间段必须大于1s.
func (circuit *CircuitBreaker) Init(lasttime int, requestThreshold uint32, interval time.Duration) {
	circuit.initCycle(lasttime)
	circuit.status = OPEN
	circuit.requestThreshold = requestThreshold
	circuit.interval = interval
	go circuit.timer()
}

func (circuit *CircuitBreaker) Destroy() {
	circuit.stop = true
}

//并发安全的函数.执行
func (circuit *CircuitBreaker) RunCommand(command Commander) {
	var err error
	status := circuit.getStatus()
	switch status {
	case OPEN:
		err = command.Run()
		circuit.openHandleCommandResult(err)
		if err != nil {
			command.FallBack()
		}
	case HAFLOPEN:
		err = command.Run()
		circuit.halfHandleCommandResult(err)
		if err != nil {
			command.FallBack()
		}
	case CLOSE:
		command.FallBack()
	}

}

func (circuit *CircuitBreaker) openHandleCommandResult(err error) {
	switch err {
	case nil:
		circuit.setSuccess()
	case FAIL:
		circuit.setFail()
	case TIMEOUT:
		circuit.setTimeout()
	case REJECT:
		circuit.setReject()
	default:
		panic("command interface run return unknown error")
	}
}
func (circuit *CircuitBreaker) halfHandleCommandResult(err error) {
	switch err {
	case nil:
		//进入open 状态
		circuit.lock4status.Lock()
		circuit.status = OPEN
		circuit.lock4status.Unlock()
		fmt.Println("enter into open from halfopen")
	default:
		//进入close状态
		circuit.lock4status.Lock()
		if circuit.status != CLOSE {

			go func() {
				select {
				case <-time.After(circuit.interval):
					circuit.lock4status.Lock()
					circuit.status = HAFLOPEN
					circuit.lock4status.Unlock()
					fmt.Println("enter into halfopen from halopen close")
				}
			}()
		}
		circuit.status = CLOSE
		circuit.lock4status.Unlock()
		fmt.Println("enter into close status from halfopen")
	}
}

func (circuit *CircuitBreaker) timer() {
	go func() {
		select {
		case <-time.After(time.Second):
			circuit.totalRequstCount = circuit.totalRequstCount - circuit.lastrecord.Next().Value.(*record).total
			circuit.lastrecord.Next().Value.(*record).reset()
			fmt.Println("next buket")
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&circuit.lastrecord)), unsafe.Pointer(circuit.lastrecord.Next()))
		}

		if circuit.stop {
			return
		} else {
			go circuit.timer()
		}
	}()
}

//circuit status, 判断条件,临时计算,不用考虑并发的问题.
func (circuit *CircuitBreaker) getStatus() circuitstatus {

	if circuit.status != OPEN {
		return circuit.status
	}

	var rd *record
	rd = circuit.lastrecord.Value.(*record)

	//timeout
	timeoutpercent := float64(rd.timeout) / float64(rd.total)
	failPercent := float64(rd.fail) / float64(rd.total)

	if (timeoutpercent > 0.5 || failPercent > 0.5) && circuit.totalRequstCount > circuit.requestThreshold {

		circuit.lock4status.Lock()
		if circuit.status != CLOSE {

			go func() {
				select {
				case <-time.After(circuit.interval):
					circuit.lock4status.Lock()
					circuit.status = HAFLOPEN
					circuit.lock4status.Unlock()
					fmt.Println("enter into halfopen status from close")
				}
			}()
		}
		circuit.status = CLOSE
		circuit.lock4status.Unlock()
		fmt.Println("enter into close from too many error")
	}
	return circuit.status

}

func (circuit *CircuitBreaker) setSuccess() {
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).total), 1)
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).success), 1)
	atomic.AddUint32(&(circuit.totalRequstCount), 1)
}

func (circuit *CircuitBreaker) setFail() {
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).total), 1)
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).fail), 1)
	atomic.AddUint32(&(circuit.totalRequstCount), 1)
}

func (circuit *CircuitBreaker) setTimeout() {
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).total), 1)
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).timeout), 1)
	atomic.AddUint32(&(circuit.totalRequstCount), 1)
}

func (circuit *CircuitBreaker) setReject() {
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).total), 1)
	atomic.AddUint32(&(circuit.lastrecord.Value.(*record).reject), 1)
	atomic.AddUint32(&(circuit.totalRequstCount), 1)
}

func (circuit *CircuitBreaker) initCycle(lasttime int) {
	circuit.lastrecord = ring.New(lasttime)

	for i := 0; i < lasttime; i++ {
		r := new(record)
		r.seq = uint32(i)
		r.reset()
		circuit.lastrecord.Value = r
		circuit.lastrecord = circuit.lastrecord.Next()
	}
}
