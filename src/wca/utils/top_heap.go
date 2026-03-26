package utils_tool

type RankItem struct {
	Value int

	Data interface{}
}

func (r *RankItem) SetData(data interface{}) {
	r.Data = data
}

type MaxHeap []RankItem

func (h MaxHeap) Len() int           { return len(h) }
func (h MaxHeap) Less(i, j int) bool { return h[i].Value > h[j].Value }
func (h MaxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(RankItem))
}

func (h *MaxHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func (h *MaxHeap) Copy() []RankItem {
	topList := make([]RankItem, h.Len())
	copy(topList, *h)
	return topList
}
