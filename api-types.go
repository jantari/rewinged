package main

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
    Versions []Versions
}

type Versions struct {
    PackageVersion string
    DefaultLocale Locale
    Channel string
    Locales []Locale
    Installers []Installer
}

type Locale struct {
    PackageLocale string `yaml:"PackageLocale"`
//    Moniker
    Publisher string `yaml:"Publisher"`
//    PublisherUrl
//    PublisherSupportUrl
//    PrivacyUrl
//    Author
    PackageName string `yaml:"PackageName"`
//    PackageUrl
    License string `yaml:"License"`
//    LicenseUrl
//    Copyright
//    CopyrightUrl
    ShortDescription string `yaml:"ShortDescription"`
//    Description
//    Tags
//    ReleaseNotes
//    ReleaseNotesUrl
//    Agreements
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
    Exact MatchType = "Exact"
    CaseInsensitive = "CaseInsensitive"
    StartsWith      = "StartsWith"
    Substring       = "Substring"
    Wildcard        = "Wildcard"
    Fuzzy           = "Fuzzy"
    FuzzySubstring  = "FuzzySubstring"
)

type PackageMatchField string

const (
    PackageIdentifier PackageMatchField = "PackageIdentifier"
    PackageName = "PackageName"
    Moniker = "Moniker"
    Command = "Command"
    Tag = "Tag"
    PackageFamilyName = "PackageFriendlyName"
    ProductCode = "ProductCode"
    NormalizedPackageNameAndPublisher = "NormalizedPackageNameAndPublisher"
    Market = "Market"
)

type QueryParameter string

const (
    Version QueryParameter = "Version"
    Channel = "Channel"
//    Market = "Market" // Already declared in PackageMatchField enum
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

type ManifestSearchVersion struct {
    PackageVersion string
    Channel string //maxlength: 16, unused
    PackageFamilyNames []string
    ProductCodes []string
}

type ManifestSearchResponse struct {
    PackageIdentifier string
    PackageName string
    Publisher string
    Versions []ManifestSearchVersion
}

type ManifestSingleResponse struct {
    Data *Manifest
    RequiredQueryParameters []QueryParameter
    UnsupportedQueryParameters []QueryParameter
}

type ManifestSearchResult struct {
    Data []ManifestSearchResponse
    RequiredPackageMatchFields []PackageMatchField
    UnsupportedPackageMatchFields []PackageMatchField
}

