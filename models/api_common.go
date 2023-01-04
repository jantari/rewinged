package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

type WingetApiError struct {
    ErrorCode    int
    ErrorMessage string
}

type Package struct {
    PackageIdentifier string
}

type Manifest struct {
    PackageIdentifier string
    Versions []API_ManifestVersionInterface
}

type API_ManifestVersionInterface interface {
    GetDefaultLocalePackageName() string
    GetDefaultLocalePublisher() string
    GetDefaultLocaleShortDescription() string
    GetPackageVersion() string
    GetInstallerProductCodes() []string
}

type PackageMultipleResponse struct {
    Data []Package
    ContinuationToken string
}

type Information struct {
    Data struct {
        SourceIdentifier        string
        ServerSupportedVersions []string
    }
}

type MatchType string

const (
    Exact MatchType           = "Exact"
    CaseInsensitive MatchType = "CaseInsensitive"
    StartsWith MatchType      = "StartsWith"
    Substring MatchType       = "Substring"
    Wildcard MatchType        = "Wildcard"
    Fuzzy MatchType           = "Fuzzy"
    FuzzySubstring MatchType  = "FuzzySubstring"
)

type PackageMatchField string

const (
    PackageIdentifier PackageMatchField = "PackageIdentifier"
    PackageName PackageMatchField = "PackageName"
    Moniker PackageMatchField = "Moniker"
    Command PackageMatchField = "Command"
    Tag PackageMatchField = "Tag"
    PackageFamilyName PackageMatchField = "PackageFriendlyName"
    ProductCode PackageMatchField = "ProductCode"
    NormalizedPackageNameAndPublisher PackageMatchField = "NormalizedPackageNameAndPublisher"
    Market PackageMatchField = "Market"
)

type QueryParameter string

const (
    QueryParameterVersion QueryParameter = "Version"
    QueryParameterChannel QueryParameter = "Channel"
    QueryParameterMarket QueryParameter = "Market"
)

type SearchRequestMatch struct {
    KeyWord string
    MatchType MatchType
}

type SearchRequestPackageMatchFilter struct {
    PackageMatchField PackageMatchField
    RequestMatch SearchRequestMatch
}

type ManifestSearch struct {
    MaximumResults int
    FetchAllManifests bool
    Query SearchRequestMatch
    Inclusions []SearchRequestPackageMatchFilter
    Filters []SearchRequestPackageMatchFilter
}

type ManifestSearchVersionInterface interface {
    ManifestSearchVersion_1_1_0 | ManifestSearchVersion_1_4_0
}

type ManifestSearchResponse[MSVI ManifestSearchVersionInterface] struct {
    PackageIdentifier string
    PackageName string
    Publisher string
    Versions []MSVI
}

type ManifestSingleResponse struct {
    Data *Manifest
    RequiredQueryParameters []QueryParameter
    UnsupportedQueryParameters []QueryParameter
}

type ManifestSearchResult[MSVI ManifestSearchVersionInterface] struct {
    Data []ManifestSearchResponse[MSVI]
    RequiredPackageMatchFields []PackageMatchField
    UnsupportedPackageMatchFields []PackageMatchField
}

type Architecture string

const (
    neutral Architecture = "neutral"
    x86 Architecture = "x86"
    x64 Architecture = "x64"
    arm Architecture = "arm"
    arm64 Architecture = "arm64"
)

type Scope string

const (
    user Scope = "user"
    machine Scope = "machine"
)

type InstallMode string

const (
    interactive InstallMode = "interactive"
    silent InstallMode = "silent"
    silentWithProgress InstallMode = "silentWithProgress"
)

