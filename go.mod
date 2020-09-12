module github.com/willscott/memphis

go 1.14

require (
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/polydawn/refmt v0.0.0-20190807091052-3d65705ee9f1 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/warpfork/go-errcat v0.0.0-20180917083543-335044ffc86e // indirect
	go.polydawn.net/go-timeless-api v0.0.0-00010101000000-000000000000 // indirect
	go.polydawn.net/rio v0.0.0-00010101000000-000000000000
)

replace go.polydawn.net/rio => github.com/polydawn/rio v0.0.0-20200325050149-e97d9995e350

replace go.polydawn.net/go-timeless-api => github.com/polydawn/go-timeless-api v0.0.0-20190707220600-0ece408663ed
