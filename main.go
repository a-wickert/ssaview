// ssaview is a small utlity that renders SSA code alongside input Go code

//
// Runs via HTTP on :8080 or the PORT environment variable
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/yosssi/ace"
)

const indexPage = "index.html"
const stdPort = "8080"

type members []ssa.Member
type Cb struct {
	Description string
	Name        string
	Checked     bool
}

type SSA struct {
	Funcs []Func
}
type Func struct {
	Name     string
	Params   []Value
	PString  string
	FreeVars []Value
	FString  string
	Locals   []Value
	LString  string
	Blocks   []BB
	BString  string
	//	AnonFuncs []Func
}

type Instr struct {
	Name string
	Type string
}

type Value struct {
	Name string
	Type string
}

type BB struct {
	Index  int
	Instrs []Instr
	Preds  []int
	Succs  []int
}

var content = map[string]interface{}{
	"Expl":          "Converts a valid go source file into the SSA represenation.",
	"scPlaceHolder": "Enter a pure go program without errors.",
	"sch3":          "Source Code",
	"scRender":      "Render source code",
	"sc":            "Source Code",
	"ssah3":         "SSA representation",
	//"ssa":           "Example SSA",
	"pagename": "SSA view",
	"cbs": []Cb{
		Cb{"Show call information", "functions", false},
		//		Cb{"Show SSA type of each instruction", "ssaType", false},
		Cb{"Show Idom of each basic block", "idom", false},
		Cb{"Build with the build mode: SanityCheckFunctions", "ssabuild", true},
	},
}

func main() {

	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = stdPort
		fmt.Printf("INFO: Port environment variable is not set. Use port %s", port)
	}
	port = ":" + port
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func (m members) Len() int           { return len(m) }
func (m members) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m members) Less(i, j int) bool { return m[i].Pos() < m[j].Pos() }

// toSSA converts go source to SSA
/*func toSSA(source io.Reader, fileName, packageName string, debug bool) ([]byte, error) {
	// adopted from saa package example

	var conf loader.Config
	var true = "true"

	file, err := conf.ParseFile(fileName, source)
	if err != nil {
		return nil, err
	}

	conf.CreateFromFiles("main.go", file)

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}
	buildsanity := content["ssabuild"] == true
	var ssaProg *ssa.Program
	if buildsanity {
		ssaProg = ssautil.CreateProgram(prog, ssa.SanityCheckFunctions)
	} else {
		ssaProg = ssautil.CreateProgram(prog, ssa.NaiveForm)
	}

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
		types := content["ssaType"] == true
		functions := content["functions"] == true
		idom := content["idom"] == true
		if types || functions || idom {
			for _, block := range bb {
				if idom {
					printIdom(block, out)
				}
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
*/

func toSSA(src io.Reader, file, pkg string) (SSA, error) {
	var fs []Func
	var conf loader.Config

	// Parse the file into a ssa file
	f, _ := conf.ParseFile(file, src)
	conf.CreateFromFiles("main.go", f)
	p, _ := conf.Load()
	buildsanity := content["ssabuild"] == true
	var ssap *ssa.Program
	if buildsanity {
		ssap = ssautil.CreateProgram(p, ssa.SanityCheckFunctions)
	} else {
		ssap = ssautil.CreateProgram(p, ssa.NaiveForm)
	}

	// Build ssa prog to retrieve all information and the main pkg
	ssap.Build()
	mainpkg := ssap.Package(p.InitialPackages()[0].Pkg)

	for _, m := range mainpkg.Members {
		if m.Token() == token.FUNC {
			f, ok := m.(*ssa.Function)
			if ok {
				var params []Value
				for _, p := range f.Params {
					v := Value{p.Name(), reflect.TypeOf(p).String()}
					params = append(params, v)
				}
				var freevars []Value
				for _, fv := range f.FreeVars {
					v := Value{fv.Name(), reflect.TypeOf(fv).String()}
					freevars = append(freevars, v)
				}
				var locals []Value
				for _, l := range f.Locals {
					v := Value{l.Name(), reflect.TypeOf(l).String()}
					locals = append(locals, v)
				}
				var blocks []BB
				for _, b := range f.Blocks {
					var instrs []Instr
					for _, i := range b.Instrs {
						in := Instr{i.String(), reflect.TypeOf(i).String()}
						instrs = append(instrs, in)
					}
					var preds []int
					for _, p := range b.Preds {
						preds = append(preds, p.Index)
					}
					var succs []int
					for _, s := range b.Succs {
						succs = append(succs, s.Index)
					}
					bb := BB{b.Index, instrs, preds, succs}
					blocks = append(blocks, bb)
				}
				fn := Func{f.Name(), params, "par_" + f.Name(), freevars, "freevars_" + f.Name(), locals, "locals_" + f.Name(), blocks, "blocks_" + f.Name()}
				fs = append(fs, fn)
			}
		}
	}
	return SSA{fs}, nil
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
	if err != nil {
		fmt.Printf("error %s", err.Error())
	}
	handleError(err, w)

	// Generate the SSA representation
	if r.Method == "POST" {
		err = r.ParseForm()
		handleError(err, w)

		// Iterate over the checkboxes
		// content[cb.Name] is used in the toSSA algorithm
		cbs := content["cbs"].([]Cb)
		for i, cb := range cbs {
			if r.PostFormValue(cb.Name) == "true" {
				content[cb.Name] = "true"
				cbs[i].Checked = true
			} else {
				content[cb.Name] = "false"
				cbs[i].Checked = false
			}
		}

		ssaBytes := bytes.NewBufferString(r.PostFormValue("source"))
		ssafs, err := toSSA(ssaBytes, "main.go", "main")
		handleError(err, w)
		content["sourceCode"] = r.PostFormValue("source")
		content["ssa"] = ssafs
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

func printIdom(b *ssa.BasicBlock, out *bytes.Buffer) {
	if b.Index == 0 {
		out.WriteString("Basic Block has no idom because it is a entry node.")
		return
	}
	if b == b.Parent().Recover {
		out.WriteString("Basic Block has no idom because it is a recover node")
		return
	}
	out.WriteString("  Idom of ")
	out.WriteString(strconv.Itoa(b.Index))
	out.WriteString(" is: ")
	out.WriteString(strconv.Itoa(b.Idom().Index))
	out.WriteString("\n")
}
