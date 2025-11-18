// Package init provides template initialization and scaffolding for Nix projects.
//
// This package implements the omnix-init functionality for creating new projects
// from Nix flake templates with parameter substitution and file manipulation.
//
// # Overview
//
// The init package provides:
//   - Template scaffolding from Nix flake templates
//   - Parameter-based text replacement in files and filenames
//   - Conditional file/directory retention and pruning
//   - Action-based template customization
//
// # Usage
//
// Basic template scaffolding:
//
//	ctx := context.Background()
//	template := &init.Template{
//	    Path: "/path/to/template",
//	    Params: []init.Param{
//	        {
//	            Name: "project-name",
//	            Action: &init.ReplaceAction{
//	                Placeholder: "MYPROJECT",
//	                Value:       strPtr("my-app"),
//	            },
//	        },
//	    },
//	}
//
//	outPath, err := template.ScaffoldAt(ctx, "/path/to/output")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Actions
//
// The package supports two types of actions:
//
//   - ReplaceAction: Replaces placeholder text in file contents and filenames
//   - RetainAction: Deletes files/directories matching glob patterns (when value is false)
//
// Actions are applied in priority order: Retain actions run before Replace actions
// to ensure files are pruned before text replacement occurs.
//
// # Example
//
// Creating a new project with parameter substitution:
//
//	template := &init.Template{
//	    Path: "./template",
//	    Params: []init.Param{
//	        {
//	            Name:        "name",
//	            Description: "Project name",
//	            Action: &init.ReplaceAction{
//	                Placeholder: "PROJECT_NAME",
//	                Value:       strPtr("awesome-app"),
//	            },
//	        },
//	        {
//	            Name:        "include-ci",
//	            Description: "Include CI configuration",
//	            Action: &init.RetainAction{
//	                Paths: []string{".github/**"},
//	                Value: boolPtr(true),  // Keep the files
//	            },
//	        },
//	    },
//	}
//
//	outPath, _ := template.ScaffoldAt(context.Background(), "./my-project")
//	fmt.Println("Project created at:", outPath)
//
// # Features
//
// The package handles:
//   - Recursive directory copying with symlink preservation
//   - Text replacement in file contents
//   - File and directory renaming based on placeholders
//   - Glob pattern matching for file pruning
//   - Action ordering to ensure correct operation
package init
