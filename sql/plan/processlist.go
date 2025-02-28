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

type process struct {
	id      int64
	user    string
	host    string
	db      string
	command string
	time    int64
	state   string
	info    string
}

func (p process) toRow() sql.Row {
	var db interface{}
	if p.db != "" {
		db = p.db
	}
	return sql.NewRow(
		p.id,
		p.user,
		p.host,
		db,
		p.command,
		p.time,
		p.state,
		p.info,
	)
}

var processListSchema = sql.Schema{
	{Name: "Id", Type: types.Int64},
	{Name: "User", Type: types.LongText},
	{Name: "Host", Type: types.LongText},
	{Name: "db", Type: types.LongText},
	{Name: "Command", Type: types.LongText},
	{Name: "Time", Type: types.Int64},
	{Name: "State", Type: types.LongText},
	{Name: "Info", Type: types.LongText},
}

// ShowProcessList shows a list of all current running processes.
type ShowProcessList struct {
	Database string
}

var _ sql.Node = (*ShowProcessList)(nil)
var _ sql.CollationCoercible = (*ShowProcessList)(nil)

// NewShowProcessList creates a new ProcessList node.
func NewShowProcessList() *ShowProcessList { return new(ShowProcessList) }

// Children implements the Node interface.
func (p *ShowProcessList) Children() []sql.Node { return nil }

// Resolved implements the Node interface.
func (p *ShowProcessList) Resolved() bool { return true }

// WithChildren implements the Node interface.
func (p *ShowProcessList) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(p, len(children), 0)
	}

	return p, nil
}

// CheckPrivileges implements the interface sql.Node.
func (p *ShowProcessList) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return opChecker.UserHasPrivileges(ctx,
		sql.NewPrivilegedOperation("", "", "", sql.PrivilegeType_Process))
}

// CollationCoercibility implements the interface sql.CollationCoercible.
func (*ShowProcessList) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 7
}

// Schema implements the Node interface.
func (p *ShowProcessList) Schema() sql.Schema { return processListSchema }

// RowIter implements the Node interface.
func (p *ShowProcessList) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	processes := ctx.ProcessList.Processes()
	var rows = make([]sql.Row, len(processes))

	for i, proc := range processes {
		var status []string
		var names []string
		for name := range proc.Progress {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			progress := proc.Progress[name]

			printer := sql.NewTreePrinter()
			_ = printer.WriteNode("\n" + progress.String())
			children := []string{}
			for _, partitionProgress := range progress.PartitionsProgress {
				children = append(children, partitionProgress.String())
			}
			sort.Strings(children)
			_ = printer.WriteChildren(children...)

			status = append(status, printer.String())
		}

		if len(status) == 0 && proc.Command == sql.ProcessCommandQuery {
			status = []string{"running"}
		}

		rows[i] = process{
			id:      int64(proc.Connection),
			user:    proc.User,
			time:    int64(proc.Seconds()),
			state:   strings.Join(status, ""),
			command: string(proc.Command),
			host:    proc.Host,
			info:    proc.Query,
			db:      proc.Database,
		}.toRow()
	}

	return sql.RowsToRowIter(rows...), nil
}

func (p *ShowProcessList) String() string { return "ProcessList" }
