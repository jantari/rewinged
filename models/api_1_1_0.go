package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

type ManifestVersion_1_1_0 struct {
    PackageVersion string
    DefaultLocale DefaultLocale
    Channel string
    Locales []Locale
    Installers []Installer
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

