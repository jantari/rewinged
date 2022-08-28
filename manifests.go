package main

import (
  "fmt"
  "errors"
  "io/ioutil"
  "strings"
  "os"
  "reflect"

  "gopkg.in/yaml.v3"
  "github.com/mitchellh/mapstructure"
)

// These properties are required in all manifest types:
// singleton, version, locale and installer
type BaseManifest struct {
  PackageIdentifier string `yaml:"PackageIdentifier"`
  PackageVersion string `yaml:"PackageVersion"`
  ManifestType string `yaml:"ManifestType"`
  ManifestVersion string `yaml:"ManifestVersion"`
}

type SingletonManifest struct {
  PackageIdentifier string `yaml:"PackageIdentifier"`
  PackageVersion string `yaml:"PackageVersion"`
  PackageLocale string `yaml:"PackageLocale"`
  Publisher string `yaml:"Publisher"`
  PackageName string `yaml:"PackageName"`
  License string `yaml:"License"`
  ShortDescription string `yaml:"ShortDescription"`
  Installers []Installer `yaml:"Installers"`
  ManifestType string `yaml:"ManifestType"`
  ManifestVersion string `yaml:"ManifestVersion"`
}

func GetManifests (path string) []SingletonManifest {
  var manifests = []SingletonManifest{}
  var nonSingletonsMap = make(map[string][]string)

  files, err := os.ReadDir(path)
  if err != nil {
    fmt.Println(err)
  }

  for _, file := range files {
    fmt.Printf("Path: %s IsDir: %v\n", path + "/" + file.Name(), file.IsDir())
    if file.IsDir() {
      var manifests_from_dir = GetManifests(path + "/" + file.Name())
      manifests = append(manifests, manifests_from_dir...)
    } else {
      if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
        var basemanifest = ParseFileAsBaseManifest(path + "/" + file.Name())
        fmt.Printf("  BaseManifest: %+v\n", basemanifest)

        if basemanifest.ManifestType == "singleton" {
          var manifest = ParseManifestFile(path + "/" + file.Name())
          manifests = append(manifests, *manifest)
        } else {
          nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion] = append(nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion], path + "/" + file.Name())
        }
      }
    }
  }

  if len(nonSingletonsMap) > 0 {
    fmt.Println("  There are multi-file manifests in this directory")
    fmt.Printf("%+v\n", nonSingletonsMap)
    for key, value := range nonSingletonsMap {
      fmt.Println("    Merging manifest for package", key)
      var merged_manifest, err = ParseManifestMultiFile(value...)
      if err != nil {
        fmt.Println(err)
      } else {
        fmt.Printf("\n%+v\n", merged_manifest)
      }
      manifests = append(manifests, *merged_manifest)
    }
  }

  return manifests
}


func ParseManifestMultiFile (filenames ...string) (*SingletonManifest, error) {
  if len(filenames) <= 0 {
    return nil, errors.New("You must provide at least one filename for reading Values")
  }
  var resultValues map[string]interface{}

  for _, filename := range filenames {
    var override map[string]interface{}
    bs, err := ioutil.ReadFile(filename)
    if err != nil {
      //log.Info(err)
      continue
    }
    if err := yaml.Unmarshal(bs, &override); err != nil {
      //log.Info(err)
      continue
    }

    //check if is nil. This will only happen for the first filename
    if resultValues == nil {
      resultValues = override
    } else {
      for k, v := range override {
        resultValues[k] = v
      }
    }
  }

  result := &SingletonManifest{}
  err := mapstructure.Decode(resultValues, &result)
  return result, err
}

func ParseFileAsBaseManifest (path string) *BaseManifest {
  yamlFile, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Printf("yamlFile.Get err   #%v ", err)
  }

  manifest := &BaseManifest{}
  err = yaml.Unmarshal(yamlFile, manifest)
  if err != nil {
    fmt.Printf("Unmarshal: %v", err)
  }

  return manifest
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

func CaseInsensitiveContains(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.Contains(s, substr)
}

func CaseInsensitiveHasSuffix(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.HasSuffix(s, substr)
}

// Modified from: https://stackoverflow.com/a/38407429
func findField(v interface{}, name string) reflect.Value {
  // create queue of values to search. Start with the function arg.
  queue := []reflect.Value{reflect.ValueOf(v)}
  for len(queue) > 0 {
    v := queue[0]
    queue = queue[1:]
    // dereference pointers
    for v.Kind() == reflect.Ptr {
        v = v.Elem()
    }
    // check all elements in slices
    if v.Kind() == reflect.Slice {
      if v.Len() > 0 {
        for i := 0; i < v.Len(); i++ {
          queue = append(queue, v.Index(i))
        }
      } else {
        //fmt.Println("CONTINUE (EMPTY SLICE)")
        continue
      }
    }
    // ignore if this is not a struct
    if v.Kind() != reflect.Struct {
        //fmt.Println("CONTINUE (NOT STRUCT)")
        continue
    }
    // iterate through fields looking for match on name
    t := v.Type()
    for i := 0; i < v.NumField(); i++ {
        //fmt.Println("TESTING FIELD", t.Field(i).Name)
        if t.Field(i).Name == name {
            // found it!
            fmt.Println("FOUND THE FIELD!")
            return v.Field(i)
        }
        // push field to queue
        queue = append(queue, v.Field(i))
    }
  }
  return reflect.Value{}
}


func GetPackagesByMatchFilter (manifests []SingletonManifest, searchfilters []SearchRequestPackageMatchFilter) []SingletonManifest {
  var manifestResults = []SingletonManifest{}

  NEXT_MANIFEST:
  for _, manifest := range manifests {
    for _, matchfilter := range searchfilters {
      // Get the value of a nested struct field passing in the field name to search for as a string
      // Source: https://stackoverflow.com/a/38407429
      f := findField(manifest, string(matchfilter.PackageMatchField))
      //fmt.Printf("findfield returns:\n%+v\n", f)
      var fieldvalue = string(f.String())

      if CaseInsensitiveContains(fieldvalue, matchfilter.RequestMatch.KeyWord) {
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
    if CaseInsensitiveContains(manifest.PackageName, keyword) || CaseInsensitiveContains(manifest.ShortDescription, keyword) {
      manifestResults = append(manifestResults, manifest)
    }
  }

  return manifestResults
}

func GetPackageByIdentifier (manifests []SingletonManifest, packageidentifier string) *SingletonManifest {
  for _, manifest := range manifests {
    if manifest.PackageIdentifier == packageidentifier {
      fmt.Println("searched ID", packageidentifier, "equals a package:", manifest.PackageIdentifier)
      return &manifest
    }
  }

  return nil
}

