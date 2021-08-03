package gen

import (
	"fmt"
	"go/token"
	"math"
	"strconv"

	"golang.org/x/tools/go/ssa"
)

var Fset = token.NewFileSet()

type Func struct {
	*ssa.Function
	Blocks       []*Block
	Values       []*Value
	OutterValues []*Value
}

func newFunc(f *ssa.Function) (r *Func) {
	r = &Func{Function: f}

	iSet := make(map[ssa.Instruction]bool)
	for _, b := range f.Blocks {
		r.Blocks = append(r.Blocks, newBlock(b))

		for _, i := range b.Instrs {
			//TODO: ast link
			if _, ok := i.(*ssa.DebugRef); ok {
				continue
			}
			iSet[i] = true
		}
	}

	var nodes []ssa.Node
	for bi := len(f.Blocks) - 1; bi >= 0; bi-- {
		b := f.Blocks[bi]
		for ii := len(b.Instrs) - 1; ii >= 0; ii-- {
			i := b.Instrs[ii]
			if _, ok := i.(*ssa.DebugRef); ok {
				continue
			}
			nodes = append(nodes, i.(ssa.Node))

			vs := i.Operands(nil)
			for vi := len(vs) - 1; vi >= 0; vi-- {
				v := vs[vi]
				if v == nil || *v == nil {
					continue
				}

				if i, ok := (*v).(ssa.Instruction); ok && iSet[i] {
					continue
				}

				nodes = append(nodes, (*v).(ssa.Node))
			}
		}
	}

	seen := make(map[ssa.Node]bool)
	for ni := len(nodes) - 1; ni >= 0; ni-- {
		n := nodes[ni]
		if seen[n] {
			continue
		}
		seen[n] = true

		v := newValue(n)
		var add bool
		for _, b := range r.Blocks {
			if b.addValue(v) {
				add = true
				break
			}
		}
		if !add {
			if v.Parent() == f {
				r.Values = append(r.Values, v)
			} else {
				r.OutterValues = append(r.OutterValues, v)
			}
		}
	}

	return
}

func posToLine(pos token.Pos) int {
	return Fset.Position(pos).Line
}

type XPos struct {
	token.Pos
}

func (p XPos) IsKnown() bool {
	return p.IsValid()
}

func (p XPos) LineNumber() string {
	if !p.IsKnown() {
		return "?"
	}

	return strconv.Itoa(posToLine(p.Pos))
}

func (p XPos) LineNumberHTML() string {
	return p.LineNumber()
}

type Block struct {
	*ssa.BasicBlock
	Values             []*Value
	StartLine, EndLine int
}

func newBlock(b *ssa.BasicBlock) (r *Block) {
	r = &Block{
		BasicBlock: b,
		StartLine:  math.MaxInt32,
	}

	for _, i := range b.Instrs {
		if !i.Pos().IsValid() {
			continue
		}
		l := posToLine(i.Pos())
		if l < r.StartLine {
			r.StartLine = l
		}
		if l > r.EndLine {
			r.EndLine = l
		}
	}

	return
}

func (b *Block) addValue(v *Value) (add bool) {
	if i, ok := v.Node.(ssa.Instruction); ok {
		if i.Block() == b.BasicBlock {
			b.Values = append(b.Values, v)
			return true
		}
	} else if v.Pos.IsKnown() {
		l := posToLine(v.Pos.Pos)
		if l >= b.StartLine && l <= b.EndLine {
			b.Values = append(b.Values, v)
			return true
		}
	}
	return false
}

func (b *Block) String() string {
	return fmt.Sprintf("b%d", b.Index)
}

type Value struct {
	ssa.Node
	Pos XPos
	vx  string
}

var vmap = make(map[ssa.Node]*Value)
var vcnt, icnt int

func (v *Value) String() string {
	if v == nil {
		return "nil"
	}
	return v.vx
}

func newValue(n ssa.Node) (r *Value) {
	r = &Value{
		Node: n,
		Pos:  XPos{n.Pos()},
	}

	_, isI := n.(ssa.Instruction)
	_, isV := n.(ssa.Value)

	if isI {
		var ext string
		if isV {
			ext = "v"
		}
		icnt++
		r.vx = fmt.Sprintf("i%d%s", icnt, ext)
	} else {
		vcnt++
		r.vx = fmt.Sprintf("v%d", vcnt)
	}

	vmap[n] = r
	return
}

func (v *Value) Operands() (ret []*Value) {
	for _, vv := range v.Node.Operands(nil) {
		if vv == nil || *vv == nil {
			ret = append(ret, nil)
		} else {
			ret = append(ret, vmap[(*vv).(ssa.Node)])
		}
	}
	return
}
