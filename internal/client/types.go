package client

import "time"

// ── Permission sets ──────────────────────────────────────────────

// PermissionInput is used when creating/updating permissions within a set.
type PermissionInput struct {
	ID              string `json:"id,omitempty"`
	PermissionSetID string `json:"permission_set_id,omitempty"`
	Name            string `json:"name"`
	Label           string `json:"label"`
	Description     string `json:"description,omitempty"`
	Dangerous       bool   `json:"dangerous"`
}

// Permission is the full permission object returned by the server (includes audit timestamps).
type Permission struct {
	ID              string    `json:"id"`
	PermissionSetID string    `json:"permission_set_id"`
	Name            string    `json:"name"`
	Label           string    `json:"label"`
	Description     string    `json:"description,omitempty"`
	Dangerous       bool      `json:"dangerous"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// PermissionSetInput is the request body for creating or updating a permission set.
// Uses PermissionInput (not Permission) to match the server DTO.
type PermissionSetInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Permissions []PermissionInput `json:"permissions"`
}

// PermissionSet is the response shape for a permission set resource.
type PermissionSet struct {
	ID          string       `json:"id"`
	WorkspaceID string       `json:"workspace_id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// ── Key specs ────────────────────────────────────────────────────

type OpaqueConfig struct {
	Prefix         string `json:"prefix"`
	LookupResponse string `json:"lookup_response"`
}

type JWTConfig struct {
	Algorithm string `json:"algorithm"`
	Issuer    string `json:"issuer"`
	Audience  string `json:"audience"`
}

type ExpiryPolicy struct {
	DefaultTTL           string `json:"default_ttl"`
	MaxTTL               string `json:"max_ttl"`
	AllowNever           bool   `json:"allow_never"`
	RotationReminderDays int    `json:"rotation_reminder_days"`
}

type ClaimField struct {
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type ClaimsConfig struct {
	Standard []string              `json:"standard"`
	Custom   map[string]ClaimField `json:"custom"`
}

type OverridePolicy struct {
	AllowCustomExpiry bool `json:"allow_custom_expiry"`
	AllowCustomScopes bool `json:"allow_custom_scopes"`
	AllowCustomClaims bool `json:"allow_custom_claims"`
}

type WebhookConfig struct {
	Endpoint string   `json:"endpoint,omitempty"`
	Events   []string `json:"events,omitempty"`
}

type KeySpecInput struct {
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 string         `json:"format"` // "opaque" | "jwt"
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	SigningKeySetID         string         `json:"signing_key_set_id,omitempty"`
	Opaque                 OpaqueConfig   `json:"opaque,omitempty"`
	JWT                    JWTConfig      `json:"jwt,omitempty"`
	Expiry                 ExpiryPolicy   `json:"expiry"`
	Claims                 ClaimsConfig   `json:"claims"`
	OverridePolicy         OverridePolicy `json:"override_policy"`
	Webhooks               WebhookConfig  `json:"webhooks"`
	WebhookSigningKeySetID string         `json:"webhook_signing_key_set_id,omitempty"`
}

// KeySpec is the full response shape, matching the server DTO exactly (includes nested config fields).
type KeySpec struct {
	ID                     string         `json:"id"`
	WorkspaceID            string         `json:"workspace_id"`
	Name                   string         `json:"name"`
	Description            string         `json:"description,omitempty"`
	Format                 string         `json:"format"`
	PermissionSetID        string         `json:"permission_set_id,omitempty"`
	PermissionSet          *PermissionSet `json:"permission_set,omitempty"`
	SigningKeySetID         string         `json:"signing_key_set_id,omitempty"`
	Opaque                 OpaqueConfig   `json:"opaque,omitempty"`
	JWT                    JWTConfig      `json:"jwt,omitempty"`
	Expiry                 ExpiryPolicy   `json:"expiry"`
	Claims                 ClaimsConfig   `json:"claims"`
	OverridePolicy         OverridePolicy `json:"override_policy"`
	Webhooks               WebhookConfig  `json:"webhooks"`
	WebhookSigningKeySetID string         `json:"webhook_signing_key_set_id,omitempty"`
	WidgetURL              string         `json:"widget_url"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

// ── Issued keys ──────────────────────────────────────────────────

type IssueKeyInput struct {
	Subject   string         `json:"subject"`
	Name      string         `json:"name"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresIn string         `json:"expires_in,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// IssueKeyResult is returned by the issue and rotate endpoints. Includes the plaintext key.
type IssueKeyResult struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	KeyHint   string     `json:"key_hint"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Subject   string     `json:"subject"`
	SpecID    string     `json:"spec_id"`
	Scopes    []string   `json:"scopes"`
	CreatedAt time.Time  `json:"created_at"`
}

// IssuedKey is the persistent key record (no plaintext key value).
type IssuedKey struct {
	ID             string         `json:"id"`
	SpecID         string         `json:"spec_id"`
	WorkspaceID    string         `json:"workspace_id"`
	Subject        string         `json:"subject"`
	Name           string         `json:"name"`
	KeyHint        string         `json:"key_hint"`
	Scopes         []string       `json:"scopes"`
	Claims         map[string]any `json:"claims"`
	Metadata       map[string]any `json:"metadata"`
	Status         string         `json:"status"`
	ExpiresAt      *time.Time     `json:"expires_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	LastVerifiedAt *time.Time     `json:"last_verified_at,omitempty"`
	RevokedAt      *time.Time     `json:"revoked_at,omitempty"`
}

// UpdateKeyInput is the PATCH body for updating a key. Name is a plain string (not pointer) per server DTO.
type UpdateKeyInput struct {
	Name      string         `json:"name,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
}

type RotateKeyInput struct {
	OverlapSeconds int `json:"overlap_seconds,omitempty"`
}

type VerifyKeyInput struct {
	Key string `json:"key"`
}

// VerifyKeyResult matches the server DTO from dto.VerifyKeyResult.
type VerifyKeyResult struct {
	Valid     bool           `json:"valid"`
	Code      string         `json:"code,omitempty"`
	KeyID     string         `json:"key_id,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Scopes    []string       `json:"scopes,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
	ExpiresAt *time.Time     `json:"expires_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	JWT       string         `json:"jwt,omitempty"`
}

type KeyEvent struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	SpecID    string    `json:"spec_id"`
	Subject   string    `json:"subject"`
	Type      string    `json:"type"`
	IP        string    `json:"ip,omitempty"`
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
}

type RevokeBySubjectInput struct {
	Subject string `json:"subject"`
}

// WidgetSessionInput matches domain.CreateWidgetSessionInput.
type WidgetSessionInput struct {
	Subject     string         `json:"subject"`
	Claims      map[string]any `json:"claims,omitempty"`
	Scopes      []string       `json:"scopes,omitempty"`
	RedirectURL string         `json:"redirect_url,omitempty"`
	TTL         string         `json:"ttl"`
}

type WidgetSessionToken struct {
	SessionToken string `json:"session_token"`
}

// ── Signing key sets ─────────────────────────────────────────────

// SigningKeySetInput is the request body for creating a signing key set.
// The server handler uses {name, kind, algorithm} fields from signingkeysets.CreateSigningKeySetBody.
type SigningKeySetInput struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`      // "asymmetric" | "symmetric"
	Algorithm string `json:"algorithm"` // RS256/ES256/HS256/...
}

// SigningKeySetUpdateInput is the PATCH body for renaming a signing key set.
type SigningKeySetUpdateInput struct {
	Name string `json:"name"`
}

type SigningKeySet struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Kind      string     `json:"kind"`
	Algorithm string     `json:"algorithm"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// AsymmetricPublicMaterial carries JWK public key material. JSON tags are required for correct
// serialization/deserialization — these match the server dto.AsymmetricPublicMaterial exactly.
type AsymmetricPublicMaterial struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
	Crv string `json:"crv,omitempty"`
}

type SigningKey struct {
	ID                       string                    `json:"id"`
	SetID                    string                    `json:"set_id"`
	Status                   string                    `json:"status"`
	AsymmetricPublicMaterial *AsymmetricPublicMaterial `json:"asymmetric_public_material,omitempty"`
	SecretRef                string                    `json:"secret_ref,omitempty"`
	CreatedAt                time.Time                 `json:"created_at"`
	RevokedAt                *time.Time                `json:"revoked_at,omitempty"`
}

type SigningKeySetDetail struct {
	Set  SigningKeySet `json:"set"`
	Keys []SigningKey  `json:"keys"`
}

// SigningKeySecretState is returned by the signing-key-set secret endpoints.
// Matches dto.SigningKeySecretState from the server.
type SigningKeySecretState struct {
	Secret         string `json:"secret"`
	ActiveKeyID    string `json:"active_key_id"`
	PreviousSecret string `json:"previous_secret,omitempty"`
	PreviousKeyID  string `json:"previous_key_id,omitempty"`
}

type SignJWTInput struct {
	Payload map[string]any `json:"payload"`
	KID     string         `json:"kid,omitempty"`
	Header  map[string]any `json:"header,omitempty"`
}

type SignJWTResult struct {
	JWT string `json:"jwt"`
}

// ── Trust exchanges ──────────────────────────────────────────────

type TrustExchangeSubjectRules struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
}

type TrustExchangeClaimRule struct {
	Equals   string   `json:"equals,omitempty"`
	OneOf    []string `json:"one_of,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	Required bool     `json:"required,omitempty"`
}

type TrustExchangeClaimMapping struct {
	SubjectClaim string            `json:"subject_claim,omitempty"`
	Scopes       []string          `json:"scopes,omitempty"`
	Claims       map[string]string `json:"claims,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type TrustExchangeInput struct {
	Name             string                            `json:"name"`
	Description      string                            `json:"description,omitempty"`
	Type             string                            `json:"type"`     // "oidc"
	Provider         string                            `json:"provider"` // github|google|auth0|custom_oidc
	Issuer           string                            `json:"issuer"`
	DiscoveryURL     string                            `json:"discovery_url,omitempty"`
	JWKSURL          string                            `json:"jwks_url,omitempty"`
	AllowedAudiences []string                          `json:"allowed_audiences"`
	SubjectRules     TrustExchangeSubjectRules         `json:"subject_rules"`
	ClaimRules       map[string]TrustExchangeClaimRule `json:"claim_rules"`
	ClaimMapping     TrustExchangeClaimMapping         `json:"claim_mapping"`
	OutputMode       string                            `json:"output_mode"` // "claims"|"jwt"
	OutputKeySpecID  string                            `json:"output_key_spec_id,omitempty"`
	Active           bool                              `json:"active"`
}

// TrustExchange is the response shape. Defined as a FLAT struct (not embedding TrustExchangeInput)
// because the server dto.TrustExchange uses explicit fields — embedding would cause duplicate json tags
// and possible field-name conflicts with future server changes.
type TrustExchange struct {
	ID               string                            `json:"id"`
	Name             string                            `json:"name"`
	Description      string                            `json:"description,omitempty"`
	Type             string                            `json:"type"`
	Provider         string                            `json:"provider"`
	Issuer           string                            `json:"issuer"`
	DiscoveryURL     string                            `json:"discovery_url,omitempty"`
	JWKSURL          string                            `json:"jwks_url,omitempty"`
	AllowedAudiences []string                          `json:"allowed_audiences"`
	SubjectRules     TrustExchangeSubjectRules         `json:"subject_rules"`
	ClaimRules       map[string]TrustExchangeClaimRule `json:"claim_rules"`
	ClaimMapping     TrustExchangeClaimMapping         `json:"claim_mapping"`
	OutputMode       string                            `json:"output_mode"`
	OutputKeySpecID  string                            `json:"output_key_spec_id,omitempty"`
	Active           bool                              `json:"active"`
	CreatedAt        time.Time                         `json:"created_at"`
	UpdatedAt        time.Time                         `json:"updated_at"`
}

// ── Trust providers ──────────────────────────────────────────────

type TrustProviderClaimPolicy struct {
	Constant map[string]any `json:"constant,omitempty"`
	Allowed  []string       `json:"allowed,omitempty"`
	Required []string       `json:"required,omitempty"`
}

type TrustProviderInput struct {
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	Type              string                   `json:"type"` // "oidc"
	SigningKeySetID    string                   `json:"signing_key_set_id"`
	IssuerHost        string                   `json:"issuer_host,omitempty"`
	AllowedAudiences  []string                 `json:"allowed_audiences"`
	DefaultAudience   string                   `json:"default_audience,omitempty"`
	ClaimPolicy       TrustProviderClaimPolicy `json:"claim_policy"`
	SubjectRequired   bool                     `json:"subject_required"`
	TTLDefaultSeconds int64                    `json:"ttl_default_seconds,omitempty"`
	TTLMaxSeconds     int64                    `json:"ttl_max_seconds,omitempty"`
	Active            bool                     `json:"active"`
}

// TrustProvider is the response shape. Defined as a FLAT struct (not embedding TrustProviderInput)
// because the server dto.TrustProvider uses explicit fields and includes non-omitempty
// TTL fields in the response (ttl_default_seconds/ttl_max_seconds always present).
type TrustProvider struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description,omitempty"`
	Type              string                   `json:"type"`
	SigningKeySetID    string                   `json:"signing_key_set_id"`
	IssuerHost        string                   `json:"issuer_host,omitempty"`
	AllowedAudiences  []string                 `json:"allowed_audiences"`
	DefaultAudience   string                   `json:"default_audience,omitempty"`
	ClaimPolicy       TrustProviderClaimPolicy `json:"claim_policy"`
	SubjectRequired   bool                     `json:"subject_required"`
	TTLDefaultSeconds int64                    `json:"ttl_default_seconds"`
	TTLMaxSeconds     int64                    `json:"ttl_max_seconds"`
	Active            bool                     `json:"active"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

// MintTokenInput is the request body for minting a trust provider token.
// Matches domain.MintTrustProviderTokenInput (Subject is required per server).
type MintTokenInput struct {
	Subject    string         `json:"subject"`
	Audience   string         `json:"audience,omitempty"`
	Claims     map[string]any `json:"claims,omitempty"`
	TTLSeconds int64          `json:"ttl_seconds,omitempty"`
}

// MintTokenResult matches domain.MintTrustProviderTokenResult.
type MintTokenResult struct {
	Valid     bool           `json:"valid"`
	Code      string         `json:"code,omitempty"`
	Token     string         `json:"token,omitempty"`
	Issuer    string         `json:"issuer,omitempty"`
	Subject   string         `json:"subject,omitempty"`
	Audience  string         `json:"audience,omitempty"`
	ExpiresIn int64          `json:"expires_in,omitempty"`
	Claims    map[string]any `json:"claims,omitempty"`
}
