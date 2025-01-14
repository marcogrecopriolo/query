//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package datastore

import (
	"fmt"
	"strings"

	"github.com/couchbase/query/auth"
)

func CredsArray(creds *auth.Credentials) []string {
	if creds != nil {
		if len(creds.CbauthCredentialsList) == 0 {
			GetDatastore().Authorize(nil, creds)
		}
		return []string(creds.AuthenticatedUsers)
	}
	return []string{}
}

func CredsString(creds *auth.Credentials) string {
	return strings.Join(CredsArray(creds), ",")
}

func FirstCred(creds *auth.Credentials) (string, string) {
	ds := GetDatastore()
	if ds != nil && creds != nil {
		return ds.CredsString(creds)
	}
	return "", ""
}

func EncodeName(user, domain string) string {
	if domain == "" {
		return "local:" + user
	}
	return domain + ":" + user
}

func DecodeName(encodedName string) (string, string) {
	fields := strings.Split(encodedName, ":")
	switch len(fields) {
	case 1:
		return fields[0], "local"
	case 2:
		return fields[1], fields[0]
	}
	return "", ""
}

// Checks if the user is an Administrator but does NOT audit the authorization check.
func IsAdmin(creds *auth.Credentials) bool {
	ds := GetDatastore()
	if ds != nil && creds != nil {

		privs := auth.NewPrivileges()
		privs.Add("", auth.PRIV_ADMIN, 0)

		// avoid logging an audit for an internal authorization action
		return ds.AuthorizeInternal(privs, creds) == nil
	}
	return false
}

func AdminCreds(node string) (*auth.Credentials, error) {
	if _DATASTORE == nil {
		return nil, fmt.Errorf("datastore not initialized")
	}
	u, p, err := _DATASTORE.AdminUser(node)
	if err != nil {
		return nil, err
	}
	creds := &auth.Credentials{}
	users := map[string]string{u: p}
	creds.Users = users
	return creds, nil
}

func GetUserUUID(creds *auth.Credentials) string {
	if _DATASTORE == nil || creds == nil {
		return ""
	}
	return _DATASTORE.GetUserUUID(creds)
}

func GetUserBuckets(creds *auth.Credentials) []string {
	if _DATASTORE == nil || creds == nil {
		return []string{}
	}
	return _DATASTORE.GetUserBuckets(creds)
}

func GetImpersonateBuckets(user, domain string) []string {
	if _DATASTORE == nil || len(user) == 0 {
		return []string{}
	}
	return _DATASTORE.GetImpersonateBuckets(user, domain)
}
