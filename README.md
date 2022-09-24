# rewinged

rewinged is an implementation of the winget source REST API as a single, portable executable.

It allows you to easily self-host private winget package repositories from a directory of package manifests.

It is currently in [pre-1.0](https://semver.org/#spec-item-4) development so command-line args, output and behavior may change at any time!

I'm also using it as an opportunity to learn some Go.

### Already Working

- ✅ Directly serve [unmodified winget package manifests](https://github.com/microsoft/winget-pkgs/tree/master/manifests)
- ✅ Add your own manifests for internal software
- ✅ Search, list, show and install software - the core winget features
- ✅ Runs on Windows, Linux or in [Docker](https://github.com/jantari/rewinged/blob/main/Dockerfile)

### Not Yet Working

- ⚠️ Package manifest versions other than 1.1.0 have not been tested
- ⚠️ Correlation of installed programs and programs in the repository is not perfect (in part due to [this](https://github.com/microsoft/winget-cli-restsource/issues/59) and [this](https://github.com/microsoft/winget-cli-restsource/issues/166))
- ⚠️ Live reload of manifests (works for new and changed manifests, but removals are only picked up on restart)
- ❌ Authentication (it's currently [not supported by winget](https://github.com/microsoft/winget-cli-restsource/issues/100))
- 🤔 Probably other stuff? It's work-in-progress - please submit an issue and/or PR if you notice anything!

### Usage

The following command-line arguments are available:

```
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

Please note that winget **requires** REST-sources to use HTTPS - plaintext HTTP is not allowed,
so you will have to either use the `-https*` options of rewinged to configure HTTPS directly or
front it with a proxy such as nginx, HAProxy, caddy etc. that handles the encryption.

You can however test the API without HTTPS, for example by opening `http://<listenaddress>:<port>/information`
or `http://<listenaddress>:<port>/packages` in a browser or with `curl` / `Invoke-RestMethod`.

### Using a local rewinged instance as a package source

```
~
❯ winget source add -n rewinged-local -a https://localhost:8443 -t "Microsoft.Rest"
Adding source:
  rewinged-local -> https://localhost:8443
Done

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

### Helpful reference documentation

winget-cli-restsource: https://github.com/microsoft/winget-cli-restsource

winget restsource API: https://github.com/microsoft/winget-cli-restsource/blob/main/documentation/WinGet-1.1.0.yaml
