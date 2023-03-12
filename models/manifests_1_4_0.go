package models

// All of these definitions are based on the v1.4.0 manifest schema specifications:
// https://github.com/microsoft/winget-cli/tree/master/schemas/JSON/manifests/v1.4.0

// A singleton manifest can only describe one package version and contain only one locale and one installer
// Schema: https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.4.0/manifest.singleton.1.4.0.json
type Manifest_SingletonManifest_1_4_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    PackageLocale string `yaml:"PackageLocale"`
    Publisher string `yaml:"Publisher"`
    PackageName string `yaml:"PackageName"`
    License string `yaml:"License"`
    ShortDescription string `yaml:"ShortDescription"`
    Description string `yaml:"Description"`
    Moniker string `yaml:"Moniker"`
    Tags []string `yaml:"Tags"`
    PurchaseUrl string `yaml:"PurchaseUrl"`
    InstallationNotes string `yaml:"InstallationNotes"`
    Documentations []Manifest_Documentation_1_4_0 `yaml:"Documentations"`
    NestedInstallerType string `yaml:"NestedInstallerType" json:",omitempty"`
    NestedInstallerFiles []Manifest_NestedInstallerFile_1_4_0 `yaml:"NestedInstallerFiles" json:",omitempty"`
    ReleaseDate string `yaml:"ReleaseDate" json:",omitempty"`
    InstallationMetadata Manifest_InstallationMetadata_1_4_0 `yaml:"InstallationMetadata"`
    Installers [1]Manifest_Installer_1_4_0 `yaml:"Installers"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate version manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.4.0/manifest.version.1.4.0.json
type Manifest_VersionManifest_1_4_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    DefaultLocale string `yaml:"DefaultLocale"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// Implement Manifest_VersionManifestInterface
func (vm Manifest_VersionManifest_1_4_0) GetPackageVersion() string {
    return vm.PackageVersion
}

// The struct for a separate installer manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json
type Manifest_InstallerManifest_1_4_0 struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    Channel string `yaml:"Channel" json:",omitempty"`
    InstallerLocale string `yaml:"InstallerLocale" json:",omitempty"`
    Platform []string `yaml:"Platform"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    InstallerType string `yaml:"InstallerType"`
    NestedInstallerType string `yaml:"NestedInstallerType" json:",omitempty"`
    NestedInstallerFiles []Manifest_NestedInstallerFile_1_4_0 `yaml:"NestedInstallerFiles" json:",omitempty"`
    Scope string `yaml:"Scope" json:",omitempty"`
    InstallModes []string `yaml:"InstallModes" json:",omitempty"`
    InstallerSwitches Manifest_InstallerSwitches_1_4_0 `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []Manifest_ExpectedReturnCode_1_4_0 `yaml:"ExpectedReturnCodes" json:",omitempty"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"` // enum of either install or uninstallPrevious
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"`
    Dependencies Manifest_Dependencies_1_4_0 `yaml:"Dependencies" json:",omitempty"`
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
    DisplayInstallWarnings bool `yaml:"DisplayInstallWarnings"`
    UnsupportedOSArchitectures []string `yaml:"UnsupportedOSArchitectures" json:",omitempty"`
    UnsupportedArguments []string `yaml:"UnsupportedArguments"`
    AppsAndFeaturesEntries []struct {
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
        Publisher string `yaml:"Publisher" json:",omitempty"`
        DisplayVersion string `yaml:"DisplayVersion" json:",omitempty"`
        ProductCode string `yaml:"ProductCode" json:",omitempty"`
        UpgradeCode string `yaml:"UpgradeCode" json:",omitempty"`
        InstallerType string `yaml:"InstallerType" json:",omitempty"`
    } `yaml:"AppsAndFeaturesEntries" json:",omitempty"`
    ElevationRequirement string `yaml:"ElevationRequirement"`
    InstallationMetadata Manifest_InstallationMetadata_1_4_0 `yaml:"InstallationMetadata"`
    Installers []Manifest_Installer_1_4_0 `yaml:"Installers"`

    PurchaseUrl string `yaml:"PurchaseUrl"`
    InstallationNotes string `yaml:"InstallationNotes"`
    Documentations []Manifest_Documentation_1_4_0 `yaml:"Documentations"`

    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// implement Manifest_InstallerManifestInterface interface
func (instm Manifest_InstallerManifest_1_4_0) ToApiInstallers() []API_InstallerInterface {
  var apiInstallers []API_InstallerInterface

  for _, installer := range instm.Installers {
    var installer_API_ExpectedReturnCodes []API_ExpectedReturnCode_1_4_0
    for _, erc := range installer.ExpectedReturnCodes {
      installer_API_ExpectedReturnCodes = append(installer_API_ExpectedReturnCodes, API_ExpectedReturnCode_1_4_0(erc))
    }
    var instm_API_ExpectedReturnCodes []API_ExpectedReturnCode_1_4_0
    for _, erc := range instm.ExpectedReturnCodes {
      instm_API_ExpectedReturnCodes = append(instm_API_ExpectedReturnCodes, API_ExpectedReturnCode_1_4_0(erc))
    }

    var installer_API_NestedInstallerFiles []API_NestedInstallerFile_1_4_0
    for _, nif := range installer.NestedInstallerFiles {
        installer_API_NestedInstallerFiles = append(installer_API_NestedInstallerFiles, API_NestedInstallerFile_1_4_0(nif))
    }
    var instm_API_NestedInstallerFiles []API_NestedInstallerFile_1_4_0
    for _, nif := range instm.NestedInstallerFiles {
        instm_API_NestedInstallerFiles = append(instm_API_NestedInstallerFiles, API_NestedInstallerFile_1_4_0(nif))
    }

    apiInstallers = append(apiInstallers, API_Installer_1_4_0 {
      InstallerIdentifier: "", // This is in the API schema but idk where to get it from
      InstallerLocale: nonDefault(installer.InstallerLocale, instm.InstallerLocale),
      Architecture: installer.Architecture, // Already mandatory per-Installer
      MinimumOSVersion: nonDefault(installer.MinimumOSVersion, instm.MinimumOSVersion), // Already mandatory per-Installer
      Platform: nonDefault(installer.Platform, instm.Platform),
      InstallerType: nonDefault(installer.InstallerType, instm.InstallerType),
      Scope: nonDefault(installer.Scope, instm.Scope),
      InstallerUrl: installer.InstallerUrl, // Already mandatory per-Installer
      InstallerSha256: installer.InstallerSha256, // Already mandatory per-Installer
      SignatureSha256: installer.SignatureSha256, // Can only be set per-Installer, impossible to copy from global manifest properties
      InstallerSwitches: API_InstallerSwitches_1_4_0(nonDefault(installer.InstallerSwitches, instm.InstallerSwitches)), // Can be converted directly as they're identical structs
      InstallModes: nonDefault(installer.InstallModes, instm.InstallModes),
      InstallerSuccessCodes: nonDefault(installer.InstallerSuccessCodes, instm.InstallerSuccessCodes),
      ExpectedReturnCodes: nonDefault(installer_API_ExpectedReturnCodes, instm_API_ExpectedReturnCodes),
      UpgradeBehavior: nonDefault(installer.UpgradeBehavior, instm.UpgradeBehavior),
      Commands: nonDefault(installer.Commands, instm.Commands),
      Protocols: nonDefault(installer.Protocols, instm.Protocols),
      FileExtensions: nonDefault(installer.FileExtensions, instm.FileExtensions),
      Dependencies: API_Dependencies_1_4_0(nonDefault(installer.Dependencies, instm.Dependencies)),
      PackageFamilyName: nonDefault(installer.PackageFamilyName, instm.PackageFamilyName),
      ProductCode: nonDefault(installer.ProductCode, instm.ProductCode),
      Capabilities: nonDefault(installer.Capabilities, instm.Capabilities),
      RestrictedCapabilities: nonDefault(installer.RestrictedCapabilities, instm.RestrictedCapabilities),
      MSStoreProductIdentifier: "", // This is in the API schema but idk where to get it from
      Markets: nonDefault(installer.Markets, instm.Markets),
      InstallerAbortsTerminal: nonDefault(installer.InstallerAbortsTerminal, instm.InstallerAbortsTerminal),
      ReleaseDate: nonDefault(installer.ReleaseDate, instm.ReleaseDate),
      InstallLocationRequired: nonDefault(installer.InstallLocationRequired, instm.InstallLocationRequired),
      RequireExplicitUpgrade: nonDefault(installer.RequireExplicitUpgrade, instm.RequireExplicitUpgrade),
      UnsupportedOSArchitectures: nonDefault(installer.UnsupportedOSArchitectures, instm.UnsupportedOSArchitectures), // field is nullable in 1.4.0 API spec, workaround from 1.1.0 not needed
      AppsAndFeaturesEntries: nonDefault(installer.AppsAndFeaturesEntries, instm.AppsAndFeaturesEntries),
      ElevationRequirement: nonDefault(installer.ElevationRequirement, instm.ElevationRequirement),
      NestedInstallerType: nonDefault(installer.NestedInstallerType, instm.NestedInstallerType),
      NestedInstallerFiles: nonDefault(installer_API_NestedInstallerFiles, instm_API_NestedInstallerFiles),
      DisplayInstallWarnings: nonDefault(installer.DisplayInstallWarnings, instm.DisplayInstallWarnings),
      UnsupportedArguments: nonDefault(installer.UnsupportedArguments, instm.UnsupportedArguments),
      InstallationMetadata: API_InstallationMetadata_1_4_0(nonDefault(installer.InstallationMetadata, instm.InstallationMetadata)),
    })
  }

  return apiInstallers
}

type Manifest_Installer_1_4_0 struct {
    InstallerLocale string `yaml:"InstallerLocale" json:",omitempty"`
    Architecture string `yaml:"Architecture"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    Platform []string `yaml:"Platform"`
    InstallerType string `yaml:"InstallerType"`
    NestedInstallerType string `yaml:"NestedInstallerType" json:",omitempty"`
    NestedInstallerFiles []Manifest_NestedInstallerFile_1_4_0 `yaml:"NestedInstallerFiles" json:",omitempty"`
    Scope string `yaml:"Scope"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
    SignatureSha256 string `yaml:"SignatureSha256" json:",omitempty"` // winget runs into an exception internally when this is an empty string (ParseFromHexString: Invalid value size), so omit in API responses if empty
    InstallModes []string `yaml:"InstallModes"`
    InstallerSwitches Manifest_InstallerSwitches_1_4_0 `yaml:"InstallerSwitches"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes" json:",omitempty"`
    ExpectedReturnCodes []Manifest_ExpectedReturnCode_1_4_0 `yaml:"ExpectedReturnCodes"`
    UpgradeBehavior string `yaml:"UpgradeBehavior" json:",omitempty"`
    Commands []string `yaml:"Commands" json:",omitempty"`
    Protocols []string `yaml:"Protocols" json:",omitempty"`
    FileExtensions []string `yaml:"FileExtensions" json:",omitempty"` 
    Dependencies Manifest_Dependencies_1_4_0 `yaml:"Dependencies"`
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
    DisplayInstallWarnings bool `yaml:"DisplayInstallWarnings"`
    UnsupportedOSArchitectures []string `yaml:"UnsupportedOSArchitectures"`
    UnsupportedArguments []string `yaml:"UnsupportedArguments"`
    AppsAndFeaturesEntries []struct {
        DisplayName string `yaml:"DisplayName" json:",omitempty"`
        Publisher string `yaml:"Publisher" json:",omitempty"`
        DisplayVersion string `yaml:"DisplayVersion" json:",omitempty"`
        ProductCode string `yaml:"ProductCode" json:",omitempty"`
        UpgradeCode string `yaml:"UpgradeCode" json:",omitempty"`
        InstallerType string `yaml:"InstallerType" json:",omitempty"`
    } `yaml:"AppsAndFeaturesEntries"`
    ElevationRequirement string `yaml:"ElevationRequirement" json:",omitempty"`
    InstallationMetadata Manifest_InstallationMetadata_1_4_0 `yaml:"InstallationMetadata"`
}

func (mi Manifest_Installer_1_4_0) ToApiInstaller() API_Installer_1_4_0 {
  var installer_API_ExpectedReturnCodes []API_ExpectedReturnCode_1_4_0
  for _, erc := range mi.ExpectedReturnCodes {
    installer_API_ExpectedReturnCodes = append(installer_API_ExpectedReturnCodes, API_ExpectedReturnCode_1_4_0(erc))
  }

  var installer_API_NestedInstallerFiles []API_NestedInstallerFile_1_4_0
  for _, nif := range mi.NestedInstallerFiles {
    installer_API_NestedInstallerFiles = append(installer_API_NestedInstallerFiles, API_NestedInstallerFile_1_4_0(nif))
  }

  return API_Installer_1_4_0 {
    Architecture: mi.Architecture,
    MinimumOSVersion: mi.MinimumOSVersion,
    Platform: mi.Platform,
    InstallerType: mi.InstallerType,
    Scope: mi.Scope,
    InstallerUrl: mi.InstallerUrl,
    InstallerSha256: mi.InstallerSha256,
    SignatureSha256: mi.SignatureSha256,
    InstallModes: mi.InstallModes,
    InstallerSwitches: API_InstallerSwitches_1_4_0(mi.InstallerSwitches), // Can be converted directly as they're identical structs
    InstallerSuccessCodes: mi.InstallerSuccessCodes,
    ExpectedReturnCodes: installer_API_ExpectedReturnCodes,
    UpgradeBehavior: mi.UpgradeBehavior,
    Commands: mi.Commands,
    Protocols: mi.Protocols,
    FileExtensions: mi.FileExtensions,
    Dependencies: API_Dependencies_1_4_0(mi.Dependencies),
    PackageFamilyName: mi.PackageFamilyName,
    ProductCode: mi.ProductCode,
    Capabilities: mi.Capabilities,
    RestrictedCapabilities: mi.RestrictedCapabilities,
    MSStoreProductIdentifier: "", // This is in the API schema but idk where to get it from
    Markets: mi.Markets,
    InstallerAbortsTerminal: mi.InstallerAbortsTerminal,
    ReleaseDate: mi.ReleaseDate,
    InstallLocationRequired: mi.InstallLocationRequired,
    RequireExplicitUpgrade: mi.RequireExplicitUpgrade,
    UnsupportedOSArchitectures: mi.UnsupportedOSArchitectures,
    AppsAndFeaturesEntries: mi.AppsAndFeaturesEntries,
    ElevationRequirement: mi.ElevationRequirement,
    NestedInstallerType: mi.NestedInstallerType,
    NestedInstallerFiles: installer_API_NestedInstallerFiles,
    DisplayInstallWarnings: mi.DisplayInstallWarnings,
    UnsupportedArguments: mi.UnsupportedArguments,
    InstallationMetadata: API_InstallationMetadata_1_4_0(mi.InstallationMetadata),
  }
}

// The struct for a separate locale manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.locale.1.1.0.json
type Manifest_LocaleManifest_1_4_0 struct {
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
    Agreements []Manifest_Agreement_1_4_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
}

func (locm Manifest_LocaleManifest_1_4_0) ToApiLocale() API_LocaleInterface {
  var apiAgreements []API_Agreement_1_4_0
  for _, ma := range locm.Agreements {
    apiAgreements = append(apiAgreements, API_Agreement_1_4_0(ma))
  }

  return API_Locale_1_4_0{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Tags: locm.Tags,
    Agreements: apiAgreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
}

// The struct for a separate defaultlocale manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.locale.1.1.0.json
// It is the same as Locale except with an added Moniker
type Manifest_DefaultLocaleManifest_1_4_0 struct {
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
    Agreements []Manifest_Agreement_1_4_0 `yaml:"Agreements"`
    ReleaseNotes string `yaml:"ReleaseNotes"`
    ReleaseNotesUrl string `yaml:"ReleaseNotesUrl"`
}

func (locm Manifest_DefaultLocaleManifest_1_4_0) ToApiDefaultLocale() API_DefaultLocaleInterface {
  var apiAgreements []API_Agreement_1_4_0
  for _, ma := range locm.Agreements {
    apiAgreements = append(apiAgreements, API_Agreement_1_4_0(ma))
  }

  return API_DefaultLocale_1_4_0{
    PackageLocale: locm.PackageLocale,
    Publisher: locm.Publisher,
    PublisherUrl: locm.PublisherUrl,
    PublisherSupportUrl: locm.PublisherSupportUrl,
    PrivacyUrl: locm.PrivacyUrl,
    Author: locm.Author,
    PackageName: locm.PackageName,
    PackageUrl: locm.PackageUrl,
    License: locm.License,
    LicenseUrl: locm.LicenseUrl,
    Copyright: locm.Copyright,
    CopyrightUrl: locm.CopyrightUrl,
    ShortDescription: locm.ShortDescription,
    Description: locm.Description,
    Moniker: locm.Moniker,
    Tags: locm.Tags,
    Agreements: apiAgreements,
    ReleaseNotes: locm.ReleaseNotes,
    ReleaseNotesUrl: locm.ReleaseNotesUrl,
  }
}

type Manifest_Agreement_1_4_0 struct {
    AgreementLabel string `yaml:"AgreementLabel"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

type Manifest_InstallerSwitches_1_4_0 struct {
    Silent string `yaml:"Silent" json:",omitempty"`
    SilentWithProgress string `yaml:"SilentWithProgress" json:",omitempty"`
    Interactive string `yaml:"Interactive" json:",omitempty"`
    InstallLocation string `yaml:"InstallLocation" json:",omitempty"`
    Log string `yaml:"Log" json:",omitempty"`
    Upgrade string `yaml:"Upgrade" json:",omitempty"`
    Custom string `yaml:"Custom" json:",omitempty"`
}

type Manifest_ExpectedReturnCode_1_4_0 struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse string `yaml:"ReturnResponse"`
    ReturnResponseUrl string `yaml:"ReturnResponseUrl"`
}

// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.2.0/manifest.installer.1.2.0.json#L251
type Manifest_Dependencies_1_4_0 struct {
    WindowsFeatures []string `yaml:"WindowsFeatures" json:",omitempty"`
    WindowsLibraries []string `yaml:"WindowsLibraries" json:",omitempty"`
    PackageDependencies []struct {
        PackageIdentifier string `yaml:"PackageIdentifier"`
        MinimumVersion string `yaml:"MinimumVersion"`
    } `yaml:"PackageDependencies" json:",omitempty"`
    ExternalDependencies []string `yaml:"ExternalDependencies" json:",omitempty"`
}

type Manifest_Documentation_1_4_0 struct {
    DocumentLabel string `yaml:"DocumentLabel"`
    DocumentUrl string `yaml:"DocumentUrl"`
}

type Manifest_InstallationMetadata_1_4_0 struct {
    DefaultInstallLocation string `yaml:"DefaultInstallLocation"` // nullable
    Files []struct {
        RelativeFilePath string `yaml:"RelativeFilePath"`
        FileSha256 string `yaml:"FileSha256" json:",omitempty"`
        FileType string `yaml:"FileType" json:",omitempty"`
        InvocationParameter string `yaml:"InvocationParameter" json:",omitempty"`
        DisplayName  string `yaml:"DisplayName" json:",omitempty"`
    } `yaml:"Files" json:",omitempty"`
}

type Manifest_NestedInstallerFile_1_4_0 struct {
    RelativeFilePath string `yaml:"RelativeFilePath"`
    PortableCommandAlias string `yaml:"PortableCommandAlias" json:",omitempty"`
}

