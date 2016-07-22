// ssaview is a small utlity that renders SSA code alongside input Go code

//
// Runs via HTTP on :8080 or the PORT environment variable
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"go/token"
	"io"
	"net/http"
	"sort"
	"strconv"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/yosssi/ace"
)

const indexPage = "index.html"

type members []ssa.Member

var content = map[string]string{
	"Expl":          "Converts a valid go source file into the golang represenation.",
	"scPlaceHolder": "Enter a pure go program without errors.",
	"sch3":          "Source Code",
	"scRender":      "Render source code",
	"sc":            "Source Code",
	"ssah3":         "SSA representation",
	"ssa":           "Example SSA",
	"cbFunctions":   "Show Call information",
	"cbssaType":     "Show SSA type of each instruction",
	"functions":     "",
	"ssatypes":      "",
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func (m members) Len() int           { return len(m) }
func (m members) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m members) Less(i, j int) bool { return m[i].Pos() < m[j].Pos() }

// toSSA converts go source to SSA
func toSSA(source io.Reader, fileName, packageName string, debug bool) ([]byte, error) {
	// adopted from saa package example

	conf := loader.Config{
		Build: &build.Default,
	}

	file, err := conf.ParseFile(fileName, source)
	if err != nil {
		return nil, err
	}

	conf.CreateFromFiles("main.go", file)

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	ssaProg := ssautil.CreateProgram(prog, ssa.NaiveForm|ssa.BuildSerially)
	ssaProg.Build()
	mainPkg := ssaProg.Package(prog.InitialPackages()[0].Pkg)

	out := new(bytes.Buffer)
	mainPkg.SetDebugMode(debug)
	mainPkg.WriteTo(out)
	mainPkg.Build()

	// grab just the functions
	funcs := members([]ssa.Member{})
	for _, obj := range mainPkg.Members {
		if obj.Token() == token.FUNC {
			funcs = append(funcs, obj)
		}
	}

	sort.Sort(funcs)
	for _, f := range funcs {
		mainPkg.Func(f.Name()).WriteTo(out)
		packageFun := mainPkg.Func(f.Name())
		bb := packageFun.Blocks
		// only iterate if special information are wanted
		types := content["ssatypes"] == "true"
		functions := content["functions"] == "true"
		if types || functions {
			for _, block := range bb {
				for _, instr := range block.Instrs {
					out.WriteString(instr.String() + "\n")
					if types {
						ssaType(instr, out)
					}
					if functions {
						call(instr, out)
					}
				}
			}
		}
	}

	return out.Bytes(), nil
}

// writeJSON attempts to serialize data and write it to w
// On error it will write an HTTP status of 400
func writeJSON(w http.ResponseWriter, data interface{}) error {
	if err, ok := data.(error); ok {
		data = struct{ Error string }{err.Error()}
		w.WriteHeader(400)
	}
	o, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return err
	}
	_, err = w.Write(o)
	return err
}

func handler(w http.ResponseWriter, r *http.Request) {
	tpl, err := ace.Load("base", "inner", nil)
	fmt.Printf("tpl == nil %t | err == nil %t", tpl == nil, err == nil)
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
	handleError(err, w)

	// Regenerate the SSA representation
	if r.Method == "POST" {
		err = r.ParseForm()
		handleError(err, w)

		content["functions"] = r.PostFormValue("functions")
		content["ssatypes"] = r.PostFormValue("ssaType")
		fmt.Printf("functions: " + content["functions"] + " ssaTypes " + content["ssatypes"])

		ssaBytes := bytes.NewBufferString(r.PostFormValue("source"))
		var ssa []byte
		ssa, err = toSSA(ssaBytes, "main.go", "main", false)
		handleError(err, w)
		content["sourceCode"] = r.PostFormValue("source")
		content["ssa"] = string(ssa)
		err = tpl.Execute(w, content)
		handleError(err, w)
	}

	err = tpl.Execute(w, content)
	handleError(err, w)
}

func handleError(e error, w http.ResponseWriter) {
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}
}

func ssaType(i ssa.Instruction, out *bytes.Buffer) {
	switch i := i.(type) {
	case *ssa.Alloc:
		out.WriteString("  *ssa.Alloc  ")
		out.WriteString("    Comment: " + i.Comment + " Heap: ")
		if i.Heap {
			out.WriteString("   true \n")
		} else {
			out.WriteString("   false \n")
		}
	case *ssa.BinOp:
		out.WriteString("  *ssa.BinOp   ")
		out.WriteString("    Op: " + i.Op.String() + " X: " + i.X.String() + " Y: " + i.Y.String() + "\n")
		/*				case *ssa.Builtin:
						out.WriteString("*ssa.Builtin  \n") */
	case *ssa.ChangeInterface:
		out.WriteString("  *ssa.ChangeInterface ")
		out.WriteString("   X: " + i.X.String() + "\n")
	case *ssa.ChangeType:
		out.WriteString("  *ssa.ChangeType ")
		out.WriteString("   X: " + i.X.String() + "\n")
		/*				case *ssa.Const:
						out.WriteString("*ssa.Const ")
						out.WriteString("Value: " + i.Value.String() + "\n") */
	case *ssa.Convert:
		out.WriteString("  ssa.Convert ")
		out.WriteString("   X:  " + i.X.String() + "\n")
	case *ssa.Extract:
		out.WriteString("  ssa.Extract")
		out.WriteString("   Tuple: " + i.Tuple.String() + " Index ")
		out.WriteString(strconv.Itoa(i.Index))
	case *ssa.Field:
		out.WriteString("  ssa.Field")
		out.WriteString("   X: " + i.X.String() + " Field " + strconv.Itoa(i.Field) + "\n")
	case *ssa.FieldAddr:
		out.WriteString("  ssa.FieldAddr")
		out.WriteString("    X: " + i.X.String() + " Field " + strconv.Itoa(i.Field) + "\n")
		/*				case *ssa.FreeVar:
						out.WriteString("ssa.FreeVar") */
		/*				case *ssa.Global:
						out.WriteString("*ssa.Global")
						out.WriteString("Pkg " + i.Pkg.String() + "\n") */
	case *ssa.Index:
		out.WriteString("  *ssa.Index ")
		out.WriteString("    X: " + i.X.String() + " Index " + i.Index.String() + " \n")
	case *ssa.IndexAddr:
		out.WriteString("  *ssa.IndexAddr")
		out.WriteString("    X: " + i.X.String() + " Index " + i.Index.String() + "\n")
	case *ssa.Lookup:
		out.WriteString("  *ssa.LookUp")
		out.WriteString("  x: " + i.X.String() + " Index " + i.Index.String() + " CommaOk " + strconv.FormatBool(i.CommaOk) + "\n")
	case *ssa.MakeChan:
		out.WriteString("  *ssa.MakeChan")
		out.WriteString("    Size: " + i.Size.String() + "\n")
	case *ssa.MakeClosure:
		out.WriteString("  *ssa.MakeClosure")
		out.WriteString("    FN: " + i.Fn.String() + " Bindings ")
		for _, i := range i.Bindings {
			out.WriteString("   " + i.String() + " ")
		}
		out.WriteString("\n")
	case *ssa.MakeInterface:
		out.WriteString("  *ssa.MakeInterface")
		out.WriteString("    X: " + i.X.String() + "\n")
	case *ssa.MakeMap:
		out.WriteString("  *ssa.MakeMap")
		if i.Reserve != nil {
			out.WriteString("    Reserve: " + i.Reserve.String() + "\n")
		}
	case *ssa.MakeSlice:
		out.WriteString("  *ssa.MakeSlice")
		out.WriteString("  Len: " + i.Len.String() + " Cap : " + i.Cap.String() + " \n")
	case *ssa.Next:
		out.WriteString("  *ssa.Next")
		out.WriteString("    Iter " + i.Iter.String() + " isString " + strconv.FormatBool(i.IsString) + "\n")
		/*				case *ssa.Parameter:
						out.WriteString("*ssa.Parameter") */
	case *ssa.Phi:
		out.WriteString("  *ssa.Phi")
		out.WriteString("    Comment: " + i.Comment + " Edges ")
		for _, i := range i.Edges {
			out.WriteString("   " + i.String() + " ")
		}
		out.WriteString(" \n")
	case *ssa.Range:
		out.WriteString("  *ssa.Range")
		out.WriteString("    X " + i.X.String() + "\n")
	case *ssa.Select:
		out.WriteString("  *ssa.Select")
		out.WriteString("    States: ")
		for _, i := range i.States {
			out.WriteString("   Channel: " + i.Chan.String() + " Send : " + i.Send.String())
		}
		out.WriteString("   Blocking: " + strconv.FormatBool(i.Blocking) + "\n")
	case *ssa.Send:
		out.WriteString("  *ssa.Send")
		out.WriteString("    Chan: " + i.Chan.String() + " X: " + i.X.String() + "\n")
	case *ssa.Slice:
		out.WriteString("  *ssa.Slice")
		if i.X != nil {
			out.WriteString("    x: " + i.X.String())
		}
		if i.Low != nil {
			out.WriteString("    Low: " + i.Low.String())
		}
		if i.Max != nil {
			out.WriteString("    Max: " + i.Max.String())
		}
		out.WriteString("\n")
	case *ssa.Store:
		out.WriteString("  *ssa.Store")
		out.WriteString("      Addr: " + i.Addr.String() + "  Val " + i.Addr.String() + "\n")
	case *ssa.TypeAssert:
		out.WriteString("  *ssa.TypeAssert")
		out.WriteString("    X: " + i.X.String() + " AssertedType: " + i.AssertedType.String() + " CommaOk " + strconv.FormatBool(i.CommaOk) + "\n")
	case *ssa.UnOp:
		out.WriteString("  *ssa.UnoOp")
		out.WriteString("    Op: " + i.Op.String() + " X: " + i.X.String() + " CommaOk: " + strconv.FormatBool(i.CommaOk) + "\n")
	}
}

func call(i ssa.Instruction, out *bytes.Buffer) {
	var callCom ssa.CallCommon
	switch i := i.(type) {
	case *ssa.Call:
		callCom = i.Call
	case *ssa.Go:
		callCom = i.Call
	case *ssa.Defer:
		callCom = i.Call
	default:
		return
	}
	// invoked call
	out.WriteString("  call is invoked")
	if callCom.IsInvoke() {
		out.WriteString(" true\n")
	} else {
		out.WriteString(" false\n")
	}
	// method
	if callCom.Method == nil {
		out.WriteString("  value is a *Builtin ")
		if _, ok := callCom.Value.(*ssa.Builtin); ok {
			out.WriteString(" true\n")
		} else {
			out.WriteString(" false\n")
		}
	}
	// signatue
	if callCom.Value != nil {
		out.WriteString("  signature(Value):" + callCom.Value.Name() + " " + callCom.Signature().String() + "\n")
	}
	if callCom.StaticCallee() != nil {
		out.WriteString("  signature:" + callCom.StaticCallee().Signature.String() + "\n")
	}
	// arguments
	for i, arg := range callCom.Args {
		if i == 1 {
			out.WriteString("  args:")
		}
		out.WriteString(arg.String() + "|")
	}
	out.WriteString("\n")
}
