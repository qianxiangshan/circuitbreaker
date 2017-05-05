package circuitbreaker

import (
	"testing"
	"time"
)

func TestRunCommandopen(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	var req Requset

	req.Err = nil
	cir.RunCommand(&req)
	if req.RunFlag == false {
		t.Fail()
	}

	req.Err = FAIL
	cir.RunCommand(&req)
	if req.Fallback == false {
		t.Fail()
	}
	req.Fallback = false
	req.Err = FAIL
	cir.RunCommand(&req)
	if req.Fallback == false {
		t.Fail()
	}

	req.Fallback = false
	req.Err = FAIL
	cir.RunCommand(&req)
	if req.Fallback == false {
		t.Fail()
	}
	t.Log("gg")
}

func TestRunCommandHalfOpen(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	cir.status = HAFLOPEN

	var req Requset

	//halopen to open
	req.Fallback = false
	req.Err = nil
	cir.RunCommand(&req)
	if req.RunFlag == false || cir.status != OPEN {
		t.Fail()
	}

	//halfopen to close
	req.Fallback = false
	cir.status = HAFLOPEN
	req.Err = FAIL
	cir.RunCommand(&req)
	if req.Fallback == false || cir.status != CLOSE {
		t.Fail()
	}
	time.Sleep(time.Second * 6)
	if cir.status != HAFLOPEN {
		t.Fail()
	}

	t.Log("TestRunCommandHalfOpen")

}

func TestRunCommandClose(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	cir.status = CLOSE

	var req Requset

	//close
	req.Fallback = false
	req.Err = nil
	cir.RunCommand(&req)
	if req.RunFlag == true || req.Fallback == false {
		t.Fail()
	}

	t.Log("TestRunCommandClose")
}

func TestOpenHandleCommandResult(t *testing.T) {
	var cir CircuitBreaker
	cir.Init(10, 5, time.Second*5)

	cir.status = CLOSE

	var req Requset

	//close
	req.Fallback = false
	req.Err = TIMEOUT
	cir.RunCommand(&req)
	if req.RunFlag == true || req.Fallback == false {
		t.Fail()
	}

	//close
	req.Fallback = false
	req.Err = REJECT
	cir.RunCommand(&req)
	if req.RunFlag == true || req.Fallback == false {
		t.Fail()
	}

	t.Log("openHandleCommandResult")

}
