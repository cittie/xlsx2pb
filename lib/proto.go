package lib

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// INDENT Use 2 spaces
	INDENT               = "  "
	curIndent            string
	defaultStructureName = "Default"
)

// GenProto genenate proto file content
func (pr *ProtoSheet) GenProto() {
	pr.AddPreHead()

	// Head
	pr.AddMessageHead(pr.Name)

	// Vars
	for _, val := range pr.vars {
		pr.AddVal(val)
	}

	// Optional Struct
	for _, optS := range pr.optStructs {
		pr.AddOptionalStruct(optS)
	}

	// Repeats
	for _, repeat := range pr.repeats {
		pr.AddRepeat(repeat)
	}

	pr.AddMessageTail()

	// MessageArray
	pr.AddMessageArray()
}

// Hash generate hash of proto content
func (pr *ProtoSheet) ProtoHash() {
	hash := md5.New()
	for _, line := range pr.outProto {
		hash.Write([]byte(line))
	}
	pr.protoHash = hash.Sum(nil)
}

// IncreaseIndent increase indent
func (pr *ProtoSheet) IncreaseIndent() {
	pr.tabCount++
	curIndent = strings.Repeat(INDENT, pr.tabCount)
}

// DecreaseIndent decrease indent
func (pr *ProtoSheet) DecreaseIndent() {
	if pr.tabCount > 0 {
		pr.tabCount--
	}
	curIndent = strings.Repeat(INDENT, pr.tabCount)
}

// AddOneEmptyLine add an empty line
func (pr *ProtoSheet) AddOneEmptyLine() {
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s", curIndent))
}

// AddPreHead add syntax and package info
func (pr *ProtoSheet) AddPreHead() {
	ver := 2
	if pr.isProto3 {
		ver = 3
	}
	pr.outProto = append(pr.outProto, fmt.Sprintf("syntax = \"proto%d\";", ver))
	pr.outProto = append(pr.outProto, fmt.Sprintf("package %s;", cfg.PackageName))
	pr.AddOneEmptyLine()
}

// AddMessageHead add a proto message head
func (pr *ProtoSheet) AddMessageHead(name string) {
	if name == "" {
		return
	}
	pr.outProto = append(pr.outProto, fmt.Sprintf("%smessage %s {", curIndent, name))
	pr.IncreaseIndent()
}

// AddOneDefine add a proto defination
func (pr *ProtoSheet) AddOneDefine(isRepeat bool, comment, p2type, typ, name, defaultValStr string, idx int) {
	// default value
	defaultStr := fmt.Sprintf(" [default = %v]", defaultValStr)
	if defaultValStr == "" {
		defaultStr = ""
	}

	// comment
	if comment != "" {
		pr.outProto = append(pr.outProto, fmt.Sprintf("%s/* %s */", curIndent, comment)) // comment
	}

	// name
	if name == "" {
		name = title2Lowercase(strings.TrimSpace(typ)) // struct use type name as name
		if name == "" {                                // if name is still missing, use a default name
			name = defaultStructureName
		}
	}

	// type
	if typ == "float32" { // convert go varient name to proto varient name
		typ = "float"
	}
	if typ == "float64" { // convert go varient name to proto varient name
		typ = "double"
	}
	if !pr.isProto3 && p2type != "" { // only for proto2
		typ = fmt.Sprintf("%s %s", p2type, typ)
	}
	if isRepeat {
		typ = "repeated " + typ
	}

	// generate line
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s%s %s = %d%v;", curIndent, typ, name, idx, defaultStr)) // define
}

// AddMessageTail add a proto message tail
func (pr *ProtoSheet) AddMessageTail() {
	pr.DecreaseIndent()
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s}", curIndent))
	pr.AddOneEmptyLine()
}

// AddVal add a proto variable define
func (pr *ProtoSheet) AddVal(sh *Val) {
	pr.AddOneDefine(false, sh.comment, sh.proto2Type, sh.typ, sh.name, sh.defaultValueStr, sh.fieldNum)
}

func (pr *ProtoSheet) AddOptionalStruct(optS *OptStruct) {
	pr.AddOneEmptyLine()

	pr.AddMessageHead(optS.name)

	for _, val := range optS.fields {
		pr.AddVal(val)
	}

	pr.AddMessageTail()

	pr.AddOptStructTail(optS)
}

// AddRepeat add repeat struct
func (pr *ProtoSheet) AddRepeat(repeat *Repeat) {
	pr.AddOneEmptyLine()

	if repeat.opts != nil {
		pr.AddMessageHead(repeat.opts.name)

		for _, val := range repeat.opts.fields {
			pr.AddVal(val)
		}

		pr.AddMessageTail()
	}

	// Add repeat tail
	pr.AddRepeatTail(repeat)
}

// AddRepeatTail add repeat declare
func (pr *ProtoSheet) AddRepeatTail(repeat *Repeat) {
	if repeat.val != nil {
		pr.AddOneDefine(true, repeat.val.comment, "", repeat.val.typ, repeat.val.name, "", repeat.fieldNum)
		return
	}

	pr.AddOneDefine(true, "", "", repeat.name, repeat.comment, "", repeat.fieldNum)
}

func (pr *ProtoSheet) AddOptStructTail(opts *OptStruct) {
	pr.AddOneDefine(false, "", "optional", opts.name, opts.comment, "", opts.fieldNum)
}

// AddMessageArray add an array for current message as XXX_ARRAY
func (pr *ProtoSheet) AddMessageArray() {
	pr.AddMessageHead(pr.Name + "_ARRAY")
	pr.outProto = append(pr.outProto, fmt.Sprintf("%srepeated %s items = 1;", curIndent, pr.Name))
	pr.AddMessageTail()
}

// WriteProto output proto file to "./proto/sheetname.proto"
func (pr *ProtoSheet) WriteProto() {
	f, err := os.Create(filepath.Join(cfg.ProtoOutPath, strings.ToLower(pr.Name)+cfg.ProtoOutExt))
	defer f.Close()

	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)
	for _, line := range pr.outProto {
		_, err := w.WriteString(line + "\n")
		if err != nil {
			panic(err)
		}
	}

	err = w.Flush()
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		fmt.Println(err.Error())
	}
}

func title2Lowercase(title string) string {
	if title == "" {
		return ""
	}

	return strings.ToLower(title) + "s"
}
