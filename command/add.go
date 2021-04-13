package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/configs/configschema"
)

// AddCommand is a Command implementation that writes a template resource config block to a given file.
type AddCommand struct {
	Meta
}

func (c *AddCommand) Run(args []string) int {
	// This command takes a resource address and output filename. Some flags to consider
	// adding:
	//  * -verbose, to control inclusion of provider's descriptions
	//  * -something to toggle required only vs. required + optional attrs
	//  * -provider, to override the resource's implied provider (required in case of collision in config)
	//  * -json, to write json instead of hcl
	// It would also be neat to write out a provider configuration block from
	// info in required_providers - the arg could be a provider or resource
	args = c.Meta.process(args)
	addrStr := args[0]
	filename := "import.tf"
	if len(args) > 1 {
		filename = args[1]
	}

	absAddr, diags := addrs.ParseAbsResourceStr(addrStr)
	if diags.HasErrors() {
		fmt.Printf("%q is not a valid resource address", addrStr)
		return 1
	}

	// Load the backend
	b, backendDiags := c.Backend(nil)
	diags = diags.Append(backendDiags)
	if backendDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.showDiagnostics(diags) // in case of any warnings in here
		c.Ui.Error(ErrUnsupportedLocalOp)
		return 1
	}

	cwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting cwd: %s", err))
		return 1
	}

	// Build the operation
	opReq := c.Operation(b)
	opReq.AllowUnsetVariables = true
	opReq.ConfigDir = cwd

	opReq.ConfigLoader, err = c.initConfigLoader()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing config loader: %s", err))
		return 1
	}

	// Get the context
	ctx, _, ctxDiags := local.Context(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// TODO: load the configuration and check that the resource address doesn't
	// already exist in the config

	// Get the schemas from the context
	schemas := ctx.Schemas()

	// For the sake of a quick prototype I will assume the implied type is the correct type.
	provider := absAddr.Resource.ImpliedProvider()
	absProvider := addrs.ImpliedProviderForUnqualifiedType(provider)

	if _, exists := schemas.Providers[absProvider]; !exists {
		c.Ui.Error(fmt.Sprintf("# missing schema for provider %q\n\n", absProvider.String()))
	}

	schema, _ := schemas.ResourceTypeConfig(absProvider, absAddr.Resource.Mode, absAddr.Resource.Type)

	// For now, we'll required a new file, but there's no reason we couldn't append.
	if _, err := os.Stat(filename); err == nil {
		c.Ui.Error(fmt.Sprintf("file %s already exists", filename))
	}
	f, err := os.Create(filename)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("error creating file %s", filename))
		return 1
	}
	defer f.Close()

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("resource %q %q {\n", absAddr.Resource.Type, absAddr.Resource.Name))
	writeConfigAttributes(&buf, schema.Attributes, 2)
	writeConfigBlocks(&buf, schema.BlockTypes, 2)

	buf.WriteString("}\n")
	f.Write([]byte(buf.String()))

	return 0
}

func (c *AddCommand) Help() string {
	helpText := `
Usage: terraform [global options] add [options] ADDRESS`
	return strings.TrimSpace(helpText)
}

func (c *AddCommand) Synopsis() string {
	return "Add a resource configuration to a file"
}

func writeConfigAttributes(buf *strings.Builder, attrs map[string]*configschema.Attribute, indent int) {
	if len(attrs) == 0 {
		return
	}
	for name, attrS := range attrs {
		if attrS.Required || attrS.Optional {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("# %s\n", attrS.Description))
			buf.WriteString(strings.Repeat(" ", indent))
		}
		if attrS.Required {
			buf.WriteString(fmt.Sprintf("%s = <REQUIRED %s>\n\n", name, attrS.Type.FriendlyName()))
		} else if attrS.Optional {
			buf.WriteString(fmt.Sprintf("%s = <OPTIONAL %s>\n\n", name, attrS.Type.FriendlyName()))
		}
	}
}

func writeConfigBlocks(buf *strings.Builder, blocks map[string]*configschema.NestedBlock, indent int) {
	if len(blocks) == 0 {
		return
	}
	// required == min items > 0, but I think there's more?
	for name, blockS := range blocks {
		if blockS.MinItems > 0 {
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString(fmt.Sprintf("%s {", name))
			if len(blockS.Attributes) > 0 {
				writeConfigAttributes(buf, blockS.Attributes, indent+2)
			}
			if len(blockS.BlockTypes) > 0 {
				writeConfigBlocks(buf, blockS.BlockTypes, indent+2)
			}
			buf.WriteString(strings.Repeat(" ", indent))
			buf.WriteString("}\n")
		}
	}
}
