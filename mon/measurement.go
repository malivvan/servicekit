package mon

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type SampleOption int

const (
	MAX SampleOption = iota
	MIN
	AVG
	SUM
	COUNT
	LAST
)

type Measurement struct {
	name string
	tags map[string]string

	fields_ sync.Mutex
	fields  map[string]measurementField

	running   atomic.Value
	waitGroup sync.WaitGroup
}

type measurementField interface {
	sample(interface{})
	get(string) map[string]interface{}
}

func New(name string, tags map[string]string) *Measurement {

	if tags == nil {
		tags = make(map[string]string)
	}
	measurement := &Measurement{
		name:   name,
		tags:   tags,
		fields: make(map[string]measurementField),
	}
	measurement.running.Store(false)

	// register measurement for graceful shutdown
	measurementsRegister_.Lock()
	measurementsRegister = append(measurementsRegister, measurement)
	measurementsRegister_.Unlock()

	return measurement
}

func (m *Measurement) Int(name string, options ...SampleOption) *Measurement {
	if len(options) == 0 {
		options = append(options, LAST)
	}
	m.fields[name] = &intField{
		samples: []int64{},
		options: options,
	}
	return m
}

func (m *Measurement) Float(name string, options ...SampleOption) *Measurement {
	if len(options) == 0 {
		options = append(options, LAST)
	}
	m.fields[name] = &floatField{
		samples: []float64{},
		options: options,
	}
	return m
}

func (m *Measurement) Sample(data map[string]interface{}) {
	m.fields_.Lock()
	defer m.fields_.Unlock()

	for name, value := range data {
		if field, ok := m.fields[name]; ok {
			field.sample(value)
		}
	}
}

func (m *Measurement) Commit(t time.Time) {
	m.fields_.Lock()
	defer m.fields_.Unlock()

	values := make(map[string]interface{})
	for fieldName, field := range m.fields {
		for k, v := range field.get(fieldName) {
			values[k] = v
		}
	}

	Write(m.name, t, m.tags, values)
}

func (m *Measurement) Start(commitMs int, sampleMs int, sampleFunc func(map[string]interface{})) error {

	// stop old routine if running
	m.Stop()

	// start appropriate routine
	if sampleFunc != nil {
		return m.sampleAndCommitRoutine(commitMs, sampleMs, sampleFunc)
	}
	return m.commitRoutine(commitMs)
}

func (m *Measurement) Stop() {
	if m.running.Load().(bool) {
		m.running.Store(false)
		m.waitGroup.Wait()
	}
}

func (m *Measurement) commitRoutine(commitMs int) error {
	m.waitGroup.Add(1)
	m.running.Store(true)
	go func() {
		defer m.waitGroup.Done()

		completeSamples := false
		commitInterval := time.Duration(commitMs) * time.Millisecond
		nextCommit := time.Now().Truncate(commitInterval).Add(commitInterval)

		for m.running.Load().(bool) {

			// wait for sample run, schedule next sample run
			timeUntilNextCommit := nextCommit.Sub(time.Now())
			for timeUntilNextCommit > 0 {

				// early out if this routine is being stopped
				if !m.running.Load().(bool) {
					return
				}

				// sleep interval is time until next sample, or one second
				sleepDuration := timeUntilNextCommit
				if sleepDuration > time.Second {
					sleepDuration = time.Second
				}

				time.Sleep(sleepDuration)
				timeUntilNextCommit -= sleepDuration
			}
			nextCommit = time.Now().Truncate(commitInterval).Add(commitInterval)

			// commit samples if the set is complete (first one is always discarded)
			if completeSamples {
				m.Commit(time.Now().Truncate(commitInterval))
			} else {

				// this will reset all fields
				for fieldName := range m.fields {
					m.fields[fieldName].get(fieldName)
				}

				// the next interval will yield complete samples
				completeSamples = true
			}
		}
	}()

	return nil
}

func (m *Measurement) sampleAndCommitRoutine(commitMs int, sampleMs int, sampleFunc func(map[string]interface{})) error {

	// multiples check
	multCheck := commitMs
	for multCheck > 0 {
		multCheck -= sampleMs
	}
	if multCheck != 0 {
		return errors.New("commit interval must be multiple of sample interval")
	}

	m.waitGroup.Add(1)
	m.running.Store(true)
	go func() {
		defer m.waitGroup.Done()

		completeSamples := false
		sampleInterval := time.Duration(sampleMs) * time.Millisecond
		commitInterval := time.Duration(commitMs) * time.Millisecond
		nextSample := time.Now().Truncate(sampleInterval).Add(sampleInterval)
		nextCommit := time.Now().Truncate(commitInterval).Add(commitInterval)

		for m.running.Load().(bool) {

			// wait for sample run, schedule next sample run
			timeUntilNextSample := nextSample.Sub(time.Now())
			for timeUntilNextSample > 0 {

				// early out if this routine is being stopped
				if !m.running.Load().(bool) {
					return
				}

				// sleep interval is time until next sample, or one second
				sleepDuration := timeUntilNextSample
				if sleepDuration > time.Second {
					sleepDuration = time.Second
				}

				time.Sleep(sleepDuration)
				timeUntilNextSample -= sleepDuration
			}
			nextSample = time.Now().Truncate(sampleInterval).Add(sampleInterval)

			// execute sample function and commit samples
			values := make(map[string]interface{})
			sampleFunc(values)
			m.Sample(values)

			// if next commit timestamp is reached perform commit, schedule next commit
			if !time.Now().Before(nextCommit) {

				if completeSamples {
					m.Commit(time.Now().Truncate(commitInterval))
					nextCommit = time.Now().Truncate(commitInterval).Add(commitInterval)

				} else {

					// this will reset all fields
					for fieldName := range m.fields {
						m.fields[fieldName].get(fieldName)
					}

					// the next interval will yield complete samples
					completeSamples = true
				}

				nextCommit = time.Now().Truncate(commitInterval).Add(commitInterval)
			}
		}
	}()

	return nil
}
