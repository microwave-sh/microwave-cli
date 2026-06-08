package client

import (
	"context"
	"errors"
	"net/url"
)

var (
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrDeviceCodeExpired    = errors.New("device code expired")
)

// ── Permission sets ──────────────────────────────────────────────

func (c *Client) CreatePermissionSet(ctx context.Context, in PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	return &out, c.Do(ctx, "POST", "/api/permission-sets", in, &out)
}

func (c *Client) UpdatePermissionSet(ctx context.Context, id string, in PermissionSetInput) (*PermissionSet, error) {
	var out PermissionSet
	return &out, c.Do(ctx, "PATCH", "/api/permission-sets/"+url.PathEscape(id), in, &out)
}

func (c *Client) DeletePermissionSet(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/permission-sets/"+url.PathEscape(id), nil, nil)
}

func (c *Client) SearchPermissionSets(ctx context.Context, req SearchRequest) (*SearchResponse[PermissionSet], error) {
	return Search[PermissionSet](ctx, c, "/api/permission-sets", req)
}

// ── Key specs ────────────────────────────────────────────────────

func (c *Client) CreateKeySpec(ctx context.Context, in KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	return &out, c.Do(ctx, "POST", "/api/key-specs", in, &out)
}

func (c *Client) UpdateKeySpec(ctx context.Context, id string, in KeySpecInput) (*KeySpec, error) {
	var out KeySpec
	return &out, c.Do(ctx, "PATCH", "/api/key-specs/"+url.PathEscape(id), in, &out)
}

func (c *Client) DeleteKeySpec(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/key-specs/"+url.PathEscape(id), nil, nil)
}

func (c *Client) SearchKeySpecs(ctx context.Context, req SearchRequest) (*SearchResponse[KeySpec], error) {
	return Search[KeySpec](ctx, c, "/api/key-specs", req)
}

func (c *Client) KeySpecEvents(ctx context.Context, id, subject string) ([]KeyEvent, error) {
	path := "/api/key-specs/" + url.PathEscape(id) + "/events"
	if subject != "" {
		path += "?subject=" + url.QueryEscape(subject)
	}
	var out []KeyEvent
	return out, c.Do(ctx, "GET", path, nil, &out)
}

func (c *Client) IssueKey(ctx context.Context, specID string, in IssueKeyInput) (*IssueKeyResult, error) {
	var out IssueKeyResult
	return &out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/keys", in, &out)
}

func (c *Client) SearchSpecKeys(ctx context.Context, specID string, req SearchRequest) (*SearchResponse[IssuedKey], error) {
	return Search[IssuedKey](ctx, c, "/api/key-specs/"+url.PathEscape(specID)+"/keys", req)
}

func (c *Client) RevokeKeysBySubject(ctx context.Context, specID, subject string) (map[string]int, error) {
	var out map[string]int
	return out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/keys/revoke-by-subject", RevokeBySubjectInput{Subject: subject}, &out)
}

func (c *Client) CreateWidgetSession(ctx context.Context, specID string, in WidgetSessionInput) (*WidgetSessionToken, error) {
	var out WidgetSessionToken
	return &out, c.Do(ctx, "POST", "/api/key-specs/"+url.PathEscape(specID)+"/widget-sessions", in, &out)
}

// ── Keys ─────────────────────────────────────────────────────────

func (c *Client) SearchKeys(ctx context.Context, req SearchRequest) (*SearchResponse[IssuedKey], error) {
	return Search[IssuedKey](ctx, c, "/api/keys", req)
}

func (c *Client) GetKey(ctx context.Context, id string) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "GET", "/api/keys/"+url.PathEscape(id), nil, &out)
}

func (c *Client) UpdateKey(ctx context.Context, id string, in UpdateKeyInput) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "PATCH", "/api/keys/"+url.PathEscape(id), in, &out)
}

func (c *Client) RevokeKey(ctx context.Context, id string) (*IssuedKey, error) {
	var out IssuedKey
	return &out, c.Do(ctx, "POST", "/api/keys/"+url.PathEscape(id)+"/revoke", nil, &out)
}

func (c *Client) RotateKey(ctx context.Context, id string, in RotateKeyInput) (*IssueKeyResult, error) {
	var out IssueKeyResult
	return &out, c.Do(ctx, "POST", "/api/keys/"+url.PathEscape(id)+"/rotate", in, &out)
}

func (c *Client) KeyEvents(ctx context.Context, id string) ([]KeyEvent, error) {
	var out []KeyEvent
	return out, c.Do(ctx, "GET", "/api/keys/"+url.PathEscape(id)+"/events", nil, &out)
}

func (c *Client) VerifyKey(ctx context.Context, key string) (*VerifyKeyResult, error) {
	var out VerifyKeyResult
	return &out, c.Do(ctx, "POST", "/api/keys/verify", VerifyKeyInput{Key: key}, &out)
}

// ── Signing key sets ({kind}/{set_name}) ─────────────────────────

// sksPath builds the base path for a signing key set, with an optional suffix.
func sksPath(kind, name, suffix string) string {
	return "/api/signing-key-sets/" + url.PathEscape(kind) + "/" + url.PathEscape(name) + suffix
}

func (c *Client) CreateSigningKeySet(ctx context.Context, in SigningKeySetInput) (*SigningKeySet, error) {
	var out SigningKeySet
	return &out, c.Do(ctx, "POST", "/api/signing-key-sets", in, &out)
}

func (c *Client) SearchSigningKeySets(ctx context.Context, req SearchRequest) (*SearchResponse[SigningKeySet], error) {
	return Search[SigningKeySet](ctx, c, "/api/signing-key-sets", req)
}

func (c *Client) GetSigningKeySet(ctx context.Context, kind, name string) (*SigningKeySetDetail, error) {
	var out SigningKeySetDetail
	return &out, c.Do(ctx, "GET", sksPath(kind, name, ""), nil, &out)
}

func (c *Client) UpdateSigningKeySet(ctx context.Context, kind, name, newName string) (*SigningKeySet, error) {
	var out SigningKeySet
	return &out, c.Do(ctx, "PATCH", sksPath(kind, name, ""), SigningKeySetUpdateInput{Name: newName}, &out)
}

func (c *Client) DeleteSigningKeySet(ctx context.Context, kind, name string) error {
	return c.Do(ctx, "DELETE", sksPath(kind, name, ""), nil, nil)
}

func (c *Client) GenerateSigningKey(ctx context.Context, kind, name string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/generate"), nil, &out)
}

func (c *Client) ActivateSigningKey(ctx context.Context, kind, name, keyID string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/activate"), nil, &out)
}

func (c *Client) RevokeSigningKey(ctx context.Context, kind, name, keyID string) (*SigningKey, error) {
	var out SigningKey
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/revoke"), nil, &out)
}

// SigningKeySecret returns the plaintext secret for a specific symmetric signing key.
// The server returns map[string]string for the per-key secret endpoint (SecretKeyOutput).
func (c *Client) SigningKeySecret(ctx context.Context, kind, name, keyID string) (map[string]string, error) {
	var out map[string]string
	return out, c.Do(ctx, "GET", sksPath(kind, name, "/keys/"+url.PathEscape(keyID)+"/secret"), nil, &out)
}

func (c *Client) SignJWT(ctx context.Context, kind, name string, in SignJWTInput) (*SignJWTResult, error) {
	var out SignJWTResult
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/sign"), in, &out)
}

// SigningKeySetSecret returns the active secret state for a symmetric signing key set.
// Returns SigningKeySecretState (matching dto.SigningKeySecretState from the server).
func (c *Client) SigningKeySetSecret(ctx context.Context, kind, name string) (*SigningKeySecretState, error) {
	var out SigningKeySecretState
	return &out, c.Do(ctx, "GET", sksPath(kind, name, "/secret"), nil, &out)
}

// RotateSigningKeySetSecret rotates the active secret for a symmetric signing key set.
func (c *Client) RotateSigningKeySetSecret(ctx context.Context, kind, name string) (*SigningKeySecretState, error) {
	var out SigningKeySecretState
	return &out, c.Do(ctx, "POST", sksPath(kind, name, "/secret/rotate"), nil, &out)
}

// ── Trust exchanges ──────────────────────────────────────────────

func (c *Client) CreateTrustExchange(ctx context.Context, in TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "POST", "/api/trust-exchanges", in, &out)
}

func (c *Client) SearchTrustExchanges(ctx context.Context, req SearchRequest) (*SearchResponse[TrustExchange], error) {
	return Search[TrustExchange](ctx, c, "/api/trust-exchanges", req)
}

func (c *Client) GetTrustExchange(ctx context.Context, id string) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "GET", "/api/trust-exchanges/"+url.PathEscape(id), nil, &out)
}

func (c *Client) UpdateTrustExchange(ctx context.Context, id string, in TrustExchangeInput) (*TrustExchange, error) {
	var out TrustExchange
	return &out, c.Do(ctx, "PATCH", "/api/trust-exchanges/"+url.PathEscape(id), in, &out)
}

func (c *Client) DeleteTrustExchange(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/trust-exchanges/"+url.PathEscape(id), nil, nil)
}

// ── Trust providers ──────────────────────────────────────────────

func (c *Client) CreateTrustProvider(ctx context.Context, in TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "POST", "/api/trust-providers", in, &out)
}

func (c *Client) SearchTrustProviders(ctx context.Context, req SearchRequest) (*SearchResponse[TrustProvider], error) {
	return Search[TrustProvider](ctx, c, "/api/trust-providers", req)
}

func (c *Client) GetTrustProvider(ctx context.Context, id string) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "GET", "/api/trust-providers/"+url.PathEscape(id), nil, &out)
}

func (c *Client) UpdateTrustProvider(ctx context.Context, id string, in TrustProviderInput) (*TrustProvider, error) {
	var out TrustProvider
	return &out, c.Do(ctx, "PATCH", "/api/trust-providers/"+url.PathEscape(id), in, &out)
}

func (c *Client) DeleteTrustProvider(ctx context.Context, id string) error {
	return c.Do(ctx, "DELETE", "/api/trust-providers/"+url.PathEscape(id), nil, nil)
}

// ── Auth-plane ───────────────────────────────────────────────────

// MintTrustProviderToken calls the public auth-plane mint endpoint.
// apiKey is used as the Bearer token for this call (call on AuthClient with the management key).
func (c *Client) MintTrustProviderToken(ctx context.Context, providerID, apiKey string, in MintTokenInput) (*MintTokenResult, error) {
	prev := c.token
	c.token = apiKey
	defer func() { c.token = prev }()
	var out MintTokenResult
	return &out, c.Do(ctx, "POST", "/trust-providers/"+url.PathEscape(providerID)+"/token", in, &out)
}

// ── Device flow / identity / management keys ─────────────────────

func (c *Client) RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	var out DeviceCodeResponse
	return &out, c.Do(ctx, "POST", "/auth/device", map[string]any{}, &out)
}

func (c *Client) PollDeviceToken(ctx context.Context, deviceCode string) (*DeviceTokenResponse, error) {
	var out DeviceTokenResponse
	form := url.Values{
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		"device_code": {deviceCode},
	}
	if err := c.DoForm(ctx, "POST", "/auth/device/token", form, &out); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			switch apiErr.OAuthError {
			case "authorization_pending":
				return nil, ErrAuthorizationPending
			case "expired_token":
				return nil, ErrDeviceCodeExpired
			}
		}
		return nil, err
	}
	return &out, nil
}

func (c *Client) Me(ctx context.Context) (*Identity, error) {
	var out Identity
	return &out, c.Do(ctx, "GET", "/api/me", nil, &out)
}

func (c *Client) CreateManagementKey(ctx context.Context, in ManagementKeyInput) (*ManagementKeyResult, error) {
	var out ManagementKeyResult
	return &out, c.Do(ctx, "POST", "/api/management-keys", in, &out)
}

func (c *Client) SearchManagementKeys(ctx context.Context, req SearchRequest) (*SearchResponse[IssuedKey], error) {
	return Search[IssuedKey](ctx, c, "/api/management-keys", req)
}

func (c *Client) RevokeManagementKey(ctx context.Context, id string) error {
	return c.Do(ctx, "POST", "/api/management-keys/"+url.PathEscape(id)+"/revoke", map[string]any{}, nil)
}
