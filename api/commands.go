package api

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// // get function
// func get_function(name string) func() {

// 	funcs_registry := map[string]func(interface{}){
// 		"get_weather": get_weather,
// 	}

// 	return funcs_registry[name]()
// }

func get_tools() []openai.Tool {
	return []openai.Tool{
		// Get Weather Tool
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the weather for a location",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"location": {
							Type:        jsonschema.String,
							Description: "The city and state, e.g Johannesburg, Gauteng",
						},
						"unit": {
							Type: jsonschema.String,
							Enum: []string{"metric", "imperial"},
						},
					},
					Required: []string{"location"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "create_file",
				Description: "Creates a file locally on the machine. This function can be used for multiple linux related file",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"file_name": {
							Type:        jsonschema.String,
							Description: "The name of the file to create. Ensure you use camel case for naming files. Ensure you use lower case for all",
						},
						"file_path": {
							Type:        jsonschema.String,
							Description: "The path of the file to create. Ensure you use camel case for naming files",
						},
						"file_content": {
							Type:        jsonschema.String,
							Description: "The content of the file to create. Ensure you use camel case for naming files. Ensure new line characters are escaped",
						},
					},
					Required: []string{"file_name", "file_path", "file_content"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionDefinition{
				Name:        "create_directory",
				Description: "Creates a directory locally on the machine. This function can be used for multiple linux related file",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"dir_name": {
							Type:        jsonschema.String,
							Description: "The name of the directory to create. Ensure you use camel case for naming files. Ensure you use lower case for all",
						},
						"dir_path": {
							Type:        jsonschema.String,
							Description: "The path of the directory to create. Ensure you use camel case for naming files",
						},
					},
					Required: []string{"dir_name", "dir_path"},
				},
			},
		},
	}

}

type stubMapping map[string]interface{}

var StubStorage = stubMapping{}

// Reflect to Call our functions
func Call(funcName string, args ...interface{}) interface{} {
	f := reflect.ValueOf(StubStorage[funcName])
	fmethod := f.MethodByName(funcName)
	fmt.Printf("fmethod: %v\n", fmethod)
	if len(args) != f.Type().NumIn() {
		return nil
	}
	fmt.Printf("args in call: %v\n", args)
	in := make([]reflect.Value, len(args))
	fmt.Printf("In: %v\n", in)
	for k, arg := range args {
		in[k] = reflect.ValueOf(arg)
		fmt.Printf("In value: %v\n", reflect.ValueOf(arg))
		fmt.Printf("In type: %v\n", reflect.TypeOf(arg))
	}
	instance := makeInstance(funcName)
	fmt.Printf("Instance: %v\n", reflect.TypeOf(instance))
	var res []reflect.Value
	fmt.Println("Calling function")
	res = f.Call(in)
	result := res[0].Interface()
	return result
}

func makeInstance(name string) interface{} {
	return reflect.New(reflect.TypeOf(StubStorage[name])).Interface()
}

// Our functions
func get_weather(m map[string]string) string {
	location := (m)["location"]
	result := "The weather in " + location + " is 30 degrees"
	return result
}

func create_file(m map[string]string) string {
	file_name := (m)["file_name"]
	file_path := (m)["file_path"]
	file_content := (m)["file_content"]
	// cd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Printf("Error getting working directory in client: %v", err)
	// 	return "", err
	// }
	f_err := os.WriteFile(filepath.Join(file_path, file_name), []byte(file_content), 0644)
	if f_err != nil {
		fmt.Printf("Error creating file: %v", f_err)
		return "Error creating file: " + reflect.ValueOf(f_err).String()
	}
	return "File created successfully"
}

func create_directory(m map[string]string) string {
	dir_name := (m)["dir_name"]
	dir_path := (m)["dir_path"]
	// cd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Printf("Error getting working directory in client: %v", err)
	// 	return "", err
	// }
	f_err := os.Mkdir(filepath.Join(dir_path, dir_name), 0644)
	if f_err != nil {
		fmt.Printf("Error creating directory: %v", f_err)
		return "Error creating directory: " + reflect.ValueOf(f_err).String()
	}
	return "Directory created successfully"
}
