package cohesion

import (
	"fmt"
	"strings"
)

func FormatDependencies(v Visitor) string {
	c := v.(*cohesionAstVisitor)
	var b strings.Builder
	nodes := c.dependencies.Nodes()
	for nodes.Next() {
		node := nodes.Node().(objectNode)
		b.WriteString(fmt.Sprintf("%s %s\n", c.fileSet.Position(node.Pos()), node.Name()))
		deps := c.dependencies.From(node.ID())
		for deps.Next() {
			dep := deps.Node().(objectNode)
			b.WriteString(fmt.Sprintf("\t%s %s\n", c.fileSet.Position(dep.Pos()), dep.Name()))
		}
	}
	return b.String()
}

func FormatCohesionStats(v Visitor) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Connected components: %d\n", v.ConnectedComponents()))
	b.WriteString(fmt.Sprintf("Average degree: %f\n", v.AverageDegree()))
	b.WriteString(fmt.Sprintf("Density: %f\n", v.Density()))
	return b.String()
}