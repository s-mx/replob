package consensuser

import (
	"github.com/s-mx/replob/nodes"
)

type MasterlessConfiguration struct {
	Info nodes.NodesInfo
}

func NewMasterlessConfiguration(numberNodes uint32) *MasterlessConfiguration {
	conf := new(MasterlessConfiguration)
	conf.Info = *nodes.NewNodesInfo(numberNodes)
	return conf
}

func (conf *MasterlessConfiguration) GetNumberNodes() uint32 {
	return conf.Info.Size()
}
