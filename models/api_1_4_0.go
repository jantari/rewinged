package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml

type API_Installer_1_4_0 struct {
    InstallerIdentifier string `yaml:"InstallerIdentifier"`
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
    InstallerSwitches InstallerSwitches_1_4_0 `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []ExpectedReturnCode_1_4_0 `yaml:"ExpectedReturnCodes"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"`
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"` 
    Dependencies Dependencies_1_4_0 `yaml:"Dependencies"`
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
    NestedInstallerFiles []struct {
        RelativeFilePath string `yaml:"RelativeFilePath"`
        PortableCommandAlias string `yaml:"PortableCommandAlias" json:",omitempty"`
    } `yaml:"NestedInstallerFiles" json:",omitempty"`
    DisplayInstallWarnings bool `yaml:"DisplayInstallWarnings" json:",omitempty"`
    UnsupportedArguments []string `yaml:"UnsupportedArguments" json:",omitempty"`
    InstallationMetadata struct {
        DefaultInstallLocation string `yaml:"DefaultInstallLocation" json:",omitempty"`
        Files []struct {
            RelativeFilePath string `yaml:"RelativeFilePath"`
            FileSha256 string `yaml:"FileSha256" json:",omitempty"`
            FileType string `yaml:"FileType" json:",omitempty"`
            InvocationParameter string `yaml:"InvocationParameter" json:",omitempty"`
            DisplayName string `yaml:"DisplayName" json:",omitempty"`
        } `yaml:"Files" json:",omitempty"`
    } `yaml:"InstallationMetadata"`
}

// API Locale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml
type Locale_1_4_0 struct {
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
    Agreements []Agreement_1_4_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
    PurchaseUrl string `yaml:"PurchaseUrl"`
    InstallationNotes string `yaml:"InstallationNotes"`
    Documentations []Documentation_1_4_0 `yaml:"Documentations"`
}

// Only exists in 1.4.0+, not in 1.1.0
type Documentation_1_4_0 struct {
    DocumentLabel string `yaml:"DocumentLabel"`
    DocumentUrl string `yaml:"DocumentUrl"`
}

type ManifestSearchVersion_1_4_0 struct {
    PackageVersion string
    Channel string //maxlength: 16, unused
    PackageFamilyNames []string
    ProductCodes []string
    AppsAndFeaturesEntryVersions string
    UpgradeCodes []string
}

type ExpectedReturnCode_1_4_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
    ReturnResponseUrl string `yaml:"ReturnResponseUrl"`
}

type Agreement_1_4_0 struct {
    AgreementLabel string `yaml:"AgreementLabel" json:",omitempty"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.4.0/manifest.installer.1.4.0.json#L294
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml#L1015
type Dependencies_1_4_0 struct {
    WindowsFeatures []string `yaml:"WindowsFeatures" json:",omitempty"`
    WindowsLibraries []string `yaml:"WindowsLibraries" json:",omitempty"`
    PackageDependencies []struct {
        PackageIdentifier string `yaml:"PackageIdentifier"`
        MinimumVersion string `yaml:"MinimumVersion"`
    } `yaml:"PackageDependencies" json:",omitempty"`
    ExternalDependencies []string `yaml:"ExternalDependencies" json:",omitempty"`
}

// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml#L856
type InstallerSwitches_1_4_0 struct {
    Silent string `yaml:"Silent" json:",omitempty"`
    SilentWithProgress string `yaml:"SilentWithProgress" json:",omitempty"`
    Interactive string `yaml:"Interactive" json:",omitempty"`
    InstallLocation string `yaml:"InstallLocation" json:",omitempty"`
    Log string `yaml:"Log" json:",omitempty"`
    Upgrade string `yaml:"Upgrade" json:",omitempty"`
    Custom string `yaml:"Custom" json:",omitempty"`
}

