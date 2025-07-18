package controllers

import (
	"fmt"
	"context"
	"net/http"
	"strings"
	"encoding/json"

	"rewinged/models"
	"rewinged/logging"
	"rewinged/settings"

	"github.com/coreos/go-oidc/v3/oidc"
)

func JWTAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rawAuthHeader := r.Header.Get("Authorization")
        if rawAuthHeader == "" {
            logging.Logger.Info().Msg("client request missing Authorization header")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }

        idToken := strings.TrimSpace(strings.Replace(rawAuthHeader, "Bearer", "", 1))
        ctxBg := context.Background()

        provider, err := oidc.NewProvider(ctxBg, settings.SourceAuthenticationEntraIDAuthorityURL)
        if err != nil {
            logging.Logger.Err(err).Msg("could not create OIDC provider")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized) // maybe this is more of a server-error but let's not tell the client this happened
            return
        }
        verifier := provider.Verifier(&oidc.Config{
            ClientID: settings.SourceAuthenticationEntraIDResource,
            SkipIssuerCheck: false, // Validate iss / issuer, see: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-mix-up-mitigation-01
            SkipClientIDCheck: false, // Validate aud / audience / client_id, see: https://datatracker.ietf.org/doc/html/draft-ietf-oauth-mix-up-mitigation-01
        })

        parsedToken, err := verifier.Verify(ctxBg, idToken)
        if err != nil {
            logging.Logger.Err(err).Msg("jwt didn't check out / no valid auth")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }

        // Auth checked out!
        logging.Logger.Debug().Msgf("OIDC token info: User (sub) '%v' from IdP (iss) '%v' authenticated", parsedToken.Subject, parsedToken.Issuer)

        // dump claims for debugging
        allclaims := map[string]interface{}{}
        if err:= parsedToken.Claims(&allclaims); err != nil {
            logging.Logger.Err(err).Msg("failed to parse JWT claims")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }
        allclaimsJSON, _ := json.MarshalIndent(allclaims, "", "  ")
        logging.Logger.Debug().Msg(string(allclaimsJSON))

        // Get groups for authorization
        var claims struct {
            Groups []string `json:"groups"`
        }
        if err := parsedToken.Claims(&claims); err != nil {
            logging.Logger.Err(err).Msg("failed to parse JWT groups claim")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }

        r = r.WithContext(
            context.WithValue(r.Context(), "groups", claims.Groups),
        )

        next.ServeHTTP(w, r)
    })
}

// result:
// -1 denied
//  0 not allowed
//  1 allowed
func EVALUATE_RULE(rule models.AuthorizationRuleset_1, groups []string) (result int) {
  if rule.DenyAll || arraysIntersect(rule.Deny, groups) {
    return -1
  }
  if rule.AllowAll || arraysIntersect(rule.Allow, groups) {
    return 1
  }
  return 0
}

func arraysIntersect[A comparable](arr1 []A, arr2 []A) (bool) {
    set := make(map[A]bool)
    for _, elem := range arr1 {
        set[elem] = true
    }
    for _, elem := range arr2 {
        if set[elem] {
            return true
        }
    }
    return false
}

func GetFilterInitialValue(groups []string) (bool, bool) {
    var globallyDenied bool = false
    var initialAllowValue bool = false

    switch EVALUATE_RULE(settings.PackageAuthorizationConfig.Global, groups) {
    case -1:
        logging.Logger.Debug().Str("PackageIdentifier", "").Str("rule", "global").Msg("all packages denied for user")
        // If we hit a global deny, there's no coming back from that. Skip all rule processing.
        globallyDenied = true
    case 0:
        // If the global rule doesn't affect this user, we start all packages with the decision of the default rule
        initialAllowValue = EVALUATE_RULE(settings.PackageAuthorizationConfig.Default, groups) == 1
    case 1:
        logging.Logger.Debug().Str("PackageIdentifier", "").Str("rule", "global").Msg("all packages allowed for user")
        initialAllowValue = true
    }

    return globallyDenied, initialAllowValue
}

func FilterAuthorizedPackage(decision *bool, packageIdentifier string, groups []string) {
    for ruleIdx, rule := range settings.PackageAuthorizationConfig.Rules {
        if rule.PackageIdentifier == packageIdentifier && rule.PackageVersion == "" {
            switch EVALUATE_RULE(rule.AuthorizationRuleset_1, groups) {
            case -1:
                logging.Logger.Debug().Str("PackageIdentifier", packageIdentifier).Str("rule", fmt.Sprint(ruleIdx)).Msg("package denied for user")
                // On deny, mark the package as not allowed and stop processing it further
                *decision = false
                return
            case 1:
                logging.Logger.Debug().Str("PackageIdentifier", packageIdentifier).Str("rule", fmt.Sprint(ruleIdx)).Msg("package allowed for user")
                *decision = true
            }
        }
    }
}

/*
// How to use the above 2 functions:

allowedPackages := []models.API_Package{}
globallyDenied, initialAllowValue := GetFilterInitialValue(groups)

if !globallyDenied {
    for _, pkg := range models.Manifests.GetAllPackageIdentifiers() {
        var packageAllowed bool = initialAllowValue
        FilterAuthorizedPackage(&packageAllowed, pkg.PackageIdentifier, groups)
        if packageAllowed {
            allowedPackages = append(allowedPackages, pkg)
        }
    }
}
*/
