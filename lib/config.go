package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

type CConfig struct {
	Config Config `toml:"Config"`
}

type Config struct {
	ConfigRegExp string `toml:"config_reg_exp"`
	XlsxPath     string `toml:"xlsx_path"`
	XlsxExt      string `toml:"xlsx_ext"`
	PackageName  string `toml:"package_name"`
	UseProto3    bool   `toml:"use_proto3"`
	ProtoOutPath string `toml:"proto_path"`
	ProtoOutExt  string `toml:"proto_ext"`
	DataOutPath  string `toml:"data_path"`
	DataOutExt   string `toml:"data_ext"`
	CacheFile    string `toml:"cache_file"`
}

var (
	cfg          *Config             // config reading from <package>/conf/config.toml
	sheetNames   map[string]struct{} // check if duplicate sheet name exists
	sheetFileMap map[string][]string // map[filename][]sheetnames, e.g. ["Sheet1", "Sheet2, Sheet3"]
	fileHashMap  map[string][16]byte // map[filename]MD5
)

func init() {
	cfg = new(Config)
	cfg.LoadConfig()
	cfg.ReplaceRelPaths()

	ResetConfigCache()
	for _, cfg := range getConfigFiles(cfg.XlsxPath) {
		fmt.Printf("Config file %s found\n", cfg)
		readCfgFile(cfg)
	}
}

// ResetConfigCache clear current config data
func ResetConfigCache() {
	sheetNames = make(map[string]struct{})
	sheetFileMap = make(map[string][]string)
	fileHashMap = make(map[string][16]byte)
}

func getConfigFiles(tarPath string) []string {
	files, err := filepath.Glob(filepath.Join(tarPath, cfg.ConfigRegExp))
	if err != nil {
		panic(err)
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
			return fmt.Errorf("%s name duplicates\n", sheet)
		}

		sheetNames[sheet] = struct{}{}
	}

	sheetFileMap[filename] = append(sheetFileMap[filename], parts[0])

	return nil
}

func (c *Config) LoadConfig() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information!")
	}

	cfgFile := filepath.Join(path.Dir(filename), "../conf/config.toml")

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		panic("Config file does not exits!")
	}

	ccfg := new(CConfig)
	if _, err := toml.DecodeFile(cfgFile, ccfg); err != nil {
		panic(err)
	}

	*c = ccfg.Config
}

func (c *Config) ReplaceRelPaths() {
	replaceRelPath := func(path *string) {
		absPath, err := filepath.Abs(*path)
		if err != nil {
			panic(err)
		}
		*path = absPath
	}
	replaceRelPath(&c.XlsxPath)
	replaceRelPath(&c.ProtoOutPath)
	replaceRelPath(&c.DataOutPath)
	replaceRelPath(&c.CacheFile)
}
