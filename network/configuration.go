package network

import (
	"strconv"
	"github.com/s-mx/replob/consensuser"
)

type Configuration struct {
	numberNodes		int
	serviceServer	[]string
}

func NewLocalNetConfiguration(numberNodes int) *Configuration {
	ptr := &Configuration{
		numberNodes:numberNodes,
		serviceServer:make([]string, numberNodes),
	}

	for ind := 0; ind < numberNodes; ind++ {
		ptr.serviceServer[ind] = ":" + strconv.Itoa(2048 + ind)
	}

	return ptr
}

func (config *Configuration) GetMasterlessConfiguration() consensuser.Configuration {
	return consensuser.NewMasterlessConfiguration(uint32(config.numberNodes))
}
