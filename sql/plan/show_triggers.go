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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

type ShowTriggers struct {
	db       sql.Database
	Triggers []*CreateTrigger
}

var _ sql.Databaser = (*ShowTriggers)(nil)
var _ sql.Node = (*ShowTriggers)(nil)
var _ sql.CollationCoercible = (*ShowTriggers)(nil)

var showTriggersSchema = sql.Schema{
	&sql.Column{Name: "Trigger", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Event", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Table", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Statement", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Timing", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Created", Type: types.Datetime, Nullable: false},
	&sql.Column{Name: "sql_mode", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Definer", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "character_set_client", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "collation_connection", Type: types.LongText, Nullable: false},
	&sql.Column{Name: "Database Collation", Type: types.LongText, Nullable: false},
}

// NewShowCreateTrigger creates a new ShowCreateTrigger node for SHOW TRIGGER statements.
func NewShowTriggers(db sql.Database) *ShowTriggers {
	return &ShowTriggers{
		db: db,
	}
}

// String implements the sql.Node interface.
func (s *ShowTriggers) String() string {
	return "SHOW TRIGGERS"
}

// Resolved implements the sql.Node interface.
func (s *ShowTriggers) Resolved() bool {
	_, ok := s.db.(sql.UnresolvedDatabase)
	return !ok
}

// Children implements the sql.Node interface.
func (s *ShowTriggers) Children() []sql.Node {
	return nil
}

// Schema implements the sql.Node interface.
func (s *ShowTriggers) Schema() sql.Schema {
	return showTriggersSchema
}

// RowIter implements the sql.Node interface.
func (s *ShowTriggers) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	var rows []sql.Row
	for _, trigger := range s.Triggers {
		triggerEvent := strings.ToUpper(trigger.TriggerEvent)
		triggerTime := strings.ToUpper(trigger.TriggerTime)
		tableName := trigger.Table.(*UnresolvedTable).Name()
		characterSetClient, err := ctx.GetSessionVariable(ctx, "character_set_client")
		if err != nil {
			return nil, err
		}
		collationConnection, err := ctx.GetSessionVariable(ctx, "collation_connection")
		if err != nil {
			return nil, err
		}
		collationServer, err := ctx.GetSessionVariable(ctx, "collation_server")
		if err != nil {
			return nil, err
		}
		rows = append(rows, sql.Row{
			trigger.TriggerName, // Trigger
			triggerEvent,        // Event
			tableName,           // Table
			trigger.BodyString,  // Statement
			triggerTime,         // Timing
			trigger.CreatedAt,   // Created
			"",                  // sql_mode
			"",                  // Definer
			characterSetClient,  // character_set_client
			collationConnection, // collation_connection
			collationServer,     // Database Collation
		})
	}
	return sql.RowsToRowIter(rows...), nil
}

// WithChildren implements the sql.Node interface.
func (s *ShowTriggers) WithChildren(children ...sql.Node) (sql.Node, error) {
	return NillaryWithChildren(s, children...)
}

// CheckPrivileges implements the interface sql.Node.
func (s *ShowTriggers) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	//TODO: figure out what privileges are needed here
	return true
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*ShowTriggers) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 7
}

// Database implements the sql.Databaser interface.
func (s *ShowTriggers) Database() sql.Database {
	return s.db
}

// WithDatabase implements the sql.Databaser interface.
func (s *ShowTriggers) WithDatabase(db sql.Database) (sql.Node, error) {
	ns := *s
	ns.db = db
	return &ns, nil
}
