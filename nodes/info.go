package nodes

import "github.com/s-mx/replob/containers"

type NodeId uint32

type NodesInfo struct {
	numberNodes uint32
	Set         containers.Set
}

func NewNodesInfo(numberNodes uint32) *NodesInfo {
	info := new(NodesInfo)
	info.numberNodes = numberNodes
	info.Set = *containers.NewSet(numberNodes)
	return info
}

func (info *NodesInfo) Size() uint32 {
	return info.numberNodes
}

func (info *NodesInfo) ConsistsId(id NodeId) bool {
	return info.Set.Consist(uint32(id))
}

func (info *NodesInfo) NodesEqual(set *containers.Set) bool {
	return info.Set.Equal(set)
}

func (info *NodesInfo) NodesNotEqual(set *containers.Set) bool {
	return !info.NodesEqual(set)
}

func (info *NodesInfo) IntersectNodes(set *containers.Set) {
	info.Set.Intersect(set)
}

func (info *NodesInfo) GetSet() containers.Set {
	return info.Set
}

func (info *NodesInfo) Erase(id NodeId) {
	info.Set.Erase(uint32(id))
}
