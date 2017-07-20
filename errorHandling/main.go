package errorHandling

import (
	"fmt"
	"net/http"
)

// RespondWithErrorStack : Responds with err.Error()
func RespondWithErrorStack(w http.ResponseWriter, err error) {
	RespondWithError(w, err.Error())
}

// RespondOnlyXAccepted : Responds with only GET, POST ... accepted
func RespondOnlyXAccepted(w http.ResponseWriter, x string) {
	RespondWithError(w, "only "+x+" accepted")
}

// RespondWithError : Responds with an error given as a parameter
func RespondWithError(w http.ResponseWriter, reason string) {
	fmt.Println("Responding with bad request because:", reason)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error : ", reason)
}
