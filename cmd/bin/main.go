package main

import (
	"fmt"
	"manifest/manifest"
	"manifest/model"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/gosuri/uitable"
	"github.com/guumaster/logsymbols"
	"github.com/integrii/flaggy"
)

var Version string
var Build string

func main() {
	checkFolder := "."

	flaggy.SetName("Manifest")
	flaggy.SetDescription("Orphoros Snapshot Manifest Utility")

	var curVersion = formatVersion()
	flaggy.SetVersion(curVersion)

	flaggy.DefaultParser.ShowHelpOnUnexpected = true
	flaggy.DefaultParser.AdditionalHelpAppend = "A subcommand is required"
	flaggy.DefaultParser.AdditionalHelpPrepend = "OSMU is a tool for managing Snapshot Manifests files"

	makeCommand := flaggy.NewSubcommand("make")
	makeCommand.Description = "Create a new manifest file in the current directory"

	checkCommand := flaggy.NewSubcommand("check")
	checkCommand.Description = "Check the integrity of a manifest file and the bundle contents in the current directory"
	checkCommand.AddPositionalValue(&checkFolder, "folder", 1, false, "Folder to check")

	flaggy.AttachSubcommand(makeCommand, 1)
	flaggy.AttachSubcommand(checkCommand, 1)

	flaggy.Parse()

	if makeCommand.Used {
		conf := model.Config{
			BundleName:     "OSMU CLI Utility",
			SnapshotName:   "snapshot",
			Version:        "1.0",
			Critical:       true,
			Root:           ".",
			Signed:         false,
			PrivateKeyPath: "./certs/private.pem",
			PublicKeyPath:  "./certs/public.pem",
			Ignore: []string{
				"*.manifest",
				"*.git*",
				"*node_modules*",
				"*.venv*",
				"*__pycache__*",
				"*.angular*",
				"*.DS_Store*",
				"*dist*",
			},
		}
		man, err := manifest.Create(conf)

		if err != nil {
			fmt.Println(err)
		}

		man.AddFiles()
		if err := man.WriteToFile(); err != nil {
			fmt.Println(err)
		}
	} else if checkCommand.Used {
		// read manifest file
		man, err := manifest.FromFile(checkFolder + "/snapshot.manifest")
		if err != nil {
			fmt.Println(err)
		}

		bundle := man.GetBundle()

		Box := box.New(box.Config{Px: 1, Py: 1, Type: "Round", Color: "Cyan", TitleColor: "Cyan", TitlePos: "Inside"})

		table := uitable.New()
		table.AddRow("Name:", bundle.NA)
		table.AddRow("Version:", bundle.VE)
		table.AddRow("Created at:", time.Unix(bundle.TS, 0).Format(time.RFC822Z))
		table.AddRow("Content count:", len(bundle.FL))

		var diskSize string
		if bundle.SZ < 1024 {
			diskSize = fmt.Sprintf("%d B", bundle.SZ)
		} else if bundle.SZ < 1024*1024 {
			diskSize = fmt.Sprintf("%.2f KB", float64(bundle.SZ)/1024)
		} else if bundle.SZ < 1024*1024*1024 {
			diskSize = fmt.Sprintf("%.2f MB", float64(bundle.SZ)/(1024*1024))
		} else {
			diskSize = fmt.Sprintf("%.2f GB", float64(bundle.SZ)/(1024*1024*1024))
		}

		table.AddRow("Disk size:", diskSize)

		err = man.Check()

		if bundle.SN {
			table.AddRow("Signature:", logsymbols.Success+" Valid")
		} else {
			table.AddRow("Signature:", logsymbols.Warn+" Unsigned")
		}

		if bundle.CR {
			table.AddRow("Critical:", logsymbols.Success+" Full validation")
		} else {
			table.AddRow("Critical:", logsymbols.Warn+" Partial validation")
		}

		if err != nil {
			fmt.Println(err)
			table.AddRow("State:", logsymbols.Error+" Invalid")
		} else {
			table.AddRow("State:", logsymbols.Success+" Valid")
		}

		Box.Print("Manifest Bundle", fmt.Sprint(table))
	} else {
		flaggy.ShowHelpAndExit("Error: No subcommand provided")
	}

}

func formatVersion() string {
	var curVersion string
	if Version == "" {
		curVersion = "dev"
	} else {
		curVersion = Version
		if Build != "" {
			curVersion += " (" + Build + ")"
		}
	}
	return curVersion
}
