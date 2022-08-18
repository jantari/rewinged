package main

import (
  "fmt"
  "io/ioutil"
  "strings"
  "os"

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

func FindManifestFiles() []os.DirEntry {
  var manifestFiles = []os.DirEntry{}

  files, err := os.ReadDir("./packages")
  if err != nil {
    fmt.Println(err)
  }

  for _, file := range files {
    if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
      fmt.Println("Found package manifest:", file.Name())
      manifestFiles = append(manifestFiles, file)
    }

    //var bottommanifest = ParseManifest("./packages/" + file.Name())
    //fmt.Printf("%+v\n", bottommanifest)
  }

  return manifestFiles
}

func ParseManifestFile(path string) *SingletonManifest {
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
