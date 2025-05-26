package controllers

import (
    "net/http"
    "encoding/json"

    "rewinged/models"
    "rewinged/settings"
)

func GetInformation(w http.ResponseWriter, r *http.Request) {
    // It's okay to always return the API 1.7.0+ version of the InformationSchema
    // because the only new property since 1.1.0, Authentication, is optional and
    // will be omitted if unset making the response identical to API 1.1.0
    response := new(models.API_Information_1_7_0)
    response.Data.SourceIdentifier = "rewinged"
    // New API schema versions have to be included here or winget CLI client won't pick
    // up the features / data fields from newer packages even if they are returned
    response.Data.ServerSupportedVersions = []string{"1.1.0", "1.4.0", "1.5.0", "1.6.0", "1.7.0", "1.9.0"}

    switch settings.SourceAuthenticationType {
    case "microsoftEntraId":
        // 1.7.0 is the minimum schema version that supports source-authentication
        response.Data.ServerSupportedVersions = []string{"1.7.0", "1.9.0"}
        response.Data.Authentication = &models.API_Authentication_1_7_0{
            AuthenticationType: "microsoftEntraId",
            MicrosoftEntraIdAuthenticationInfo: struct {
                Resource string `yaml:"Resource"`
                Scope string `yaml:"Scope" json:",omitempty"`
            }{
                Resource: settings.SourceAuthenticationEntraIDResource, // Entra Application ID
                Scope: "user_impersonation",
            },
        }
    default:
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

