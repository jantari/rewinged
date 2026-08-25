package main

// Tests live beside the code they cover, as `<file>_test.go` in the same
// package, and run in CI on every push and pull request via test.yml. A change
// that fixes a defect is expected to carry a test that fails without the fix -
// several parsing paths here run only on malformed or unreachable input, which
// is where a regression goes unnoticed longest.

import (
  "os"
  "path/filepath"
  "testing"

  "rewinged/logging"
  "rewinged/models"
)

// The package logs through a shared zerolog logger that main() initialises, so
// tests have to stand it up before touching any code that logs.
func TestMain(m *testing.M) {
  logging.InitLogger("error", true)
  os.Exit(m.Run())
}

// manifestNodesFromYAML writes each document to a file and reads it back the
// way ingestion does, so a test exercises the same decoding path as production
// rather than a hand-built node the real code would never see.
func manifestNodesFromYAML(t *testing.T, docs ...string) []models.ManifestNode {
  t.Helper()

  dir := t.TempDir()
  nodes := []models.ManifestNode{}

  for i, doc := range docs {
    path := filepath.Join(dir, string(rune('a'+i))+".yaml")
    if err := os.WriteFile(path, []byte(doc), 0o600); err != nil {
      t.Fatalf("writing manifest %d: %v", i, err)
    }

    parsed, err := parseFileAsBaseManifests(path)
    if err != nil {
      t.Fatalf("parsing manifest %d: %v", i, err)
    }
    for _, node := range parsed {
      nodes = append(nodes, *node)
    }
  }

  return nodes
}

const versionManifest = `
PackageIdentifier: Contoso.Probe
PackageVersion: 1.0.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
`

const defaultLocaleManifest = `
PackageIdentifier: Contoso.Probe
PackageVersion: 1.0.0
PackageLocale: en-US
Publisher: Contoso
PackageName: Probe Tool
License: Proprietary
ShortDescription: A package used to exercise the manifest parser.
Tags:
  - probe
ManifestType: defaultLocale
ManifestVersion: 1.6.0
`

const installerManifest = `
PackageIdentifier: Contoso.Probe
PackageVersion: 1.0.0
Installers:
  - Architecture: x64
    InstallerType: exe
    InstallerUrl: https://example.invalid/probe-1.0.0.exe
    InstallerSha256: 5B04177B0CE3A6B56A744A6F6910D506F295C0463415E69C023DF77A0C07F14E
ManifestType: installer
ManifestVersion: 1.6.0
`

// A second installer manifest for the same version, to cover the documented
// case of more than one InstallerManifest per package.
const secondInstallerManifest = `
PackageIdentifier: Contoso.Probe
PackageVersion: 1.0.0
Installers:
  - Architecture: arm64
    InstallerType: exe
    InstallerUrl: https://example.invalid/probe-1.0.0-arm64.exe
    InstallerSha256: 9F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A08
ManifestType: installer
ManifestVersion: 1.6.0
`

const localeManifest = `
PackageIdentifier: Contoso.Probe
PackageVersion: 1.0.0
PackageLocale: de-DE
Publisher: Contoso
PackageName: Probe Werkzeug
ManifestType: locale
ManifestVersion: 1.6.0
`

func TestParseMultiFileManifest(t *testing.T) {
  tests := []struct {
    name            string
    docs            []string
    wantPackageName string
    wantPublisher   string
    wantInstallers  int
    wantLocales     int
  }{
    {
      name:            "version, defaultLocale and installer",
      docs:            []string{versionManifest, defaultLocaleManifest, installerManifest},
      wantPackageName: "Probe Tool",
      wantPublisher:   "Contoso",
      wantInstallers:  1,
      wantLocales:     0,
    },
    {
      name:            "an additional locale manifest",
      docs:            []string{versionManifest, defaultLocaleManifest, installerManifest, localeManifest},
      wantPackageName: "Probe Tool",
      wantPublisher:   "Contoso",
      wantInstallers:  1,
      wantLocales:     1,
    },
    {
      name:            "two installer manifests for one version",
      docs:            []string{versionManifest, defaultLocaleManifest, installerManifest, secondInstallerManifest},
      wantPackageName: "Probe Tool",
      wantPublisher:   "Contoso",
      wantInstallers:  2,
      wantLocales:     0,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      nodes := manifestNodesFromYAML(t, tt.docs...)

      manifest, err := parseMultiFileManifest(nodes...)
      if err != nil {
        t.Fatalf("parseMultiFileManifest returned an error: %v", err)
      }
      if manifest == nil {
        t.Fatal("parseMultiFileManifest returned a nil manifest and no error")
      }

      if got := manifest.GetPackageIdentifier(); got != "Contoso.Probe" {
        t.Errorf("PackageIdentifier = %q, want %q", got, "Contoso.Probe")
      }

      versions := manifest.GetVersions()
      if len(versions) != 1 {
        t.Fatalf("got %d versions, want 1", len(versions))
      }
      version := versions[0]

      if got := version.GetPackageVersion(); got != "1.0.0" {
        t.Errorf("PackageVersion = %q, want %q", got, "1.0.0")
      }
      if got := version.GetDefaultLocalePackageName(); got != tt.wantPackageName {
        t.Errorf("PackageName = %q, want %q", got, tt.wantPackageName)
      }
      if got := version.GetDefaultLocalePublisher(); got != tt.wantPublisher {
        t.Errorf("Publisher = %q, want %q", got, tt.wantPublisher)
      }
      if got := len(version.GetInstallers()); got != tt.wantInstallers {
        t.Errorf("got %d installers, want %d", got, tt.wantInstallers)
      }
      if got := len(version.GetLocales()); got != tt.wantLocales {
        t.Errorf("got %d additional locales, want %d", got, tt.wantLocales)
      }
    })
  }
}

// parseMultiFileManifest is variadic and documents the empty case as an error
// rather than a nil result, so that contract is pinned here too.
func TestParseMultiFileManifestRejectsNoNodes(t *testing.T) {
  if _, err := parseMultiFileManifest(); err == nil {
    t.Error("expected an error when called with no manifest nodes, got nil")
  }
}
