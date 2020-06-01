module github.com/lade-io/lade

go 1.13

require (
	github.com/docker/docker v1.13.1
	github.com/dustin/go-humanize v1.0.0
	github.com/iancoleman/orderedmap v0.0.0-20180606015914-fec04b9a4f6d
	github.com/jinzhu/configor v1.1.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lade-io/go-lade v0.1.1
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-isatty v0.0.10
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/ts v0.0.0-20171002115256-78ecb04241c0
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7 // indirect
	github.com/rodaine/table v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6
	gopkg.in/AlecAivazis/survey.v1 v1.8.7
	gopkg.in/yaml.v2 v2.2.5
)

replace github.com/rodaine/table => github.com/beornf/table v1.0.1-0.20180415234414-08cbc594e511
