# rewinged

rewinged is a self-hosted winget package source. It's portable and can run on Linux, Windows, in Docker, locally and in any cloud.
rewinged reads your package manifests from a directory and makes them searchable and accessable to winget via a REST API.

It is currently in [pre-1.0](https://semver.org/#spec-item-4) development so configuration options, output and behavior may change at any time!

## 🚀 Features

- Directly serve [unmodified winget package manifests](https://github.com/microsoft/winget-pkgs/tree/master/manifests)
- Add your own manifests for internal or customized software
- Search, list, show and install software - the core winget features
- Package manifest versions 1.1.0, 1.2.0 and 1.4.0 are all supported simultaneously
- Runs on Windows, Linux and in Docker

## 🚧 Not Yet Working or Complete

- Correlation of installed programs and programs in the repository is not perfect (in part due to [this](https://github.com/microsoft/winget-cli-restsource/issues/59) and [this](https://github.com/microsoft/winget-cli-restsource/issues/166))
- Live reload of manifests (works for new and changed manifests, but removals are only picked up on restart)
- Authentication (it's currently [not supported by winget](https://github.com/microsoft/winget-cli-restsource/issues/100))
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
  ghcr.io/jantari/rewinged:0.3.1-rc.2
```

### ⚙️ Configuration

rewinged can be configured through commandline arguments, environment variables and a JSON configuration file.

<details>
<summary><b>Commandline Arguments</b></summary>

Commandline arguments have the highest priority and take precedence over both environment variables and the configuration file.

```
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
  -manifestPath string
        The directory to search for package manifest files (default "./packages")
  -version
        Print the version information and exit
```

</details>

<details>
<summary><b>Environment Variables</b></summary>

Environment variables take precedence over the configuration file, but are overridden by any commandline arguments if passed.

```
REWINGED_CONFIGFILE (string)
REWINGED_HTTPS (bool)
REWINGED_HTTPSCERTIFICATEFILE (string)
REWINGED_HTTPSPRIVATEKEYFILE (string)
REWINGED_LISTEN (string)
REWINGED_MANIFESTPATH (string)
```

</details>

<details>
<summary><b>Configuration File</b></summary>

Use the `-configFile` argument or `REWINGED_CONFIGFILE` environment variable to enable the config file option.
rewinged will not look for any configuration file by default. Config file must be valid JSON.

```json
{
  "https": false,
  "httpsCertificateFile": "./cert.pem",
  "httpsPrivateKeyFile": "./private.key",
  "listen": "localhost:8080",
  "manifestPath": "./packages"
}
```

</details>

### 🪄 Using rewinged

You can run rewinged and test the API by opening `http://localhost:8080/information`
or `http://localhost:8080/packages` in a browser or with `curl` / `Invoke-RestMethod`.

But to use it with winget you will have to set up HTTPS because winget **requires**
REST-sources to use HTTPS - plaintext HTTP is not allowed. If you do not have an internal
PKI or a certificate from a publicly trusted CA like Let's Encrypt you can use then you
can generate and trust a new self-signed certificate for testing purposes:

<details>
<summary><b>Generate and trust a certificate and private key in PowerShell 5.1</b></summary>

```powershell
# Because we are adding a certificate to the local machine store, this has to be run in an elevated PowerShell session

$SelfSignedCertificateParameters = @{
    'DnsName'         = 'localhost'
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
winget source add -n rewinged-local -a https://localhost:8443 -t "Microsoft.Rest"
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

## Helpful reference documentation

rewinged: Run `./rewinged -help` to see all available command-line options.

winget-cli-restsource: https://github.com/microsoft/winget-cli-restsource

winget restsource API: https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml

If you have trouble, questions or suggestions about rewinged please open an issue!
