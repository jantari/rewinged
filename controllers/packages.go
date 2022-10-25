package controllers

import (
    "fmt"
    "github.com/gin-gonic/gin"

    "rewinged/models"
)

func GetPackages(c *gin.Context) {
    response := &models.PackageMultipleResponse{
        Data: models.Manifests.GetAllPackageIdentifiers(),
    }

    fmt.Println(response)
    c.JSON(200, response)
}
