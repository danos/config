// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package commit

type PrioNodes []*PrioNode

func (pn PrioNodes) Empty() bool { return len(pn) == 0 }

func (pn PrioNodes) Len() int { return len(pn) }

func (pn PrioNodes) Less(i, j int) bool {
	return pn[i].Priority < pn[j].Priority
}

func (pn PrioNodes) Swap(i, j int) {
	pn[i], pn[j] = pn[j], pn[i]
}

func (pn *PrioNodes) Push(x interface{}) {
	item := x.(*PrioNode)
	*pn = append(*pn, item)
}

func (pn *PrioNodes) Pop() interface{} {
	old := *pn
	n := len(old)
	item := old[n-1]
	*pn = old[0 : n-1]
	return item
}

type MinHeap struct {
	PrioNodes
}

type MaxHeap struct {
	PrioNodes
}

func (heap MaxHeap) Less(i, j int) bool {
	return heap.PrioNodes[i].Priority > heap.PrioNodes[j].Priority
}
