package models

// These types / data structures are used the same by the manifests and the API.
// This means no conversion is necessary when moving this data between the two.

type Installer struct {
    Architecture Architecture `yaml:"Architecture"`
    MinimumOSVersion string `yaml:"MinimumOSVersion"`
    Platform []string `yaml:"Platform"`
    InstallerType InstallerType `yaml:"InstallerType"`
    Scope Scope `yaml:"Scope"`
    InstallerUrl string `yaml:"InstallerUrl"`
    InstallerSha256 string `yaml:"InstallerSha256"`
    SignatureSha256 string `yaml:"SignatureSha256" json:",omitempty"` // winget runs into an exception internally when this is an empty string (ParseFromHexString: Invalid value size), so omit in API responses if empty
    InstallModes []InstallMode `yaml:"InstallModes"`
    InstallerSuccessCodes []int64 `yaml:"InstallerSuccessCodes"`
    ExpectedReturnCodes []ExpectedReturnCode `yaml:"ExpectedReturnCodes"`
    ProductCode string `yaml:"ProductCode"`
    ReleaseDate string `yaml:"ReleaseDate"`
}

type ExpectedReturnCode struct {
    InstallerReturnCode int64 `yaml:"InstallerReturnCode"`
    ReturnResponse ReturnResponse `yaml:"ReturnResponse"`
}

type Agreement struct {
    AgreementLabel string `yaml:"AgreementLabel"`
    Agreement string `yaml:"Agreement"`
    AgreementUrl string `yaml:"AgreementUrl"`
}

type Architecture string

const (
    neutral Architecture = "neutral"
    x86 Architecture = "x86"
    x64 Architecture = "x64"
    arm Architecture = "arm"
    arm64 Architecture = "arm64"
)

type InstallerType string

const (
    msix InstallerType = "msix"
    msi InstallerType = "msi"
    appx InstallerType = "appx"
    exe InstallerType = "exe"
    zip InstallerType = "zip"
    inno InstallerType = "inno"
    nullsoft InstallerType = "nullsoft"
    wix InstallerType = "wix"
    burn InstallerType = "burn"
    pwa InstallerType = "pwa"
    msstore InstallerType = "msstore"
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

type ReturnResponse string

const (
    packageInUse ReturnResponse = "packageInUse"
    installInProgress ReturnResponse = "installInProgress"
    fileInUse ReturnResponse = "fileInUse"
    missingDependency ReturnResponse = "missingDependency"
    diskFull ReturnResponse = "diskFull"
    insufficientMemory ReturnResponse = "insufficientMemory"
    noNetwork ReturnResponse = "noNetwork"
    contactSupport ReturnResponse = "contactSupport"
    rebootRequiredToFinish ReturnResponse = "rebootRequiredToFinish"
    rebootRequiredForInstall ReturnResponse = "rebootRequiredForInstall"
    rebootInitiated ReturnResponse = "rebootInitiated"
    cancelledByUser ReturnResponse = "cancelledByUser"
    alreadyInstalled ReturnResponse = "alreadyInstalled"
    downgrade ReturnResponse = "downgrade"
    blockedByPolicy ReturnResponse = "blockedByPolicy"
)
