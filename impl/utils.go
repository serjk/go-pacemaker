package impl

import ."github.com/serjk/go-pacemaker"


func ForQuery(config *CibOpenConfig) {
	config.connection = Query
}

func ForCommand(config *CibOpenConfig) {
	config.connection = Command
}

func ForNoConnection(config *CibOpenConfig) {
	config.connection = NoConnection
}

func ForCommandNonBlocking(config *CibOpenConfig) {
	config.connection = CommandNonBlocking
}

func FromFile(file string) func(*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.file = file
	}
}

func FromShadow(shadow string) func(*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.shadow = shadow
	}
}

func FromRemote(server, user, passwd string, port int, encrypted bool) func(*CibOpenConfig) {
	return func(config *CibOpenConfig) {
		config.server = server
		config.user = user
		config.passwd = passwd
		config.port = port
		config.encrypted = encrypted
	}
}

func convertPMCodeToError(code int, msg string) error {
	switch code {
	case -ENXIO:
		return NewNotFoundErr(msg)
	case -ENOTCONN,
		-ECONNABORTED,
		-ECONNREFUSED,
		-ECONNRESET,
		-ECOMM:
		return NewConnectionErr(msg)
	case -ENOTUNIQ:
		return NewAlreadyExistedErr(msg)
	case -EOPNOTSUPP:
		return NewNotSupportedOpErr(msg)
	default:
		return NewCibError(msg)
	}
}
func convertCSCodeToError(code int, msg string) error {
	switch code {
	case CS_ERR_NOT_EXIST:
		return NewNotFoundErr(msg)
	case CS_ERR_LIBRARY:
		return NewConnectionErr(msg)
	default:
		return convertPMCodeToError(code, msg)
	}
}
