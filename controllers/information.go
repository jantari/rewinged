package controllers

import (
    "github.com/gin-gonic/gin"

    "rewinged/models"
)

func GetInformation(c *gin.Context) {
    response := new(models.API_Information_1_1_0)
    response.Data.SourceIdentifier = "rewinged"
    // New API schema versions have to be included here or winget CLI client won't pick
    // up the features / data fields from newer packages even if they are returned
    response.Data.ServerSupportedVersions = []string{"1.1.0", "1.4.0"}
    c.JSON(200, response)
}

