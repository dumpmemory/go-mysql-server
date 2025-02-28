// Copyright 2020-2021 Dolthub, Inc.
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

package plan

import (
	"sort"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// ShowDatabases is a node that shows the databases.
type ShowDatabases struct {
	Catalog sql.Catalog
}

var _ sql.Node = (*ShowDatabases)(nil)
var _ sql.CollationCoercible = (*ShowDatabases)(nil)

// NewShowDatabases creates a new show databases node.
func NewShowDatabases() *ShowDatabases {
	return new(ShowDatabases)
}

// Resolved implements the Resolvable interface.
func (p *ShowDatabases) Resolved() bool {
	return true
}

// Children implements the Node interface.
func (*ShowDatabases) Children() []sql.Node {
	return nil
}

// Schema implements the Node interface.
func (*ShowDatabases) Schema() sql.Schema {
	return sql.Schema{{
		Name:     "Database",
		Type:     types.LongText,
		Nullable: false,
	}}
}

// RowIter implements the Node interface.
func (p *ShowDatabases) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	dbs := p.Catalog.AllDatabases(ctx)
	var rows = make([]sql.Row, 0, len(dbs))
	for _, db := range dbs {
		rows = append(rows, sql.Row{db.Name()})
	}
	if _, err := p.Catalog.Database(ctx, "mysql"); err == nil {
		rows = append(rows, sql.Row{"mysql"})
	}

	sort.Slice(rows, func(i, j int) bool {
		return strings.Compare(rows[i][0].(string), rows[j][0].(string)) < 0
	})

	return sql.RowsToRowIter(rows...), nil
}

// WithChildren implements the Node interface.
func (p *ShowDatabases) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(p, len(children), 0)
	}

	return p, nil
}

// CheckPrivileges implements the interface sql.Node.
func (p *ShowDatabases) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	//TODO: Having the "SHOW DATABASES" privilege should allow one to see all databases
	// Currently, only shows databases that the user has access to
	return true
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*ShowDatabases) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 7
}

func (p ShowDatabases) String() string {
	return "ShowDatabases"
}
