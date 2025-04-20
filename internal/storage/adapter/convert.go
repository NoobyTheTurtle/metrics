package adapter

import "encoding/json"

func convertToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return 0, false
		}
		var f float64
		if err := json.Unmarshal(jsonData, &f); err != nil {
			return 0, false
		}
		return f, true
	}
}

func convertToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return 0, false
		}
		var i int64
		if err := json.Unmarshal(jsonData, &i); err != nil {
			return 0, false
		}
		return i, true
	}
}
