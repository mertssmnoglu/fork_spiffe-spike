//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/spiffe"
	"github.com/spiffe/spike-sdk-go/validation"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
	"github.com/spiffe/spike/internal/net"
)

func guardPutPolicyRequest(
	request reqres.PolicyCreateRequest, w http.ResponseWriter, r *http.Request,
) error {
	name := request.Name
	spiffeIdPattern := request.SpiffeIdPattern
	pathPattern := request.PathPattern
	permissions := request.Permissions

	spiffeid, err := spiffe.IdFromRequest(r)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateSpiffeId(spiffeid.String())
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	// Request "write" access to the ACL system for the SPIFFE ID.
	allowed := state.CheckAccess(
		spiffeid.String(), "spike/system/acl",
		[]data.PolicyPermission{data.PermissionWrite},
	)
	if !allowed {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrUnauthorized,
		}, w)
		net.Respond(http.StatusUnauthorized, responseBody, w)
		return apiErr.ErrUnauthorized
	}

	err = validation.ValidateName(name)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidateSpiffeIdPattern(spiffeIdPattern)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidatePathPattern(pathPattern)
	if err != nil {
		responseBody :=
			net.MarshalBody(reqres.PolicyCreateResponse{
				Err: data.ErrBadInput,
			}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	err = validation.ValidatePermissions(permissions)
	if err != nil {
		responseBody := net.MarshalBody(reqres.PolicyCreateResponse{
			Err: data.ErrBadInput,
		}, w)
		net.Respond(http.StatusBadRequest, responseBody, w)
		return apiErr.ErrInvalidInput
	}

	return nil
}
