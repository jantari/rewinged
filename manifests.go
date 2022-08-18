package main

import (
  "fmt"
  "io/ioutil"
  "strings"
  "os"
  "reflect"

  "gopkg.in/yaml.v3"
)

type SingletonInstaller struct {
  Architecture string `yaml:"Architecture"`
  InstallerType string `yaml:"InstallerType"`
  InstallerUrl string `yaml:"InstallerUrl"`
  InstallerSha256 string `yaml:"InstallerSha256"`
  SignatureSha256 string `yaml:"SignatureSha256"`
}

type SingletonManifest struct {
  PackageIdentifier string `yaml:"PackageIdentifier"`
  PackageVersion string `yaml:"PackageVersion"`
  PackageLocale string `yaml:"PackageLocale"`
  Publisher string `yaml:"Publisher"`
  PackageName string `yaml:"PackageName"`
  License string `yaml:"License"`
  ShortDescription string `yaml:"ShortDescription"`
  Installers []SingletonInstaller `yaml:"Installers"`
  ManifestType string `yaml:"ManifestType"`
  ManifestVersion string `yaml:"ManifestVersion"`
}

func FindManifestFiles () []os.DirEntry {
  var manifestFiles = []os.DirEntry{}

  files, err := os.ReadDir("./packages")
  if err != nil {
    fmt.Println(err)
  }

  for _, file := range files {
    if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
      manifestFiles = append(manifestFiles, file)
    }
  }

  return manifestFiles
}

func ParseManifestFile (path string) *SingletonManifest {
  yamlFile, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Printf("yamlFile.Get err   #%v ", err)
  }

  manifest := &SingletonManifest{}
  err = yaml.Unmarshal(yamlFile, manifest)
  if err != nil {
    fmt.Printf("Unmarshal: %v", err)
  }

  return manifest
}

func GetPackagesByMatchFilter (manifests []SingletonManifest, searchfilters []SearchRequestPackageMatchFilter) []SingletonManifest {
  var manifestResults = []SingletonManifest{}

  NEXT_MANIFEST:
  for _, manifest := range manifests {
    for _, matchfilter := range searchfilters {
      // Get the value of a struct field passing in the field name as a string
      // Kinda like PowerShells $Variable.PSObject.Properties['Name'].Value
      // Source: https://stackoverflow.com/a/18931036
      r := reflect.ValueOf(manifest)
      f := reflect.Indirect(r).FieldByName(string(matchfilter.PackageMatchField))
      var fieldvalue = string(f.String())

      if strings.Contains(fieldvalue, matchfilter.RequestMatch.KeyWord) {
        manifestResults = append(manifestResults, manifest)
        // Jump to the next manifest to prevent returning the same one multiple times if it matched more than 1 search criteria
        continue NEXT_MANIFEST
      }
    }
  }

  return manifestResults
}

func GetPackagesByKeyword (manifests []SingletonManifest, keyword string) []SingletonManifest {
  var manifestResults = []SingletonManifest{}
  for _, manifest := range manifests {
    if strings.Contains(manifest.PackageName, keyword) || strings.Contains(manifest.ShortDescription, keyword) {
      manifestResults = append(manifestResults, manifest)
    }
  }

  return manifestResults
}

func GetPackageByIdentifier (manifests []SingletonManifest, packageidentifier string) *SingletonManifest {
  var manifestResult = &SingletonManifest{}
  manifestResult = nil
  for _, manifest := range manifests {
    if manifest.PackageIdentifier == packageidentifier {
      manifestResult = &manifest
    }
  }

  return manifestResult
}

