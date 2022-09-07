package main

import (
  "fmt"
  "errors"
  "io/ioutil"
  "strings"
  "os"
  "reflect"

  "gopkg.in/yaml.v3"
  //"github.com/mitchellh/mapstructure"
)

func GetManifests (path string) []Manifest {
  var manifests = []Manifest{}
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
          nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion] =
            append(nonSingletonsMap[basemanifest.PackageIdentifier + "/" + basemanifest.PackageVersion], path + "/" + file.Name())
        }
      }
    }
  }

  if len(nonSingletonsMap) > 0 {
    fmt.Println("  There are multi-file manifests in this directory")
    fmt.Printf("%+v\n", nonSingletonsMap)
    for key, value := range nonSingletonsMap {
      fmt.Println("    Merging manifest for package", key)
      var merged_manifest, err = ParseMultiFileManifest(value...)
      if err != nil {
        fmt.Println("    Could not parse the manifest files for this package", err)
      } else {
        manifests = append(manifests, *merged_manifest)
      }
    }
  }

  return manifests
}

//
func ParseMultiFileManifest (filenames ...string) (*Manifest, error) {
  if len(filenames) <= 0 {
    return nil, errors.New("You must provide at least one filename for reading Values")
  }

  var packageidentifier string
  versions      := []VersionManifest{}
  installers    := []InstallerManifest{}
  locales       := []LocaleManifest{}
  defaultlocale := &DefaultLocaleManifest{}

  for _, file := range filenames {
    var basemanifest = ParseFileAsBaseManifest(file)
    packageidentifier = basemanifest.PackageIdentifier

    yamlFile, err := ioutil.ReadFile(file)
    if err != nil {
      fmt.Printf("yamlFile.Get err   #%v ", err)
    }
    switch basemanifest.ManifestType {
      case "version":
        fmt.Println("Parsing version manifest ...")
        version := &VersionManifest{}
        err = yaml.Unmarshal(yamlFile, version)
        if err != nil {
          fmt.Printf("unmarshal version err   #%v ", err)
        }
        fmt.Printf("unmarshalled version %+v\n", *version)
        versions = append(versions, *version)
      case "installer":
        fmt.Println("Parsing installer manifest ...")
        installer := &InstallerManifest{}
        err = yaml.Unmarshal(yamlFile, installer)
        if err != nil {
          fmt.Printf("unmarshal installer err   #%v ", err)
        }
        fmt.Printf("unmarshalled installer %+v\n", installer)
        installers = append(installers, *installer)
      case "locale":
        fmt.Println("Parsing locale manifest ...")
        locale := &LocaleManifest{}
        err = yaml.Unmarshal(yamlFile, locale)
        if err != nil {
          fmt.Printf("unmarshal locale err   #%v ", err)
        }
        fmt.Printf("unmarshalled locale %+v\n", locale)
        locales = append(locales, *locale)
      case "defaultLocale":
        fmt.Println("Parsing defaultlocale manifest ...")
        err = yaml.Unmarshal(yamlFile, defaultlocale)
        if err != nil {
          fmt.Printf("unmarshal defaultlocale err   #%v ", err)
        }
        fmt.Printf("unmarshalled defaultlocale %+v\n", defaultlocale)
      default:
    }
  }

  // This transforms the manifest data into the format the API will return.
  // This logic should probably be moved out of this function, so that it returns
  // the full unaltered data from the combined manifests - and restructuring to
  // API-format will happen somewhere else
  var apiLocales []Locale
  for _, locale := range locales {
    apiLocales = append(apiLocales, *localeManifestToAPILocale(locale))
  }

  versions_api := []Versions{
    {
      PackageVersion: versions[0].PackageVersion,
      DefaultLocale: *defaultLocaleManifestToAPIDefaultLocale(*defaultlocale),
      Channel: "",
      Locales: apiLocales,
      Installers: installers[0].Installers,
    },
  }

  manifest := &Manifest {
    PackageIdentifier: packageidentifier,
    Versions: versions_api[:],
  }

  return manifest, nil //err
}

func defaultLocaleManifestToAPIDefaultLocale (locm DefaultLocaleManifest) *DefaultLocale {
  return &DefaultLocale{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Moniker: locm.Moniker,
    Tags: locm.Tags,
    Agreements: locm.Agreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
}

func localeManifestToAPILocale (locm LocaleManifest) *Locale {
  return &Locale{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Tags: locm.Tags,
    Agreements: locm.Agreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
}

func GetLocaleByName (locales []Locale, localename string) *Locale {
  fmt.Println("GetLocaleByName: looking for the locale", localename, "in", len(locales), "total locales")
  for _, locale := range locales {
    if locale.PackageLocale == localename {
      return &locale
    }
  }

  return nil
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

func ParseManifestFile (path string) *Manifest {
  yamlFile, err := ioutil.ReadFile(path)
  if err != nil {
    fmt.Printf("yamlFile.Get err   #%v ", err)
  }

  singleton := &SingletonManifest{}
  err = yaml.Unmarshal(yamlFile, singleton)
  if err != nil {
    fmt.Printf("Unmarshal singleton error: %v", err)
  }

  manifest := SingletonToStandardManifest(singleton)

  return manifest
}

func SingletonToStandardManifest (singleton *SingletonManifest) *Manifest {
  manifest := &Manifest {
    PackageIdentifier: singleton.PackageIdentifier,
    Versions: []Versions {
      {
        PackageVersion: singleton.PackageVersion,
        DefaultLocale: DefaultLocale {
          PackageLocale: singleton.PackageLocale,
          PackageName: singleton.PackageName,
          Publisher: singleton.Publisher,
          ShortDescription: singleton.ShortDescription,
          License: singleton.License,
        },
        Channel: "",
        Locales: []Locale{},
        Installers: singleton.Installers[:],
      },
    },
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
            fmt.Println("FOUND THE FIELD", name, "!")
            return v.Field(i)
        }
        // push field to queue
        queue = append(queue, v.Field(i))
    }
  }
  return reflect.Value{}
}


func GetPackagesByMatchFilter (manifests []Manifest, searchfilters []SearchRequestPackageMatchFilter) []Manifest {
  var manifestResults = []Manifest{}

  NEXT_MANIFEST:
  for _, manifest := range manifests {
    for _, matchfilter := range searchfilters {
      var requestMatchValue string

      switch matchfilter.PackageMatchField {
        case NormalizedPackageNameAndPublisher:
          // winget only ever sends the package / software name, the publisher isn't included so to
          // enable proper matching we also only compare against the normalized packagename.
          requestMatchValue = strings.ReplaceAll(strings.ToLower(manifest.Versions[0].DefaultLocale.PackageName), " ", "")
        case PackageIdentifier:
          fallthrough
        case PackageName:
          fallthrough
        case Moniker:
          fallthrough
        case Command:
          fallthrough
        case Tag:
          fallthrough
        case PackageFamilyName:
          fallthrough
        case ProductCode:
          fallthrough
        case Market:
          fallthrough
        default:
          // Just search the whole struct for a field with the right name
          // Get the value of a nested struct field passing in the field name to search for as a string
          // Source: https://stackoverflow.com/a/38407429
          f := findField(manifest, string(matchfilter.PackageMatchField))
          requestMatchValue = string(f.String())
      }

      switch matchfilter.RequestMatch.MatchType {
        // TODO: `winget list -s rewinged-local -q lapce` searches for the ProductCode with MatchType Exact
        // Why does it use MatchType Exact?? Does the reference / official source normalize all ProductCodes on ingest??
        case Exact:
          if requestMatchValue == matchfilter.RequestMatch.KeyWord {
            manifestResults = append(manifestResults, manifest)
            // Jump to the next manifest to prevent returning the same one multiple times if it matched more than 1 search criteria
            continue NEXT_MANIFEST
          }
        case CaseInsensitive:
          if strings.EqualFold(requestMatchValue, matchfilter.RequestMatch.KeyWord) {
            manifestResults = append(manifestResults, manifest)
            // Jump to the next manifest to prevent returning the same one multiple times if it matched more than 1 search criteria
            continue NEXT_MANIFEST
          }
        case StartsWith:
          // StartsWith is implemented as case-sensitive, because it is that way in the reference implementation as well:
          // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
          if strings.HasPrefix(requestMatchValue, matchfilter.RequestMatch.KeyWord) {
            manifestResults = append(manifestResults, manifest)
            // Jump to the next manifest to prevent returning the same one multiple times if it matched more than 1 search criteria
            continue NEXT_MANIFEST
          }
        case Substring:
          // Substring comparison is case-insensitive, because it is that way in the reference implementation as well:
          // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
          if CaseInsensitiveContains(requestMatchValue, matchfilter.RequestMatch.KeyWord) {
            manifestResults = append(manifestResults, manifest)
            // Jump to the next manifest to prevent returning the same one multiple times if it matched more than 1 search criteria
            continue NEXT_MANIFEST
          }
        default:
          // Unimplemented: Wildcard, Fuzzy, FuzzySubstring
      }
    }
  }

  fmt.Printf("ADVANCED QUERY RESULTS: %+v\n", manifestResults)

  return manifestResults
}

func GetPackagesByKeyword (manifests []Manifest, keyword string) []Manifest {
  var manifestResults = []Manifest{}
  for _, manifest := range manifests {
    if CaseInsensitiveContains(manifest.Versions[0].DefaultLocale.PackageName, keyword) || CaseInsensitiveContains(manifest.Versions[0].DefaultLocale.ShortDescription, keyword) {
      manifestResults = append(manifestResults, manifest)
    }
  }

  return manifestResults
}

func GetPackageByIdentifier (manifests []Manifest, packageidentifier string) *Manifest {
  for _, manifest := range manifests {
    if manifest.PackageIdentifier == packageidentifier {
      fmt.Println("searched ID", packageidentifier, "equals a package:", manifest.PackageIdentifier)
      return &manifest
    }
  }

  return nil
}

