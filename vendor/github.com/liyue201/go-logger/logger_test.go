package logger

import (
	"strconv"
	"testing"
	"sync"
)

var wg sync.WaitGroup

func log(i int) {
	Debug("Debug>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Debugf("Debugf>>>>>>>>>>>>>>>>>>>>>> %d", i)

	Info("Info>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Infof("Infof>>>>>>>>>>>>>>>>>>>>>> %d", i)

	Warn("Warn>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Warnf("Warnf>>>>>>>>>>>>>>>>>>>>>> %d", i)

	Error("Error>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Errorf("Errorf>>>>>>>>>>>>>>>>>>>>>> %d", i)

	//	Fatal("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	//	Fatalf("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>%d", i)

	wg.Done()
}

func TestMain(t *testing.M) {
	Init("/Users/kexin/go/src/github.com/codinl/go-logger", "filename", DEBUG, true, true, false)

	for i := 100; i > 0; i-- {
		wg.Add(1)
		log(i)
	}

	wg.Wait()
}
