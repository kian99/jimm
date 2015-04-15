package v1

import (
	"net/http"
	"strings"

	"github.com/CanonicalLtd/jem/params"
	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v0/bakery"
	"gopkg.in/macaroon-bakery.v0/bakery/checkers"
	"gopkg.in/macaroon-bakery.v0/httpbakery"
	"gopkg.in/macaroon.v1"
)

const (
	usernameAttr = "username"
	groupsAttr   = "groups"
)

// authorization conatains authorization information extracted from an HTTP request.
// The zero value for a authorization contains no privileges.
type authorization struct {
	Username string
	Groups   []string
}

// checkRequest checks for any authorization tokens in the request and returns any
// found as an authorization. If no suitable credentials are found, or an error occurs,
// then a zero valued authorization is returned.
func (h *Handler) checkRequest(req *http.Request) (authorization, error) {
	attrMap, verr := httpbakery.CheckRequest(h.db.Bakery, req, nil, checkers.New())
	if verr == nil {
		return authorization{
			Username: attrMap[usernameAttr],
			Groups:   strings.Fields(attrMap[groupsAttr]),
		}, nil
	}
	if _, ok := errgo.Cause(verr).(*bakery.VerificationError); !ok {
		return authorization{}, errgo.Mask(verr, errgo.Is(params.ErrUnauthorized))
	}
	// Macaroon verification failed: mint a new macaroon.
	m, err := h.newMacaroon()
	if err != nil {
		return authorization{}, errgo.Notef(err, "cannot mint macaroon")
	}
	// Request that this macaroon be supplied for all requests
	// to the whole handler.
	// TODO use a relative URL here: router.RelativeURLPath(req.RequestURI, "/")
	cookiePath := "/"
	return authorization{}, httpbakery.NewDischargeRequiredError(m, cookiePath, verr)
}

func (h *Handler) newMacaroon() (*macaroon.Macaroon, error) {
	return h.db.Bakery.NewMacaroon("", nil, []checkers.Caveat{
		checkers.NeedDeclaredCaveat(
			checkers.Caveat{
				Location:  h.config.IdentityLocation,
				Condition: "is-authenticated-user",
			},
			usernameAttr,
			groupsAttr,
		),
	})
}
