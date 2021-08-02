package gen

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ssa"
)

type HTMLWriter struct {
	w    io.WriteCloser
	Func *Func
}

func NewHTMLWriter(path string, f *ssa.Function) *HTMLWriter {
	fp, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	html := &HTMLWriter{
		w:    fp,
		Func: newFunc(f),
	}
	html.start()
	return html
}

// FuncLines contains source code for a function to be displayed
// in sources column.
type FuncLines struct {
	Filename    string
	StartLineno uint
	Lines       []string
}

// WriteSources writes lines as source code in a column headed by title.
// phase is used for collapsing columns and should be unique across the table.
func (w *HTMLWriter) WriteSources(phase string, all []*FuncLines) {
	if w == nil {
		return // avoid generating HTML just to discard it
	}
	var buf bytes.Buffer
	fmt.Fprint(&buf, "<div class=\"lines\" style=\"width: 8%\">")
	filename := ""
	for _, fl := range all {
		fmt.Fprint(&buf, "<div>&nbsp;</div>")
		if filename != fl.Filename {
			fmt.Fprint(&buf, "<div>&nbsp;</div>")
			filename = fl.Filename
		}
		for i := range fl.Lines {
			ln := int(fl.StartLineno) + i
			fmt.Fprintf(&buf, "<div class=\"l%v line-number\">%v</div>", ln, ln)
		}
	}
	fmt.Fprint(&buf, "</div><div style=\"width: 92%\"><pre>")
	filename = ""
	for _, fl := range all {
		fmt.Fprint(&buf, "<div>&nbsp;</div>")
		if filename != fl.Filename {
			fmt.Fprintf(&buf, "<div><strong>%v</strong></div>", fl.Filename)
			filename = fl.Filename
		}
		for i, line := range fl.Lines {
			ln := int(fl.StartLineno) + i
			var escaped string
			if strings.TrimSpace(line) == "" {
				escaped = "&nbsp;"
			} else {
				escaped = html.EscapeString(line)
			}
			fmt.Fprintf(&buf, "<div class=\"l%v line-number\">%v</div>", ln, escaped)
		}
	}
	fmt.Fprint(&buf, "</pre></div>")
	w.WriteColumn(phase, phase, "allow-x-scroll", buf.String())
}

func (w *HTMLWriter) WriteAST(phase string, buf *bytes.Buffer) {
	if w == nil {
		return // avoid generating HTML just to discard it
	}
	lines := strings.Split(buf.String(), "\n")
	var out bytes.Buffer

	fmt.Fprint(&out, "<div>")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		var escaped string
		var lineNo string
		if l == "" {
			escaped = "&nbsp;"
		} else {
			if strings.HasPrefix(l, "buildssa") {
				escaped = fmt.Sprintf("<b>%v</b>", l)
			} else {
				// Parse the line number from the format l(123).
				idx := strings.Index(l, " l(")
				if idx != -1 {
					subl := l[idx+3:]
					idxEnd := strings.Index(subl, ")")
					if idxEnd != -1 {
						if _, err := strconv.Atoi(subl[:idxEnd]); err == nil {
							lineNo = subl[:idxEnd]
						}
					}
				}
				escaped = html.EscapeString(l)
			}
		}
		if lineNo != "" {
			fmt.Fprintf(&out, "<div class=\"l%v line-number ast\">%v</div>", lineNo, escaped)
		} else {
			fmt.Fprintf(&out, "<div class=\"ast\">%v</div>", escaped)
		}
	}
	fmt.Fprint(&out, "</div>")
	w.WriteColumn(phase, phase, "allow-x-scroll", out.String())
}

// WriteColumn writes raw HTML in a column headed by title.
// It is intended for pre- and post-compilation log output.
func (w *HTMLWriter) WriteColumn(phase, title, class, html string) {
	w.WriteMultiTitleColumn(phase, []string{title}, class, html)
}

func (w *HTMLWriter) WriteMultiTitleColumn(phase string, titles []string, class, html string) {
	if w == nil {
		return
	}
	id := strings.Replace(phase, " ", "-", -1)
	// collapsed column
	w.Printf("<td id=\"%v-col\" class=\"collapsed\"><div>%v</div></td>", id, phase)

	if class == "" {
		w.Printf("<td id=\"%v-exp\">", id)
	} else {
		w.Printf("<td id=\"%v-exp\" class=\"%v\">", id, class)
	}
	for _, title := range titles {
		w.WriteString("<h2>" + title + "</h2>")
	}
	w.WriteString(html)
	w.WriteString("</td>\n")
}

func (w *HTMLWriter) Printf(msg string, v ...interface{}) {
	if _, err := fmt.Fprintf(w.w, msg, v...); err != nil {
		panic(err)
	}
}

func (w *HTMLWriter) WriteString(s string) {
	if _, err := io.WriteString(w.w, s); err != nil {
		panic(err)
	}
}

func (v *Value) HTML() string {
	// TODO: Using the value ID as the class ignores the fact
	// that value IDs get recycled and that some values
	// are transmuted into other values.
	s := v.String()
	return fmt.Sprintf("<span class=\"%s ssa-value\">%s</span>", s, s)
}

func (v *Value) LongHTML() string {
	// TODO: Any intra-value formatting?
	// I'm wary of adding too much visual noise,
	// but a little bit might be valuable.
	// We already have visual noise in the form of punctuation
	// maybe we could replace some of that with formatting.
	s := fmt.Sprintf("<span class=\"%s ssa-long-value\">", v.String())

	linenumber := "<span class=\"no-line-number\">(?)</span>"
	if v.Pos.IsKnown() {
		linenumber = fmt.Sprintf("<span class=\"l%v line-number\">(%s)</span>", v.Pos.LineNumber(), v.Pos.LineNumberHTML())
	}

	s += fmt.Sprintf("%s %s = %s", v.HTML(), linenumber, v.Node.String())

	s += " &lt;" + html.EscapeString(reflect.TypeOf(v.Node).Elem().Name()) + "&gt;"
	// s += html.EscapeString(v.auxString())
	for _, a := range v.Operands() {
		s += fmt.Sprintf(" %s", a.HTML())
	}
	// r := v.Block.Func.RegAlloc
	// if int(v.ID) < len(r) && r[v.ID] != nil {
	// 	s += " : " + html.EscapeString(r[v.ID].String())
	// }
	// var names []string
	// for name, values := range v.Block.Func.NamedValues {
	// 	for _, value := range values {
	// 		if value == v {
	// 			names = append(names, name.String())
	// 			break // drop duplicates.
	// 		}
	// 	}
	// }
	// if len(names) != 0 {
	// 	s += " (" + strings.Join(names, ", ") + ")"
	// }

	s += "</span>"
	return s
}

func blockHTML(b *ssa.BasicBlock) string {
	// TODO: Using the value ID as the class ignores the fact
	// that value IDs get recycled and that some values
	// are transmuted into other values.
	s := html.EscapeString(fmt.Sprintf("b%d", b.Index))
	return fmt.Sprintf("<span class=\"%s ssa-block\">%s</span>", s, s)
}

func (b *Block) HTML() string {
	return blockHTML(b.BasicBlock)
}

func (b *Block) LongHTML() string {
	// TODO: improve this for HTML?
	s := fmt.Sprintf("<span class=\"%s ssa-block\">%s</span>", html.EscapeString(b.String()), html.EscapeString(b.String()))
	// if b.Aux != nil {
	// 	s += html.EscapeString(fmt.Sprintf(" {%v}", b.Aux))
	// }
	// if t := b.AuxIntString(); t != "" {
	// 	s += html.EscapeString(fmt.Sprintf(" [%v]", t))
	// }
	// for _, c := range b.ControlValues() {
	// 	s += fmt.Sprintf(" %s", c.HTML())
	// }
	if len(b.Succs) > 0 {
		s += " &#8594;" // right arrow
		for _, e := range b.Succs {
			s += " " + blockHTML(e)
		}
	}
	// switch b.Likely {
	// case BranchUnlikely:
	// 	s += " (unlikely)"
	// case BranchLikely:
	// 	s += " (likely)"
	// }
	if b.StartLine != math.MaxInt32 {
		// TODO does not begin to deal with the full complexity of line numbers.
		// Maybe we want a string/slice instead, of outer-inner when inlining.
		s += fmt.Sprintf(" <span class=\"l%d line-number\">(%d)</span>", b.StartLine, b.StartLine)
	}
	return s
}

func (f *Func) HTML(phase string) string {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, "<code>")
	p := htmlFuncPrinter{w: buf}
	fprintFunc(p, f)

	// fprintFunc(&buf, f) // TODO: HTML, not text, <br /> for line breaks, etc.
	fmt.Fprint(buf, "</code>")
	return buf.String()
}

type htmlFuncPrinter struct {
	w io.Writer
}

func (p htmlFuncPrinter) header(f *Func) {
	io.WriteString(p.w, "<ul>")
	for _, v := range f.Values {
		p.value(v, true)
	}
	io.WriteString(p.w, "</ul>")
}

func (p htmlFuncPrinter) footer(f *Func) {
	io.WriteString(p.w, "<ul>")
	for _, v := range f.OutterValues {
		p.value(v, true)
	}
	io.WriteString(p.w, "</ul>")
}

func (p htmlFuncPrinter) startBlock(b *Block, reachable bool) {
	var dead string
	if !reachable {
		dead = "dead-block"
	}
	fmt.Fprintf(p.w, "<ul class=\"%s ssa-print-func %s\">", b, dead)
	fmt.Fprintf(p.w, "<li class=\"ssa-start-block\">%s:", b.HTML())
	if len(b.Preds) > 0 {
		io.WriteString(p.w, " &#8592;") // left arrow
		for _, e := range b.Preds {
			fmt.Fprintf(p.w, " %s", blockHTML(e))
		}
	}
	if len(b.Values) > 0 {
		io.WriteString(p.w, `<button onclick="hideBlock(this)">-</button>`)
	}
	io.WriteString(p.w, "</li>")
	if len(b.Values) > 0 { // start list of values
		io.WriteString(p.w, "<li class=\"ssa-value-list\">")
		io.WriteString(p.w, "<ul>")
	}
}

func (p htmlFuncPrinter) endBlock(b *Block) {
	if len(b.Values) > 0 { // end list of values
		io.WriteString(p.w, "</ul>")
		io.WriteString(p.w, "</li>")
	}
	io.WriteString(p.w, "<li class=\"ssa-end-block\">")
	fmt.Fprint(p.w, b.LongHTML())
	io.WriteString(p.w, "</li>")
	io.WriteString(p.w, "</ul>")
}

func (p htmlFuncPrinter) value(v *Value, live bool) {
	var dead string
	if !live {
		dead = "dead-value"
	}
	fmt.Fprintf(p.w, "<li class=\"ssa-long-value %s\">", dead)
	fmt.Fprint(p.w, v.LongHTML())
	io.WriteString(p.w, "</li>")
}

func (p htmlFuncPrinter) startDepCycle() {
	fmt.Fprintln(p.w, "<span class=\"depcycle\">")
}

func (p htmlFuncPrinter) endDepCycle() {
	fmt.Fprintln(p.w, "</span>")
}

// func (p htmlFuncPrinter) named(n LocalSlot, vals []*Value) {
// 	fmt.Fprintf(p.w, "<li>name %s: ", n)
// 	for _, val := range vals {
// 		fmt.Fprintf(p.w, "%s ", val.HTML())
// 	}
// 	fmt.Fprintf(p.w, "</li>")
// }

type funcPrinter interface {
	header(f *Func)
	startBlock(b *Block, reachable bool)
	endBlock(b *Block)
	value(v *Value, live bool)
	startDepCycle()
	endDepCycle()
	// named(n LocalSlot, vals []*Value)
	footer(f *Func)
}

func fprintFunc(p funcPrinter, f *Func) {
	p.header(f)
	for _, b := range f.Blocks {
		p.startBlock(b, true)
		for _, v := range b.Values {
			p.value(v, true)
		}
		p.endBlock(b)
	}
	p.footer(f)
}
