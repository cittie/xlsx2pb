package xlsx2pb

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

var (
	// INDENT Use 2 spaces
	INDENT    = "  "
	curIndent string
	outPrefix = "./proto/"
	outSuffix = ".proto"
)

// GenProto genenate proto file content
func (pr *ProtoRow) GenProto() {
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

// IncreaseIndent increase indent
func (pr *ProtoRow) IncreaseIndent() {
	pr.tabCount++
	curIndent = strings.Repeat(INDENT, pr.tabCount)
}

// DecreaseIndent decrease indent
func (pr *ProtoRow) DecreaseIndent() {
	if pr.tabCount > 0 {
		pr.tabCount--
	}
	curIndent = strings.Repeat(INDENT, pr.tabCount)
}

// AddOneEmptyLine add an empty line
func (pr *ProtoRow) AddOneEmptyLine() {
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s", curIndent))
}

// AddMessageHead add a proto message head
func (pr *ProtoRow) AddMessageHead(name string) {
	pr.outProto = append(pr.outProto, fmt.Sprintf("%smessage %s {", curIndent, name))
	pr.IncreaseIndent()
}

// AddOneDefination add a proto defination
func (pr *ProtoRow) AddOneDefination(isRepeat bool, comment, typ, name string, idx *int) {
	if comment != "" {
		pr.outProto = append(pr.outProto, fmt.Sprintf("%s/* %s */", curIndent, comment)) // comment
	}
	if name == "" {
		name = firstLetter2Lowercase(typ)
	}
	if isRepeat {
		typ = "repeated " + typ
	}
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s%s %s = %d;", curIndent, typ, name, *idx+1)) // define
	*idx++
}

// AddMessageTail add a proto message tail
func (pr *ProtoRow) AddMessageTail() {
	pr.DecreaseIndent()
	pr.outProto = append(pr.outProto, fmt.Sprintf("%s}", curIndent))
	pr.AddOneEmptyLine()
}

// AddVal add a proto variable define
func (pr *ProtoRow) AddVal(sh *Val) {
	pr.AddOneDefination(false, sh.comment, sh.typ, sh.name, &pr.varIdx)
	sh.fieldNum = pr.varIdx
}

// AddRepeat add repeat struct
func (pr *ProtoRow) AddRepeat(repeat *Repeat) {
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
func (pr *ProtoRow) AddRepeatDefine(repeat *Repeat) {
	for _, sh := range repeat.fields {
		pr.AddOneDefination(false, sh.comment, sh.typ, sh.name, &repeat.repeatIdx)
	}
}

// AddRepeatTail add repeat declare
func (pr *ProtoRow) AddRepeatTail(repeat *Repeat) {
	pr.AddOneDefination(true, repeat.comment, repeat.fieldName, repeat.comment, &pr.varIdx)
	repeat.fieldNum = pr.varIdx
}

// AddMessageArray add an array for current message as XXX_ARRAY
func (pr *ProtoRow) AddMessageArray() {
	pr.AddMessageHead(pr.Name + "_ARRAY")
	pr.outProto = append(pr.outProto, fmt.Sprintf("%srepeated %s %s = 1;", curIndent, pr.Name, firstLetter2Lowercase(pr.Name)))
	pr.AddMessageTail()
}

// Write output proto file to "./proto/sheetname.proto"
func (pr *ProtoRow) Write() {
	f, err := os.Create(outPrefix + strings.ToLower(pr.Name) + outSuffix)
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
