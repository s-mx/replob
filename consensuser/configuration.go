package consensuser

import (
	cont "github.com/s-mx/replob/containers"
)

type MasterlessConfiguration struct {
	Info cont.Set
}

func NewMasterlessConfiguration(numberNodes uint32) *MasterlessConfiguration {
	conf := new(MasterlessConfiguration)
	conf.Info = cont.NewSet(numberNodes)
	return conf
}

func (conf *MasterlessConfiguration) GetNumberNodes() uint32 {
	return conf.Info.Size()
}
