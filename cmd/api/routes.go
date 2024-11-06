// Filename: cmd/api/routes.go
package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(a.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	//Product part
	router.HandlerFunc(http.MethodGet, "/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/product", a.listProductHandler)
	router.HandlerFunc(http.MethodPost, "/product", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/product/:pid", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/product/:pid", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/product/:pid", a.deleteProductHandler)

	// //Review part
	router.HandlerFunc(http.MethodGet, "/review", a.listReviewHandler)
	router.HandlerFunc(http.MethodPost, "/review", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/review/:rid", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/review/:rid", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/review/:rid", a.deleteReviewHandler)

	router.HandlerFunc(http.MethodGet, "/product-review/:rid", a.listProductReviewHandler)
	router.HandlerFunc(http.MethodGet, "/product/:pid/review/:rid", a.getProductReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/helpful-count/:rid", a.HelpfulCountHandler)

	return a.recoverPanic(router)

}
