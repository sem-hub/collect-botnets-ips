package utils

import "encoding/json"

type JsonResponse struct {
	Status string   `json:"status"`
	ErrMsg string   `json:"error"`
	Data   []string `json:"data"`
}

func SendAsJson(resp JsonResponse) string {
	out, _ := json.Marshal(resp)
	return string(out)
}
