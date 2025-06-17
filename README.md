# rewinged

rewinged is a self-hosted winget package source. It's portable and can run on Linux, Windows, in Docker, locally and in any cloud.
rewinged reads your package manifests from a directory and makes them searchable and accessable to winget via a REST API.

It is currently in [pre-1.0](https://semver.org/#spec-item-4) development so configuration options, output and behavior may change at any time!

## 🚀 Features

- Directly serve [unmodified winget package manifests](https://github.com/microsoft/winget-pkgs/tree/master/manifests)
- Add your own manifests for internal or customized software
- Search, list, show and install software - the core winget features
- Automatically internalize package installers to serve them to machines without internet
- Restrict access to the package source with Entra ID authentication
- Package manifest versions from 1.1.0 to 1.10.0 are all supported simultaneously
- Runs on Windows, Linux and in Docker

## 🚧 Not Yet Working or Complete

- Correlation of installed programs and programs in the repository is not perfect (in part due to [this](https://github.com/microsoft/winget-cli-restsource/issues/59) and [this](https://github.com/microsoft/winget-cli-restsource/issues/166))
- Live reload of manifests (works for new and changed manifests, but removals are only picked up on restart)
- Probably other stuff? It's work-in-progress - please submit an issue and/or PR if you notice anything!

## 🧭 Getting Started

### 💾 Binaries

Download the latest [release](https://github.com/jantari/rewinged/releases), extract and run it!

### 🐋 Docker

```
docker run \
  -e REWINGED_LISTEN='0.0.0.0:8080' \
  -p 8080:8080 \
  -v ${PWD}/packages:/packages:ro \
  ghcr.io/jantari/rewinged:stable
```

### ⚙️ Configuration

rewinged can be configured through commandline arguments, environment variables and a JSON configuration file.

<details>
<summary><b>Commandline Arguments</b></summary>

Commandline arguments have the highest priority and take precedence over both environment variables and the configuration file.

```
  -autoInternalize
        Turn on the auto-internalization feature
  -autoInternalizePath string
        The directory where auto-internalized installers will be stored (default "./installers")
  -autoInternalizeSkip string
        List of hostnames excluded from auto-internalization (comma or space to separate)
  -configFile string
        Path to a json configuration file (optional)
  -https
        Serve encrypted HTTPS traffic directly from rewinged without the need for a proxy
  -httpsCertificateFile string
        The webserver certificate to use if HTTPS is enabled (default "./cert.pem")
  -httpsPrivateKeyFile string
        The private key file to use if HTTPS is enabled (default "./private.key")
  -listen string
        The address and port for the REST API to listen on (default "localhost:8080")
  -logLevel string
        Set log verbosity: disable, error, warn, info, debug or trace (default "info")
  -manifestPath string
        The directory to search for package manifest files (default "./packages")
  -sourceAuthEntraIDAuthorityURL string
        Authority/Issuer URL of the EntraID App used for authenticating clients
  -sourceAuthEntraIDResource string
        ApplicationID of the EntraID App used for authenticating clients
  -sourceAuthType string
        Require authentication to interact with the REST API: none, microsoftEntraId (default "none")
  -trustedProxies string
        List of IPs from which to trust Client-IP headers (comma or space to separate)
  -version
        Print the version information and exit
```

</details>

<details>
<summary><b>Environment Variables</b></summary>

Environment variables take precedence over the configuration file, but are overridden by any commandline arguments if passed.

```
REWINGED_CONFIGFILE (string)
REWINGED_AUTOINTERNALIZE (bool)
REWINGED_AUTOINTERNALIZEPATH (string)
REWINGED_AUTOINTERNALIZESKIP (string)
REWINGED_HTTPS (bool)
REWINGED_HTTPSCERTIFICATEFILE (string)
REWINGED_HTTPSPRIVATEKEYFILE (string)
REWINGED_LISTEN (string)
REWINGED_LOGLEVEL (string)
REWINGED_MANIFESTPATH (string)
REWINGED_SOURCEAUTHENTRAIDAUTHORITYURL (string)
REWINGED_SOURCEAUTHENTRAIDRESOURCE (string)
REWINGED_SOURCEAUTHTYPE (string)
REWINGED_TRUSTEDPROXIES (string)
```

</details>

<details>
<summary><b>Configuration File</b></summary>

Use the `-configFile` argument or `REWINGED_CONFIGFILE` environment variable to enable the config file option.
rewinged will not look for any configuration file by default. Config file must be valid JSON.

```json
{
  "autoInternalize": false,
  "autoInternalizePath": "./installers",
  "autoInternalizeSkip": "",
  "https": false,
  "httpsCertificateFile": "./cert.pem",
  "httpsPrivateKeyFile": "./private.key",
  "listen": "localhost:8080",
  "logLevel": "info",
  "manifestPath": "./packages",
  "sourceAuthEntraIDAuthorityURL": "",
  "sourceAuthEntraIDResource": "",
  "sourceAuthType": "none",
  "trustedProxies": ""
}
```

</details>

### 🪄 Using rewinged

You can run rewinged and test the API by opening `http://localhost:8080/api/information`
or `http://localhost:8080/api/packages` in a browser or with `curl` / `Invoke-RestMethod`.

But to use it with winget you will have to set up HTTPS because winget **requires**
REST-sources to use HTTPS - plaintext HTTP is not allowed. If you do not have an internal
PKI or a certificate from a publicly trusted CA like Let's Encrypt you can use then you
can generate and trust a new self-signed certificate for testing purposes:

<details>
<summary><b>Generate and trust a certificate and private key in PowerShell 5.1</b></summary>

```powershell
# Because we are adding a certificate to the local machine store, this has to be run in an elevated PowerShell session

$IPs = [System.Net.NetworkInformation.NetworkInterface]::GetAllNetworkInterfaces() |
    Foreach-Object GetIPProperties |
    Foreach-Object UnicastAddresses |
    Foreach-Object Address |
    Foreach-Object {
        "&IPAddress=$( [System.Net.IPAddress]::new($_.GetAddressBytes() ))"
    }

[string]$SanIPs = -join $IPs

$SelfSignedCertificateParameters = @{
    'Subject'         = 'localhost'
    'TextExtension'   = @("2.5.29.17={text}DNS=localhost${SanIPs}")
    'NotAfter'        = (Get-Date).AddYears(1)
    'FriendlyName'    = 'rewinged HTTPS'
    'KeyAlgorithm'    = 'RSA'
    'KeyExportPolicy' = 'Exportable'
}
$cert = New-SelfSignedCertificate @SelfSignedCertificateParameters

$RSAPrivateKey    = [System.Security.Cryptography.X509Certificates.RSACertificateExtensions]::GetRSAPrivateKey($cert)
$PrivateKeyBytes  = $RSAPrivateKey.Key.Export([System.Security.Cryptography.CngKeyBlobFormat]::Pkcs8PrivateBlob)
$PrivateKeyBase64 = [System.Convert]::ToBase64String($PrivateKeyBytes, [System.Base64FormattingOptions]::InsertLineBreaks)

$CertificateBase64 = [System.Convert]::ToBase64String($cert.Export('Cert'), [System.Base64FormattingOptions]::InsertLineBreaks)

Set-Content -Path private.key -Encoding Ascii -Value @"
-----BEGIN RSA PRIVATE KEY-----`r`n${PrivateKeyBase64}`r`n-----END RSA PRIVATE KEY-----
"@

Set-Content -Path cert.pem -Encoding Ascii -Value @"
-----BEGIN CERTIFICATE-----`r`n${CertificateBase64}`r`n-----END CERTIFICATE-----
"@

$store = [System.Security.Cryptography.X509Certificates.X509Store]::new('Root', 'LocalMachine')
$store.Open('ReadWrite')
$store.Add($cert)
$store.Close()

Remove-Item $cert.PSPath
```
</details>

Then, run rewinged with HTTPS enabled:

```
./rewinged -https -listen localhost:8443
```

add it as a package source in winget:

```
winget source add -n rewinged-local -a https://localhost:8443/api -t "Microsoft.Rest"
```

and query it!

```
~
❯ winget search -s rewinged-local -q bottom
Name   Id              Version
------------------------------
bottom rewinged.bottom 0.6.8

~
❯ winget install -s rewinged-local -q bottom
Found bottom [rewinged.bottom] Version 0.6.8
This application is licensed to you by its owner.
Microsoft is not responsible for, nor does it grant any licenses to, third-party packages.
Downloading https://github.com/ClementTsang/bottom/releases/download/0.6.8/bottom_x86_64_installer.msi
  ██████████████████████████████  1.56 MB / 1.56 MB
Successfully verified installer hash
Starting package install...
Successfully installed

~ took 8s
❯
```

## 🤖 Auto-Internalization

With auto-internalization enabled, rewinged will automatically:

1. Download all installers referenced in your package manifests (InstallerUrl fields)
2. Serve all of the downloaded installer files itself, locally, on the `/installers/` URL-path
3. Dynamically rewrite all InstallerUrls returned from its APIs to point to itself instead of the original source

You can choose a path where rewinged will store the downloaded installers with `autoInternalizePath`
and you can exempt a list of hostnames from being auto-internalized with `autoInternalizeSkip`.
This is useful if you have custom manifests that already point to internal sources and there is no
need to re-internalize them again, or when certain installers are very large and you just don't want
to store them locally. For example:

```
./rewinged -autoInternalize -autoInternalizeSkip "internal.example.org github.com"
```

## 🔒 Entra ID Authentication

You can optionally enable Entra ID authentication for rewinged. This means only authorized users will be able to
get software from the repository. Currently, this is an all-or-nothing setting meaning the entire source repository
and all packages will require authentication to access, and you cannot restrict individual packages to specific
groups of users yet. If a user is able to successfully authenticate with your Entra ID tenant they will be able
to access everything in the repository. If they fail to authenticate, for example because you have not assigned the
user aceess to rewinged or because of Conditional Access policies, they will not be able to access anything.

On a hybrid- or cloud-only Entra ID-joined Windows device, the users authentication through winget to rewinged
should be transparent and automatic (SSO). Otherwise the user will see authentication windows.

To configure Entra ID Authentication, an application must be registered in your Entra ID tenant.

<details>
<summary><b>Register an application for use with winget REST source authentication in Entra ID</b></summary>

The following is based on Microsofts [`New-MicrosoftEntraIdApp.ps1`](https://github.com/microsoft/winget-cli-restsource/blob/main/Tools/PowershellModule/src/Library/New-MicrosoftEntraIdApp.ps1) script.

```powershell
$ScopeId = [Guid]::NewGuid().ToString()
$app = New-AzADApplication -DisplayName "rewinged" -SignInAudience AzureADMyOrg -RequestedAccessTokenVersion 2 -Api @{
    oauth2PermissionScopes = @(
        @{
            adminConsentDescription = "Sign in to access rewinged REST source"
            adminConsentDisplayName = "Access rewinged REST source"
            userConsentDescription  = "Sign in to access rewinged REST source"
            userConsentDisplayName  = "Access rewinged REST source"
            id = $ScopeId
            isEnabled = $true
            type = "User"
            value = "user_impersonation"
        }
    )
    preAuthorizedApplications = @(
        @{
            # "App Installer"
            appId = "7b8ea11a-7f45-4b3a-ab51-794d5863af15"
            delegatedPermissionIds = @($ScopeId)
        },
        @{
            # "Microsoft Azure CLI"
            appId = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
            delegatedPermissionIds = @($ScopeId)
        },
        @{
            # "Microsoft Azure PowerShell"
            appId = "1950a258-227b-4e31-a9cf-717495945fc2"
            delegatedPermissionIds = @($ScopeId)
        }
    )
}

Update-AzADApplication -ApplicationId $app.AppId -IdentifierUri "api://$($app.AppId)"

Write-Output "Done. Application Id: $($app.AppId)"
```
</details>

Then start rewinged with the authentication-related configuration set:

```
./rewinged -https -sourceAuthType microsoftEntraId -sourceAuthEntraIDAuthorityURL "https://login.microsoftonline.com/<Your-Tenant-Id>/v2.0" -sourceAuthEntraIDResource "<Entra-Application-Id>"
```

<table>
  <tr>
    <th>⚠️</th>
    <td>When you enable authentication, ensure all of your package manifests are ManifestVersion 1.10.0 or higher. Previous ManifestVersions did not fully support authentication and winget will not be able to correctly retrieve such packages when authentication is enabled.</td>
  </tr>
</table>

## Helpful reference documentation

rewinged: Run `./rewinged -help` to see all available command-line options.

winget-cli-restsource: https://github.com/microsoft/winget-cli-restsource

winget restsource API: https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

If you have trouble, questions or suggestions about rewinged please open an issue!
