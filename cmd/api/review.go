package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	// import the data package which contains the definition for Comment
	"github.com/mtechguy/test2/internal/data"
	"github.com/mtechguy/test2/internal/validator"
)

// Struct to hold incoming review data
var incomingReviewData struct {
	ProductID    *int64  `json:"product_id"` // foreign key referencing products
	Author       *string `json:"author"`
	Rating       *int64  `json:"rating"` // integer with a constraint (1-5)
	HelpfulCount *int32  `json:"helpful_count"`
	ReviewText   *string `json:"review_text"` // non-null text field

}

// Updated createReviewHandler with product existence check
func (a *applicationDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Create a local instance of incomingReviewData
	var incomingReviewData struct {
		ProductID    *int64  `json:"product_id"` // foreign key referencing products
		Author       *string `json:"author"`
		Rating       *int64  `json:"rating"` // integer with a constraint (1-5)
		HelpfulCount *int32  `json:"helpful_count"`
		ReviewText   *string `json:"review_text"` // non-null text field
	}

	// Decode the incoming JSON into the struct
	err := a.readJSON(w, r, &incomingReviewData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Check if product_id is provided
	if incomingReviewData.ProductID == nil {
		a.badRequestResponse(w, r, errors.New("product_id is required"))
		return
	}

	// Check if the product exists in the database
	exists, err := a.productModel.ProductExists(*incomingReviewData.ProductID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if !exists {
		a.PRIDnotFound(w, r, *incomingReviewData.ProductID) // Respond with a 404 if product is not found
		return
	}

	if incomingReviewData.HelpfulCount == nil {
		incomingReviewData.HelpfulCount = new(int32) // Default to 0 if not provided
	}

	// Create the review object based on the incoming data
	review := &data.Review{
		ProductID:    int64(*incomingReviewData.ProductID),
		Author:       *incomingReviewData.Author,
		Rating:       int64(*incomingReviewData.Rating),
		ReviewText:   *incomingReviewData.ReviewText,
		HelpfulCount: int32(*incomingReviewData.HelpfulCount),
		CreatedAt:    time.Now(),
	}

	// Initialize a Validator instance
	v := validator.New()

	// Validate the review object
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the review into the database
	err = a.reviewModel.InsertReview(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Set a Location header. The path to the newly created review
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/reviews/%d", review.ReviewID))

	data := envelope{
		"Review": review,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) displayReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query teh comments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Call Get() to retrieve the comment with the specified id
	review, err := a.reviewModel.GetReview(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the comment
	data := envelope{
		"Review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Read the review ID from the URL parameter
	id, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Retrieve the review from the database
	review, err := a.reviewModel.GetReview(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Define a struct to hold incoming JSON data
	var incomingReviewData struct {
		Author     *string `json:"author"`
		Rating     *int64  `json:"rating"`      // integer with a constraint (1-5)
		ReviewText *string `json:"review_text"` // non-null text field
	}

	// Decode the incoming JSON into the struct
	err = a.readJSON(w, r, &incomingReviewData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Update the fields if provided in the incoming JSON
	if incomingReviewData.Author != nil {
		review.Author = *incomingReviewData.Author
	}
	if incomingReviewData.Rating != nil {
		review.Rating = *incomingReviewData.Rating
	}
	if incomingReviewData.ReviewText != nil {
		review.ReviewText = *incomingReviewData.ReviewText
	}

	// Validate the updated review
	v := validator.New()
	data.ValidateReview(v, review) // Assuming ValidateReview is the correct validation function for reviews
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the review in the database
	err = a.reviewModel.UpdateReview(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated review as a JSON response
	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.reviewModel.DeleteReview(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.RIDnotFound(w, r, id) // Pass the ID to the custom message handler
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "Review successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listReviewHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Author string
		data.Filters
	}

	queryParameters := r.URL.Query()

	// Get author and rating from query parameters
	queryParametersData.Author = a.getSingleQueryParameter(queryParameters, "author", "")

	v := validator.New()

	// Get pagination and sorting filters
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "review_id")
	queryParametersData.Filters.SortSafeList = []string{"review_id", "author", "-review_id", "-author"}

	// Validate filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch reviews
	reviews, metadata, err := a.reviewModel.GetAllReviews(
		queryParametersData.Author,
		queryParametersData.Filters,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Prepare and write response
	responseData := envelope{
		"Reviews":   reviews,
		"@metadata": metadata,
	}
	if err := a.writeJSON(w, http.StatusOK, responseData, nil); err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listProductReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the URL /v1/comments/:id so that we
	// can use it to query teh comments table. We will
	// implement the readIDParam() function later
	id, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Check if the review exists
	exists, err := a.productModel.ProductExists(id) // Assuming you have an Exists method in reviewModel
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if !exists {
		a.PRIDnotFound(w, r, id)
		return
	}

	// Call Get() to retrieve the comment with the specified id
	review, err := a.reviewModel.GetAllProductReviews(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the comment
	data := envelope{
		"Review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

func (a *applicationDependencies) HelpfulCountHandler(w http.ResponseWriter, r *http.Request) {
	// Read the review ID from the URL parameter
	id, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Check if the review exists
	exists, err := a.reviewModel.Exists(id) // Assuming you have an Exists method in reviewModel
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if !exists {
		a.RRIDnotFound(w, r, id)
		return
	}

	// Retrieve and update the review's helpful count in the database
	review, err := a.reviewModel.UpdateHelpfulCount(id)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated review as a JSON response
	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

	// Log a confirmation message for the incremented helpful count
	confirmationMessage := fmt.Sprintf("\nHelpful count incremented by 1 for the review with id = %d", id)
	fmt.Fprintln(w, confirmationMessage)
}

func (a *applicationDependencies) getProductReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Read the product ID from the request
	pid, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Read the review ID from the request
	rid, err := a.readIDParam(r, "rid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Retrieve the review from the model using the new GetProductReview function
	review, err := a.reviewModel.GetProductReview(rid, pid)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Send the updated review as a JSON response
	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
