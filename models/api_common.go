package models

// All of these definitions are based on the v1.1.0 API specification:
// https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

type API_WingetApiError struct {
    ErrorCode    int
    ErrorMessage string
}

type API_Package struct {
    PackageIdentifier string
}

type API_PackageMultipleResponse struct {
    Data []API_Package
    ContinuationToken string
}

type API_ManifestInterface interface {
    GetPackageIdentifier() string
    GetVersions() []API_ManifestVersionInterface
}

type API_ManifestVersionInterface interface {
    GetDefaultLocalePackageName() string
    GetDefaultLocalePublisher() string
    GetDefaultLocaleShortDescription() string
    GetPackageVersion() string
    GetDefaultLocale() API_DefaultLocaleInterface
    GetLocales() []API_LocaleInterface
    GetInstallers() []API_InstallerInterface
    GetInstallerProductCodes() []string
}

type API_InstallerInterface interface {
    GetInstallerSha() string
    GetInstallerUrl() string
    SetInstallerUrl(newUrl string)
    dummyFunc() bool
}

type API_InstallerWithAuthInterface interface {
    API_InstallerInterface
    SetInstallerAuthentication(auth *API_Authentication_1_7_0) // temporary, needs to be an interface
}

type API_LocaleInterface interface {
    dummyFunc() bool
}

type API_DefaultLocaleInterface interface {
    dummyFunc() bool
}

// TODO: Decide whether generic union or implementing
// a non-empty interface is best for the below types

type API_ManifestSearchVersionInterface interface {
    API_ManifestSearchVersion_1_1_0 | API_ManifestSearchVersion_1_4_0
}

type API_ManifestSearchResponse[MSVI API_ManifestSearchVersionInterface] struct {
    PackageIdentifier string
    PackageName string
    Publisher string
    Versions []MSVI
}

type API_ManifestSearchResult[MSVI API_ManifestSearchVersionInterface] struct {
    Data []API_ManifestSearchResponse[MSVI]
    RequiredPackageMatchFields []string
    UnsupportedPackageMatchFields []string
}

