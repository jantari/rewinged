package settings

import (
    "net/netip"
)

var (
    TrustedProxies []netip.Prefix = []netip.Prefix{}
    SourceAuthenticationType = "none"
    SourceAuthenticationEntraIDResource = ""
    SourceAuthenticationEntraIDAuthorityURL = ""
)
