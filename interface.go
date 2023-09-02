package archive

type Pack interface {
	Pack() error
}

type Unpack interface {
	Unpack() error
}

type Archive interface {
	Pack
	Unpack
}
