package main

// All of these definitions are based on the v1.1.0 manifest schema specifications:
// https://github.com/microsoft/winget-cli/tree/master/schemas/JSON/manifests/v1.1.0

// These properties are required in all manifest types:
// singleton, version, locale and installer
type BaseManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// A singleton manifest can only contain one locale and one installer
// Schema: https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.singleton.1.1.0.json
type SingletonManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    PackageLocale string `yaml:"PackageLocale"`
    Publisher string `yaml:"Publisher"`
    PackageName string `yaml:"PackageName"`
    License string `yaml:"License"`
    ShortDescription string `yaml:"ShortDescription"`
    Installers [1]Installer `yaml:"Installers"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate version manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.version.1.1.0.json
type VersionManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    DefaultLocale string `yaml:"DefaultLocale"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// The struct for a separate installer manifest file
// https://github.com/microsoft/winget-cli/blob/master/schemas/JSON/manifests/v1.1.0/manifest.installer.1.1.0.json
type InstallerManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    Platform []string `yaml:"Platform"`
    ReleaseDate string `yaml:"ReleaseDate"`
    ElevationRequirement string `yaml:"ElevationRequirement"`
    Installers []Installer `yaml:"Installers"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

