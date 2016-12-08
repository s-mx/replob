package nodes

import "github.com/s-mx/replob/containers"

type NodeId uint32

type NodesInfo struct {
    numberNodes uint32
    Set *containers.Set
}

func NewNodesInfo(numberNodes uint32) *NodesInfo {
    info := new(NodesInfo)
    info.numberNodes = numberNodes
    info.Set = containers.NewSet(numberNodes)
    return info
}

func (info *NodesInfo) Size() uint32 {
    return info.numberNodes
}

func (info* NodesInfo) ConsistsId(id NodeId) bool {
    return id < NodeId(info.numberNodes) // это не правильно!
}

func (info* NodesInfo) NodesNotEqual(set *containers.Set) bool {
    return true // dummy
}

func (info* NodesInfo) IntersectNodes(setNodes *containers.Set) {
    // dummy
}

func (info* NodesInfo) GetSet() containers.Set {
    return *info.Set // Кажется, здесь оно разыменуется и вернется по значению
}