package lib

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var (
	// INDENT Use 2 spaces
	INDENT    = "  "
	curIndent string
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

	// Repeats
	for _, repeat := range pr.repeats {
		pr.AddRepeat(repeat)
	}

	pr.AddMessageTail()

	// MessageArray
	pr.AddMessageArray()
}

// Hash generate hash of proto content
func (pr *ProtoSheet) Hash() {
	hash := md5.New()
	for _, line := range pr.outProto {
		hash.Write([]byte(line))
	}
	pr.hash = hash.Sum(nil)
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
	pr.outProto = append(pr.outProto, fmt.Sprintf("%smessage %s {", curIndent, name))
	pr.IncreaseIndent()
}

// AddOneDefination add a proto defination
func (pr *ProtoSheet) AddOneDefination(isRepeat bool, comment, p2type, typ, name string, idx *int) {
	if comment != "" {
		pr.outProto = append(pr.outProto, fmt.Sprintf("%s/* %s */", curIndent, comment)) // comment
	}
	if name == "" {
		name = firstLetter2Lowercase(typ) // struct use type name as name
	}
	if typ == "float" || typ == "float64" { // convert go varient name to proto varient name
		typ = "double"
	}
	if !pr.isProto3 && p2type != "" { // only for proto2
		typ = fmt.Sprintf("%s %s", p2type, typ)
	}
	if isRepeat {
		typ = "repeated " + typ
	}
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s%s %s = %d;", curIndent, typ, name, *idx+1)) // define
	*idx++
}

// AddMessageTail add a proto message tail
func (pr *ProtoSheet) AddMessageTail() {
	pr.DecreaseIndent()
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s}", curIndent))
	pr.AddOneEmptyLine()
}

// AddVal add a proto variable define
func (pr *ProtoSheet) AddVal(sh *Val) {
	pr.AddOneDefination(false, sh.comment, sh.proto2Type, sh.typ, sh.name, &pr.varIdx)
	sh.fieldNum = pr.varIdx
}

// AddRepeat add repeat struct
func (pr *ProtoSheet) AddRepeat(repeat *Repeat) {
	pr.AddOneEmptyLine()
	// Add message head
	pr.AddMessageHead(repeat.fieldName)

	// Add repeat vals
	pr.AddRepeatDefine(repeat)

	// Add message tail
	pr.AddMessageTail()

	// Add repeat tail
	pr.AddRepeatTail(repeat)
}

// AddRepeatDefine add repeat inner define
func (pr *ProtoSheet) AddRepeatDefine(repeat *Repeat) {
	for _, sh := range repeat.fields {
		pr.AddOneDefination(false, sh.comment, sh.proto2Type, sh.typ, sh.name, &repeat.repeatIdx)
	}
}

// AddRepeatTail add repeat declare
func (pr *ProtoSheet) AddRepeatTail(repeat *Repeat) {
	pr.AddOneDefination(true, repeat.comment, "", repeat.fieldName, repeat.comment, &pr.varIdx)
	repeat.fieldNum = pr.varIdx
}

// AddMessageArray add an array for current message as XXX_ARRAY
func (pr *ProtoSheet) AddMessageArray() {
	pr.AddMessageHead(pr.Name + "_ARRAY")
	pr.outProto = append(pr.outProto, fmt.Sprintf("%srepeated %s %s = 1;", curIndent, pr.Name, firstLetter2Lowercase(pr.Name)))
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
		w.WriteString(line + "\n")
	}

	w.Flush()
}

func firstLetter2Lowercase(title string) string {
	runes := []rune(title)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}