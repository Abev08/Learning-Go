package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

// Adds comments to SDL3 Go wrapper.

const (
	WIKI_URL string =
	//
	"https://wiki.libsdl.org/SDL3" // Url for main SDL package wiki
	// "https://wiki.libsdl.org/SDL3_ttf" // Url for SDL TTF package wiki
	// "https://wiki.libsdl.org/SDL3_image" // Url for SDL Image package wiki

	NAME_PREFIX string =
	//
	"SDL" // Prefix for main SDL package
	// "TTF" // Prefix for SDL TTF package
	// "IMG" // Prefix for SDL Image package
)

var FilePath string = "sdl/sdl_keyboard.go" // Path to the file that should be parsed (also can be provided in command line args).

var previousResponse string // Previous HTTP response from wiki.libsdl.org
var previousType string     // Previously requested type

func main() {
	if len(os.Args) >= 2 {
		FilePath = os.Args[1]
	}

	if len(FilePath) == 0 {
		slog.Error("provide path to the file")
		return
	} else if len(WIKI_URL) == 0 || len(NAME_PREFIX) == 0 {
		slog.Error("fix the const global variables")
		return
	}

	// Open the file
	file, err := os.Open(FilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	parsingStruct, parsingConst, constType := false, false, ""
	sb := strings.Builder{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "func") {
			AddFunctionComment(&sb, line) // Add comment to function definition
		} else if strings.HasPrefix(line, "type") {
			if strings.HasSuffix(line, "struct {") {
				AddTypeComment(&sb, line) // Add comment to a type definition
				fmt.Fprintf(&sb, "%s\n", line)

				// Check if it's empty struct
				if !strings.HasSuffix(line, "}") {
					parsingStruct = true
				}
			} else {
				AddTypeComment(&sb, line) // Add comment to a type definition
			}
		} else if strings.HasPrefix(line, "const (") {
			fmt.Fprintf(&sb, "%s\n", line)
			parsingConst = true
			constType = ""
		} else if strings.HasPrefix(line, ")") {
			parsingConst = false
		} else if strings.HasPrefix(line, "}") {
			parsingStruct = false
		} else if parsingStruct {
			AddStructComment(&sb, line) // Add comment to a struct field definition
		} else if parsingConst {
			if len(constType) == 0 {
				// Get wiki entry of the const type
				idx := strings.Index(line, "=")
				fields := strings.Fields(line[:idx])
				if len(fields) == 1 { // int type? use last parsed type?
					constType = previousType
				} else if len(fields) >= 2 {
					constType = fields[1]
				}

				url := WIKI_URL + "/" + NAME_PREFIX + "_" + constType
				previousResponse = getWikiEntry(url, line)
				if len(previousResponse) == 0 {
					parsingConst = false // Some error happened, stop parsing the const block
				}
			}

			if parsingConst {
				AddConstComment(&sb, line)
			}
		}

		if !parsingStruct && !parsingConst {
			fmt.Fprintf(&sb, "%s\n", line)
		}
	}

	err = os.WriteFile(FilePath, []byte(sb.String()), os.ModePerm)
	if err != nil {
		panic(err)
	}

	// Run gofmt on the file
	cmd := exec.Command("gofmt", "-w", FilePath)
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}

func AddFunctionComment(sb *strings.Builder, line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}

	var name string
	count := strings.Count(line, ")")
	switch count {
	case 1:
		// Simple function definition "func Name(xxx) {"
		idx := strings.Index(fields[1], "(")
		if idx < 0 {
			idx = strings.Index(fields[1], "[")
		}
		if idx < 0 {
			idx = len(fields[1])
		}
		name = fields[1][:idx]
	case 3:
		// Function definition has struct pointer and tuple return
		idx := strings.Index(fields[3], "(")
		if idx < 0 {
			idx = strings.Index(fields[3], "[")
		}
		if idx < 0 {
			idx = len(fields[3])
		}
		name = fields[3][:idx]
	default:
		// Function definition has struct pointer "func (p *P) Name() {" or tuple as return value
		for i := 1; i < len(fields); i++ {
			if strings.HasSuffix(fields[i], "()") || strings.HasSuffix(fields[i], "(") {
				idx := strings.Index(fields[i], "(")
				if idx < 0 {
					idx = strings.Index(fields[i], "[")
				}
				if idx < 0 {
					idx = len(fields[i])
				}
				name = fields[i][:idx]
				break
			}
		}
	}

	if len(name) == 0 {
		slog.Warn("error couldn't get function name", "function", line)
		return
	}
	slog.Info("Requesting comment for", "function", name)
	time.Sleep(time.Millisecond * 500) // Wait at least 500 ms between HTTP requests

	url := WIKI_URL + "/" + NAME_PREFIX + "_" + name
	response := getWikiEntry(url, line)
	previousResponse = response
	if len(response) == 0 {
		return
	}
	idx := strings.Index(response, name+"</h1>")
	if idx <= 0 {
		slog.Error("error, response from wiki.libsdl.org is missing function name?")
		return
	}
	idx += strings.Index(response[idx:], "<p>") + 3
	endIdx := idx + strings.Index(response[idx:], "</p>")

	// Modify beginning of the comment
	commentFields := strings.Fields(response[idx:endIdx])
	commentFields[0] = strings.ToLower(commentFields[0])
	if commentFields[0] == "use" {
		// Don't modify
	} else if commentFields[0] == "copy" {
		commentFields[0] = "copies"
	} else if commentFields[0] == "query" {
		commentFields[0] = "queries"
	} else if commentFields[0] == "dismiss" {
		commentFields[0] = "dismisses"
	} else if !strings.HasSuffix(commentFields[0], "s") {
		commentFields[0] += "s"
	}
	comment := prettifyComment(strings.Join(commentFields, " "), false)

	// Append comment lines
	fmt.Fprintf(sb, "// [%s] %s\n//\n", name, comment)
	fmt.Fprintf(sb, "// [%s]: %s\n", name, url)
}

func AddTypeComment(sb *strings.Builder, line string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}
	name := fields[1]
	previousType = name
	if len(name) == 0 {
		slog.Warn("error couldn't get type name", "type", line)
		return
	}

	slog.Info("Requesting comment for", "type", name)
	time.Sleep(time.Millisecond * 500) // Wait at least 500 ms between HTTP requests

	url := WIKI_URL + "/" + NAME_PREFIX + "_" + name
	response := getWikiEntry(url, line)
	previousResponse = response
	if len(response) == 0 {
		return
	}

	idx := strings.Index(response, name+"</h1>")
	if idx <= 0 {
		slog.Error("error, response from wiki.libsdl.org is missing function name?")
		return
	}
	idx += strings.Index(response[idx:], "<p>") + 3
	endIdx := idx + strings.Index(response[idx:], "</p>")

	// Modify beginning of the comment
	commentFields := strings.Fields(response[idx:endIdx])
	switch commentFields[0] {
	case "The":
		commentFields[0] = "is a structure specifying the"

		if !strings.HasSuffix(commentFields[len(commentFields)-1], "s.") {
			commentFields[len(commentFields)-1] = commentFields[len(commentFields)-1][:len(commentFields[len(commentFields)-1])-1]
			commentFields[len(commentFields)-1] += "s."
		}
	case "A":
		commentFields[0] = "is a"
	case "An":
		commentFields[0] = "is an"
	case "This":
		commentFields = slices.Delete(commentFields, 0, 1)
	default:
		commentFields[0] = "defines the " + strings.ToLower(commentFields[0])
	}
	comment := prettifyComment(strings.Join(commentFields, " "), false)

	// Append comment lines
	fmt.Fprintf(sb, "// [%s] %s\n//\n", name, comment)
	fmt.Fprintf(sb, "// [%s]: %s\n", name, url)
}

// Add struct fields comments
func AddStructComment(sb *strings.Builder, line string) {
	fields := strings.Fields(line)
	fieldName := strings.TrimPrefix(fields[0], "*")
	fieldName = strings.ToLower(string(fieldName[0])) + fieldName[1:] + ";"

	var lineStart, lineEnd, commentStart, commentEnd int
	lineStart = strings.Index(previousResponse, fieldName)
	if lineStart >= 0 {
		lineEnd = lineStart + strings.Index(previousResponse[lineStart:], "\n")
		commentStart = lineStart + strings.Index(previousResponse[lineStart:], "/**&lt;") + 7
		if commentStart >= 0 {
			commentEnd = commentStart + strings.Index(previousResponse[commentStart:], "*/")
		}
	}
	if lineStart < 0 || lineEnd < 0 || commentStart < 0 || commentEnd < 0 {
		slog.Error("error reading field comment of a", "struct", previousType, "field", strings.TrimSpace(line))
		fmt.Fprintf(sb, "%s\n", line)
		return
	} else if commentStart > lineEnd || commentEnd > lineEnd || commentEnd <= commentStart {
		// Empty comment
		fmt.Fprintf(sb, "%s\n", line)
		return
	}

	comment := prettifyComment(strings.TrimSpace(previousResponse[commentStart:commentEnd]), false)

	fmt.Fprintf(sb, "%s // %s\n", line, comment)
}

func AddConstComment(sb *strings.Builder, line string) {
	fields := strings.Fields(line)
	fieldName := strings.TrimPrefix(fields[0], "*")
	// Add underscores before capital letters and numbers
	previousWasANumber := false
	fsb := strings.Builder{}
	for i, s := range fieldName {
		isANumber := (s >= '0' && s <= '9')

		if i > 0 && (s >= 'A' && s <= 'Z') {
			fsb.WriteRune('_')
		} else if i > 0 && isANumber && !previousWasANumber {
			fsb.WriteRune('_')
		}

		previousWasANumber = isANumber
		fsb.WriteRune(s)
	}
	fieldName = NAME_PREFIX + "_" + strings.ToUpper(fsb.String()) // Usually const are ALL_UPPER_CASE

	var lineStart, lineEnd, commentStart, commentEnd int
	lineStart = strings.Index(previousResponse, fieldName)
	if lineStart >= 0 {
		lineEnd = lineStart + strings.Index(previousResponse[lineStart:], "\n")
		commentStart = lineStart + strings.Index(previousResponse[lineStart:], "/**&lt;") + 7
		if commentStart >= 0 {
			commentEnd = commentStart + strings.Index(previousResponse[commentStart:], "*/")
		}
	}
	if lineStart < 0 || lineEnd < 0 || commentStart < 0 || commentEnd < 0 {
		slog.Error("error reading field comment of a", "struct", previousType, "field", strings.TrimSpace(line))
		fmt.Fprintf(sb, "%s\n", line)
		return
	} else if commentStart > lineEnd || commentEnd > lineEnd || commentEnd <= commentStart {
		// Empty comment
		fmt.Fprintf(sb, "%s\n", line)
		return
	}

	comment := prettifyComment(strings.TrimSpace(previousResponse[commentStart:commentEnd]), true)

	fmt.Fprintf(sb, "%s // %s\n", line, comment)
}

func prettifyComment(comment string, startUpperCase bool) string {
	comment = html.UnescapeString(comment)

	// Remove <a> </a> tags
	for {
		idx := strings.Index(comment, "<a")
		if idx < 0 {
			break
		}

		end := idx + strings.Index(comment[idx:], ">") + 1
		comment = comment[:idx] + comment[end:]
	}
	comment = strings.ReplaceAll(comment, "</a>", "")

	// Remove <code> </code> tags
	comment = strings.ReplaceAll(comment, "<code>", "")
	comment = strings.ReplaceAll(comment, "</code>", "")

	// Remove empty brackets
	comment = strings.ReplaceAll(comment, "()", "")
	comment = strings.ReplaceAll(comment, "[]", "")
	comment = strings.ReplaceAll(comment, "{}", "")

	// Remove "SDL_" and surround that type with [] brackets
	for {
		idx := strings.Index(comment, NAME_PREFIX+"_")
		if idx < 0 {
			break
		}
		end := strings.IndexAny(comment[idx:], " ,.)>")
		if end < 0 {
			end = len(comment)
		} else {
			end += idx
		}
		comment = comment[:idx] + "[" + comment[idx+4:end] + "]" + comment[end:]
	}

	// Replace "NULL" with "nil"
	comment = strings.ReplaceAll(comment, "NULL", "nil")

	// Capitalize "SDL"
	comment = strings.ReplaceAll(comment, "sdl", "SDL")

	// Check if the comment ends with "."
	if !strings.HasSuffix(comment, ".") {
		comment += "."
	}

	if startUpperCase {
		comment = strings.ToUpper(string(comment[0])) + comment[1:]
	}

	// Replace double spaces with single one
	comment = strings.ReplaceAll(comment, "  ", " ")

	return comment
}

// Get function comment from wiki.libsdl.org
func getWikiEntry(url, line string) string {
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("error getting response from wiki.libsdl.org", "err", err)
		return ""
	} else if resp.StatusCode != http.StatusOK {
		slog.Error("error getting response from wiki.libsdl.org", "status", resp.Status, "function", strings.TrimSpace(line))
		return ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("error reading response body from wiki.libsdl.org", "err", err)
		return ""
	}

	return string(data)
}
