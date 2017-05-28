package consensuser

import (
	cont "github.com/s-mx/replob/containers"
)

type Configuration struct {
	Info cont.Set
}

func NewMasterlessConfiguration(numberNodes int) Configuration {
	conf := new(Configuration)
	conf.Info = cont.NewSet(uint32(numberNodes))
	return *conf
}

func (conf *Configuration) Size() uint {
	return uint(conf.Info.Size())
}
