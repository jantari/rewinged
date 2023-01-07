package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

type API_Information_1_1_0 struct {
    Data struct {
        SourceIdentifier        string
        ServerSupportedVersions []string
    }
}

// API_ManifestVersion_1_1_0 implements all of the API_ManifestVersionInterface interface methods
type API_ManifestVersion_1_1_0 struct {
    PackageVersion string
    DefaultLocale API_DefaultLocale_1_1_0
    Channel string
    Locales []API_Locale_1_1_0
    Installers []API_Installer_1_1_0
}

func (ver API_ManifestVersion_1_1_0) GetDefaultLocalePackageName() string {
    return ver.DefaultLocale.PackageName
}

func (ver API_ManifestVersion_1_1_0) GetDefaultLocalePublisher() string {
    return ver.DefaultLocale.Publisher
}

func (ver API_ManifestVersion_1_1_0) GetDefaultLocaleShortDescription() string {
    return ver.DefaultLocale.ShortDescription
}

func (ver API_ManifestVersion_1_1_0) GetPackageVersion() string {
    return ver.PackageVersion
}

func (ver API_ManifestVersion_1_1_0) GetInstallerProductCodes() []string {
    var productCodes []string

    for _, installer := range ver.Installers {
      productCodes = append(productCodes, installer.ProductCode)
    }

    return productCodes
}

type API_Manifest_1_1_0 struct {
    PackageIdentifier string
    Versions []API_ManifestVersionInterface
}

// API_Manifest_1_1_0 implements all of the API_ManifestInterface interface methods
func (in API_Manifest_1_1_0) GetPackageIdentifier() string {
    return in.PackageIdentifier
}

func (in API_Manifest_1_1_0) GetVersions() []API_ManifestVersionInterface {
    return in.Versions
}

type API_Installer_1_1_0 struct {
    InstallerIdentifier string `yaml:"InstallerIdentifier"`
    InstallerLocale string `yaml:"InstallerLocale" json:",omitempty"`
    Architecture string `yaml:"Architecture"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    Platform []string `yaml:"Platform"`
    InstallerType string `yaml:"InstallerType"`
    Scope string `yaml:"Scope"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
    SignatureSha256 string `yaml:"SignatureSha256" json:",omitempty"` // winget runs into an exception internally when this is an empty string (ParseFromHexString: Invalid value size), so omit in API responses if empty
    InstallModes []string `yaml:"InstallModes"`
    InstallerSwitches InstallerSwitches_1_1_0 `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []ExpectedReturnCode_1_1_0 `yaml:"ExpectedReturnCodes"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"`
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"` 
    Dependencies Dependencies_1_1_0 `yaml:"Dependencies"`
    PackageFamilyName string `yaml:"PackageFamilyName" json:",omitempty"`
    ProductCode string `yaml:"ProductCode"`
    Capabilities []string `yaml:"Capabilities" json:",omitempty"`
    RestrictedCapabilities []string `yaml:"RestrictedCapabilities" json:",omitempty"`
    MSStoreProductIdentifier string `yaml:"MSStoreProductIdentifier" json:",omitempty"`
    Markets struct { // the manifest schema allows only one of AllowedMarkets or ExcludedMarkets per manifest but we don't verify that
        AllowedMarkets []string `yaml:"AllowedMarkets" json:",omitempty"`
        ExcludedMarkets []string `yaml:"ExcludedMarkets" json:",omitempty"`
    } `yaml:"Markets"`
    InstallerAbortsTerminal bool `yaml:"InstallerAbortsTerminal"`
    ReleaseDate string `yaml:"ReleaseDate"`
    InstallLocationRequired bool `yaml:"InstallLocationRequired"`
    RequireExplicitUpgrade bool `yaml:"RequireExplicitUpgrade"`
    UnsupportedOSArchitectures []string `yaml:"UnsupportedOSArchitectures"`
    AppsAndFeaturesEntries []struct {
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
        Publisher string `yaml:"Publisher" json:",omitempty"`
        DisplayVersion string `yaml:"DisplayVersion" json:",omitempty"`
        ProductCode string `yaml:"ProductCode" json:",omitempty"`
        UpgradeCode string `yaml:"UpgradeCode" json:",omitempty"`
        InstallerType string `yaml:"InstallerType" json:",omitempty"`
    } `yaml:"AppsAndFeaturesEntries"`
    ElevationRequirement string `yaml:"ElevationRequirement" json:",omitempty"`
}

func (in API_Installer_1_1_0) dummyFunc() bool {
    return false
}

type ExpectedReturnCode_1_1_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
}

type Agreement_1_1_0 struct {
    AgreementLabel string `yaml:"AgreementLabel"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

// API Locale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L1336
type API_Locale_1_1_0 struct {
    PackageLocale string `yaml:"PackageLocale"`
    Publisher string `yaml:"Publisher"`
    PublisherUrl string `yaml:"PublisherUrl"`
    PublisherSupportUrl string `yaml:"PublisherSupportUrl"`
    PrivacyUrl string `yaml:"PrivacyUrl"`
    Author string `yaml:"Author"`
    PackageName string `yaml:"PackageName"`
    PackageUrl string `yaml:"PackageUrl"`
    License string `yaml:"License"`
    LicenseUrl string `yaml:"LicenseUrl"`
    Copyright string `yaml:"Copyright"`
    CopyrightUrl string `yaml:"CopyrightUrl"`
    ShortDescription string `yaml:"ShortDescription"`
    Description string `yaml:"Description"`
    Tags []string `yaml:"Tags"`
    Agreements []Agreement_1_1_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
}

func (in API_Locale_1_1_0) dummyFunc() bool {
    return false
}

// API DefaultLocale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L1421
// It is the same as Locale except with an added Moniker
type API_DefaultLocale_1_1_0 struct {
    PackageLocale string `yaml:"PackageLocale"`
    Publisher string `yaml:"Publisher"`
    PublisherUrl string `yaml:"PublisherUrl"`
    PublisherSupportUrl string `yaml:"PublisherSupportUrl"`
    PrivacyUrl string `yaml:"PrivacyUrl"`
    Author string `yaml:"Author"`
    PackageName string `yaml:"PackageName"`
    PackageUrl string `yaml:"PackageUrl"`
    License string `yaml:"License"`
    LicenseUrl string `yaml:"LicenseUrl"`
    Copyright string `yaml:"Copyright"`
    CopyrightUrl string `yaml:"CopyrightUrl"`
    ShortDescription string `yaml:"ShortDescription"`
    Description string `yaml:"Description"`
    Moniker string `yaml:"Moniker"`
    Tags []string `yaml:"Tags"`
    Agreements []Agreement_1_1_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
}

func (in API_DefaultLocale_1_1_0) dummyFunc() bool {
    return false
}

type API_ManifestSearchVersion_1_1_0 struct {
    PackageVersion string
    Channel string //maxlength: 16, unused
    PackageFamilyNames []string
    ProductCodes []string
}

// https://github.com/microsoft/winget-cli/blob/56df5adb2f974230c3db8fb7f84d2fe3150eb859/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json#L229
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L969
type Dependencies_1_1_0 struct {
    WindowsFeatures []string `yaml:"WindowsFeatures" json:",omitempty"`
    WindowsLibraries []string `yaml:"WindowsLibraries" json:",omitempty"`
    PackageDependencies []struct {
        PackageIdentifier string `yaml:"PackageIdentifier"`
        MinimumVersion string `yaml:"MinimumVersion"`
    } `yaml:"PackageDependencies" json:",omitempty"`
    ExternalDependencies []string `yaml:"ExternalDependencies" json:",omitempty"`
}

// All properties of this struct are nullable strings, so set omitempty to make responses smaller
// https://github.com/microsoft/winget-cli/blob/56df5adb2f974230c3db8fb7f84d2fe3150eb859/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json#L88
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L816
type InstallerSwitches_1_1_0 struct {
    Silent string `yaml:"Silent" json:",omitempty"`
    SilentWithProgress string `yaml:"SilentWithProgress" json:",omitempty"`
    Interactive string `yaml:"Interactive" json:",omitempty"`
    InstallLocation string `yaml:"InstallLocation" json:",omitempty"`
    Log string `yaml:"Log" json:",omitempty"`
    Upgrade string `yaml:"Upgrade" json:",omitempty"`
    Custom string `yaml:"Custom" json:",omitempty"`
}

type API_ManifestSingleResponse_1_1_0 struct {
    Data *API_Manifest_1_1_0
    RequiredQueryParameters []API_QueryParameter_1_1_0
    UnsupportedQueryParameters []API_QueryParameter_1_1_0
}

type API_SearchRequestMatch_1_1_0 struct {
    KeyWord string
    MatchType API_MatchType_1_1_0
}

type API_SearchRequestPackageMatchFilter_1_1_0 struct {
    PackageMatchField API_PackageMatchField_1_1_0
    RequestMatch API_SearchRequestMatch_1_1_0
}

type API_ManifestSearchRequest_1_1_0 struct {
    MaximumResults int
    FetchAllManifests bool
    Query API_SearchRequestMatch_1_1_0
    Inclusions []API_SearchRequestPackageMatchFilter_1_1_0
    Filters []API_SearchRequestPackageMatchFilter_1_1_0
}

// "Enums":

type API_QueryParameter_1_1_0 string

const (
    QueryParameterVersion API_QueryParameter_1_1_0 = "Version"
    QueryParameterChannel API_QueryParameter_1_1_0 = "Channel"
    QueryParameterMarket API_QueryParameter_1_1_0 = "Market"
)

type API_MatchType_1_1_0 string

const (
    Exact API_MatchType_1_1_0           = "Exact"
    CaseInsensitive API_MatchType_1_1_0 = "CaseInsensitive"
    StartsWith API_MatchType_1_1_0      = "StartsWith"
    Substring API_MatchType_1_1_0       = "Substring"
    Wildcard API_MatchType_1_1_0        = "Wildcard"
    Fuzzy API_MatchType_1_1_0           = "Fuzzy"
    FuzzySubstring API_MatchType_1_1_0  = "FuzzySubstring"
)

type API_PackageMatchField_1_1_0 string

const (
    PackageIdentifier API_PackageMatchField_1_1_0 = "PackageIdentifier"
    PackageName API_PackageMatchField_1_1_0 = "PackageName"
    Moniker API_PackageMatchField_1_1_0 = "Moniker"
    Command API_PackageMatchField_1_1_0 = "Command"
    Tag API_PackageMatchField_1_1_0 = "Tag"
    PackageFamilyName API_PackageMatchField_1_1_0 = "PackageFriendlyName"
    ProductCode API_PackageMatchField_1_1_0 = "ProductCode"
    NormalizedPackageNameAndPublisher API_PackageMatchField_1_1_0 = "NormalizedPackageNameAndPublisher"
    Market API_PackageMatchField_1_1_0 = "Market"
)

