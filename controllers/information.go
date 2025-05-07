package controllers

import (
    "net/http"
    "encoding/json"

    "rewinged/models"
)

func GetInformation(w http.ResponseWriter, r *http.Request) {
    response := new(models.API_Information_1_1_0)
    response.Data.SourceIdentifier = "rewinged"
    // New API schema versions have to be included here or winget CLI client won't pick
    // up the features / data fields from newer packages even if they are returned
    response.Data.ServerSupportedVersions = []string{"1.1.0", "1.4.0", "1.5.0", "1.6.0", "1.7.0", "1.9.0"}

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

