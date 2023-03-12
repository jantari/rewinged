package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml

type API_Information_1_4_0 struct {
    Data struct {
        SourceIdentifier        string
        ServerSupportedVersions []string
    }
}

// API_ManifestVersion_1_4_0 implements all of the API_ManifestVersionInterface interface methods
type API_ManifestVersion_1_4_0 struct {
    PackageVersion string
    DefaultLocale API_DefaultLocale_1_4_0
    Channel string
    Locales []API_Locale_1_4_0
    Installers []API_Installer_1_4_0
}

func (ver API_ManifestVersion_1_4_0) GetDefaultLocalePackageName() string {
    return ver.DefaultLocale.PackageName
}

func (ver API_ManifestVersion_1_4_0) GetDefaultLocalePublisher() string {
    return ver.DefaultLocale.Publisher
}

func (ver API_ManifestVersion_1_4_0) GetDefaultLocaleShortDescription() string {
    return ver.DefaultLocale.ShortDescription
}

func (ver API_ManifestVersion_1_4_0) GetPackageVersion() string {
    return ver.PackageVersion
}

func (ver API_ManifestVersion_1_4_0) GetInstallerProductCodes() []string {
    var productCodes []string

    for _, installer := range ver.Installers {
      productCodes = append(productCodes, installer.ProductCode)
    }

    return productCodes
}

type API_Manifest_1_4_0 struct {
    PackageIdentifier string
    Versions []API_ManifestVersionInterface
}

// API_Manifest_1_4_0 implements all of the API_ManifestInterface interface methods
func (in API_Manifest_1_4_0) GetPackageIdentifier() string {
    return in.PackageIdentifier
}

func (in API_Manifest_1_4_0) GetVersions() []API_ManifestVersionInterface {
    return in.Versions
}

type API_Installer_1_4_0 struct {
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
    InstallerSwitches API_InstallerSwitches_1_4_0 `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []API_ExpectedReturnCode_1_4_0 `yaml:"ExpectedReturnCodes"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"`
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"` 
    Dependencies API_Dependencies_1_4_0 `yaml:"Dependencies"`
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
    UnsupportedOSArchitectures []string `yaml:"UnsupportedOSArchitectures" json:",omitempty"`
    AppsAndFeaturesEntries []struct {
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
        Publisher string `yaml:"Publisher" json:",omitempty"`
        DisplayVersion string `yaml:"DisplayVersion" json:",omitempty"`
        ProductCode string `yaml:"ProductCode" json:",omitempty"`
        UpgradeCode string `yaml:"UpgradeCode" json:",omitempty"`
        InstallerType string `yaml:"InstallerType" json:",omitempty"`
    } `yaml:"AppsAndFeaturesEntries" json:",omitempty"`
    ElevationRequirement string `yaml:"ElevationRequirement" json:",omitempty"`
    NestedInstallerType string `yaml:"NestedInstallerType" json:",omitempty"`
    NestedInstallerFiles []API_NestedInstallerFile_1_4_0 `yaml:"NestedInstallerFiles" json:",omitempty"`
    DisplayInstallWarnings bool `yaml:"DisplayInstallWarnings" json:",omitempty"`
    UnsupportedArguments []string `yaml:"UnsupportedArguments" json:",omitempty"`
    InstallationMetadata API_InstallationMetadata_1_4_0 `yaml:"InstallationMetadata"`
}

func (in API_Installer_1_4_0) dummyFunc() bool {
    return false
}

// API Locale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml
type API_Locale_1_4_0 struct {
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
    Agreements []API_Agreement_1_4_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
    PurchaseUrl string `yaml:"PurchaseUrl"`
    InstallationNotes string `yaml:"InstallationNotes"`
    Documentations []Documentation_1_4_0 `yaml:"Documentations"`
}

func (in API_Locale_1_4_0) dummyFunc() bool {
    return false
}

// API DefaultLocale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L1421
// It is the same as Locale except with an added Moniker
type API_DefaultLocale_1_4_0 struct {
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
    Agreements []API_Agreement_1_4_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
    PurchaseUrl string `yaml:"PurchaseUrl"`
    InstallationNotes string `yaml:"InstallationNotes"`
    Documentations []Documentation_1_4_0 `yaml:"Documentations"`
}

func (in API_DefaultLocale_1_4_0) dummyFunc() bool {
    return false
}

type API_ManifestSearchVersion_1_4_0 struct {
    PackageVersion string
    Channel string //maxlength: 16, unused
    PackageFamilyNames []string
    ProductCodes []string
    AppsAndFeaturesEntryVersions string
    UpgradeCodes []string
}

// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.4.0/manifest.installer.1.4.0.json#L294
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml#L1015
type API_Dependencies_1_4_0 struct {
    WindowsFeatures []string `yaml:"WindowsFeatures" json:",omitempty"`
    WindowsLibraries []string `yaml:"WindowsLibraries" json:",omitempty"`
    PackageDependencies []struct {
        PackageIdentifier string `yaml:"PackageIdentifier"`
        MinimumVersion string `yaml:"MinimumVersion"`
    } `yaml:"PackageDependencies" json:",omitempty"`
    ExternalDependencies []string `yaml:"ExternalDependencies" json:",omitempty"`
}

// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml#L856
type API_InstallerSwitches_1_4_0 struct {
    Silent string `yaml:"Silent" json:",omitempty"`
    SilentWithProgress string `yaml:"SilentWithProgress" json:",omitempty"`
    Interactive string `yaml:"Interactive" json:",omitempty"`
    InstallLocation string `yaml:"InstallLocation" json:",omitempty"`
    Log string `yaml:"Log" json:",omitempty"`
    Upgrade string `yaml:"Upgrade" json:",omitempty"`
    Custom string `yaml:"Custom" json:",omitempty"`
}

type API_ExpectedReturnCode_1_4_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
    ReturnResponseUrl string `yaml:"ReturnResponseUrl"`
}

type API_Agreement_1_4_0 struct {
    AgreementLabel string `yaml:"AgreementLabel" json:",omitempty"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

type API_ManifestSingleResponse_1_4_0 struct {
    Data *API_Manifest_1_4_0
    RequiredQueryParameters []string
    UnsupportedQueryParameters []string
}

type API_ManifestSearchRequest_1_4_0 struct {
    MaximumResults int
    FetchAllManifests bool
    Query API_SearchRequestMatch_1_4_0
    Inclusions []API_SearchRequestPackageMatchFilter_1_4_0
    Filters []API_SearchRequestPackageMatchFilter_1_4_0
}

type API_SearchRequestPackageMatchFilter_1_4_0 struct {
    PackageMatchField string
    RequestMatch API_SearchRequestMatch_1_4_0
}

type API_SearchRequestMatch_1_4_0 struct {
    KeyWord string
    MatchType string
}

// Only exists in 1.4.0+, not in 1.1.0
type Documentation_1_4_0 struct {
    DocumentLabel string `yaml:"DocumentLabel"`
    DocumentUrl string `yaml:"DocumentUrl"`
}

type API_InstallationMetadata_1_4_0 struct {
    DefaultInstallLocation string `yaml:"DefaultInstallLocation" json:",omitempty"`
    Files []struct {
        RelativeFilePath string `yaml:"RelativeFilePath"`
        FileSha256 string `yaml:"FileSha256" json:",omitempty"`
        FileType string `yaml:"FileType" json:",omitempty"`
        InvocationParameter string `yaml:"InvocationParameter" json:",omitempty"`
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
    } `yaml:"Files" json:",omitempty"`
}

type API_NestedInstallerFile_1_4_0 struct {
    RelativeFilePath string `yaml:"RelativeFilePath"`
    PortableCommandAlias string `yaml:"PortableCommandAlias" json:",omitempty"`
}
