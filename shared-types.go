package main

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
    x86 = "x86"
    x64 = "x64"
    arm = "arm"
    arm64 = "arm64"
)

type InstallerType string

const (
    msix InstallerType = "msix"
    msi = "msi"
    appx = "appx"
    exe = "exe"
    zip = "zip"
    inno = "inno"
    nullsoft = "nullsoft"
    wix = "wix"
    burn = "burn"
    pwa = "pwa"
    msstore = "msstore"
)

type Scope string

const (
    user Scope = "user"
    machine = "machine"
)

type InstallMode string

const (
    interactive InstallMode = "interactive"
    silent = "silent"
    silentWithProgress = "silentWithProgress"
)

type ReturnResponse string

const (
    packageInUse ReturnResponse = "packageInUse"
    installInProgress = "installInProgress"
    fileInUse = "fileInUse"
    missingDependency = "missingDependency"
    diskFull = "diskFull"
    insufficientMemory = "insufficientMemory"
    noNetwork = "noNetwork"
    contactSupport = "contactSupport"
    rebootRequiredToFinish = "rebootRequiredToFinish"
    rebootRequiredForInstall = "rebootRequiredForInstall"
    rebootInitiated = "rebootInitiated"
    cancelledByUser = "cancelledByUser"
    alreadyInstalled = "alreadyInstalled"
    downgrade = "downgrade"
    blockedByPolicy = "blockedByPolicy"
)
