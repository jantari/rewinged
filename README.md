# rewinged

rewinged is a work-in-progress, experimental implementation of the winget source REST API.

It allows to easily self-host winget package repositories.

I'm also using it as an opportunity to learn some Go.

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
