package response

import (
	"encoding/json"
	"net/http"
)

type Meta struct {
	Page      int `json:"page,omitempty"`
	PerPage   int `json:"per_page,omitempty"`
	Total     int `json:"total,omitempty"`
	TotalPage int `json:"total_page,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, map[string]any{"success": true, "data": data})
}

func Paginated(w http.ResponseWriter, status int, data any, meta Meta) {
	JSON(w, status, map[string]any{"success": true, "data": data, "meta": meta})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{"success": false, "error": message})
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}
