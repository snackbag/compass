package compass

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Component struct {
	name    string
	vars    map[string]interface{}
	content string
}

func (server *Server) ReloadComponents() error {
	files, err := getFilesWithExtension(server.ComponentsDirectory, ".html")
	if err != nil {
		return err
	}

	server.components = make(map[string]*Component, 0)

	for _, file := range files {
		_fn := strings.Split(file, string(filepath.Separator))
		fileName := strings.TrimSuffix(_fn[len(_fn)-1], ".html")

		read, _ := os.ReadFile(file)
		content := string(read)
		split := strings.Split(content, "// CONTENT //")

		if len(split) != 2 {
			return errors.New("component '" + file + fmt.Sprintf("' is not following component standards - require // CONTENT // split length of 2, got %d", len(split)))
		}

		vars := make(map[string]interface{})
		err = json.Unmarshal([]byte(split[0]), &vars)
		if err != nil {
			return errors.New("at '" + file + "' " + err.Error())
		}

		server.components[fileName] = &Component{name: strings.TrimSuffix(file, ".html"), vars: vars, content: split[1]}
	}

	return nil
}

func (server *Server) StylizeComponent(name string, vars map[string]interface{}, ctx *TemplateContext) (string, error) {
	component, ok := server.components[name]
	if !ok {
		return "", errors.New("cannot find token by name '" + name + "'")
	}

	for k, v := range vars {
		component.vars[k] = v
	}

	newDat := *ctx
	for k, v := range component.vars {
		newDat.SetVariable("&"+k, v)
	}

	return FillRaw(component.content, newDat, server).Content, nil
}

func getFilesWithExtension(dir string, extension string) ([]string, error) {
	var files []string

	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(strings.ToLower(entry.Name()), strings.ToLower(extension)) {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	return files, nil
}
