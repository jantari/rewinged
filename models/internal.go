package models

import "sync"

func getMapValues[M ~map[K]V, K comparable, V any](m M) []V {
    r := make([]V, 0, len(m))
    for _, v := range m {
        r = append(r, v)
    }
    return r
}

// Internal in-memory data store of all manifest data
type ManifestsStore struct {
    sync.RWMutex
    internal map[string]map[string]Versions
}

func (ms *ManifestsStore) Set(packageidentifier string, packageversion string, value Versions) {
    ms.Lock()
    vmap, ok := ms.internal[packageidentifier]
    if !ok {
        vmap = make(map[string]Versions)
        ms.internal[packageidentifier] = vmap
    }
    vmap[packageversion] = value
    ms.Unlock()
}

func (ms *ManifestsStore) GetAllVersions(packageidentifier string) (value []Versions) {
    ms.RLock()
    result := getMapValues(ms.internal[packageidentifier])
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) Get(packageidentifier string, packageversion string) (value Versions) {
    ms.RLock()
    result := ms.internal[packageidentifier][packageversion]
    ms.RUnlock()
    return result
}

func (ms *ManifestsStore) GetAll() (value map[string][]Versions) {
    ms.RLock()
    var m = make(map[string][]Versions)
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

// Global variable that will hold all in-memory manifest data
var Manifests = ManifestsStore{
    internal: make(map[string]map[string]Versions),
}

