package models

// All of these definitions are based on the v1.1.0 manifest schema specifications:
// https://github.com/microsoft/winget-cli/tree/master/schemas/JSON/manifests/v1.1.0

// A singleton manifest can only describe one package version and contain only one locale and one installer
// Schema: https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.singleton.1.1.0.json
type Manifest_SingletonManifest_1_1_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    PackageLocale string `yaml:"PackageLocale"`
    Publisher string `yaml:"Publisher"`
    PackageName string `yaml:"PackageName"`
    License string `yaml:"License"`
    ShortDescription string `yaml:"ShortDescription"`
    Installers [1]Manifest_Installer_1_1_0 `yaml:"Installers"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate version manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.version.1.1.0.json
type Manifest_VersionManifest_1_1_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    DefaultLocale string `yaml:"DefaultLocale"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate installer manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json
type Manifest_InstallerManifest_1_1_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    Channel string `yaml:"Channel" json:",omitempty"`
    InstallerLocale string `yaml:"InstallerLocale" json:",omitempty"`
    Platform []string `yaml:"Platform"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    InstallerType string `yaml:"InstallerType"`
    Scope Scope `yaml:"Scope" json:",omitempty"`
    InstallModes []InstallMode `yaml:"InstallModes" json:",omitempty"`
    InstallerSwitches InstallerSwitches `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []ExpectedReturnCode_1_1_0 `yaml:"ExpectedReturnCodes" json:",omitempty"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"` // enum of either install or uninstallPrevious
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"`
    Dependencies Dependencies_1_1_0 `yaml:"Dependencies" json:",omitempty"`
    PackageFamilyName string `yaml:"PackageFamilyName" json:",omitempty"`
    ProductCode string `yaml:"ProductCode" json:",omitempty"`
    Capabilities []string `yaml:"Capabilities" json:",omitempty"`
    RestrictedCapabilities []string `yaml:"RestrictedCapabilities" json:",omitempty"`
    Markets struct { // the manifest schema allows only one of AllowedMarkets or ExcludedMarkets per manifest but we don't verify that
        AllowedMarkets []string `yaml:"AllowedMarkets" json:",omitempty"`
        ExcludedMarkets []string `yaml:"ExcludedMarkets" json:",omitempty"`
    } `yaml:"Markets"`
    InstallerAbortsTerminal bool `yaml:"InstallerAbortsTerminal" json:",omitempty"`
    ReleaseDate string `yaml:"ReleaseDate" json:",omitempty"`
    InstallLocationRequired bool `yaml:"InstallLocationRequired" json:",omitempty"`
    RequireExplicitUpgrade bool `yaml:"RequireExplicitUpgrade" json:",omitempty"`
    UnsupportedOSArchitectures []string `yaml:"UnsupportedOSArchitectures" json:",omitempty"`
    AppsAndFeaturesEntries []struct {
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
        Publisher string `yaml:"Publisher" json:",omitempty"`
        DisplayVersion string `yaml:"DisplayVersion" json:",omitempty"`
        ProductCode string `yaml:"ProductCode" json:",omitempty"`
        UpgradeCode string `yaml:"UpgradeCode" json:",omitempty"`
        InstallerType string `yaml:"InstallerType" json:",omitempty"`
    } `yaml:"AppsAndFeaturesEntries" json:",omitempty"`
    ElevationRequirement string `yaml:"ElevationRequirement"`
    Installers []Manifest_Installer_1_1_0 `yaml:"Installers"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

type Manifest_Installer_1_1_0 struct {
    InstallerLocale string `yaml:"InstallerLocale" json:",omitempty"`
    Architecture Architecture `yaml:"Architecture"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    Platform []string `yaml:"Platform"`
    InstallerType string `yaml:"InstallerType"`
    Scope Scope `yaml:"Scope"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
    SignatureSha256 string `yaml:"SignatureSha256" json:",omitempty"` // winget runs into an exception internally when this is an empty string (ParseFromHexString: Invalid value size), so omit in API responses if empty
    InstallModes []InstallMode `yaml:"InstallModes"`
    InstallerSwitches InstallerSwitches `yaml:"InstallerSwitches"`
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

// The struct for a separate locale manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.locale.1.1.0.json
type Manifest_LocaleManifest_1_1_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
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

// The struct for a separate defaultlocale manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.locale.1.1.0.json
// It is the same as Locale except with an added Moniker
type Manifest_DefaultLocaleManifest_1_1_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
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

