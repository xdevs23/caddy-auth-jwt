// Copyright 2020 Paul Greenberg greenpau@outlook.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package user

import (
	"encoding/json"
	"fmt"
	"github.com/greenpau/caddy-authorize/pkg/errors"
	"github.com/greenpau/caddy-authorize/pkg/utils/cfgutils"
	"strings"
	"time"
)

/*
var reservedFields = map[string]interface{}{
	"email":        true,
	"role":         true,
	"groups":       true,
	"group":        true,
	"app_metadata": true,
	"realm_access": true,
	"paths":        true,
	"acl":          true,
}
*/

// User is a user with claims and status.
type User struct {
	Claims          *Claims       `json:"claims,omitempty" xml:"claims,omitempty" yaml:"claims,omitempty"`
	Token           string        `json:"token,omitempty" xml:"token,omitempty" yaml:"token,omitempty"`
	TokenName       string        `json:"token_name,omitempty" xml:"token_name,omitempty" yaml:"token_name,omitempty"`
	TokenSource     string        `json:"token_source,omitempty" xml:"token_source,omitempty" yaml:"token_source,omitempty"`
	Authenticator   Authenticator `json:"authenticator,omitempty" xml:"authenticator,omitempty" yaml:"authenticator,omitempty"`
	Checkpoints     []*Checkpoint `json:"checkpoints,omitempty" xml:"checkpoints,omitempty" yaml:"checkpoints,omitempty"`
	Authorized      bool          `json:"authorized,omitempty" xml:"authorized,omitempty" yaml:"authorized,omitempty"`
	FrontendLinks   []string      `json:"frontend_links,omitempty" xml:"frontend_links,omitempty" yaml:"frontend_links,omitempty"`
	Locked          bool          `json:"locked,omitempty" xml:"locked,omitempty" yaml:"locked,omitempty"`
	requestHeaders  map[string]string
	requestIdentity map[string]interface{}
	Cached          bool `json:"cached,omitempty" xml:"cached,omitempty" yaml:"cached,omitempty"`
	// Holds the map for all the claims parsed from a token.
	mkv map[string]interface{}
	// Holds the map for a subset of claims necessary for ACL evaluation.
	tkv map[string]interface{}
	// Holds the map of the user roles.
	rkv map[string]interface{}
}

// Checkpoint represents additional checks that a user needs to pass. Once
// a user passes the checks, the Authorized is set to true. The checks
// could be the acceptance of the terms of use, multi-factor authentication,
// etc.
type Checkpoint struct {
	ID             int    `json:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`
	Name           string `json:"name,omitempty" xml:"name,omitempty" yaml:"name,omitempty"`
	Type           string `json:"type,omitempty" xml:"type,omitempty" yaml:"type,omitempty"`
	Parameters     string `json:"parameters,omitempty" xml:"parameters,omitempty" yaml:"parameters,omitempty"`
	Passed         bool   `json:"passed,omitempty" xml:"passed,omitempty" yaml:"passed,omitempty"`
	FailedAttempts int    `json:"failed_attempts,omitempty" xml:"failed_attempts,omitempty" yaml:"failed_attempts,omitempty"`
}

// Authenticator represents authentication backend
type Authenticator struct {
	Name          string `json:"name,omitempty" xml:"name,omitempty" yaml:"name,omitempty"`
	Realm         string `json:"realm,omitempty" xml:"realm,omitempty" yaml:"realm,omitempty"`
	Method        string `json:"method,omitempty" xml:"method,omitempty" yaml:"method,omitempty"`
	TempSecret    string `json:"temp_secret,omitempty" xml:"temp_secret,omitempty" yaml:"temp_secret,omitempty"`
	TempSessionID string `json:"temp_session_id,omitempty" xml:"temp_session_id,omitempty" yaml:"temp_session_id,omitempty"`
	URL           string `json:"url,omitempty" xml:"url,omitempty" yaml:"url,omitempty"`
}

// Claims represents custom and standard JWT claims associated with User.
type Claims struct {
	Audience      []string               `json:"aud,omitempty" xml:"aud,omitempty" yaml:"aud,omitempty"`
	ExpiresAt     int64                  `json:"exp,omitempty" xml:"exp,omitempty" yaml:"exp,omitempty"`
	ID            string                 `json:"jti,omitempty" xml:"jti,omitempty" yaml:"jti,omitempty"`
	IssuedAt      int64                  `json:"iat,omitempty" xml:"iat,omitempty" yaml:"iat,omitempty"`
	Issuer        string                 `json:"iss,omitempty" xml:"iss,omitempty" yaml:"iss,omitempty"`
	NotBefore     int64                  `json:"nbf,omitempty" xml:"nbf,omitempty" yaml:"nbf,omitempty"`
	Subject       string                 `json:"sub,omitempty" xml:"sub,omitempty" yaml:"sub,omitempty"`
	Name          string                 `json:"name,omitempty" xml:"name,omitempty" yaml:"name,omitempty"`
	Email         string                 `json:"email,omitempty" xml:"email,omitempty" yaml:"email,omitempty"`
	Roles         []string               `json:"roles,omitempty" xml:"roles,omitempty" yaml:"roles,omitempty"`
	Origin        string                 `json:"origin,omitempty" xml:"origin,omitempty" yaml:"origin,omitempty"`
	Scopes        []string               `json:"scopes,omitempty" xml:"scopes,omitempty" yaml:"scopes,omitempty"`
	Organizations []string               `json:"org,omitempty" xml:"org,omitempty" yaml:"org,omitempty"`
	AccessList    *AccessListClaim       `json:"acl,omitempty" xml:"acl,omitempty" yaml:"acl,omitempty"`
	Address       string                 `json:"addr,omitempty" xml:"addr,omitempty" yaml:"addr,omitempty"`
	PictureURL    string                 `json:"picture,omitempty" xml:"picture,omitempty" yaml:"picture,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" xml:"metadata,omitempty" yaml:"metadata,omitempty"`
	Username      string		     `json:"username,omitempty" xml:"username,omitempty" yaml:"username,omitempty"`
}

// AccessListClaim represents custom acl/paths claim
type AccessListClaim struct {
	Paths map[string]interface{} `json:"paths,omitempty" xml:"paths,omitempty" yaml:"paths,omitempty"`
}

// Valid validates user claims.
func (c Claims) Valid() error {
	if c.ExpiresAt < time.Now().Unix() {
		return errors.ErrExpiredToken
	}
	return nil
}

/*
func (c *Claims) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if len(c.Audience) > 0 {
		m["aud"] = c.Audience
	}
	if c.ExpiresAt > 0 {
		m["exp"] = c.ExpiresAt
	}
	if c.IssuedAt > 0 {
		m["iat"] = c.IssuedAt
	}
	if c.NotBefore > 0 {
		m["nbf"] = c.NotBefore
	}
	if c.ID != "" {
		m["jti"] = c.ID
	}
	if c.Issuer != "" {
		m["iss"] = c.Issuer
	}
	if c.Subject != "" {
		m["sub"] = c.Subject
	}
	if c.Name != "" {
		m["sub"] = c.Name
	}
	if c.Email != "" {
		m["email"] = c.Email
	}
	if len(c.Roles) > 0 {
		m["roles"] = c.Roles
	}
	if c.Origin != "" {
		m["origin"] = c.Origin
	}
	if len(c.Scopes) > 0 {
		m["scopes"] = c.Scopes
	}
	if len(c.Organizations) > 0 {
		m["org"] = c.Organizations
	}
	if c.AccessList != nil {
		m["acl"] = c.AccessList
	}
	if c.Address != "" {
		m["addr"] = c.Address
	}
	if c.PictureURL != "" {
		m["picture"] = c.PictureURL
	}
	if c.Metadata != nil {
		m["metadata"] = c.Metadata
	}
	if c.custom != nil {
		for k, v := range c.custom {
			m[k] = v
		}
	}
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return j, nil
}
*/

// AsMap converts Claims struct to dictionary.
func (u *User) AsMap() map[string]interface{} {
	return u.mkv
}

// GetData return user claim felds and their values for the evaluation by an ACL.
func (u *User) GetData() map[string]interface{} {
	return u.tkv
}

// SetRequestHeaders sets request headers associated with the user.
func (u *User) SetRequestHeaders(m map[string]string) {
	u.requestHeaders = m
	return
}

// GetRequestHeaders returns request headers associated with the user.
func (u *User) GetRequestHeaders() map[string]string {
	return u.requestHeaders
}

// SetRequestIdentity sets request identity associated with the user.
func (u *User) SetRequestIdentity(m map[string]interface{}) {
	u.requestIdentity = m
	return
}

// GetRequestIdentity returns request identity associated with the user.
func (u *User) GetRequestIdentity() map[string]interface{} {
	return u.requestIdentity
}

// HasRole checks whether a user has any of the provided roles.
func (u *User) HasRole(roles ...string) bool {
	for _, role := range roles {
		if _, exists := u.rkv[role]; exists {
			return true
		}
	}
	return false
}

// HasRoles checks whether a user has all of the provided roles.
func (u *User) HasRoles(roles ...string) bool {
	for _, role := range roles {
		if _, exists := u.rkv[role]; !exists {
			return false
		}
	}
	return true
}

// NewUser returns a user with associated claims.
func NewUser(data interface{}) (*User, error) {
	var m map[string]interface{}
	u := &User{}

	switch v := data.(type) {
	case string:
		m = make(map[string]interface{})
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			return nil, err
		}
	case []uint8:
		m = make(map[string]interface{})
		if err := json.Unmarshal(v, &m); err != nil {
			return nil, err
		}
	case map[string]interface{}:
		m = v
	}

	if len(m) == 0 {
		return nil, errors.ErrInvalidUserDataType
	}

	c := &Claims{}
	mkv := make(map[string]interface{})
	tkv := make(map[string]interface{})

	if _, exists := m["aud"]; exists {
		switch audiences := m["aud"].(type) {
		case string:
			c.Audience = append(c.Audience, audiences)
		case []interface{}:
			for _, audience := range audiences {
				switch audience.(type) {
				case string:
					c.Audience = append(c.Audience, audience.(string))
				default:
					return nil, errors.ErrInvalidAudience.WithArgs(audience)
				}
			}
		case []string:
			for _, audience := range audiences {
				c.Audience = append(c.Audience, audience)
			}
		default:
			return nil, errors.ErrInvalidAudienceType.WithArgs(m["aud"])
		}
		switch len(c.Audience) {
		case 0:
		case 1:
			tkv["aud"] = c.Audience
			mkv["aud"] = c.Audience[0]
		default:
			tkv["aud"] = c.Audience
			mkv["aud"] = c.Audience
		}
	}

	if _, exists := m["exp"]; exists {
		switch exp := m["exp"].(type) {
		case float64:
			c.ExpiresAt = int64(exp)
		case int:
			c.ExpiresAt = int64(exp)
		case int64:
			c.ExpiresAt = exp
		case json.Number:
			v, _ := exp.Int64()
			c.ExpiresAt = v
		default:
			return nil, errors.ErrInvalidClaimExpiresAt.WithArgs(m["exp"])
		}
		mkv["exp"] = c.ExpiresAt
	}

	if _, exists := m["jti"]; exists {
		switch m["jti"].(type) {
		case string:
			c.ID = m["jti"].(string)
		default:
			return nil, errors.ErrInvalidIDClaimType.WithArgs(m["jti"])
		}
		tkv["jti"] = c.ID
		mkv["jti"] = c.ID
	}

	if _, exists := m["iat"]; exists {
		switch exp := m["iat"].(type) {
		case float64:
			c.IssuedAt = int64(exp)
		case int:
			c.IssuedAt = int64(exp)
		case int64:
			c.IssuedAt = exp
		case json.Number:
			v, _ := exp.Int64()
			c.IssuedAt = v
		default:
			return nil, errors.ErrInvalidClaimIssuedAt.WithArgs(m["iat"])
		}
		mkv["iat"] = c.IssuedAt
	}

	if _, exists := m["iss"]; exists {
		switch m["iss"].(type) {
		case string:
			c.Issuer = m["iss"].(string)
		default:
			return nil, errors.ErrInvalidIssuerClaimType.WithArgs(m["iss"])
		}
		tkv["iss"] = c.Issuer
		mkv["iss"] = c.Issuer
	}

	if _, exists := m["nbf"]; exists {
		switch exp := m["nbf"].(type) {
		case float64:
			c.NotBefore = int64(exp)
		case int:
			c.NotBefore = int64(exp)
		case int64:
			c.NotBefore = exp
		case json.Number:
			v, _ := exp.Int64()
			c.NotBefore = v
		default:
			return nil, errors.ErrInvalidClaimNotBefore.WithArgs(m["nbf"])
		}
		mkv["nbf"] = c.NotBefore
	}

	if _, exists := m["sub"]; exists {
		switch m["sub"].(type) {
		case string:
			c.Subject = m["sub"].(string)
		default:
			return nil, errors.ErrInvalidSubjectClaimType.WithArgs(m["sub"])
		}
		tkv["sub"] = c.Subject
		mkv["sub"] = c.Subject
	}

	for _, ma := range []string{"email", "mail"} {
		if _, exists := m[ma]; exists {
			switch m[ma].(type) {
			case string:
				c.Email = m[ma].(string)
			default:
				return nil, errors.ErrInvalidEmailClaimType.WithArgs(ma, m[ma])
			}
		}
	}
	if c.Email != "" {
		tkv["mail"] = c.Email
		mkv["mail"] = c.Email
	}

	if _, exists := m["name"]; exists {
		switch names := m["name"].(type) {
		case string:
			c.Name = names
		case []interface{}:
			packedNames := []string{}
			for _, n := range names {
				switch n.(type) {
				case string:
					parsedName := n.(string)
					if parsedName == c.Email {
						continue
					}
					packedNames = append(packedNames, parsedName)
				default:
					return nil, errors.ErrInvalidNameClaimType.WithArgs(m["name"])
				}
			}
			c.Name = strings.Join(packedNames, " ")
		default:
			return nil, errors.ErrInvalidNameClaimType.WithArgs(m["name"])
		}
		tkv["name"] = c.Name
		mkv["name"] = c.Name
	}

	for _, ra := range []string{"roles", "role", "groups", "group"} {
		if mra, exists := m[ra]; exists {
			switch roles := mra.(type) {
			case []interface{}:
				for _, role := range roles {
					switch role.(type) {
					case string:
						c.Roles = append(c.Roles, role.(string))
					default:
						return nil, errors.ErrInvalidRole.WithArgs(role)
					}
				}
			case []string:
				for _, role := range roles {
					c.Roles = append(c.Roles, role)
				}
			case string:
				for _, role := range strings.Split(roles, " ") {
					c.Roles = append(c.Roles, role)
				}
			default:
				return nil, errors.ErrInvalidRoleType.WithArgs(m[ra])
			}
		}
	}

	if _, exists := m["app_metadata"]; exists {
		switch m["app_metadata"].(type) {
		case map[string]interface{}:
			appMetadata := m["app_metadata"].(map[string]interface{})
			if _, authzExists := appMetadata["authorization"]; authzExists {
				switch appMetadata["authorization"].(type) {
				case map[string]interface{}:
					appMetadataAuthz := appMetadata["authorization"].(map[string]interface{})
					if _, rolesExists := appMetadataAuthz["roles"]; rolesExists {
						switch roles := appMetadataAuthz["roles"].(type) {
						case []interface{}:
							for _, role := range roles {
								switch role.(type) {
								case string:
									c.Roles = append(c.Roles, role.(string))
								default:
									return nil, errors.ErrInvalidRole.WithArgs(role)
								}
							}
						case []string:
							for _, role := range roles {
								c.Roles = append(c.Roles, role)
							}
						default:
							return nil, errors.ErrInvalidAppMetadataRoleType.WithArgs(appMetadataAuthz["roles"])
						}
					}
				}
			}
		}
	}

	if _, exists := m["realm_access"]; exists {
		switch m["realm_access"].(type) {
		case map[string]interface{}:
			realmAccess := m["realm_access"].(map[string]interface{})
			if _, rolesExists := realmAccess["roles"]; rolesExists {
				switch roles := realmAccess["roles"].(type) {
				case []interface{}:
					for _, role := range roles {
						switch role.(type) {
						case string:
							c.Roles = append(c.Roles, role.(string))
						default:
							return nil, errors.ErrInvalidRole.WithArgs(role)
						}
					}
				case []string:
					for _, role := range roles {
						c.Roles = append(c.Roles, role)
					}
				}
			}
		}
	}

	for _, ra := range []string{"scopes", "scope"} {
		if _, exists := m[ra]; exists {
			switch scopes := m[ra].(type) {
			case []interface{}:
				for _, scope := range scopes {
					switch scope.(type) {
					case string:
						c.Scopes = append(c.Scopes, scope.(string))
					default:
						return nil, errors.ErrInvalidScope.WithArgs(scope)
					}
				}
			case []string:
				for _, scope := range scopes {
					c.Scopes = append(c.Scopes, scope)
				}
			case string:
				for _, scope := range strings.Split(scopes, " ") {
					c.Scopes = append(c.Scopes, scope)
				}
			default:
				return nil, errors.ErrInvalidScopeType.WithArgs(m[ra])
			}
		}
	}

	if len(c.Scopes) > 0 {
		tkv["scopes"] = c.Scopes
		mkv["scopes"] = strings.Join(c.Scopes, " ")
	}

	if _, exists := m["paths"]; exists {
		switch m["paths"].(type) {
		case []interface{}:
			paths := m["paths"].([]interface{})
			for _, path := range paths {
				switch path.(type) {
				case string:
					if c.AccessList == nil {
						c.AccessList = &AccessListClaim{}
					}
					if c.AccessList.Paths == nil {
						c.AccessList.Paths = make(map[string]interface{})
					}
					c.AccessList.Paths[path.(string)] = make(map[string]interface{})
				default:
					return nil, errors.ErrInvalidAccessListPath.WithArgs(path)
				}
			}
		}
	}

	if _, exists := m["acl"]; exists {
		switch m["acl"].(type) {
		case map[string]interface{}:
			acl := m["acl"].(map[string]interface{})
			if _, pathsExists := acl["paths"]; pathsExists {
				switch acl["paths"].(type) {
				case map[string]interface{}:
					paths := acl["paths"].(map[string]interface{})
					for path := range paths {
						if c.AccessList == nil {
							c.AccessList = &AccessListClaim{}
						}
						if c.AccessList.Paths == nil {
							c.AccessList.Paths = make(map[string]interface{})
						}
						c.AccessList.Paths[path] = make(map[string]interface{})
					}
				case []interface{}:
					paths := acl["paths"].([]interface{})
					for _, path := range paths {
						switch path.(type) {
						case string:
							if c.AccessList == nil {
								c.AccessList = &AccessListClaim{}
							}
							if c.AccessList.Paths == nil {
								c.AccessList.Paths = make(map[string]interface{})
							}
							c.AccessList.Paths[path.(string)] = make(map[string]interface{})
						default:
							return nil, errors.ErrInvalidAccessListPath.WithArgs(path)
						}
					}
				}
			}
		}
	}

	if c.AccessList != nil && c.AccessList.Paths != nil {
		tkv["acl"] = map[string]interface{}{
			"paths": c.AccessList.Paths,
		}
		mkv["acl"] = map[string]interface{}{
			"paths": c.AccessList.Paths,
		}
	}

	if _, exists := m["origin"]; exists {
		switch m["origin"].(type) {
		case string:
			c.Origin = m["origin"].(string)
		default:
			return nil, errors.ErrInvalidOriginClaimType.WithArgs(m["origin"])
		}
		tkv["origin"] = c.Origin
		mkv["origin"] = c.Origin
	}

	if _, exists := m["org"]; exists {
		switch orgs := m["org"].(type) {
		case []interface{}:
			for _, org := range orgs {
				switch org.(type) {
				case string:
					c.Organizations = append(c.Organizations, org.(string))
				default:
					return nil, errors.ErrInvalidOrg.WithArgs(org)
				}
			}
		case []string:
			for _, org := range orgs {
				c.Organizations = append(c.Organizations, org)
			}
		case string:
			for _, org := range strings.Split(orgs, " ") {
				c.Organizations = append(c.Organizations, org)
			}
		default:
			return nil, errors.ErrInvalidOrgType.WithArgs(m["org"])
		}
		tkv["org"] = c.Organizations
		mkv["org"] = strings.Join(c.Organizations, " ")
	}

	if _, exists := m["addr"]; exists {
		switch m["addr"].(type) {
		case string:
			c.Address = m["addr"].(string)
		default:
			return nil, errors.ErrInvalidAddrType.WithArgs(m["addr"])
		}
		tkv["addr"] = c.Address
		mkv["addr"] = c.Address
	}

	if _, exists := m["picture"]; exists {
		switch m["picture"].(type) {
		case string:
			c.PictureURL = m["picture"].(string)
		default:
			return nil, errors.ErrInvalidPictureClaimType.WithArgs(m["picture"])
		}
		mkv["picture"] = c.PictureURL
	}

	if _, exists := m["metadata"]; exists {
		switch m["metadata"].(type) {
		case map[string]interface{}:
			c.Metadata = m["metadata"].(map[string]interface{})
		default:
			return nil, errors.ErrInvalidMetadataClaimType.WithArgs(m["metadata"])
		}
		mkv["metadata"] = c.Metadata
	}
	
	if _, exists := m["username"]; exists {
		switch m["username"].(type) {
		case string:
			c.Username = m["username"].(string)
		default:
			return nil, errors.ErrInvalidMetadataClaimType.WithArgs(m["username"])
		}
		mkv["username"] = c.Username
	}

	if len(c.Roles) == 0 {
		c.Roles = append(c.Roles, "anonymous")
		c.Roles = append(c.Roles, "guest")
	}
	tkv["roles"] = c.Roles
	mkv["roles"] = c.Roles

	u.rkv = make(map[string]interface{})
	for _, roleName := range c.Roles {
		u.rkv[roleName] = true
	}

	/*
		for k, v := range m {
			if _, exists := mkv[k]; exists {
				continue
			}
			if _, exists := reservedFields[k]; exists {
				continue
			}
			if c.custom == nil {
				c.custom = make(map[string]interface{})
			}
			mkv[k] = v
			c.custom[k] = v
		}
	*/

	u.Claims = c
	u.mkv = mkv
	u.tkv = tkv
	return u, nil
}

// AddFrontendLinks adds frontend links to User instance.
func (u *User) AddFrontendLinks(v interface{}) error {
	var entries []string
	switch data := v.(type) {
	case string:
		entries = append(entries, data)
	case []string:
		entries = data
	case []interface{}:
		for _, entry := range data {
			switch entry.(type) {
			case string:
				entries = append(entries, entry.(string))
			default:
				return errors.ErrCheckpointInvalidType.WithArgs(data, data)
			}
		}
	default:
		return errors.ErrFrontendLinkInvalidType.WithArgs(data, data)
	}
	m := make(map[string]bool)
	for _, entry := range entries {
		m[entry] = true
	}
	for _, link := range u.FrontendLinks {
		if _, exists := m[link]; exists {
			m[link] = false
		}
	}
	for _, entry := range entries {
		if m[entry] {
			u.FrontendLinks = append(u.FrontendLinks, entry)
		}
	}
	return nil
}

// GetClaimValueByField returns the value of the provides claims field.
func (u *User) GetClaimValueByField(k string) string {
	if u.mkv == nil {
		return ""
	}
	if v, exists := u.mkv[k]; exists {
		switch data := v.(type) {
		case string:
			return data
		case []string:
			return strings.Join(data, " ")
		default:
			return fmt.Sprintf("%v", data)
		}
	}
	return ""
}

// NewCheckpoints returns Checkpoint instances.
func NewCheckpoints(v interface{}) ([]*Checkpoint, error) {
	var entries []string
	checkpoints := []*Checkpoint{}
	switch data := v.(type) {
	case string:
		entries = append(entries, data)
	case []string:
		entries = data
	case []interface{}:
		for _, entry := range data {
			switch entry.(type) {
			case string:
				entries = append(entries, entry.(string))
			default:
				return nil, errors.ErrCheckpointInvalidType.WithArgs(data, data)
			}
		}
	default:
		return nil, errors.ErrCheckpointInvalidType.WithArgs(data, data)
	}
	for i, entry := range entries {
		c, err := NewCheckpoint(entry)
		if err != nil {
			return nil, errors.ErrCheckpointInvalidInput.WithArgs(entry, err)
		}
		c.ID = i
		checkpoints = append(checkpoints, c)
	}
	if len(checkpoints) < 1 {
		return nil, errors.ErrCheckpointEmpty
	}
	return checkpoints, nil
}

// NewCheckpoint returns Checkpoint instance.
func NewCheckpoint(s string) (*Checkpoint, error) {
	c := &Checkpoint{}
	args, err := cfgutils.DecodeArgs(s)
	if err != nil {
		return nil, err
	}
	switch args[0] {
	case "require":
		if len(args) != 2 {
			return nil, fmt.Errorf("must contain two keywords")
		}
		switch args[1] {
		case "mfa":
			c.Name = "Multi-factor authentication"
			c.Type = "mfa"
		default:
			return nil, fmt.Errorf("unsupported require keyword: %s", args[1])
		}
	default:
		return nil, fmt.Errorf("unsupported keyword: %s", args[0])
	}
	return c, nil
}
