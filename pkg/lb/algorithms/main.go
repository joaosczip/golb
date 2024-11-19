package lb

import (
	"net/http"
)

type Algorithm interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}
