package main

var funcNodes = map[string]map[string]bool{
	"javascript": {
		"function_declaration":           true,
		"function_expression":            true,
		"generator_function_declaration": true,
		"arrow_function":                 true,
		"method_definition":              true,
	},
	"typescript": {
		"function_declaration":           true,
		"function_expression":            true,
		"generator_function_declaration": true,
		"arrow_function":                 true,
		"method_definition":              true,
	},
	"ruby": {
		"method":           true,
		"singleton_method": true,
	},
	"python": {
		"function_definition": true,
	},
	"go": {
		"function_declaration": true,
		"method_declaration":   true,
	},
	"rust": {
		"function_item": true,
	},
	"java": {
		"method_declaration":      true,
		"constructor_declaration": true,
	},
	"c": {
		"function_definition": true,
	},
	"cpp": {
		"function_definition": true,
	},
	"php": {
		"function_definition": true,
		"method_declaration":  true,
	},
}
