package utils

import (
	"math"
)

func Merge[T comparable](list1, list2 []T) []T {
	merged := make(map[T]bool)

	// 添加列表1中的元素到 merged map 中
	for _, v := range list1 {
		merged[v] = true
	}

	// 添加列表2中的元素到 merged map 中
	for _, v := range list2 {
		merged[v] = true
	}

	// 从 merged map 中提取不重复的元素到结果列表中
	var result []T
	for k := range merged {
		result = append(result, k)
	}

	return result
}

func Delete[T comparable](org, del []T) []T {
	// 创建一个 map 以便快速检查 del 切片中的值
	delMap := make(map[T]struct{})
	for _, v := range del {
		delMap[v] = struct{}{}
	}

	// 遍历 org 切片，将不在 delMap 中的元素添加到结果切片中
	var result []T
	for _, v := range org {
		if _, ok := delMap[v]; !ok {
			result = append(result, v)
		}
	}

	return result
}

func Has[T comparable](org, s []T) []T {
	// 创建一个 map 以便快速检查 s 切片中的值
	sMap := make(map[T]struct{})
	for _, v := range s {
		sMap[v] = struct{}{}
	}

	// 遍历 org 切片，将存在于 sMap 中的元素添加到结果切片中
	var result []T
	for _, v := range org {
		if _, ok := sMap[v]; ok {
			result = append(result, v)
		}
	}

	return result
}

func Page[T any](src []T, page, size int) (out []T, total int) {
	if len(src) == 0 {
		return src, len(src)
	}

	if page <= 1 {
		page = 1
	}
	if size > len(src) {
		size = len(src)
	}
	if size > 1000 {
		size = 1000
	}
	if size < 20 {
		size = 20
	}

	start := (page - 1) * size

	if start > len(src) {
		return nil, len(src)
	}

	end := int(math.Min(float64(start+size), float64(len(src))))
	return src[start:end], len(src)
}
