package utils

import "encoding/json"

func ConvertToStruct[T any](m any) T {
	var s T
	jsonStr, _ := json.Marshal(m)
	json.Unmarshal(jsonStr, &s)
	return s
}
