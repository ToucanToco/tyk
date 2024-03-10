package swagger

import (
	"net/http"

	"github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"

	"github.com/TykTechnologies/tyk/gateway"
)

const OAuthTag = "OAuth"

func OAuthApi(r *openapi3.Reflector) error {
	///err := revokeTokenHandler(r)
	err := rotateOauthClientHandler(r)
	if err != nil {
		return err
	}
	err = invalidateOauthRefresh(r)
	if err != nil {
		return err
	}
	err = updateOauthClient(r)
	if err != nil {
		return err
	}
	err = getApisForOauthApp(r)
	if err != nil {
		return err
	}

	err = purgeLapsedOAuthTokens(r)
	if err != nil {
		return err
	}
	err = listOAuthClients(r)
	if err != nil {
		return err
	}

	err = deleteOAuthClient(r)
	if err != nil {
		return err
	}
	err = getSingleOAuthClient(r)
	if err != nil {
		return err
	}
	return createOauthClient(r)
}

func createOauthClient(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodPost, "/tyk/oauth/clients/create")
	if err != nil {
		return err
	}
	oc.SetTags(OAuthTag)
	oc.AddReqStructure(new(gateway.NewClientRequest))
	oc.AddRespStructure(new(gateway.NewClientRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	// TODO::ask why we return 500 instead of 400 for wrong body
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.SetID("createOAuthClient")
	oc.SetSummary("Create new OAuth client")
	oc.SetDescription("Any OAuth keys must be generated with the help of a client ID. These need to be pre-registered with Tyk before they can be used (in a similar vein to how you would register your app with Twitter before attempting to ask user permissions using their API).\n        <br/><br/>\n        <h3>Creating OAuth clients with Access to Multiple APIs</h3>\n        New from Tyk Gateway 2.6.0 is the ability to create OAuth clients with access to more than one API. If you provide the api_id it works the same as in previous releases. If you don't provide the api_id the request uses policy access rights and enumerates APIs from their setting in the newly created OAuth-client.\n")
	return r.AddOperation(oc)
}

func rotateOauthClientHandler(r *openapi3.Reflector) error {
	// TODO::find summary and description for this
	oc, err := r.NewOperationContext(http.MethodPut, "/tyk/oauth/clients/{apiID}/{keyName}/rotate")
	if err != nil {
		return err
	}
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusOK))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.SetID("rotateOauthClient")
	oc.SetSummary("Rotate the oath client")
	oc.SetTags(OAuthTag)
	par := []openapi3.ParameterOrRef{keyNameParameter(), oauthApiIdParameter()}
	o3.Operation().WithParameters(par...)
	return r.AddOperation(oc)
}

func invalidateOauthRefresh(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodDelete, "/tyk/oauth/refresh/{keyName}")
	if err != nil {
		return err
	}
	oc.SetTags(OAuthTag)
	oc.SetID("invalidateOAuthRefresh")
	oc.SetSummary("Invalidate OAuth refresh token")
	oc.SetDescription("It is possible to invalidate refresh tokens in order to manage OAuth client access more robustly.")
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiModifyKeySuccess), openapi.WithHTTPStatus(http.StatusOK))
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{keyNameParameter(), requiredApiIdQuery()}
	o3.Operation().WithParameters(par...)
	return r.AddOperation(oc)
}

func updateOauthClient(r *openapi3.Reflector) error {
	// TODO:: in previous OAs this was '/tyk/oauth/clients/{apiID}' inquire
	oc, err := r.NewOperationContext(http.MethodPut, "/tyk/oauth/clients/{apiID}/{keyName}")
	if err != nil {
		return err
	}
	oc.AddReqStructure(new(gateway.NewClientRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.AddRespStructure(new(gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusOK))
	oc.SetID("updateOAuthClient")
	oc.SetSummary("Update OAuth metadata and Policy ID")
	oc.SetDescription("Allows you to update the metadata and Policy ID for an OAuth client.")
	oc.SetTags(OAuthTag)
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{keyNameParameter(), oauthApiIdParameter()}
	o3.Operation().WithParameters(par...)
	return r.AddOperation(oc)
}

func getApisForOauthApp(r *openapi3.Reflector) error {
	// TODO::This is has org_id as form value need to find a way to fix it
	// TODO:: Go over this again
	oc, err := r.NewOperationContext(http.MethodGet, "/tyk/oauth/clients/apis/{appID}")
	if err != nil {
		return err
	}
	oc.SetTags(OAuthTag)
	oc.SetID("getApisForOauthApp")
	oc.SetSummary("Get Apis for Oauth app")
	oc.SetDescription("Get Apis for Oauth app")
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new([]string), openapi.WithHTTPStatus(http.StatusOK))
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{appIDParameter()}
	o3.Operation().WithParameters(par...)
	return r.AddOperation(oc)
}

func purgeLapsedOAuthTokens(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodDelete, "/tyk/oauth/tokens")
	if err != nil {
		return err
	}
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusUnprocessableEntity))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiStatusMessage))
	oc.SetID("purgeLapsedOAuthTokens")
	oc.SetSummary("Purge lapsed OAuth tokens")
	oc.SetDescription("Purge scoped lapsed OAuth token")
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{scopeQuery()}
	o3.Operation().WithParameters(par...)

	oc.SetTags(OAuthTag)
	return r.AddOperation(oc)
}

func listOAuthClients(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodDelete, "/tyk/oauth/clients/{apiID}")
	if err != nil {
		return err
	}
	oc.SetTags(OAuthTag)
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new([]gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusOK))
	// TODO:: ask why 404 returns null
	oc.AddRespStructure(new(*[]gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new([]gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusOK))
	oc.SetID("listOAuthClients")
	oc.SetSummary("List oAuth clients")
	oc.SetDescription("OAuth Clients are organised by API ID, and therefore are queried as such.")
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{oauthApiIdParameter()}
	o3.Operation().WithParameters(par...)

	return r.AddOperation(oc)
}

func deleteOAuthClient(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodDelete, "/tyk/oauth/clients/{apiID}/{keyName}")
	if err != nil {
		return err
	}
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiModifyKeySuccess), openapi.WithHTTPStatus(http.StatusOK))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.SetTags(OAuthTag)
	oc.SetID("deleteOAuthClient")
	oc.SetSummary("Delete OAuth client")
	oc.SetDescription("Please note that tokens issued with the client ID will still be valid until they expire.")
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{oauthApiIdParameter(), keyNameParameter()}
	o3.Operation().WithParameters(par...)

	return r.AddOperation(oc)
}

func getSingleOAuthClient(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodGet, "/tyk/oauth/clients/{apiID}/{keyName}")
	if err != nil {
		return err
	}
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusNotFound))
	oc.AddRespStructure(new(gateway.NewClientRequest), openapi.WithHTTPStatus(http.StatusOK))
	// TODO::returned when basing dowritejsonfails
	// oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusInternalServerError))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.SetID("getOAuthClient")
	oc.SetSummary("Get OAuth client")
	oc.SetDescription("Get OAuth client details")
	oc.SetTags(OAuthTag)
	o3, ok := oc.(openapi3.OperationExposer)
	if !ok {
		return ErrOperationExposer
	}
	par := []openapi3.ParameterOrRef{oauthApiIdParameter(), keyNameParameter()}
	o3.Operation().WithParameters(par...)

	return r.AddOperation(oc)
}

func revokeTokenHandler(r *openapi3.Reflector) error {
	oc, err := r.NewOperationContext(http.MethodPost, "/tyk/oauth/revoke")
	if err != nil {
		return err
	}
	// TODO::This is totally wrong find out how to do it
	oc.AddReqStructure(new(struct {
		Token         string `json:"token"`
		TokenTypeHint string `json:"token_type_hint"`
		ClientID      string `json:"client_id"`
		OrgID         string `json:"org_id"`
	}), func(cu *openapi.ContentUnit) {
		cu.SetFieldMapping(openapi.InFormData, map[string]string{
			"token_type_hint": "value3",
			"token":           "upload6",
			"client_id":       "",
			"org_id":          "",
		})
	})
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusBadRequest))
	oc.AddRespStructure(new(apiStatusMessage), openapi.WithHTTPStatus(http.StatusForbidden))
	oc.AddRespStructure(new(apiStatusMessage))
	oc.SetTags(OAuthTag)
	return r.AddOperation(oc)
}

func keyNameParameter() openapi3.ParameterOrRef {
	desc := "Refresh token"
	return openapi3.Parameter{In: openapi3.ParameterInPath, Name: "keyName", Required: &isRequired, Description: &desc, Schema: stringSchema()}.ToParameterOrRef()
}

func oauthApiIdParameter() openapi3.ParameterOrRef {
	return openapi3.Parameter{In: openapi3.ParameterInPath, Name: "apiID", Required: &isRequired, Schema: stringSchema()}.ToParameterOrRef()
}

func requiredApiIdQuery() openapi3.ParameterOrRef {
	desc := "The API id"
	return openapi3.Parameter{In: openapi3.ParameterInQuery, Name: "api_id", Required: &isRequired, Description: &desc, Schema: stringSchema()}.ToParameterOrRef()
}

func appIDParameter() openapi3.ParameterOrRef {
	return openapi3.Parameter{In: openapi3.ParameterInPath, Name: "appID", Required: &isRequired, Schema: stringSchema()}.ToParameterOrRef()
}

func scopeQuery() openapi3.ParameterOrRef {
	stringType := openapi3.SchemaTypeString
	return openapi3.Parameter{In: openapi3.ParameterInQuery, Name: "scope", Required: &isRequired, Schema: &openapi3.SchemaOrRef{
		Schema: &openapi3.Schema{
			Type: &stringType,
			Enum: []interface{}{"lapsed"},
		},
	}}.ToParameterOrRef()
}
