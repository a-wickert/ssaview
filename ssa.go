package main

/*func toSSA(src io.Reader, file, pkg string) (*SSA, error) {
	var fs []*ssa.Function
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
				fs = append(fs, f)
			}
		}
	}
	return &SSA{fs}, nil
} */
