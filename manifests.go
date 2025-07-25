package main

import (
  "os"
  "errors"
  "slices"
  "strings"
  "io"
  "io/fs"
  "net/url"
  "net/http"
  "path/filepath"

  "gopkg.in/yaml.v3"

  "rewinged/logging"
  "rewinged/models"
)

func ingestManifestsWorker(autoInternalize bool, autoInternalizePath string, autoInternalizeSkipHosts []string) error {
  for path := range jobs {
    files, err := os.ReadDir(path)
    if err != nil {
      logging.Logger.Error().Err(err).Msg("ingestManifestsWorker error")
      wg.Done()
      continue
    }

    // temporary map collecting all files belonging to a particular package
    var nonSingletonsMap = make(map[models.MultiFileManifest][]models.ManifestNode)

    for _, file := range files {
      if !file.IsDir() {
        if caseInsensitiveHasSuffix(file.Name(), ".yml") || caseInsensitiveHasSuffix(file.Name(), ".yaml") {
          var basemanifests, err = parseFileAsBaseManifests(filepath.Join(path, file.Name()))
          if err != nil {
            logging.Logger.Error().Err(err).Str("file", filepath.Join(path, file.Name())).Msgf("cannot unmarshal YAML file as BaseManifest")
            continue
          }

          for _, basemanifest := range basemanifests {
            // There could be other, non winget-manifest YAML files/documents in the manifestPath as well. Skip them.
            // All valid manifests must have all basemanifest fields set as they are required by the schema
            if basemanifest.PackageIdentifier != "" && basemanifest.PackageVersion != "" &&
              basemanifest.ManifestType != "" && basemanifest.ManifestVersion != "" {
              if basemanifest.ManifestType == "singleton" {
                logging.Logger.Debug().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("found singleton manifest")
                manifest, err := parseNodeAsSingletonManifest(basemanifest.ManifestVersion, basemanifest.Node)
                if err != nil {
                  logging.Logger.Error().Err(err).Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msg("could not parse singleton manifest")
                } else {
                  // Singleton manifests can only contain version of a package each
                  var version = manifest.GetVersions()[0]

                  // Internalization logic
                  if (autoInternalize) {
                    var installers []models.API_InstallerInterface = version.GetInstallers()

                    internalizeInstallers(basemanifest.PackageIdentifier, basemanifest.PackageVersion, installers, autoInternalizePath, autoInternalizeSkipHosts)

                    // Recreate manifest object, but with overwritten values (InstallerUrl(s))
                    manifest, err = newAPIManifest(
                      basemanifest.ManifestVersion,
                      basemanifest.PackageIdentifier,
                      basemanifest.PackageVersion,
                      version.GetDefaultLocale(),
                      version.GetLocales(),
                      installers,
                    )

                    if err != nil {
                      logging.Logger.Error().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("error reconstructing package after overwrite")
                    }

                    version = manifest.GetVersions()[0]
                  }
                  // End internalization logic

                  models.Manifests.Set(manifest.GetPackageIdentifier(), basemanifest.PackageVersion, version)
                }
              } else if basemanifest.ManifestType == "merged" {
                logging.Logger.Error().Str("package", basemanifest.PackageIdentifier).Str("packageversion", basemanifest.PackageVersion).Msgf("merged manifests are not yet supported")
              } else {
                nonSingletonsMap[basemanifest.ToMultiFileManifest()] = append(nonSingletonsMap[basemanifest.ToMultiFileManifest()], *basemanifest)
              }
            } else {
              logging.Logger.Debug().Err(err).Str("file", filepath.Join(path, file.Name())).Int("line", basemanifest.Node.Line).Msg("YAML document is not a package manifest")
            }
          }
        }
      }
    }

    if len(nonSingletonsMap) > 0 {
      for key, value := range nonSingletonsMap {
        logging.Logger.Debug().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("found multi-file manifest")
        var mergedManifest, err = parseMultiFileManifest(value...)
        if err != nil {
          logging.Logger.Error().Err(err).Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("could not parse all manifest files for this package")
        } else {
          for _, version := range mergedManifest.GetVersions() {
            // Internalization logic
            if (autoInternalize) {
              var installers []models.API_InstallerInterface = version.GetInstallers()

              internalizeInstallers(key.PackageIdentifier, key.PackageVersion, installers, autoInternalizePath, autoInternalizeSkipHosts)

              // Recreate manifest object, but with overwritten values (InstallerUrl(s))
              overwrittenMergedManifest, err := newAPIManifest(
                key.ManifestVersion,
                key.PackageIdentifier,
                key.PackageVersion,
                version.GetDefaultLocale(),
                version.GetLocales(),
                installers,
              )

              if err != nil {
                logging.Logger.Error().Str("package", key.PackageIdentifier).Str("packageversion", key.PackageVersion).Msgf("error reconstructing package after overwrite")
              }

              version = overwrittenMergedManifest.GetVersions()[0]
            }
            // End internalization logic

            // Replace the existing PkgId + PkgVersion entry with this one
            models.Manifests.Set(mergedManifest.GetPackageIdentifier(), version.GetPackageVersion(), version)
          }
        }
      }
    }

    wg.Done()
  }

  return nil
}

func internalizeInstallers(
  packageIdentifier string,
  packageVersion string,
  installers []models.API_InstallerInterface,
  autoInternalizePath string,
  autoInternalizeSkipHosts []string,
) {
  for _, installer := range installers {
    var originalInstallerURL string = installer.GetInstallerUrl()
    u, err := url.Parse(originalInstallerURL)
    if err != nil {
      logging.Logger.Error().Err(err).Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("cannot parse InstallerUrl %s", originalInstallerURL)
      continue
    }
    if slices.Contains(autoInternalizeSkipHosts, u.Hostname()) {
      logging.Logger.Debug().Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("not internalizing %s", originalInstallerURL)
      continue
    }

    var destFile string = filepath.Join(autoInternalizePath, strings.ToLower(installer.GetInstallerSha()))
    // Why os.OpenFile instead of os.Create:
    // https://stackoverflow.com/a/22483001
    out, err := os.OpenFile(destFile, os.O_RDWR | os.O_CREATE | os.O_EXCL, 0666)
    if err != nil {
      if errors.Is(err, fs.ErrExist) {
        logging.Logger.Debug().Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("file already exists, not redownloading %s", destFile)
      } else {
        logging.Logger.Error().Err(err).Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("cannot create file %s", destFile)
        continue
      }
    } else {
      defer out.Close()
      // No error, we could open the file for writing and it does not exist yet - so download it
      logging.Logger.Debug().Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("downloading installer")

      resp, err := http.Get(originalInstallerURL)
      if err != nil {
        logging.Logger.Error().Err(err).Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("cannot download file %s", originalInstallerURL)
        continue
      }
      defer resp.Body.Close()
      if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        logging.Logger.Error().Err(err).Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("cannot download file %s (http status %d)", originalInstallerURL, resp.StatusCode)
        continue
      }

      n, err := io.Copy(out, resp.Body)
      if err != nil {
        logging.Logger.Error().Err(err).Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("cannot save to file %s", destFile)
        continue
      } else {
        logging.Logger.Debug().Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("downloaded installer, %d bytes written", n)
      }
    }

    logging.Logger.Debug().Str("package", packageIdentifier).Str("packageversion", packageVersion).Msgf("prepared internaliziation")
    // Remember that this installer was internalized successfully (could be or already was downloaded)
    // so we know we can rewrite its InstallerUrl later.
    models.InternalizedInstallers[installer.GetInstallerSha()] = true
  }
}

// Finds and parses all package manifest files in a directory
// recursively and returns them as a map of PackageIdentifier
// and PackageVersions
func getManifests (path string) {
  files, err := os.ReadDir(path)
  if err != nil {
    logging.Logger.Error().Err(err)
  }

  // wg.Add() before goroutine, see staticcheck check SA2000 and also
  // https://stackoverflow.com/questions/65213707/where-to-put-wg-add
  wg.Add(1)
  go func() {
    jobs <- path
  }()

  for _, file := range files {
    if file.IsDir() {
      subdirPath := filepath.Join(path, file.Name())
      logging.Logger.Trace().Msgf("searching directory %s", subdirPath)
      getManifests(subdirPath)
    }
  }
}

func unmarshalVersionManifest (manifestVersion string, node yaml.Node) (models.Manifest_VersionManifestInterface, error) {
    var version models.Manifest_VersionManifestInterface

    switch manifestVersion {
        case "1.1.0":
          version = &models.Manifest_VersionManifest_1_1_0{}
        case "1.2.0":
          version = &models.Manifest_VersionManifest_1_2_0{}
        case "1.4.0":
          version = &models.Manifest_VersionManifest_1_4_0{}
        case "1.5.0":
          version = &models.Manifest_VersionManifest_1_5_0{}
        case "1.6.0":
          version = &models.Manifest_VersionManifest_1_6_0{}
        case "1.7.0":
          version = &models.Manifest_VersionManifest_1_7_0{}
        case "1.9.0":
          version = &models.Manifest_VersionManifest_1_9_0{}
        case "1.10.0":
          version = &models.Manifest_VersionManifest_1_10_0{}
        default:
          return nil, errors.New("unsupported VersionManifest version " + manifestVersion)
    }

    err := node.Decode(version)
    if err != nil {
      return nil, err
    }

    return version, nil
}

func unmarshalInstallerManifest (manifestVersion string, node yaml.Node) (models.Manifest_InstallerManifestInterface, error) {
    var installer models.Manifest_InstallerManifestInterface

    switch manifestVersion {
        case "1.1.0":
            installer = &models.Manifest_InstallerManifest_1_1_0{}
        case "1.2.0":
            installer = &models.Manifest_InstallerManifest_1_2_0{}
        case "1.4.0":
            installer = &models.Manifest_InstallerManifest_1_4_0{}
        case "1.5.0":
            installer = &models.Manifest_InstallerManifest_1_5_0{}
        case "1.6.0":
            installer = &models.Manifest_InstallerManifest_1_6_0{}
        case "1.7.0":
            installer = &models.Manifest_InstallerManifest_1_7_0{}
        case "1.9.0":
            installer = &models.Manifest_InstallerManifest_1_9_0{}
        case "1.10.0":
            installer = &models.Manifest_InstallerManifest_1_10_0{}
        default:
            return nil, errors.New("unsupported InstallerManifest version " + manifestVersion)
    }

    err := node.Decode(installer)
    if err != nil {
      return nil, err
    }

    return installer, nil
}

func unmarshalLocaleManifest (manifestVersion string, node yaml.Node) (models.Manifest_LocaleManifestInterface, error) {
    var locale models.Manifest_LocaleManifestInterface

    switch manifestVersion {
        case "1.1.0":
            locale = &models.Manifest_LocaleManifest_1_1_0{}
        case "1.2.0":
            locale = &models.Manifest_LocaleManifest_1_2_0{}
        case "1.4.0":
            locale = &models.Manifest_LocaleManifest_1_4_0{}
        case "1.5.0":
            locale = &models.Manifest_LocaleManifest_1_5_0{}
        case "1.6.0":
            locale = &models.Manifest_LocaleManifest_1_6_0{}
        case "1.7.0":
            locale = &models.Manifest_LocaleManifest_1_7_0{}
        case "1.9.0":
            locale = &models.Manifest_LocaleManifest_1_9_0{}
        case "1.10.0":
            locale = &models.Manifest_LocaleManifest_1_10_0{}
        default:
            return nil, errors.New("unsupported LocaleManifest version " + manifestVersion)
    }

    err := node.Decode(locale)
    if err != nil {
        return nil, err
    }

    return locale, nil
}

func unmarshalDefaultLocaleManifest (manifestVersion string, node yaml.Node) (models.Manifest_DefaultLocaleManifestInterface, error) {
    var defaultlocale models.Manifest_DefaultLocaleManifestInterface

    switch manifestVersion {
        case "1.1.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_1_0{}
        case "1.2.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_2_0{}
        case "1.4.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_4_0{}
        case "1.5.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_5_0{}
        case "1.6.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_6_0{}
        case "1.7.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_7_0{}
        case "1.9.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_9_0{}
        case "1.10.0":
            defaultlocale = &models.Manifest_DefaultLocaleManifest_1_10_0{}
        default:
            return nil, errors.New("unsupported DefaultLocaleManifest version " + manifestVersion)
    }

    err := node.Decode(defaultlocale)
    if err != nil {
        return nil, err
    }

    return defaultlocale, nil
}

func parseMultiFileManifest (nodes ...models.ManifestNode) (models.API_ManifestInterface, error) {
  if len(nodes) <= 0 {
    return nil, errors.New("you must provide at least one ManifestNode for reading values")
  }

  versions   := []models.Manifest_VersionManifestInterface{}
  installers := []models.Manifest_InstallerManifestInterface{}
  locales    := []models.Manifest_LocaleManifestInterface{}
  var defaultlocale models.Manifest_DefaultLocaleManifestInterface

  for _, node := range nodes {
    var err error
    switch node.ManifestType {
      case "version":
        var version models.Manifest_VersionManifestInterface
        version, err = unmarshalVersionManifest(node.ManifestVersion, node.Node)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("cannot unmarshal version manifest")
          continue
        }
        versions = append(versions, version)
      case "installer":
        var installer models.Manifest_InstallerManifestInterface
        installer, err = unmarshalInstallerManifest(node.ManifestVersion, node.Node)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("cannot unmarshal installer manifest")
          continue
        }
        installers = append(installers, installer)
      case "locale":
        var locale models.Manifest_LocaleManifestInterface
        locale, err = unmarshalLocaleManifest(node.ManifestVersion, node.Node)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("cannot unmarshal locale manifest")
          continue
        }
        locales = append(locales, locale)
      case "defaultLocale":
        defaultlocale, err = unmarshalDefaultLocaleManifest(node.ManifestVersion, node.Node)
        if err != nil {
          logging.Logger.Error().Err(err).Msg("cannot unmarshal defaultLocale manifest")
          continue
        }
      default:
    }
  }

  // It's possible there were no installer or locale manifests or parsing them failed
  if len(installers) == 0 {
    return nil, errors.New("no (valid) installer manifest")
  }
  if len(versions) == 0 {
    return nil, errors.New("no (valid) locale manifest")
  }

  // This transforms the manifest data into the format the API will return.
  // This logic should probably be moved out of this function, so that it returns
  // the full unaltered data from the combined manifests - and restructuring to
  // API-format will happen somewhere else
  var apiLocales []models.API_LocaleInterface
  for _, locale := range locales {
    apiLocales = append(apiLocales, locale.ToApiLocale())
  }

  var apiInstallers []models.API_InstallerInterface

  // In case there are more than 1 InstallerManifests per package (not sure if officially allowed)
  // we get all the installers defined in all the InstallerManifest files.
  for _, v := range installers {
    apiInstallers = append(apiInstallers, v.ToApiInstallers()...)
  }

  // We know we have at least 1 node because otherwise we fail early, and that all nodes' PackageIdentifier,
  // PackageVersion and ManifestVersion are identical because that's what they were grouped by.
  manifest, err := newAPIManifest(
    nodes[0].ManifestVersion,
    nodes[0].PackageIdentifier,
    nodes[0].PackageVersion,
    defaultlocale.ToApiDefaultLocale(),
    apiLocales,
    apiInstallers,
  )

  return manifest, err
}

func newAPIManifest (
  ManifestVersion string,
  PackageIdentifier string,
  pv string,
  dl models.API_DefaultLocaleInterface,
  l []models.API_LocaleInterface,
  inst []models.API_InstallerInterface,
) (
  models.API_ManifestInterface,
  error,
) {
  var apiReturnManifest models.API_ManifestInterface
  var apiMvi models.API_ManifestVersionInterface

  if ManifestVersion == "1.1.0" {
    var apiLocales []models.API_Locale_1_1_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_1_0))
    }

    var apiInstallers []models.API_Installer_1_1_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_1_0))
    }

    apiMvi = models.API_ManifestVersion_1_1_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_1_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_1_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.2.0" || ManifestVersion == "1.4.0" {
    // There is no API schema 1.2.0, so both v1.2.0 and v1.4.0
    // packages are returned to clients as v1.4.0 API responses
    var apiLocales []models.API_Locale_1_4_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_4_0))
    }

    var apiInstallers []models.API_Installer_1_4_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_4_0))
    }

    apiMvi = models.API_ManifestVersion_1_4_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_4_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_4_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.5.0" {
    var apiLocales []models.API_Locale_1_5_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_5_0))
    }

    var apiInstallers []models.API_Installer_1_5_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_5_0))
    }

    apiMvi = models.API_ManifestVersion_1_5_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_5_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_5_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.6.0" {
    var apiLocales []models.API_Locale_1_6_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_6_0))
    }

    var apiInstallers []models.API_Installer_1_6_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_6_0))
    }

    apiMvi = models.API_ManifestVersion_1_6_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_6_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_6_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.7.0" {
    var apiLocales []models.API_Locale_1_7_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_7_0))
    }

    var apiInstallers []models.API_Installer_1_7_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_7_0))
    }

    apiMvi = models.API_ManifestVersion_1_7_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_7_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_7_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.9.0" {
    var apiLocales []models.API_Locale_1_9_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_9_0))
    }

    var apiInstallers []models.API_Installer_1_9_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_9_0))
    }

    apiMvi = models.API_ManifestVersion_1_9_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_9_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_9_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }
  } else if ManifestVersion == "1.10.0" {
    var apiLocales []models.API_Locale_1_10_0
    for _, locale := range l {
      apiLocales = append(apiLocales, locale.(models.API_Locale_1_10_0))
    }

    var apiInstallers []models.API_Installer_1_10_0
    for _, intf := range inst {
      apiInstallers = append(apiInstallers, *intf.(*models.API_Installer_1_10_0))
    }

    apiMvi = models.API_ManifestVersion_1_10_0{
      PackageVersion: pv,
      DefaultLocale: dl.(models.API_DefaultLocale_1_10_0),
      Channel: "",
      Locales: apiLocales,
      Installers: apiInstallers,
    }

    apiReturnManifest = &models.API_Manifest_1_10_0 {
      PackageIdentifier: PackageIdentifier,
      Versions: []models.API_ManifestVersionInterface{ apiMvi },
    }

  } else {
    return nil, errors.New("Converting manifest v" + ManifestVersion + " data for API responses is not yet supported.")
  }

  return apiReturnManifest, nil
}

// One file could contain multiple manifests, using YAML document separators ("---")
func parseFileAsBaseManifests (path string) ([]*models.ManifestNode, error) {
  manifests := []*models.ManifestNode{}

  yamlFile, err := os.Open(path)
  if err != nil {
    return manifests, err
  }
  defer yamlFile.Close()

  // Decode all YAML documents in the YAML file.
  // This allows for multiple manifests, or non-manifest
  // metadata to be in one file, separated by "---".
  fileDecoder := yaml.NewDecoder(yamlFile)
  for {
    var node yaml.Node
    if err := fileDecoder.Decode(&node); err == io.EOF {
      break
    } else if err != nil {
      return manifests, err
    }

    manifest := &models.ManifestNode{}
    if err := node.Decode(manifest); err != nil {
      return manifests, err
    }
    manifest.Node = node

    manifests = append(manifests, manifest)
  }

  return manifests, err
}

func parseNodeAsSingletonManifest (manifestVersion string, node yaml.Node) (models.API_ManifestInterface, error) {
  var manifest models.API_ManifestInterface
  var singleton models.Manifest_SingletonManifestInterface

  singleton, err := unmarshalSingletonManifest(manifestVersion, node)
  if err == nil {
    manifest = singleton.ToApiManifest()
  }

  return manifest, err
}

func unmarshalSingletonManifest (manifestVersion string, node yaml.Node) (models.Manifest_SingletonManifestInterface, error) {
    var smanifest models.Manifest_SingletonManifestInterface

    switch manifestVersion {
        case "1.1.0":
            smanifest = &models.Manifest_SingletonManifest_1_1_0{}
        case "1.2.0":
            smanifest = &models.Manifest_SingletonManifest_1_2_0{}
        case "1.4.0":
            smanifest = &models.Manifest_SingletonManifest_1_4_0{}
        case "1.5.0":
            smanifest = &models.Manifest_SingletonManifest_1_5_0{}
        case "1.6.0":
            smanifest = &models.Manifest_SingletonManifest_1_6_0{}
        case "1.7.0":
            smanifest = &models.Manifest_SingletonManifest_1_7_0{}
        case "1.9.0":
            smanifest = &models.Manifest_SingletonManifest_1_9_0{}
        case "1.10.0":
            smanifest = &models.Manifest_SingletonManifest_1_10_0{}
        default:
            return nil, errors.New("unsupported SingletonManifest version " + manifestVersion)
    }

    err := node.Decode(smanifest)
    if err != nil {
        return nil, err
    }

    return smanifest, nil
}

func caseInsensitiveHasSuffix(s, substr string) bool {
  s, substr = strings.ToUpper(s), strings.ToUpper(substr)
  return strings.HasSuffix(s, substr)
}

