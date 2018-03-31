# xlsx2pb
convert xlsx to protobuf

## Installation ##
`go get -u github.com/cittie/xlsx2pb`
`go install github.com/cittie/xlsx2pb`

## Usage ##
Generate config files as "xlsx_xxxxx.config" in xlsx directory with the following content:

`SHEETNAME4 XLSXFILENAME2.xlsx`

If sheets have the same structure, it can be written as following:

`SHEETNAME1,SHEETNAME2,SHEETNAME3 XLSXFILENAME1.xlsx`

Then run the command:

`xlsx2pb xlsx_input_path proto_output_path`
