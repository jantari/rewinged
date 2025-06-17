package controllers

import (
	"context"
	"net/http"
	"strings"

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

        //claims := map[string]interface{}{}
        //if parsedToken.Claims(&claims); err != nil {
        //    logging.Logger.Err(err).Msg("failed to parse JWT claims")
        //}
        //claimsJSON, _ := json.MarshalIndent(claims, "", "  ")
        //logging.Logger.Debug().Msg(string(claimsJSON))

        next.ServeHTTP(w, r)
    })
}

