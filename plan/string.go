package plan

import (
	"errors"

	"../exe"
	"../sql/syntax"
)

func StringPlan(stn *syntax.SyntaxTreeNode) (*exe.Relation, *Result, error) {
	result := NewResult()
	if stn.Name != "string" {
		return nil, nil, errors.New("Expect string, but get " + stn.Name)
	}
	relation := exe.NewRelation()
	value := exe.NewValue(exe.STRING, stn.Value.([]byte))
	row := exe.NewRow([]exe.Value{value})
	relation.AddRow(row)
	result.SetResult(1)
	return relation, result, nil
}
