package circuitbreaker

import (
	"math/rand"
	"testing"
	"time"
)

type Requset struct {
	Err      error
	Fallback bool
	RunFlag  bool
}

func (r *Requset) Run() error {
	r.RunFlag = true
	return r.Err
}

func (r *Requset) FallBack() error {
	r.Fallback = true
	return nil
}

func TestRandom(t *testing.T) {

	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	var req Requset

	var total, success, fail, timeout, reject uint32

	for total < 10000000 {
		req.Fallback = false
		total += 1
		random := rand.Int() % 10
		switch random {
		case 0:
			//success
			req.Err = nil
			success += 1
		case 1:
			//fail
			req.Err = FAIL
			fail += 1
		case 2:
			//timeout
			req.Err = TIMEOUT
			timeout += 1
		case 3:
			//reject
			req.Err = REJECT
			reject += 1
		default:
			req.Err = FAIL
			timeout += 1
		}

		cir.RunCommand(&req)
	}
	t.Log("gg")
}

func TestInit(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	if cir.lastrecord.Len() != 10 {
		t.Fatal("ring wrong")
	}

	if cir.interval != time.Second*5 {
		t.Fatal("interval")
	}
	if cir.requestThreshold != 5 {
		t.Fatal("totla")
	}

	for i := 0; i < 10; i++ {
		if cir.lastrecord.Value.(*record).seq != uint32(i) {
			t.Fatal("seq error")
		}
		time.Sleep(time.Millisecond * 1100)
	}
}

func TestTimer(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	for i := 0; i < 10; i++ {
		if cir.lastrecord.Value.(*record).seq != uint32(i) {
			t.Fatal("seq error")
		}
		time.Sleep(time.Millisecond * 1100)
	}

	cir.Destroy()
	time.Sleep(time.Second * 2)
	tmp := cir.lastrecord.Value.(*record).seq
	time.Sleep(time.Second * 3)
	tmp1 := cir.lastrecord.Value.(*record).seq
	if tmp != tmp1 {
		t.Fail()
	}

}
