package metrics

import (
	"io/ioutil"
	"log"
	"sync"
	"testing"
	"time"
)

const FANOUT = 128

// Stop the compiler from complaining during debugging.
var (
	_ = ioutil.Discard
	_ = log.LstdFlags
)

func BenchmarkMetrics(b *testing.B) {
	r := NewRegistry()
	c := NewRegisteredCounter("counter", r)
	h := NewRegisteredHistogram("histogram", r, NewUniformSample(100))
	b.ResetTimer()
	ch := make(chan bool)

	wgD := &sync.WaitGroup{}
	/*
		wgD.Add(1)
		go func() {
			defer wgD.Done()
			//log.Println("go CaptureDebugGCStats")
			for {
				select {
				case <-ch:
					//log.Println("done CaptureDebugGCStats")
					return
				default:
					CaptureDebugGCStatsOnce(r)
				}
			}
		}()
	//*/

	//*/

	wgW := &sync.WaitGroup{}
	/*
		wgW.Add(1)
		go func() {
			defer wgW.Done()
			//log.Println("go Write")
			for {
				select {
				case <-ch:
					//log.Println("done Write")
					return
				default:
					WriteOnce(r, ioutil.Discard)
				}
			}
		}()
	//*/

	wg := &sync.WaitGroup{}
	wg.Add(FANOUT)
	for i := 0; i < FANOUT; i++ {
		go func(i int) {
			defer wg.Done()
			//log.Println("go", i)
			for i := 0; i < b.N; i++ {
				c.Inc(1)
				h.Update(time.Now(), int64(i))
			}
			//log.Println("done", i)
		}(i)
	}
	wg.Wait()
	close(ch)
	wgD.Wait()
	wgW.Wait()
}
