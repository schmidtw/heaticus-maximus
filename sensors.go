package main

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/schmidtw/go1wire/adapters/ds2480"
	"github.com/schmidtw/go1wire/devices/ds18x20"
)

type TempSensors interface {
	// Shuts down and turns everything off.
	Shutdown()
}

type TempSensorsOpts struct {
	// The Namespace of the metrics for the thing
	Namespace string

	// Path to the adapter device
	Path string

	SamplePeriod time.Duration

	Names map[string]string
}

type tempSensors struct {
	adapter  *ds2480.Ds2480
	ticker   *time.Ticker
	devices  map[string]*ds18x20.Ds18x20
	readings map[string]float64
	wg       sync.WaitGroup
	mutex    sync.Mutex
	done     chan bool

	// Metrics
	metrics map[string]prometheus.Gauge
}

func NewTempSensors(opts TempSensorsOpts) (TempSensors, error) {
	adapter := &ds2480.Ds2480{
		Name:  opts.Path,
		Speed: "standard",
		PDSRC: 1370,
		//PPD:   time.Microsecond * 512,
		//SPUD:  time.Microsecond * 1,
		W1LT: time.Microsecond * 10,
		W0RT: time.Microsecond * 8,
		LOAD: 0,
		Baud: 9600,
		SPU:  false,
		IRP:  false,
	}

	err := adapter.Init()
	if nil != err {
		return nil, err
	}
	adapter.Open()
	adapter.Detect()

	ts := &tempSensors{
		adapter:  adapter,
		devices:  make(map[string]*ds18x20.Ds18x20),
		readings: make(map[string]float64),
		metrics:  make(map[string]prometheus.Gauge),
	}

	list, err := adapter.Search()
	if nil != err {
		return nil, err
	}

	for _, v := range list {
		if name, ok := opts.Names[v.String()]; ok {
			sensor, _ := ds18x20.New(adapter, v)
			ts.devices[name] = sensor
			ts.readings[name] = 0.0
			ts.metrics[name] = promauto.NewGauge(prometheus.GaugeOpts{
				Namespace: opts.Namespace,
				Subsystem: "physical",
				Name:      name + "_temp",
				Help:      name + " temperature (F)",
			})

		}
	}

	ts.wg.Add(1)
	go ts.run()

	ts.ticker = time.NewTicker(opts.SamplePeriod)

	return ts, nil
}

func (ts *tempSensors) Shutdown() {
	ts.done <- true
	ts.wg.Wait()
}

func (ts *tempSensors) Get(name string) float64 {
	ts.mutex.Lock()
	temp, ok := ts.readings[name]
	ts.mutex.Unlock()

	if false == ok {
		temp = -1000
	}
	return temp
}

func (ts *tempSensors) run() {
	defer ts.wg.Done()
	for {
		select {
		case <-ts.done:
			ts.adapter.Close()
			return
		case <-ts.ticker.C:
			ds18x20.ConvertAll(ts.adapter)
			for k, v := range ts.devices {
				temp, err := v.LastTemp()
				if nil == err {
					// Convert to F
					temp = temp*9/5 + 32.0

					ts.mutex.Lock()
					ts.readings[k] = temp
					ts.mutex.Unlock()
					ts.metrics[k].Set(temp)
				}
			}

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}
