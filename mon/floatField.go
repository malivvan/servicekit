package mon

import (
	"math"
)

type floatField struct {
	samples []float64
	options []SampleOption
}

func (is *floatField) sample(value interface{}) {
	switch v := value.(type) {
	case float64:
		is.samples = append(is.samples, v)
	case float32:
		is.samples = append(is.samples, float64(v))
	}
}

func (fs *floatField) get(name string) map[string]interface{} {
	values := make(map[string]interface{})
	if len(fs.samples) == 0 {
		return values
	}

	// create aggregates
	for _, t := range fs.options {
		switch t {
		case MAX:
			maxValue := float64(-math.MaxFloat64)
			for _, sample := range fs.samples {
				if sample > maxValue {
					maxValue = sample
				}
			}
			values[name+"_max"] = maxValue
		case MIN:
			minValue := float64(math.MaxFloat64)
			for _, sample := range fs.samples {
				if sample < minValue {
					minValue = sample
				}
			}
			values[name+"_min"] = minValue
		case AVG:
			var valueSum float64
			for _, sample := range fs.samples {
				valueSum += sample
			}
			values[name+"_avg"] = valueSum / float64(len(fs.samples))
		case SUM:
			var valueSum float64
			for _, sample := range fs.samples {
				valueSum += sample
			}
			values[name+"_sum"] = valueSum
		case COUNT:
			values[name+"_cnt"] = len(fs.samples)
		case LAST:
			values[name] = fs.samples[len(fs.samples)-1]
		}

	}

	// reset samples and return values
	fs.samples = []float64{}
	return values
}
