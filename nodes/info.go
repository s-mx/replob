package nodes

import cont "github.com/s-mx/replob/containers"


type NodesInfo struct {
	numberNodes uint32
	Set         cont.Set
}

func NewNodesInfo(numberNodes uint32) *NodesInfo {
	info := new(NodesInfo)
	info.numberNodes = numberNodes
	info.Set = *cont.NewSet(numberNodes)
	return info
}

func (info *NodesInfo) Size() uint32 {
	return info.numberNodes
}

func (info *NodesInfo) ConsistsId(id cont.NodeId) bool {
	return info.Set.Consist(uint32(id))
}

func (info *NodesInfo) NodesEqual(set *cont.Set) bool {
	return info.Set.Equal(set)
}

func (info *NodesInfo) NodesNotEqual(set *cont.Set) bool {
	return !info.NodesEqual(set)
}

func (info *NodesInfo) IntersectNodes(set *cont.Set) {
	info.Set.Intersect(set)
}

func (info *NodesInfo) GetSet() cont.Set {
	return info.Set
}

func (info *NodesInfo) Erase(id cont.NodeId) {
	info.Set.Erase(uint32(id))
}
