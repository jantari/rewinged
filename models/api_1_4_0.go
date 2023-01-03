package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.4.0.yaml


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
