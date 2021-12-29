package mon

import (
	"math"
)

type intField struct {
	samples []int64
	options []SampleOption
}

func (is *intField) sample(value interface{}) {
	switch v := value.(type) {
	case int:
		is.samples = append(is.samples, int64(v))
	case int64:
		is.samples = append(is.samples, v)
	case int32:
		is.samples = append(is.samples, int64(v))
	case int16:
		is.samples = append(is.samples, int64(v))
	case int8:
		is.samples = append(is.samples, int64(v))
	case uint:
		is.samples = append(is.samples, int64(v))
	case uint64:
		is.samples = append(is.samples, int64(v))
	case uint32:
		is.samples = append(is.samples, int64(v))
	case uint16:
		is.samples = append(is.samples, int64(v))
	case uint8:
		is.samples = append(is.samples, int64(v))
	}
}

func (is *intField) get(name string) map[string]interface{} {
	values := make(map[string]interface{})
	if len(is.samples) == 0 {
		return values
	}

	// create aggregates
	for _, t := range is.options {
		switch t {
		case MAX:
			maxValue := int64(-math.MaxInt64)
			for _, sample := range is.samples {
				if sample > maxValue {
					maxValue = sample
				}
			}
			values[name+"_max"] = maxValue
		case MIN:
			minValue := int64(math.MaxInt64)
			for _, sample := range is.samples {
				if sample < minValue {
					minValue = sample
				}
			}
			values[name+"_min"] = minValue
		case AVG:
			var valueSum int64
			for _, sample := range is.samples {
				valueSum += sample
			}
			values[name+"_avg"] = valueSum / int64(len(is.samples))
		case SUM:
			var valueSum int64
			for _, sample := range is.samples {
				valueSum += sample
			}
			values[name+"_sum"] = valueSum
		case COUNT:
			values[name+"_cnt"] = len(is.samples)
		case LAST:
			values[name] = is.samples[len(is.samples)-1]
		}
	}

	// reset samples and return values
	is.samples = []int64{}
	return values
}
