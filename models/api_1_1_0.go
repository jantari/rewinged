package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

type ManifestVersion_1_1_0 struct {
    PackageVersion string
    DefaultLocale DefaultLocale_1_1_0
    Channel string
    Locales []Locale_1_1_0
    Installers []Installer_1_1_0
}

// ManifestVersion_1_1_0 implements all of the ManifestVersionInterface interface methods:

func (ver ManifestVersion_1_1_0) GetDefaultLocalePackageName() string {
    return ver.DefaultLocale.PackageName
}

func (ver ManifestVersion_1_1_0) GetDefaultLocalePublisher() string {
    return ver.DefaultLocale.Publisher
}

func (ver ManifestVersion_1_1_0) GetDefaultLocaleShortDescription() string {
    return ver.DefaultLocale.ShortDescription
}

func (ver ManifestVersion_1_1_0) GetPackageVersion() string {
    return ver.PackageVersion
}

func (ver ManifestVersion_1_1_0) GetInstallerProductCodes() []string {
    var productCodes []string

    for _, installer := range ver.Installers {
      productCodes = append(productCodes, installer.ProductCode)
    }

    return productCodes
}

// API Locale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L1336
type Locale_1_1_0 struct {
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

// API DefaultLocale schema
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml#L1421
// It is the same as Locale except with an added Moniker
type DefaultLocale_1_1_0 struct {
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

type ManifestSearchVersion_1_1_0 struct {
    PackageVersion string
    Channel string //maxlength: 16, unused
    PackageFamilyNames []string
    ProductCodes []string
}

