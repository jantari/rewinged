package models

// These types / data structures are used the same by the manifests and the API.
// This means no conversion is necessary when moving this data between the two.

type Installer_1_1_0 struct {
    // InstallerIdentifier - only API
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
    Dependencies Dependencies `yaml:"Dependencies"`
    PackageFamilyName string `yaml:"PackageFamilyName" json:",omitempty"`
    ProductCode string `yaml:"ProductCode"`
    Capabilities []string `yaml:"Capabilities" json:",omitempty"`
    RestrictedCapabilities []string `yaml:"RestrictedCapabilities" json:",omitempty"`
    // MSStoreProductIdentifier - only API
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

type Installer_1_4_0 struct {
    // InstallerIdentifier - only API
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
    ExpectedReturnCodes []ExpectedReturnCode_1_4_0 `yaml:"ExpectedReturnCodes"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"`
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"` 
    Dependencies Dependencies `yaml:"Dependencies"`
    PackageFamilyName string `yaml:"PackageFamilyName" json:",omitempty"`
    ProductCode string `yaml:"ProductCode"`
    Capabilities []string `yaml:"Capabilities" json:",omitempty"`
    RestrictedCapabilities []string `yaml:"RestrictedCapabilities" json:",omitempty"`
    // MSStoreProductIdentifier - only API
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

type ExpectedReturnCode_1_1_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
}

type ExpectedReturnCode_1_4_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
    ReturnResponseUrl string `yaml:"ReturnResponseUrl"`
}

type Agreement_1_1_0 struct {
    AgreementLabel string `yaml:"AgreementLabel"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

type Agreement_1_4_0 struct {
    AgreementLabel string `yaml:"AgreementLabel" json:",omitempty"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

// All properties of this struct are nullable strings, so set omitempty to make responses smaller
// https://github.com/microsoft/winget-cli/blob/56df5adb2f974230c3db8fb7f84d2fe3150eb859/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json#L88
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L816
type InstallerSwitches struct {
    Silent string `yaml:"Silent" json:",omitempty"`
    SilentWithProgress string `yaml:"SilentWithProgress" json:",omitempty"`
    Interactive string `yaml:"Interactive" json:",omitempty"`
    InstallLocation string `yaml:"InstallLocation" json:",omitempty"`
    Log string `yaml:"Log" json:",omitempty"`
    Upgrade string `yaml:"Upgrade" json:",omitempty"`
    Custom string `yaml:"Custom" json:",omitempty"`
}

// https://github.com/microsoft/winget-cli/blob/56df5adb2f974230c3db8fb7f84d2fe3150eb859/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json#L229
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L969
type Dependencies struct {
    WindowsFeatures []string `yaml:"WindowsFeatures" json:",omitempty"`
    WindowsLibraries []string `yaml:"WindowsLibraries" json:",omitempty"`
    PackageDependencies []struct {
        PackageIdentifier string `yaml:"PackageIdentifier"`
        MinimumVersion string `yaml:"MinimumVersion"`
    } `yaml:"PackageDependencies" json:",omitempty"`
    ExternalDependencies []string `yaml:"ExternalDependencies" json:",omitempty"`
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

