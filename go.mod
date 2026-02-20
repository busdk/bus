module bus

go 1.22.2

require (
	github.com/busdk/bus-bank v0.0.0
	github.com/busdk/bus-journal v0.0.0
)

require (
	github.com/busdk/bus-accounts v0.0.0 // indirect
	github.com/busdk/bus-bfl v0.0.0-00010101000000-000000000000 // indirect
	github.com/busdk/bus-data v0.0.0 // indirect
	github.com/busdk/bus-period v0.0.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/term v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/busdk/bus-bank => ../bus-bank

replace github.com/busdk/bus-journal => ../bus-journal

replace github.com/busdk/bus-accounts => ../bus-accounts

replace github.com/busdk/bus-period => ../bus-period

replace github.com/busdk/bus-data => ../bus-data

replace github.com/busdk/bus-bfl => ../bus-bfl
