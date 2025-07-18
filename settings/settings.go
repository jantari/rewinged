package settings

import (
    "net/netip"
    "rewinged/models"
)

var (
    TrustedProxies []netip.Prefix = []netip.Prefix{}
    SourceAuthenticationType = "none"
    SourceAuthenticationEntraIDResource = ""
    SourceAuthenticationEntraIDAuthorityURL = ""
    PackageAuthorizationConfig = models.GetInitialAuthorizationConfig_1()
)
