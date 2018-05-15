package xlsx2pb

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	configRegexp = "xlsx*.config"
	sheetNames   map[string]struct{} // check if duplicate sheet name exists
	sheetFileMap map[string][]string // map[filename][]sheetnames, e.g. ["Sheet1", "Sheet2, Sheet3"]
	fileHashMap  map[string][16]byte // map[filename]MD5
)

func init() {
	ResetConfigCache()
}

// ResetConfigCache clear current config data
func ResetConfigCache() {
	sheetNames = make(map[string]struct{})
	sheetFileMap = make(map[string][]string)
	fileHashMap = make(map[string][16]byte)
}

func getConfigFiles(tarPath string) []string {
	files, err := filepath.Glob(filepath.Join(tarPath, configRegexp))
	if err != nil {
		log.Fatal(err)
	}

	return files
}

func readCfgFile(cfgFile string) {
	file, err := os.Open(cfgFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := readCfgLine(scanner.Text()); err != nil {
			log.Panicln(err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}

func readCfgLine(cfgLine string) error {
	parts := strings.Fields(cfgLine)
	if len(parts) != 2 {
		return fmt.Errorf("%v is illegel in config", cfgLine)
	}

	filename := parts[1]

	if _, ok := sheetFileMap[filename]; !ok {
		sheetFileMap[filename] = make([]string, 0)
	}

	// check if duplicate sheet name exists
	sheets := strings.Split(parts[0], ",")
	for _, sheet := range sheets {
		if _, ok := sheetNames[sheet]; ok {
			return fmt.Errorf("%v name duplicates", sheet)
		}

		sheetNames[sheet] = struct{}{}
	}

	sheetFileMap[filename] = append(sheetFileMap[filename], parts[0])

	return nil
}
