package controllers

import (
	"fmt"
	"context"
	"net/http"
	"strings"

	"rewinged/models"
	"rewinged/logging"
	"rewinged/settings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// This middleware sits first in the Auth-Stack and chooses the next handler
// to pass the request along to based on the configured authentication type,
// which currently is either "none" or "microsoftEntraId".
func AuthMiddleware(next http.Handler, authType string) http.Handler {
    switch authType {
    case "none":
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            r = r.WithContext(
                context.WithValue(r.Context(), "groups", []string{}),
            )

            next.ServeHTTP(w, r)
        })
    case "microsoftEntraId":
        return EntraIdAuthMiddleware(next)
    default:
        logging.Logger.Fatal().Msg("sourceAuthType must be either none or microsoftEntraId")
    }

    return nil
}

// Helpful linls:
// https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference
//
func EntraIdAuthMiddleware(next http.Handler) http.Handler {
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
            // maybe this is more of a server-error but let's not tell the client this happened
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
        //allclaims := map[string]interface{}{}
        //if err:= parsedToken.Claims(&allclaims); err != nil {
        //    logging.Logger.Err(err).Msg("failed to parse JWT claims")
        //    http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
        //    return
        //}
        //allclaimsJSON, _ := json.MarshalIndent(allclaims, "", "  ")
        //logging.Logger.Debug().Msg(string(allclaimsJSON))

        // Get groups for authorization
        var claims struct {
            Groups []string `json:"groups"`
            // Group overage claim 1 (if present, user is a member of more groups than fit in the JWT)
            HasGroups bool `json:"hasgroups"`
            // Group overage claim 2 (if present, user is a member of more groups than fit in the JWT)
            ClaimNames struct {
                Groups string `json:"groups"`
            } `json:"_claim_names"`
        }
        if err := parsedToken.Claims(&claims); err != nil {
            logging.Logger.Err(err).Msg("failed to parse JWT groups claim")
            http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
            return
        }

        // https://learn.microsoft.com/en-us/entra/identity-platform/id-token-claims-reference#groups-overage-claim
        if claims.HasGroups || claims.ClaimNames.Groups != "" {
            logging.Logger.Warn().Msg("token contains groups overage claim")
        }

        r = r.WithContext(
            context.WithValue(r.Context(), "groups", claims.Groups),
        )

        next.ServeHTTP(w, r)
    })
}

type AuthorizationRuleDecision int
const (
    Denied AuthorizationRuleDecision = iota
    Unchanged
    Allowed
)
func (ard AuthorizationRuleDecision) String() string {
    switch ard {
    case Denied:
        return "Denied"
    case Unchanged:
        return "Unchanged"
    case Allowed:
        return "Allowed"
    default:
        return "Unknown"
    }
}

func EVALUATE_RULE(rule models.AuthorizationRuleset_1, groups []string) (AuthorizationRuleDecision) {
  if rule.DenyAll || arraysIntersect(rule.Deny, groups) {
    return Denied
  }
  if rule.AllowAll || arraysIntersect(rule.Allow, groups) {
    return Allowed
  }
  return Unchanged
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

func EvaluateGlobalRule(groups []string) AuthorizationRuleDecision {
    switch EVALUATE_RULE(settings.PackageAuthorizationConfig.Global, groups) {
    case Denied:
        logging.Logger.Debug().Str("PackageIdentifier", "").Str("rule", "global").Msg("all packages denied for user")
        // If we hit a global deny, there's no coming back from that. Skip all rule processing.
        return Denied
    case Unchanged:
        return Unchanged
    case Allowed:
        logging.Logger.Debug().Str("PackageIdentifier", "").Str("rule", "global").Msg("all packages allowed for user")
        return Allowed
    }

    logging.Logger.Error().Str("rule", "global").Msg("unexpected return value evaluating rule, this is a bug")
    return Denied
}

// This is the final decision maker, it will return either true if the package is to be allowed or false otherwise
func FilterAuthorizedPackage(decision AuthorizationRuleDecision, packageIdentifier string, groups []string) bool {
    var ruleMatched bool = decision != Unchanged
    for ruleIdx, rule := range settings.PackageAuthorizationConfig.Rules {
        if rule.PackageIdentifier == packageIdentifier && rule.PackageVersion == "" {
            ruleMatched = true
            switch EVALUATE_RULE(rule.AuthorizationRuleset_1, groups) {
            case Denied:
                logging.Logger.Debug().Str("PackageIdentifier", packageIdentifier).Str("rule", fmt.Sprint(ruleIdx)).Msg("package denied for user")
                // On deny, mark the package as not allowed and stop processing it further
                return false
            case Allowed:
                logging.Logger.Debug().Str("PackageIdentifier", packageIdentifier).Str("rule", fmt.Sprint(ruleIdx)).Msg("package allowed for user")
                decision = Allowed
            }
        }
    }

    // If no rule has been hit and decision is Unchanged (not previously Allowed by Global), evaluate the default rule.
    // If a rule has been hit but the decision is stil Unchanged (groups didnt match) we skip the default rule and return false
    // because a per-package rules' Allow-List overrides the Default, so no match means package denied
    return decision == Allowed || !ruleMatched && EVALUATE_RULE(settings.PackageAuthorizationConfig.Default, groups) == Allowed
}

