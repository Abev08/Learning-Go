package main

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const CONFIG_FILE_PATH string = "config.ini"

type Config uint16

// Collection of config values
var ConfigData = make(map[Config]string, SomethingElse+1) // Last variable + 1 provided as a size

const (
	Variable1 Config = iota
	OtherVariable
	SomethingElse
)

func (c Config) ToString() string {
	switch c {
	case Variable1:
		return "Variable1"
	case OtherVariable:
		return "OtherVariable"
	case SomethingElse:
		return "SomethingElse"
	default:
		return ""
	}
}

// Pass '0' as first argumnet
func (c Config) FromString(s string) (Config, error) {
	switch s {
	case "Variable1":
		return Variable1, nil
	case "OtherVariable":
		return OtherVariable, nil
	case "SomethingElse":
		return SomethingElse, nil
	default:
		return 0, errors.New("index out of range")
	}
}

func main() {
	var err error
	var file *os.File
	file, err = os.Open(CONFIG_FILE_PATH)
	if file == nil {
		file, err = createConfigFile(CONFIG_FILE_PATH)
	}
	if err != nil {
		slog.Error("Couldn't open config.ini file", "Err", err)
		return
	}

	parseConfigFile(file)
}

func createConfigFile(fileName string) (*os.File, error) {
	var file, err = os.Create(fileName)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("Creating new %s file", fileName))

	// Create config file contents
	file.WriteString("; This is config file\n")
	file.WriteString("\n")
	file.WriteString("\n")
	file.WriteString(fmt.Sprintf("%s = a1\n", Variable1.ToString()))
	file.WriteString("UnrecognizedVariable = b2\n")
	file.WriteString(fmt.Sprintf("%s = c3\n", OtherVariable.ToString()))
	file.WriteString("; DeprecatedVar2 = d4\n")
	file.WriteString(fmt.Sprintf("%s = e5\n", SomethingElse.ToString()))

	return os.Open(file.Name())
}

func parseConfigFile(file *os.File) {
	var reader = bufio.NewReader(file)

	for {
		var data, err = reader.ReadBytes('\n')
		if err != nil {
			return
		}
		var s = strings.TrimSpace(string(data))
		s = strings.TrimSuffix(s, "\r\n")
		s = strings.TrimSuffix(s, "\n")

		// Skip empty and commented out lines
		if len(s) == 0 || strings.HasPrefix(s, "//") || strings.HasPrefix(s, ";") {
			continue
		}
		fmt.Printf("Parsing config line: %s\n", s)

		// Get variable and value parts
		var index = strings.Index(s, "=")
		var v = strings.TrimSpace(s[:index])
		var variable Config
		variable, err = Config.FromString(0, v)
		if err != nil {
			slog.Warn("Variable not recognized", "Var", v)
			continue
		}
		var value = s[(index + 1):]
		index = strings.Index(value, ";")
		if index > 0 {
			value = value[:index]
		}
		index = strings.Index(value, "//")
		if index > 0 {
			value = value[:index]
		}
		value = strings.TrimSpace(value)

		// Set ConfigData
		switch variable {
		case Variable1:
			// etc.
		}
		fmt.Printf("Variable: %s, Old value: %s, New value: %s\n", variable.ToString(), ConfigData[variable], value)
		ConfigData[variable] = value
	}
}
