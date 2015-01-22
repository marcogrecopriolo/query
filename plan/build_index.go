//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package plan

import (
	"fmt"
	"strings"

	"github.com/couchbaselabs/query/algebra"
	"github.com/couchbaselabs/query/datastore"
)

func (this *builder) VisitCreatePrimaryIndex(stmt *algebra.CreatePrimaryIndex) (interface{}, error) {
	ksref := stmt.Keyspace()
	keyspace, err := this.getNameKeyspace(ksref.Namespace(), ksref.Keyspace())
	if err != nil {
		return nil, err
	}

	return NewCreatePrimaryIndex(keyspace, stmt), nil
}

func (this *builder) VisitCreateIndex(stmt *algebra.CreateIndex) (interface{}, error) {
	ksref := stmt.Keyspace()
	keyspace, err := this.getNameKeyspace(ksref.Namespace(), ksref.Keyspace())
	if err != nil {
		return nil, err
	}

	return NewCreateIndex(keyspace, stmt), nil
}

func (this *builder) VisitDropIndex(stmt *algebra.DropIndex) (interface{}, error) {
	ksref := stmt.Keyspace()
	keyspace, err := this.getNameKeyspace(ksref.Namespace(), ksref.Keyspace())
	if err != nil {
		return nil, err
	}

	indexers, er := keyspace.Indexers()
	if er != nil {
		return nil, er
	}

	var index datastore.Index
	for _, indexer := range indexers {
		index, er = indexer.IndexByName(stmt.Name())
		if er == nil {
			break
		}
	}

	if er != nil {
		return nil, er
	}

	return NewDropIndex(index, stmt), nil
}

func (this *builder) VisitAlterIndex(stmt *algebra.AlterIndex) (interface{}, error) {
	ksref := stmt.Keyspace()
	keyspace, err := this.getNameKeyspace(ksref.Namespace(), ksref.Keyspace())
	if err != nil {
		return nil, err
	}

	indexers, er := keyspace.Indexers()
	if er != nil {
		return nil, er
	}

	var index datastore.Index
	for _, indexer := range indexers {
		index, er = indexer.IndexByName(stmt.Name())
		if er == nil {
			break
		}
	}

	if er != nil {
		return nil, er
	}

	return NewAlterIndex(index, stmt), nil
}

func (this *builder) VisitBuildIndexes(stmt *algebra.BuildIndexes) (interface{}, error) {
	ksref := stmt.Keyspace()
	keyspace, err := this.getNameKeyspace(ksref.Namespace(), ksref.Keyspace())
	if err != nil {
		return nil, err
	}

	return NewBuildIndexes(keyspace, stmt), nil
}

func (this *builder) getNameKeyspace(ns, ks string) (datastore.Keyspace, error) {
	if ns == "" {
		ns = this.namespace
	}

	if strings.ToLower(ns) == "#system" {
		return nil, fmt.Errorf("Index operations not allowed on system namespace.")
	}

	datastore := this.datastore
	namespace, err := datastore.NamespaceByName(ns)
	if err != nil {
		return nil, err
	}

	keyspace, err := namespace.KeyspaceByName(ks)
	if err != nil {
		return nil, err
	}

	return keyspace, nil
}
