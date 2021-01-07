package main

import (
	"encoding/json"
	"flag"
	"fmt"
	engine "go-templating-engine"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Templating engine
func renderHead() {}

type Page struct {
	Filename string
	Header   map[string]interface{}
	Body     []byte
}

func extractFrontmatterData(data []byte) (map[string]interface{}, []byte) {
	body := ""
	header := make(map[string]interface{})
	inHeader := false
	lines := strings.Split(string(data), "\n")

	for i := 0; i <= len(lines)-1; i++ {
		if lines[i] == "---" {
			inHeader = !inHeader
			continue
		}

		if inHeader {
			key := strings.Split(lines[i], ":")[0]
			value := strings.Split(lines[i], ":")[1]
			header[key] = strings.TrimSpace(value)
			continue
		}

		body += lines[i] + "\n"
	}

	return header, []byte(body)
}

func readTemplateFile(filepath string) (*Page, error) {
	fileData, err := ioutil.ReadFile(filepath)

	if err != nil {
		return nil, err
	}

	frontmatter, body := extractFrontmatterData(fileData)

	page := &Page{
		Filename: filepath,
		Header:   frontmatter,
		Body:     body,
	}

	return page, nil
}

func generateProject(projectLocation string) error {
	pages, err := ioutil.ReadDir(projectLocation + "/pages")
	if err != nil {
		return err
	}

	siteConfigFile, err := ioutil.ReadFile(projectLocation + "/config.json")
	var globalSiteContext map[string]interface{}
	json.Unmarshal(siteConfigFile, &globalSiteContext)

	var result []string
	var renderFunctionsToCall []string

	result = append(result, "package main\n\nimport (\n    \"strings\"\n    \"encoding/json\"\n    \"reflect\"\n    \"io/ioutil\"\n)\n"+strings.Join(result, ""))

	for i := 0; i < len(pages); i++ {
		templateName := strings.Split(pages[i].Name(), ".")[1]

		templateContents, err := ioutil.ReadFile(projectLocation + "/templates/" + templateName + ".html")
		if err != nil {
			return err
		}

		p, err := readTemplateFile(projectLocation + "/pages/" + pages[i].Name())
		if err != nil {
			return err
		}

		p.Header["content"] = string(p.Body)
		context := map[string]interface{}{
			"site": globalSiteContext,
			"page": p.Header,
		}

		renderFunctionName := strings.Replace(pages[i].Name(), ".", "_", -1)

		renderFunctionsToCall = append(renderFunctionsToCall, renderFunctionName)
		result = append(result, engine.RenderTemplateString(string(templateContents), renderFunctionName, context))
	}

	mainFunction := "func main() {\n"
	for i := 0; i < len(renderFunctionsToCall); i++ {
		mainFunction += "    ioutil.WriteFile(\"../site/" + renderFunctionsToCall[i] + ".html\", []byte(c_render_" + renderFunctionsToCall[i] + "()), 0644)\n"
	}
	mainFunction += "}"

	result = append(result, mainFunction)

	_ = os.Mkdir(projectLocation+"/dist", 0755)
	_ = os.Mkdir(projectLocation+"/dist/bin", 0755)
	_ = os.Mkdir(projectLocation+"/dist/site", 0755)

	finalCompiledOutput := []byte(strings.Join(result, "\n"))
	err = ioutil.WriteFile(projectLocation+"/dist/bin/out.go", finalCompiledOutput, 0755)
	if err != nil {
		return err
	}

	out, cErr := RunCMD("/bin/go", []string{"run", projectLocation + "/dist/bin/out.go"}, true)
	if cErr != nil {
		fmt.Println(out)
		fmt.Println(cErr.Error())
		return cErr
	}

	return nil
}

func RunCMD(path string, args []string, debug bool) (out string, err error) {

	cmd := exec.Command(path, args...)

	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)

	if debug {
		fmt.Println(strings.Join(cmd.Args[:], " "))

		if err != nil {
			fmt.Println("RunCMD ERROR")
			fmt.Println(out)
		}
	}

	return
}

func main() {
	var projectDirFlag = flag.String("p", ".", "Project directory")
	flag.Parse()
	generateProject(*projectDirFlag)
}
