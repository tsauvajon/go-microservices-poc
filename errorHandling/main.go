package errorHandling

import (
	"fmt"
	"net/http"
)

func RespondWithErrorStack(w http.ResponseWriter, err error) {
	RespondWithError(w, err.Error())
}

func RespondOnlyXAccepted(w http.ResponseWriter, x string) {
	RespondWithError(w, "only "+x+" accepted")
}

func RespondWithError(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error : ", reason)
}
