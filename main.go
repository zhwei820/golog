package main

import (
	"golog/log"
	"flag"
	"time"
	"sync"
	"sync/atomic"
	"os"
	"os/signal"
	"runtime/pprof"
)

func main() {

	cpuprofile := "perf.prof"
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			//logger.LFatal("%s", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	logger1 := log.GetLogger("./logs/xyz")
	defer log.Uninit(logger1)

	logger := log.GetLogger("./logs/app")
	defer log.Uninit(logger)

	log.SetLevel(log.LvTRACE)

	flag.Parse()
	d := time.Duration(0) * time.Millisecond

	running := int32(1)

	var g sync.WaitGroup

	task := func() {
		for atomic.LoadInt32(&running) != 0 {
			logger.LTrace("hello %s", "Trace")
			logger.LDebug("hello %s", "Debug")
			logger.LInfo("hello %s", "Info")
			logger.LWarn("hello %s", "Warn")
			logger.LError("hello %s", "Error")

			logger1.LTrace("hello %s", "Trace")
			logger1.LDebug("hello %s", "Debug")
			logger1.LInfo("hello %s", "Info")
			logger1.LWarn("hello %s", "Warn")
			logger1.LError("hello %s", "Error")

			if d > 0 {
				time.Sleep(d)
			}
		}
		g.Done()
	}

	n := 10
	g.Add(n)
	for i := 0; i < n; i++ {
		go task()
	}

	listenSignal(func(sig os.Signal) bool {
		atomic.StoreInt32(&running, 0)
		return true
	})

	g.Wait()
	logger.LInfo("app exit")
}

func listenSignal(handler func(sig os.Signal) (ret bool), signals ...os.Signal) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	go func() {
		for !handler(<-sigChan) {
		}
	}()
}
