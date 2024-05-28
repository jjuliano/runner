package main

import (
	"fmt"
	"kdeps/resolver"
	"os"
)

func main() {
	var command string
	var packages []string

	for _, arg := range os.Args[1:] {
		switch arg {
		case "depends", "rdepends", "add", "update", "search", "category", "show", "tree", "tree-list":
			if command == "" {
				command = arg
			} else {
				packages = append(packages, arg)
			}
		default:
			packages = append(packages, arg)
		}
	}

	if command == "" {
		fmt.Println("Usage: script [depends|rdepends|add|update|search|category|show|tree|tree-list] [package_names...]")
		return
	}

	resolver := resolver.NewDependencyResolver()
	resolver.LoadPackageEntries("setup.yml")

	switch command {
	case "add":
		resolver.HandleAddCommand(packages)
	case "update":
		resolver.HandleUpdateCommand(packages)
	case "show":
		resolver.HandleShowCommand(packages)
	case "depends":
		resolver.HandleDependsCommand(packages)
	case "rdepends":
		resolver.HandleRDependsCommand(packages)
	case "search":
		resolver.HandleSearchCommand(packages)
	case "category":
		resolver.HandleCategoryCommand(packages)
	case "tree":
		resolver.HandleTreeCommand(packages)
	case "tree-list":
		resolver.HandleTreeListCommand(packages)
	default:
		fmt.Println("Invalid command:", command)
	}
}
