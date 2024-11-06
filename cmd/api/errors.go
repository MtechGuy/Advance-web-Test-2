// Filename: cmd/api/errors.go
package main

import (
	"fmt"
	"net/http"
)

func (a *applicationDependencies) logError(r *http.Request, err error) {

	method := r.Method
	uri := r.URL.RequestURI()
	a.logger.Error(err.Error(), "method", method, "uri", uri)

}

func (a *applicationDependencies) errorResponseJSON(w http.ResponseWriter,
	r *http.Request,
	status int,
	message any) {

	errorData := envelope{"error": message}
	err := a.writeJSON(w, status, errorData, nil)
	if err != nil {
		a.logError(r, err)
		w.WriteHeader(500)
	}
}

func (a *applicationDependencies) serverErrorResponse(w http.ResponseWriter,
	r *http.Request,
	err error) {

	a.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	a.errorResponseJSON(w, r, http.StatusInternalServerError, message)
}

func (a *applicationDependencies) notFoundResponse(w http.ResponseWriter,
	r *http.Request) {

	message := "the requested resource could not be found"
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

func (a *applicationDependencies) PRIDnotFound(w http.ResponseWriter, r *http.Request, id int64) {
	message := fmt.Sprintf("Product with id = %d was not found", id)
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

func (a *applicationDependencies) RRIDnotFound(w http.ResponseWriter, r *http.Request, id int64) {
	message := fmt.Sprintf("Review with id = %d was not found", id)
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

func (a *applicationDependencies) PIDnotFound(w http.ResponseWriter, r *http.Request, id int64) {
	message := fmt.Sprintf("Product with id = %d was already deleted", id)
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}
func (a *applicationDependencies) RIDnotFound(w http.ResponseWriter, r *http.Request, id int64) {
	message := fmt.Sprintf("Review with id = %d was already deleted", id)
	a.errorResponseJSON(w, r, http.StatusNotFound, message)
}

func (a *applicationDependencies) methodNotAllowedResponse(
	w http.ResponseWriter,
	r *http.Request) {

	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	a.errorResponseJSON(w, r, http.StatusMethodNotAllowed, message)
}

func (a *applicationDependencies) badRequestResponse(w http.ResponseWriter,
	r *http.Request, err error) {

	a.errorResponseJSON(w, r, http.StatusBadRequest, err.Error())
}

func (a *applicationDependencies) failedValidationResponse(w http.ResponseWriter, r *http.Request,
	errors map[string]string) {
	a.errorResponseJSON(w, r, http.StatusUnprocessableEntity, errors)
}
