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

// A singleton manifest can only contain one locale and one installer
// Schema: https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.singleton.1.1.0.json
type SingletonManifest struct {
  PackageIdentifier string `yaml:"PackageIdentifier"`
  PackageVersion string `yaml:"PackageVersion"`
  PackageLocale string `yaml:"PackageLocale"`
  Publisher string `yaml:"Publisher"`
  PackageName string `yaml:"PackageName"`
  License string `yaml:"License"`
  ShortDescription string `yaml:"ShortDescription"`
  Installers [1]Installer `yaml:"Installers"`
  ManifestType string `yaml:"ManifestType"`
  ManifestVersion string `yaml:"ManifestVersion"`
}

type Manifest struct {
    PackageIdentifier string
    Versions []Versions
}

// The struct for a separate version manifest file
type VersionManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    DefaultLocale string `yaml:"DefaultLocale"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate installer manifest file
type InstallerManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`

    Installers []Installer `yaml:"Installers"`

    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

type Versions struct {
    PackageVersion string
    DefaultLocale Locale
    Channel string
    Locales []Locale
    Installers []Installer
}

type Installer struct {
    Architecture Architecture `yaml:"Architecture"`
    InstallerType InstallerType `yaml:"InstallerType"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
    SignatureSha256 string `yaml:"SignatureSha256" json:",omitempty"` // winget runs into an exception internally when this is an empty string (ParseFromHexString: Invalid value size), so omit in API responses if empty
    ProductCode string `yaml:"ProductCode"`
}

type Locale struct {
    PackageLocale string `yaml:"PackageLocale"`
//    Moniker // Is this needed for DefaultLocale?
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

type Architecture string

const (
    neutral Architecture = "neutral"
    x86 = "x86"
    x64 = "x64"
    arm = "arm"
    arm64 = "arm64"
)

type InstallerType string

const (
    msix InstallerType = "msix"
    msi = "msi"
    appx = "appx"
    exe = "exe"
    zip = "zip"
    inno = "inno"
    nullsoft = "nullsoft"
    wix = "wix"
    burn = "burn"
    pwa = "pwa"
    msstore = "msstore"
)

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
//    Package Package
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

