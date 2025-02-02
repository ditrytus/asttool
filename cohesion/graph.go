package cohesion

import (
	"gonum.org/v1/gonum/graph"
	"image/color"
	"log"
	"math"

	"gonum.org/v1/gonum/graph/layout"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func drawGraph(g graph.Graph, output string) {
	// Use the Eades layout algorithm with reasonable defaults.
	eades := layout.EadesR2{Repulsion: 1, Rate: 0.05, Updates: 30, Theta: 0.2}

	// Make a layout optimizer with the target graph and update function.
	optimizer := layout.NewOptimizerR2(g, eades.Update)

	// Perform layout optimization.
	for optimizer.Update() {
	}

	p := plot.New()

	// Add to plot.
	p.Add(render{optimizer})
	p.HideAxes()

	// Render graph on save.
	err := p.Save(15*vg.Centimeter, 15*vg.Centimeter, output)
	if err != nil {
		log.Fatal(err)
	}
}

const radius = vg.Length(15)

// render implements the plot.Plotter interface for graphs.
type render struct {
	layout.GraphR2
}

func (p render) Plot(c draw.Canvas, plt *plot.Plot) {
	nodes := p.GraphR2.Nodes()
	if nodes.Len() == 0 {
		return
	}
	var (
		xys plotter.XYs
		ids []string
	)
	if nodes.Len() >= 0 {
		xys = make(plotter.XYs, 0, nodes.Len())
		ids = make([]string, 0, nodes.Len())
	}
	for nodes.Next() {
		u := nodes.Node().(objectNode)
		uid := u.ID()
		ur2 := p.GraphR2.LayoutNodeR2(uid)
		xys = append(xys, plotter.XY(ur2.Coord2))
		ids = append(ids, u.Name())
		to := p.GraphR2.From(uid)
		for to.Next() {
			v := to.Node()
			vid := v.ID()
			vr2 := p.GraphR2.LayoutNodeR2(vid)

			l, err := plotter.NewLine(plotter.XYs{plotter.XY(ur2.Coord2), plotter.XY(vr2.Coord2)})
			if err != nil {
				panic(err)
			}
			l.Plot(c, plt)
			if err != nil {
				panic(err)
			}
		}
	}

	n, err := plotter.NewScatter(xys)
	if err != nil {
		panic(err)
	}
	n.GlyphStyle.Shape = nodeGlyph{}
	n.GlyphStyle.Radius = radius
	n.Plot(c, plt)

	l, err := plotter.NewLabels(plotter.XYLabels{XYs: xys, Labels: ids})
	if err != nil {
		panic(err)
	}
	fnt := font.From(plot.DefaultFont, 18)
	for i := range l.TextStyle {
		l.TextStyle[i] = draw.TextStyle{
			Font: fnt, Handler: plot.DefaultTextHandler,
			XAlign: draw.XCenter, YAlign: -0.4,
		}
	}

	l.Plot(c, plt)
}

// DataRange returns the minimum and maximum X and Y values.
func (p render) DataRange() (xmin, xmax, ymin, ymax float64) {
	nodes := p.GraphR2.Nodes()
	if nodes.Len() == 0 {
		return
	}
	var xys plotter.XYs
	if nodes.Len() >= 0 {
		xys = make(plotter.XYs, 0, nodes.Len())
	}
	for nodes.Next() {
		u := nodes.Node()
		uid := u.ID()
		ur2 := p.GraphR2.LayoutNodeR2(uid)
		xys = append(xys, plotter.XY(ur2.Coord2))
	}
	return plotter.XYRange(xys)
}

// GlyphBoxes returns a slice of plot.GlyphBoxes, implementing the
// plot.GlyphBoxer interface.
func (p render) GlyphBoxes(plt *plot.Plot) []plot.GlyphBox {
	nodes := p.GraphR2.Nodes()
	if nodes.Len() == 0 {
		return nil
	}
	var b []plot.GlyphBox
	if nodes.Len() >= 0 {
		b = make([]plot.GlyphBox, 0, nodes.Len())
	}
	for i := 0; nodes.Next(); i++ {
		u := nodes.Node()
		uid := u.ID()
		ur2 := p.GraphR2.LayoutNodeR2(uid)

		b = append(b, plot.GlyphBox{})
		b[i].X = plt.X.Norm(ur2.Coord2.X)
		b[i].Y = plt.Y.Norm(ur2.Coord2.Y)
		r := radius
		b[i].Rectangle = vg.Rectangle{
			Min: vg.Point{X: -r, Y: -r},
			Max: vg.Point{X: +r, Y: +r},
		}
	}
	return b
}

// nodeGlyph is a glyph that draws a filled circle.
type nodeGlyph struct{}

// DrawGlyph implements the GlyphDrawer interface.
func (nodeGlyph) DrawGlyph(c *draw.Canvas, sty draw.GlyphStyle, pt vg.Point) {
	var p vg.Path
	c.Push()
	c.SetColor(color.White)
	p.Move(vg.Point{X: pt.X + sty.Radius, Y: pt.Y})
	p.Arc(pt, sty.Radius, 0, 2*math.Pi)
	p.Close()
	c.Fill(p)
	c.Pop()
	c.Stroke(p)
}
