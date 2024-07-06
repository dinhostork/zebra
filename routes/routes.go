package routes

import (
	"net/http"
	handleupload "zebra/internal/handle_upload"
)

// SetupRoutes sets up the routes for the application
func SetupRoutes() {
	http.HandleFunc("/upload", handleupload.HandleUpload)

	http.ListenAndServe(":8080", nil)
}
