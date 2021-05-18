/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package database

import (
	"testing"

	"github.com/codenotary/immudb/embedded/store"
	"github.com/codenotary/immudb/pkg/api/schema"
	"github.com/stretchr/testify/require"
)

func TestSQLExecAndQuery(t *testing.T) {
	db, closer := makeDb()
	defer closer()

	md, err := db.SQLExec(&schema.SQLExecRequest{Sql: `
		CREATE TABLE table1(id INTEGER, title VARCHAR, active BOOLEAN, PRIMARY KEY id)
	`})
	require.NoError(t, err)
	require.Len(t, md.Ctxs, 1)
	require.Len(t, md.Dtxs, 0)

	res, err := db.ListTables()
	require.NoError(t, err)
	require.Len(t, res.Rows, 1)

	res, err = db.DescribeTable("table1")
	require.NoError(t, err)
	require.Len(t, res.Rows, 3)

	md, err = db.SQLExec(&schema.SQLExecRequest{Sql: `
		UPSERT INTO table1(id, title) VALUES (1, 'title1'), (2, 'title2'), (3, 'title3')
	`})
	require.NoError(t, err)
	require.Len(t, md.Ctxs, 0)
	require.Len(t, md.Dtxs, 1)

	params := make([]*schema.NamedParam, 1)
	params[0] = &schema.NamedParam{Name: "active", Value: &schema.SQLValue{Value: &schema.SQLValue_B{B: true}}}

	res, err = db.SQLQuery(&schema.SQLQueryRequest{Sql: "SELECT t.id, t.id as id2, title FROM (table1 as t) WHERE id < 3 AND active != @active", Params: params})
	require.NoError(t, err)
	require.Len(t, res.Rows, 2)

	ve, err := db.VerifiableSQLGet(&schema.VerifiableSQLGetRequest{
		SqlGetRequest: &schema.SQLGetRequest{Table: "table1", PkValue: &schema.SQLValue{Value: &schema.SQLValue_N{N: 1}}},
		ProveSinceTx:  0,
	})
	require.NoError(t, err)
	require.NotNil(t, ve)

	_, err = db.VerifiableSQLGet(&schema.VerifiableSQLGetRequest{
		SqlGetRequest: &schema.SQLGetRequest{Table: "table1", PkValue: &schema.SQLValue{Value: &schema.SQLValue_N{N: 4}}},
		ProveSinceTx:  0,
	})
	require.Equal(t, store.ErrKeyNotFound, err)

}
