package domain

import "errors"

// ErrInvalidStr Define a static error.
var ErrInvalidStr = errors.New("invalid string type")

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

func (e *MetricType) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ErrInvalidStr
	}
	*e = MetricType(str)

	return nil
}

func (e MetricType) Value() (interface{}, error) {
	return string(e), nil
}

type MetricSource string

const (
	Runtime MetricSource = "runtime"
	Custom  MetricSource = "custom"
	GopsMem MetricSource = "gopsutil/mem"
	GopsCPU MetricSource = "gopsutil/cpu"
)

type MetricAlias string

const (
	MTotal  MetricAlias = "Total"
	MFree   MetricAlias = "Free"
	MSystem MetricAlias = "System"
)

type MetricName string

// Implement the Stringer interface.
func (m MetricName) String() string {
	return string(m)
}

const (
	Alloc         MetricName = "Alloc"
	BuckHashSys   MetricName = "BuckHashSys"
	Frees         MetricName = "Frees"
	GCCPUFraction MetricName = "GCCPUFraction"
	GCSys         MetricName = "GCSys"
	HeapAlloc     MetricName = "HeapAlloc"
	HeapIdle      MetricName = "HeapIdle"
	HeapInuse     MetricName = "HeapInuse"
	HeapObjects   MetricName = "HeapObjects"
	HeapReleased  MetricName = "HeapReleased"
	HeapSys       MetricName = "HeapSys"
	LastGC        MetricName = "LastGC"
	Lookups       MetricName = "Lookups"
	MCacheInuse   MetricName = "MCacheInuse"
	MCacheSys     MetricName = "MCacheSys"
	MSpanInuse    MetricName = "MSpanInuse"
	MSpanSys      MetricName = "MSpanSys"
	Mallocs       MetricName = "Mallocs"
	NextGC        MetricName = "NextGC"
	NumForcedGC   MetricName = "NumForcedGC"
	NumGC         MetricName = "NumGC"
	OtherSys      MetricName = "OtherSys"
	PauseTotalNs  MetricName = "PauseTotalNs"
	StackInuse    MetricName = "StackInuse"
	StackSys      MetricName = "StackSys"
	Sys           MetricName = "Sys"
	TotalAlloc    MetricName = "TotalAlloc"
	RandomValue   MetricName = "RandomValue"
	PollCount     MetricName = "PollCount"
)
