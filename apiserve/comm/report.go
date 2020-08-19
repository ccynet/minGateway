package comm

import (
	"encoding/json"
	"net/http"
)

func Report(w http.ResponseWriter, code int, msg string, body interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(map[string]interface{}{"code": code, "msg": msg, "body": body})
	return err
}
