package models

// These are the criteria on which multi-file manifests
// are associated together to form one package. If these
// three match, the files belong to / form one package.
type MultiFileManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

func (multifilemanifest MultiFileManifest) ToBaseManifest(manifesttype string) BaseManifest {
    return BaseManifest{
        PackageIdentifier: multifilemanifest.PackageIdentifier,
        PackageVersion: multifilemanifest.PackageVersion,
        ManifestType: manifesttype,
        ManifestVersion: multifilemanifest.ManifestVersion,
    }
}

// These properties are required in all manifest types:
// singleton, version, locale and installer
type BaseManifest struct {
    PackageIdentifier string `yaml:"PackageIdentifier"`
    PackageVersion string `yaml:"PackageVersion"`
    ManifestType string `yaml:"ManifestType"`
    ManifestVersion string `yaml:"ManifestVersion"`
}

// This function reduces a BaseManifest to the data fields that will be
// identical for all manifest files belonging to the same package.
// This allows multi-manifest files to be correlated to a single package.
func (basemanifest BaseManifest) ToMultiFileManifest() MultiFileManifest {
    return MultiFileManifest{
        PackageIdentifier: basemanifest.PackageIdentifier,
        PackageVersion: basemanifest.PackageVersion,
        ManifestVersion: basemanifest.ManifestVersion,
    }
}

type Manifest_VersionManifestInterface interface {
    GetPackageVersion() string
}

type Manifest_InstallerManifestInterface interface {
    ToApiInstallers() []API_InstallerInterface
}

type Manifest_InstallerInterface interface {
    ToApiInstaller() API_InstallerInterface
}

type Manifest_LocaleManifestInterface interface {
    ToApiLocale() API_LocaleInterface
}

type Manifest_DefaultLocaleManifestInterface interface {
    ToApiDefaultLocale() API_DefaultLocaleInterface
}
