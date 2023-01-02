package models

import (
    "fmt"
    "sync"
    "strings"
    "reflect"
)

func getMapValues[M ~map[K]V, K comparable, V any](m M) []V {
    r := make([]V, 0, len(m))
    for _, v := range m {
        r = append(r, v)
    }
    return r
}

func caseInsensitiveContains(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.Contains(s, substr)
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
            //fmt.Println("FOUND THE FIELD", name, "!")
            return v.Field(i)
        }
        // push field to queue
        queue = append(queue, v.Field(i))
    }
  }
  return reflect.Value{}
}

// Internal in-memory data store of all manifest data
type ManifestsStore struct {
    sync.RWMutex
    internal map[string]map[string]ManifestVersionInterface
}

func (ms *ManifestsStore) Set(packageidentifier string, packageversion string, value ManifestVersionInterface) {
    ms.Lock()
    vmap, ok := ms.internal[packageidentifier]
    if !ok {
        vmap = make(map[string]ManifestVersionInterface)
        ms.internal[packageidentifier] = vmap
    }
    vmap[packageversion] = value
    ms.Unlock()
}

func (ms *ManifestsStore) GetAllVersions(packageidentifier string) (value []ManifestVersionInterface) {
    ms.RLock()
    result := getMapValues(ms.internal[packageidentifier])
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) Get(packageidentifier string, packageversion string) (value ManifestVersionInterface) {
    ms.RLock()
    result := ms.internal[packageidentifier][packageversion]
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) GetAll() (value map[string][]ManifestVersionInterface) {
    ms.RLock()
    var m = make(map[string][]ManifestVersionInterface)
    for k := range ms.internal {
        m[k] = getMapValues(ms.internal[k])
    }
    ms.RUnlock()
    return m
}

func (ms *ManifestsStore) GetAllPackageIdentifiers() (value []Package) {
    ms.RLock()
    var p []Package
    for k := range ms.internal {
        p = append(p, Package{
            PackageIdentifier: k,
        })
    }
    ms.RUnlock()
    return p
}

func (ms *ManifestsStore) GetManifestCount() (value int) {
    ms.RLock()
    var count int
    count = len(Manifests.internal)
    ms.RUnlock()
    return count
}

func (ms *ManifestsStore) GetByKeyword (keyword string) map[string][]ManifestVersionInterface {
  var manifestResultsMap = make(map[string][]ManifestVersionInterface)
  ms.RLock()
  for packageIdentifier, packageVersions := range ms.internal {
    for _, version := range packageVersions {
      if caseInsensitiveContains(version.GetDefaultLocalePackageName(), keyword) || caseInsensitiveContains(version.GetDefaultLocaleShortDescription(), keyword) {
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], version)
      }
    }
  }
  ms.RUnlock()

  return manifestResultsMap
}

func (ms *ManifestsStore) GetByMatchFilter (inclusions []SearchRequestPackageMatchFilter, filters []SearchRequestPackageMatchFilter) map[string][]ManifestVersionInterface {
  var manifestResultsMap = make(map[string][]ManifestVersionInterface)
  normalizeReplacer := strings.NewReplacer(" ", "", "-", "", "+", "")

  ms.RLock()
  for packageIdentifier, packageVersions := range ms.internal {
    // Loop through every version of the package as well because MatchFields like
    // ProductCode, PackageName etc. can change between versions.
    NEXT_VERSION:
    for _, packageVersion := range packageVersions {
      // From what I can gather from https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/CosmosDataStore.cs#L452
      // the difference between inclusions and filters are that inclusions are evaluated with a logical OR (only one of them has to match) and filters are evaluated with a logical AND
      // (all filter specified have to match) - so this is what I implemented here. But I am not 100% sure this is the correct/intended use for inclusions vs. filters.

      // process filters (if any)
      for _, filter := range filters {
        var requestMatchValue string

        switch filter.PackageMatchField {
          case NormalizedPackageNameAndPublisher:
            // winget only ever sends the package / software name, the publisher isn't included so to
            // enable proper matching we also only compare against the normalized packagename.
            requestMatchValue = normalizeReplacer.Replace(strings.ToLower(packageVersion.GetDefaultLocalePackageName()))
          case PackageIdentifier:
            // We don't need to recursively search for this field, it's easy to get to
            requestMatchValue = packageIdentifier
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
            f := findField(packageVersion, string(filter.PackageMatchField))
            requestMatchValue = string(f.String())
        }

        // Because all filters (if any) must match (logical AND)
        // we just skip to the next packageversion if any did not match
        switch filter.RequestMatch.MatchType {
          // TODO: `winget list -s rewinged-local -q lapce` searches for the ProductCode with MatchType Exact
          // Why does it use MatchType Exact?? Does the reference / official source normalize all ProductCodes on ingest??
          case Exact:
            if requestMatchValue != filter.RequestMatch.KeyWord {
              continue NEXT_VERSION
            }
          case CaseInsensitive:
            if !strings.EqualFold(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          case StartsWith:
            // StartsWith is implemented as case-sensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if !strings.HasPrefix(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          case Substring:
            // Substring comparison is case-insensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if !caseInsensitiveContains(requestMatchValue, filter.RequestMatch.KeyWord) {
              continue NEXT_VERSION
            }
          default:
            // Unimplemented: Wildcard, Fuzzy, FuzzySubstring
        }
      }

      if len(inclusions) == 0 {
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], packageVersion)
        continue NEXT_VERSION
      }

      var anyInclusionMatched bool
      anyInclusionMatched = true
      // process inclusions (if any)
      NEXT_INCLUSION:
      for _, inclusion := range inclusions {
        var requestMatchValue string
        anyInclusionMatched = false

        switch inclusion.PackageMatchField {
          case NormalizedPackageNameAndPublisher:
            // winget only ever sends the package / software name, the publisher isn't included so to
            // enable proper matching we also only compare against the normalized packagename.
            requestMatchValue = normalizeReplacer.Replace(strings.ToLower(packageVersion.GetDefaultLocalePackageName()))
          case PackageIdentifier:
            // We don't need to recursively search for this field, it's easy to get to
            requestMatchValue = packageIdentifier
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
            f := findField(packageVersion, string(inclusion.PackageMatchField))
            requestMatchValue = string(f.String())
        }

        switch inclusion.RequestMatch.MatchType {
          // TODO: `winget list -s rewinged-local -q lapce` searches for the ProductCode with MatchType Exact
          // Why does it use MatchType Exact?? Does the reference / official source normalize all ProductCodes on ingest??
          case Exact:
            if requestMatchValue == inclusion.RequestMatch.KeyWord {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case CaseInsensitive:
            if strings.EqualFold(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case StartsWith:
            // StartsWith is implemented as case-sensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if strings.HasPrefix(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          case Substring:
            // Substring comparison is case-insensitive, because it is that way in the reference implementation as well:
            // https://github.com/microsoft/winget-cli-restsource/blob/01542050d79da0efbd11c0a5be543cb970b86eb9/src/WinGet.RestSource/Cosmos/PredicateGenerator.cs#L92-L102
            if caseInsensitiveContains(requestMatchValue, inclusion.RequestMatch.KeyWord) {
              // Break out of the inclusions loop after one successful match
              anyInclusionMatched = true
              break NEXT_INCLUSION
            }
          default:
            // Unimplemented: Wildcard, Fuzzy, FuzzySubstring
        }
      }

      // All filters and inclusions have passed for this manifest, add it to the returned map
      if anyInclusionMatched {
        fmt.Println("Adding to the results map:", packageIdentifier, "version", packageVersion.GetPackageVersion())
        manifestResultsMap[packageIdentifier] = append(manifestResultsMap[packageIdentifier], packageVersion)
      }
    }
  }
  ms.RUnlock()

  return manifestResultsMap
}

// Global variable that will hold all in-memory manifest data
var Manifests = ManifestsStore{
    internal: make(map[string]map[string]ManifestVersionInterface),
}

