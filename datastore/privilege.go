//  Copyright 2014-Present Couchbase, Inc.
//
//  Use of this software is governed by the Business Source License included
//  in the file licenses/BSL-Couchbase.txt.  As of the Change Date specified
//  in that file, in accordance with the Business Source License, use of this
//  software will be governed by the Apache License, Version 2.0, included in
//  the file licenses/APL2.txt.

package datastore

import (
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

func FirstCred(creds *auth.Credentials) string {
	ds := GetDatastore()
	if ds != nil && creds != nil {
		return ds.CredsString(creds)
	}
	return ""
}

func IsAdmin(creds *auth.Credentials) bool {
	ds := GetDatastore()
	if ds != nil && creds != nil {

		// TODO TENANT agree on appropriate privilege
		privs := auth.NewPrivileges()
		privs.Add("", auth.PRIV_SYSTEM_READ, 0)
		return ds.Authorize(privs, creds) == nil
	}
	return false
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
